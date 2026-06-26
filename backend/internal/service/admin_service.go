package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
	"gorm.io/gorm"
)

// ErrAlreadyBlocked Admin-specific typed errors.
var (
	ErrAlreadyBlocked   = errors.New("ALREADY_BLOCKED")
	ErrAlreadyCancelled = errors.New("ALREADY_CANCELLED")
	ErrSlotNotEmpty     = errors.New("SLOT_NOT_EMPTY")
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

	BlockSlot(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminBlockSlotResponse, error)

	UnblockSlot(ctx context.Context, date, time, route string) (*model.AdminUnblockSlotResponse, error)

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

	CancelSlot(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminCancelSlotResponse, error)

	UncancelSlot(ctx context.Context, date, time, route string) (*model.AdminUncancelSlotResponse, error)

	UpsertSlot(ctx context.Context, date, time, route string, capacityBig, capacityMedium, capacitySmall int, adminID uuid.UUID) (*model.AdminUpsertSlotResponse, error)

	GetSlotBookings(ctx context.Context, date, time, route string) (*model.AdminSlotBookingsResponse, error)

	ListBookings(ctx context.Context, date, status string, limit, offset int) (*model.AdminBookingHistoryResponse, error)

	MoveBooking(ctx context.Context, bookingID uuid.UUID, date, slotTime, route string) (*model.AdminMoveBookingResponse, error)

	DeleteSlot(ctx context.Context, date, time, route string) error
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
	date, time, route string,
	adminID uuid.UUID,
	reason string,
) (*model.AdminBlockSlotResponse, error) {
	slot, err := s.slots.FindByDateTime(ctx, date, time, route)
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
	if err := s.slots.Block(ctx, date, time, route, adminID, reasonPtr); err != nil {
		return nil, err
	}

	updated, err := s.slots.FindByDateTime(ctx, date, time, route)
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
		RouteName: route,
		Blocked:   true,
		Reason:    updated.BlockReason,
		BlockedAt: blockedAt,
	}, nil
}

func (s *adminService) UnblockSlot(
	ctx context.Context,
	date, time, route string,
) (*model.AdminUnblockSlotResponse, error) {
	if _, err := s.slots.FindByDateTime(ctx, date, time, route); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}
	if err := s.slots.Unblock(ctx, date, time, route); err != nil {
		return nil, err
	}
	return &model.AdminUnblockSlotResponse{Date: date, Time: time, RouteName: route, Blocked: false}, nil
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
	date, slotTime, route string,
	capacityBig, capacityMedium, capacitySmall int,
	adminID uuid.UUID,
) (*model.AdminUpsertSlotResponse, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", date, err)
	}
	row := model.Slot{
		Date:           pgtype.Date{Time: t, Valid: true},
		Time:           slotTime,
		RouteName:      route,
		CapacityBig:    capacityBig,
		CapacityMedium: capacityMedium,
		CapacitySmall:  capacitySmall,
	}
	final, created, err := s.slots.Upsert(ctx, row)
	if err != nil {
		return nil, err
	}
	return &model.AdminUpsertSlotResponse{
		Date:           final.Date.Time.UTC().Format("2006-01-02"),
		Time:           final.Time,
		RouteName:      final.RouteName,
		CapacityBig:    final.CapacityBig,
		CapacityMedium: final.CapacityMedium,
		CapacitySmall:  final.CapacitySmall,
		Blocked:        final.Blocked,
		Created:        created,
	}, nil
}

func (s *adminService) GetSlotBookings(
	ctx context.Context,
	date, slotTime, route string,
) (*model.AdminSlotBookingsResponse, error) {
	if _, err := s.slots.FindByDateTime(ctx, date, slotTime, route); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}

	bookings, err := s.bookings.FindBySlot(ctx, date, slotTime, route)
	if err != nil {
		return nil, err
	}

	entries := make([]model.AdminBookingListEntry, 0, len(bookings))
	for _, b := range bookings {
		entries = append(entries, s.toAdminEntry(b))
	}

	return &model.AdminSlotBookingsResponse{
		Date:      date,
		Time:      slotTime,
		RouteName: route,
		Bookings:  entries,
	}, nil
}

func (s *adminService) ListBookings(
	ctx context.Context,
	date, status string,
	limit, offset int,
) (*model.AdminBookingHistoryResponse, error) {
	bookings, err := s.bookings.FindAllForAdmin(ctx, repository.BookingHistoryFilter{
		Date:   date,
		Status: status,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	entries := make([]model.AdminBookingListEntry, 0, len(bookings))
	for _, b := range bookings {
		entries = append(entries, s.toAdminEntry(b))
	}
	return &model.AdminBookingHistoryResponse{Bookings: entries}, nil
}

func (s *adminService) CancelSlot(
	ctx context.Context,
	date, slotTime, route string,
	adminID uuid.UUID,
	reason string,
) (*model.AdminCancelSlotResponse, error) {
	var resp *model.AdminCancelSlotResponse

	err := platform.WithTx(ctx, s.db, func(ctx context.Context) error {
		slot, err := s.slots.LockForUpdate(ctx, date, slotTime, route)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSlotNotFound
		}
		if err != nil {
			return err
		}
		if slot.Cancelled {
			return ErrAlreadyCancelled
		}

		var reasonPtr *string
		if r := strings.TrimSpace(reason); r != "" {
			reasonPtr = &r
		}
		if err := s.slots.Cancel(ctx, date, slotTime, route, adminID, reasonPtr); err != nil {
			return err
		}

		// Cascade: cancel all active bookings on this slot.
		cancelledCount, err := s.bookings.CancelBySlot(ctx, date, slotTime, route, adminID, reason)
		if err != nil {
			return err
		}

		updated, err := s.slots.FindByDateTime(ctx, date, slotTime, route)
		if err != nil {
			return err
		}
		cancelledAt := s.clock.Now()
		if updated.CancelledAt != nil {
			cancelledAt = *updated.CancelledAt
		}
		resp = &model.AdminCancelSlotResponse{
			Date:              date,
			Time:              slotTime,
			RouteName:         route,
			Cancelled:         true,
			Reason:            updated.CancelReason,
			CancelledAt:       cancelledAt,
			CancelledBookings: cancelledCount,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UncancelSlot re-opens the slot for booking. It does NOT reinstate bookings
// that were cancelled by a previous CancelSlot — those stay cancelled.
func (s *adminService) UncancelSlot(
	ctx context.Context,
	date, slotTime, route string,
) (*model.AdminUncancelSlotResponse, error) {
	if _, err := s.slots.FindByDateTime(ctx, date, slotTime, route); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSlotNotFound
		}
		return nil, err
	}
	if err := s.slots.Uncancel(ctx, date, slotTime, route); err != nil {
		return nil, err
	}
	return &model.AdminUncancelSlotResponse{Date: date, Time: slotTime, RouteName: route, Cancelled: false}, nil
}

func (s *adminService) MoveBooking(
	ctx context.Context,
	bookingID uuid.UUID,
	date, slotTime, route string,
) (*model.AdminMoveBookingResponse, error) {
	if !IsValidRoute(route) {
		return nil, ErrInvalidRoute
	}
	if !IsValidSlotTime(slotTime) {
		return nil, ErrInvalidInput
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, ErrInvalidInput
	}

	var resp *model.AdminMoveBookingResponse
	err := platform.WithTx(ctx, s.db, func(ctx context.Context) error {
		b, err := s.bookings.FindByID(ctx, bookingID)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrBookingNotFound
		}
		if err != nil {
			return err
		}

		// Only active bookings can be moved; terminal states are immutable.
		if b.Status != model.StatusPending && b.Status != model.StatusConfirmed {
			return ErrBookingNotPending
		}

		// Lock the destination slot and verify it can take this booking.
		slot, err := s.slots.LockForUpdate(ctx, date, slotTime, route)
		if errors.Is(err, repository.ErrNotFound) {
			return ErrSlotNotFound
		}
		if err != nil {
			return err
		}
		if slot.Blocked {
			return ErrSlotBlocked
		}
		if slot.Cancelled {
			return ErrSlotCancelled
		}

		bookedBig, bookedMedium, bookedSmall, err := s.bookings.SumActiveQuantitiesForSlot(ctx, date, slotTime, route)
		if err != nil {
			return err
		}
		// Exclude this booking's own quantities if it already sits on the destination slot.
		if b.DateFormatted() == date && b.Time == slotTime && b.RouteName == route {
			bookedBig -= b.QtyBig
			bookedMedium -= b.QtyMedium
			bookedSmall -= b.QtySmall
		}
		if b.QtyBig > slot.CapacityBig-bookedBig ||
			b.QtyMedium > slot.CapacityMedium-bookedMedium ||
			b.QtySmall > slot.CapacitySmall-bookedSmall {
			return ErrSlotTaken
		}

		if err := s.bookings.Move(ctx, bookingID, date, slotTime, route); err != nil {
			return err
		}

		resp = &model.AdminMoveBookingResponse{
			BookingID: b.ID.String(),
			Date:      date,
			Time:      slotTime,
			RouteName: route,
			Status:    string(b.Status),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *adminService) DeleteSlot(ctx context.Context, date, slotTime, route string) error {
	return platform.WithTx(ctx, s.db, func(ctx context.Context) error {
		if _, err := s.slots.LockForUpdate(ctx, date, slotTime, route); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrSlotNotFound
			}
			return err
		}

		// "Empty" = no confirmed bookings and no unexpired pending holds.
		bookedBig, bookedMedium, bookedSmall, err := s.bookings.SumActiveQuantitiesForSlot(ctx, date, slotTime, route)
		if err != nil {
			return err
		}
		if bookedBig > 0 || bookedMedium > 0 || bookedSmall > 0 {
			return ErrSlotNotEmpty
		}

		return s.slots.Delete(ctx, date, slotTime, route)
	})
}

func (s *adminService) toAdminEntry(b model.Booking) model.AdminBookingListEntry {
	return model.AdminBookingListEntry{
		ID:                          b.ID.String(),
		Date:                        b.DateFormatted(),
		Time:                        b.Time,
		RouteName:                   b.RouteName,
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
	}
}
