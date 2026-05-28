package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"gorm.io/gorm"
)

// ErrAlreadyBlocked Admin-specific typed errors.
var (
	ErrAlreadyBlocked = errors.New("ALREADY_BLOCKED")
)

// AdminService implements all /admin/* business logic.
type AdminService interface {
	OverridePrice(
		ctx context.Context,
		bookingID uuid.UUID,
		adminID uuid.UUID,
		amount *float64,
		reason string,
	) (*model.AdminPriceOverrideResponse, error)

	BlockSlot(
		ctx context.Context,
		date, time string,
		adminID uuid.UUID,
		reason string,
	) (*model.AdminBlockSlotResponse, error)

	UnblockSlot(ctx context.Context, date, time string) (*model.AdminUnblockSlotResponse, error)

	BlockDate(
		ctx context.Context,
		date string,
		adminID uuid.UUID,
		reason string,
	) (*model.AdminBlockDateResponse, error)

	UnblockDate(ctx context.Context, date string) (*model.AdminUnblockDateResponse, error)

	SetBookingsEnabled(
		ctx context.Context,
		enabled bool,
		reason string,
		adminID uuid.UUID,
	) (*model.AdminSetBookingsEnabledResponse, error)

	CancelBooking(
		ctx context.Context,
		bookingID uuid.UUID,
		adminID uuid.UUID,
		reason string,
	) (*model.AdminCancelBookingResponse, error)

	UpsertSlot(
		ctx context.Context,
		date, time string,
		capacityBig, capacityMedium int,
		adminID uuid.UUID,
	) (*model.AdminUpsertSlotResponse, error)

	GetSlotBookings(
		ctx context.Context,
		date, time string,
	) (*model.AdminSlotBookingsResponse, error)
}

type adminService struct {
	db         *gorm.DB
	bookings   repository.BookingRepository
	slots      repository.SlotRepository
	dateBlocks repository.DateBlockRepository
	system     repository.SystemRepository
	pricing    PricingService
	clock      platform.Clock
}

func NewAdminService(
	db *gorm.DB,
	bookings repository.BookingRepository,
	slots repository.SlotRepository,
	dateBlocks repository.DateBlockRepository,
	system repository.SystemRepository,
	pricing PricingService,
	clock platform.Clock,
) AdminService {
	return &adminService{
		db:         db,
		bookings:   bookings,
		slots:      slots,
		dateBlocks: dateBlocks,
		system:     system,
		pricing:    pricing,
		clock:      clock,
	}
}

// ----------------------------------------------------------------------------
// Price override
// ----------------------------------------------------------------------------

func (s *adminService) OverridePrice(
	ctx context.Context,
	bookingID uuid.UUID,
	adminID uuid.UUID,
	amount *float64,
	reason string,
) (*model.AdminPriceOverrideResponse, error) {
	b, err := s.bookings.FindByID(ctx, bookingID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, err
	}

	// Don't override a booking that's already paid for — refund handling is out of scope.
	if b.Status == model.StatusConfirmed {
		return nil, ErrAlreadyConfirmed
	}

	if amount != nil && *amount <= 0 {
		return nil, ErrInvalidInput
	}

	var reasonPtr *string
	if amount != nil {
		r := strings.TrimSpace(reason)
		reasonPtr = &r
	}

	if err := s.bookings.SetPriceOverride(ctx, bookingID, amount, reasonPtr, &adminID); err != nil {
		return nil, err
	}

	updated, err := s.bookings.FindByID(ctx, bookingID)
	if err != nil {
		return nil, err
	}
	effective := s.pricing.EffectiveAmount(updated.TotalAmount, updated.PriceOverride)

	return &model.AdminPriceOverrideResponse{
		BookingID:       updated.ID.String(),
		TotalAmount:     updated.TotalAmount,
		PriceOverride:   updated.PriceOverride,
		OverrideReason:  updated.OverrideReason,
		EffectiveAmount: effective,
	}, nil
}

// ----------------------------------------------------------------------------
// Slot block / unblock
// ----------------------------------------------------------------------------

func (s *adminService) BlockSlot(
	ctx context.Context,
	date, time string,
	adminID uuid.UUID,
	reason string,
) (*model.AdminBlockSlotResponse, error) {
	slot, err := s.slots.FindByDateTime(ctx, date, time)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSlotNotFound
	}
	if err != nil {
		return nil, err
	}

	if slot.Blocked {
		return nil, ErrAlreadyBlocked
	}

	var reasonPtr *string
	if r := strings.TrimSpace(reason); r != "" {
		reasonPtr = &r
	}
	if err := s.slots.Block(ctx, date, time, adminID, reasonPtr); err != nil {
		return nil, err
	}

	updated, err := s.slots.FindByDateTime(ctx, date, time)
	if err != nil {
		return nil, err
	}
	blockedAt := s.clock.Now()
	if updated.BlockedAt != nil {
		blockedAt = *updated.BlockedAt
	}
	return &model.AdminBlockSlotResponse{
		Date:      date,
		Time:      time,
		Blocked:   true,
		Reason:    updated.BlockReason,
		BlockedAt: blockedAt,
	}, nil
}

func (s *adminService) UnblockSlot(
	ctx context.Context,
	date, time string,
) (*model.AdminUnblockSlotResponse, error) {
	if _, err := s.slots.FindByDateTime(ctx, date, time); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}
	if err := s.slots.Unblock(ctx, date, time); err != nil {
		return nil, err
	}
	return &model.AdminUnblockSlotResponse{Date: date, Time: time, Blocked: false}, nil
}

// ----------------------------------------------------------------------------
// Date block / unblock
// ----------------------------------------------------------------------------

func (s *adminService) BlockDate(
	ctx context.Context,
	date string,
	adminID uuid.UUID,
	reason string,
) (*model.AdminBlockDateResponse, error) {
	// Idempotent-ish: explicit 409 if already blocked, per spec.
	if existing, err := s.dateBlocks.Find(ctx, date); err == nil {
		_ = existing
		return nil, ErrAlreadyBlocked
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	now := s.clock.Now()
	var reasonPtr *string
	if r := strings.TrimSpace(reason); r != "" {
		reasonPtr = &r
	}
	rec := &model.DateBlock{
		Date:        date,
		BlockedBy:   adminID,
		BlockedAt:   now,
		BlockReason: reasonPtr,
	}
	if err := s.dateBlocks.Create(ctx, rec); err != nil {
		return nil, err
	}
	return &model.AdminBlockDateResponse{
		Date:      date,
		Blocked:   true,
		Reason:    rec.BlockReason,
		BlockedAt: now,
	}, nil
}

func (s *adminService) UnblockDate(ctx context.Context, date string) (*model.AdminUnblockDateResponse, error) {
	if err := s.dateBlocks.Delete(ctx, date); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrBookingNotFound // 404 NOT_FOUND per spec
		}
		return nil, err
	}
	return &model.AdminUnblockDateResponse{Date: date, Blocked: false}, nil
}

// ----------------------------------------------------------------------------
// Kill-switch
// ----------------------------------------------------------------------------

func (s *adminService) SetBookingsEnabled(
	ctx context.Context,
	enabled bool,
	reason string,
	adminID uuid.UUID,
) (*model.AdminSetBookingsEnabledResponse, error) {
	var reasonPtr *string
	if r := strings.TrimSpace(reason); r != "" {
		reasonPtr = &r
	}
	settings, err := s.system.SetBookingsEnabled(ctx, enabled, reasonPtr, adminID)
	if err != nil {
		return nil, err
	}
	return &model.AdminSetBookingsEnabledResponse{
		BookingsEnabled: settings.BookingsEnabled,
		Reason:          settings.Reason,
		UpdatedAt:       settings.UpdatedAt,
	}, nil
}

// ----------------------------------------------------------------------------
// Cancel booking
// ----------------------------------------------------------------------------

func (s *adminService) CancelBooking(
	ctx context.Context,
	bookingID uuid.UUID,
	adminID uuid.UUID,
	reason string,
) (*model.AdminCancelBookingResponse, error) {
	b, err := s.bookings.FindByID(ctx, bookingID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookingNotFound
	}
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()

	// Idempotent: cancelling an already-cancelled booking returns 200.
	if b.Status == model.StatusCancelled {
		cancelledAt := now
		if b.CancelledAt != nil {
			cancelledAt = *b.CancelledAt
		}
		existingReason := reason
		if b.CancelReason != nil {
			existingReason = *b.CancelReason
		}
		return &model.AdminCancelBookingResponse{
			BookingID:   b.ID.String(),
			Status:      string(model.StatusCancelled),
			CancelledAt: cancelledAt,
			Reason:      existingReason,
		}, nil
	}

	if err := s.bookings.Cancel(ctx, bookingID, adminID, reason); err != nil {
		return nil, err
	}
	return &model.AdminCancelBookingResponse{
		BookingID:   b.ID.String(),
		Status:      string(model.StatusCancelled),
		CancelledAt: now,
		Reason:      reason,
	}, nil
}

func (s *adminService) UpsertSlot(
	ctx context.Context,
	date, slotTime string,
	capacityBig, capacityMedium int,
	adminID uuid.UUID,
) (*model.AdminUpsertSlotResponse, error) {

	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", date, err)
	}
	row := model.Slot{
		Date:           pgtype.Date{Time: t, Valid: true},
		Time:           slotTime,
		CapacityBig:    capacityBig,
		CapacityMedium: capacityMedium,
	}
	final, created, err := s.slots.Upsert(ctx, row)
	if err != nil {
		return nil, err
	}
	return &model.AdminUpsertSlotResponse{
		Date:           final.Date.Time.UTC().Format("2006-01-02"),
		Time:           final.Time,
		CapacityBig:    final.CapacityBig,
		CapacityMedium: final.CapacityMedium,
		Blocked:        final.Blocked,
		Created:        created,
	}, nil
}

func (s *adminService) GetSlotBookings(
	ctx context.Context,
	date, slotTime string,
) (*model.AdminSlotBookingsResponse, error) {
	// Verify the slot exists so we return 404 rather than an empty list for a
	// nonsense (date, time) pair.
	if _, err := s.slots.FindByDateTime(ctx, date, slotTime); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}

	bookings, err := s.bookings.FindBySlot(ctx, date, slotTime)
	if err != nil {
		return nil, err
	}

	entries := make([]model.AdminBookingListEntry, 0, len(bookings))
	for _, b := range bookings {
		entries = append(entries, model.AdminBookingListEntry{
			ID:                          b.ID.String(),
			UserEmail:                   b.UserEmail,
			FirstName:                   b.FirstName,
			LastName:                    b.LastName,
			Phone:                       b.Phone,
			Quantities:                  b.Quantities(),
			TotalAmount:                 b.TotalAmount,
			EffectiveAmount:             s.pricing.EffectiveAmount(b.TotalAmount, b.PriceOverride),
			Status:                      string(b.Status),
			CreatedAt:                   b.CreatedAt,
			PosterIncomingOrderID:       b.PosterIncomingOrderID,
			PosterIncomingTransactionID: b.PosterIncomingTransactionID,
		})
	}

	return &model.AdminSlotBookingsResponse{
		Date:     date,
		Time:     slotTime,
		Bookings: entries,
	}, nil
}
