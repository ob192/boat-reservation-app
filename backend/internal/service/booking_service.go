package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"time"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"gorm.io/gorm"
)

// HoldDuration is how long a pending booking holds slot capacity before expiring.
const HoldDuration = 15 * time.Minute

// Booking-creation typed errors. Handlers map these to HTTP error codes.
var (
	ErrBookingsDisabled  = errors.New("BOOKINGS_DISABLED")
	ErrDateBlocked       = errors.New("DATE_BLOCKED")
	ErrSlotBlocked       = errors.New("SLOT_BLOCKED")
	ErrSlotTaken         = errors.New("SLOT_TAKEN")
	ErrSlotNotFound      = errors.New("SLOT_NOT_FOUND")
	ErrInvalidInput      = errors.New("INVALID_INPUT")
	ErrValidationFailed  = errors.New("VALIDATION_FAILED")
	ErrBookingNotFound   = errors.New("BOOKING_NOT_FOUND")
	ErrBookingNotPending = errors.New("BOOKING_NOT_PENDING")
	ErrBookingExpired    = errors.New("BOOKING_EXPIRED")
	ErrAlreadyConfirmed  = errors.New("ALREADY_CONFIRMED")
	ErrForbidden         = errors.New("FORBIDDEN")
)

// CreateBookingInput holds the validated, normalised inputs passed to the booking service.
// All authentication-derived fields (user ID, email) come from the JWT, never the body.
type CreateBookingInput struct {
	UserID         uuid.UUID
	UserEmail      string
	IdempotencyKey string

	Date       string
	Time       string
	Quantities model.Quantities

	FirstName string
	LastName  string
	Phone     string // required per project decision
}

// BookingService is the orchestrator for booking creation and lookups.
type BookingService interface {
	Create(ctx context.Context, in CreateBookingInput) (*model.Booking, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	GetByIDForUser(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Booking, error)
	GetBySession(ctx context.Context, sessionID string, userID uuid.UUID) (*model.Booking, error)

	// ExpireStale runs the periodic expiry sweep.
	ExpireStale(ctx context.Context) (int64, error)
}

type bookingService struct {
	db         *gorm.DB
	bookings   repository.BookingRepository
	slots      repository.SlotRepository
	dateBlocks repository.DateBlockRepository
	system     repository.SystemRepository
	pricing    PricingService
	clock      platform.Clock
}

func NewBookingService(
	db *gorm.DB,
	bookings repository.BookingRepository,
	slots repository.SlotRepository,
	dateBlocks repository.DateBlockRepository,
	system repository.SystemRepository,
	pricing PricingService,
	clock platform.Clock,
) BookingService {
	return &bookingService{
		db:         db,
		bookings:   bookings,
		slots:      slots,
		dateBlocks: dateBlocks,
		system:     system,
		pricing:    pricing,
		clock:      clock,
	}
}

// Create implements the spec's POST /bookings business logic in order:
//  1. Validate request fields (caller has already done JSON / required-field checks).
//  2. Kill-switch check.
//  3. Date-block check.
//  4. Idempotency check.
//  5. DB tx: SELECT FOR UPDATE slot, check blocked, check capacity, INSERT.
func (s *bookingService) Create(ctx context.Context, in CreateBookingInput) (*model.Booking, error) {
	// (1) Validate inputs not already enforced by struct tags.
	if err := validateCreateInput(in, s.clock.Now()); err != nil {
		return nil, err
	}

	// (2) Kill-switch — read straight from DB so we never serve a stale "enabled".
	settings, err := s.system.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("read system settings: %w", err)
	}
	if !settings.BookingsEnabled {
		return nil, ErrBookingsDisabled
	}

	// (3) Date-block check.
	if _, err := s.dateBlocks.Find(ctx, in.Date); err == nil {
		return nil, ErrDateBlocked
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check date block: %w", err)
	}

	// (4) Idempotency: if a previous booking exists with this (user_id, idempotency_key),
	// return it as-is. Safe because we never mutate it post-creation under the same key.
	if existing, err := s.bookings.FindByIdempotencyKey(ctx, in.UserID, in.IdempotencyKey); err == nil {
		return existing, nil
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("check idempotency: %w", err)
	}

	// (5) Capacity tx — pessimistic lock on the slot row.
	total := s.pricing.ComputeTotal(in.Quantities)
	now := s.clock.Now()
	expiresAt := now.Add(HoldDuration)

	var created *model.Booking
	err = platform.WithTx(ctx, s.db, func(ctx context.Context) error {
		slot, err := s.slots.LockForUpdate(ctx, in.Date, in.Time)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSlotNotFound
		}
		if err != nil {
			return fmt.Errorf("lock slot: %w", err)
		}

		if slot.Blocked {
			return ErrSlotBlocked
		}

		bookedBig, bookedMedium, err := s.bookings.SumActiveQuantitiesForSlot(ctx, in.Date, in.Time)
		if err != nil {
			return fmt.Errorf("sum active qty: %w", err)
		}

		// Check big and medium independently — fleets are not interchangeable.
		if in.Quantities.Big > (slot.CapacityBig - bookedBig) {
			return ErrSlotTaken
		}
		if in.Quantities.Medium > (slot.CapacityMedium - bookedMedium) {
			return ErrSlotTaken
		}

		var phone *string
		if in.Phone != "" {
			p := in.Phone
			phone = &p
		}

		t, _ := time.Parse("2006-01-02", in.Date)

		b := &model.Booking{
			ID:             uuid.New(),
			UserID:         in.UserID,
			UserEmail:      in.UserEmail,
			Date:           pgtype.Date{Time: t, Valid: true},
			Time:           in.Time,
			QtyBig:         in.Quantities.Big,
			QtyMedium:      in.Quantities.Medium,
			QtyChild:       in.Quantities.Child,
			FirstName:      in.FirstName,
			LastName:       in.LastName,
			Phone:          phone,
			TotalAmount:    total,
			Status:         model.StatusPending,
			IdempotencyKey: in.IdempotencyKey,
			ExpiresAt:      expiresAt,
		}
		if err := s.bookings.Create(ctx, b); err != nil {
			return fmt.Errorf("insert booking: %w", err)
		}
		created = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (s *bookingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	b, err := s.bookings.FindByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookingNotFound
	}
	return b, err
}

func (s *bookingService) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*model.Booking, error) {
	b, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if b.UserID != userID {
		return nil, ErrForbidden
	}
	return b, nil
}

func (s *bookingService) GetBySession(ctx context.Context, sessionID string, userID uuid.UUID) (*model.Booking, error) {
	b, err := s.bookings.FindByPaymentSessionID(ctx, sessionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, err
	}
	if b.UserID != userID {
		return nil, ErrForbidden
	}
	return b, nil
}

func (s *bookingService) ExpireStale(ctx context.Context) (int64, error) {
	return s.bookings.ExpirePending(ctx, s.clock.Now())
}

// validateCreateInput enforces business rules that Gin's struct binding can't:
// non-negative ints, child-without-big, sums ≥ 1, allowed time, date not in the past, etc.
func validateCreateInput(in CreateBookingInput, now time.Time) error {
	// Time slot whitelist.
	if !IsValidSlotTime(in.Time) {
		return fmt.Errorf("%w: invalid time %q", ErrInvalidInput, in.Time)
	}

	// Date format + not in the past.
	d, err := ParseDate(in.Date)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidInput, err.Error())
	}
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if d.Before(today) {
		return fmt.Errorf("%w: date is in the past", ErrInvalidInput)
	}

	// Quantities — non-negative, sane combinations.
	if in.Quantities.Big < 0 || in.Quantities.Medium < 0 || in.Quantities.Child < 0 {
		return fmt.Errorf("%w: quantities must be non-negative", ErrInvalidInput)
	}
	if in.Quantities.Big+in.Quantities.Medium < 1 {
		return fmt.Errorf("%w: at least one boat required", ErrValidationFailed)
	}
	if in.Quantities.Child > 0 && in.Quantities.Big == 0 {
		return fmt.Errorf("%w: child seats require at least one big boat", ErrValidationFailed)
	}

	// Contact — names checked by binding, phone required per project decision.
	if in.FirstName == "" || in.LastName == "" {
		return fmt.Errorf("%w: firstName and lastName required", ErrInvalidInput)
	}
	if in.Phone == "" {
		return fmt.Errorf("%w: phone is required", ErrInvalidInput)
	}
	return nil
}
