package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

type adminFixture struct {
	bookings   *mockBookingRepo
	slots      *mockSlotRepo
	dateBlocks *mockDateBlockRepo
	system     *mockSystemRepo
}

func newAdminFixture() *adminFixture {
	return &adminFixture{
		bookings:   &mockBookingRepo{},
		slots:      &mockSlotRepo{},
		dateBlocks: &mockDateBlockRepo{},
		system:     &mockSystemRepo{},
	}
}

func (f *adminFixture) service(t *testing.T) AdminService {
	return NewAdminService(
		newTestDB(t), f.bookings, f.slots, f.dateBlocks, f.system,
		NewPricingService(), fakeClock{t: testNow},
	)
}

const (
	tDate  = "2026-08-01"
	tTime  = "07:00"
	tRoute = RouteDesna
)

var adminID = uuid.New()

// ----------------------------------------------------------------------------
// OverridePrice
// ----------------------------------------------------------------------------

func TestOverridePrice(t *testing.T) {
	t.Run("booking not found", func(t *testing.T) {
		f := newAdminFixture()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		}
		_, err := f.service(t).OverridePrice(context.Background(), uuid.New(), adminID, fptr(300), "vip")
		if !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("confirmed bookings cannot be overridden", func(t *testing.T) {
		f := newAdminFixture()
		b := mkBooking()
		b.Status = model.StatusConfirmed
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return b, nil }
		_, err := f.service(t).OverridePrice(context.Background(), b.ID, adminID, fptr(300), "vip")
		if !errors.Is(err, ErrAlreadyConfirmed) {
			t.Errorf("want ErrAlreadyConfirmed, got %v", err)
		}
	})

	t.Run("non-positive amount rejected", func(t *testing.T) {
		for _, amount := range []float64{0, -5} {
			f := newAdminFixture()
			f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return mkBooking(), nil }
			_, err := f.service(t).OverridePrice(context.Background(), uuid.New(), adminID, fptr(amount), "why")
			if !errors.Is(err, ErrInvalidInput) {
				t.Errorf("amount %v: want ErrInvalidInput, got %v", amount, err)
			}
		}
	})

	t.Run("set override and report effective amount", func(t *testing.T) {
		f := newAdminFixture()
		b := mkBooking() // total 450
		updated := *b
		updated.PriceOverride = fptr(300)
		updated.OverrideReason = sptr("vip guest")

		calls := 0
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			calls++
			if calls == 1 {
				return b, nil
			}
			return &updated, nil // refetch after the write
		}
		var gotOverride *float64
		var gotReason *string
		var gotAdmin *uuid.UUID
		f.bookings.setPriceOverride = func(_ context.Context, _ uuid.UUID, override *float64, reason *string, admin *uuid.UUID) error {
			gotOverride, gotReason, gotAdmin = override, reason, admin
			return nil
		}

		resp, err := f.service(t).OverridePrice(context.Background(), b.ID, adminID, fptr(300), "  vip guest  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotOverride == nil || *gotOverride != 300 {
			t.Errorf("override written as %v", gotOverride)
		}
		if gotReason == nil || *gotReason != "vip guest" {
			t.Errorf("reason must be trimmed, got %v", gotReason)
		}
		if gotAdmin == nil || *gotAdmin != adminID {
			t.Errorf("admin id written as %v", gotAdmin)
		}
		if resp.EffectiveAmount != 300 || resp.TotalAmount != 450 {
			t.Errorf("resp = %+v, want effective 300 over total 450", resp)
		}
		if resp.PriceOverride == nil || *resp.PriceOverride != 300 {
			t.Errorf("resp override = %v", resp.PriceOverride)
		}
	})

	t.Run("nil amount clears the override", func(t *testing.T) {
		f := newAdminFixture()
		b := mkBooking()
		cleared := *b // no override
		calls := 0
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			calls++
			if calls == 1 {
				return b, nil
			}
			return &cleared, nil
		}
		var gotOverride *float64
		reasonWritten := false
		f.bookings.setPriceOverride = func(_ context.Context, _ uuid.UUID, override *float64, reason *string, _ *uuid.UUID) error {
			gotOverride = override
			reasonWritten = reason != nil
			return nil
		}

		resp, err := f.service(t).OverridePrice(context.Background(), b.ID, adminID, nil, "ignored")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotOverride != nil {
			t.Errorf("override must be cleared with nil, got %v", gotOverride)
		}
		if reasonWritten {
			t.Error("clearing must not write a reason")
		}
		if resp.EffectiveAmount != 450 {
			t.Errorf("effective = %v, want the plain total 450", resp.EffectiveAmount)
		}
	})

	t.Run("write failure propagates", func(t *testing.T) {
		f := newAdminFixture()
		boom := errors.New("update failed")
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return mkBooking(), nil }
		f.bookings.setPriceOverride = func(context.Context, uuid.UUID, *float64, *string, *uuid.UUID) error { return boom }
		if _, err := f.service(t).OverridePrice(context.Background(), uuid.New(), adminID, fptr(300), "vip"); !errors.Is(err, boom) {
			t.Errorf("want write error, got %v", err)
		}
	})
}

// ----------------------------------------------------------------------------
// BlockSlot / UnblockSlot
// ----------------------------------------------------------------------------

func TestBlockSlot(t *testing.T) {
	t.Run("slot not found", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		_, err := f.service(t).BlockSlot(context.Background(), tDate, tTime, tRoute, adminID, "")
		if !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("already blocked conflicts", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 1, 1, 1)
			s.Blocked = true
			return s, nil
		}
		_, err := f.service(t).BlockSlot(context.Background(), tDate, tTime, tRoute, adminID, "")
		if !errors.Is(err, ErrAlreadyBlocked) {
			t.Errorf("want ErrAlreadyBlocked, got %v", err)
		}
	})

	t.Run("success writes trimmed reason and echoes the block", func(t *testing.T) {
		f := newAdminFixture()
		blockedAt := testNow.Add(-time.Hour)
		calls := 0
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			calls++
			s := mkSlot(d, tm, r, 1, 1, 1)
			if calls > 1 { // refetch after Block
				s.Blocked = true
				s.BlockReason = sptr("maintenance")
				s.BlockedAt = &blockedAt
			}
			return s, nil
		}
		var gotReason *string
		f.slots.block = func(_ context.Context, _, _, _ string, admin uuid.UUID, reason *string) error {
			if admin != adminID {
				t.Errorf("blocked by %v, want %v", admin, adminID)
			}
			gotReason = reason
			return nil
		}

		resp, err := f.service(t).BlockSlot(context.Background(), tDate, tTime, tRoute, adminID, "  maintenance ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotReason == nil || *gotReason != "maintenance" {
			t.Errorf("reason written as %v, want trimmed", gotReason)
		}
		if !resp.Blocked || resp.Date != tDate || resp.Time != tTime || resp.RouteName != tRoute {
			t.Errorf("resp = %+v", resp)
		}
		if !resp.BlockedAt.Equal(blockedAt) {
			t.Errorf("blockedAt = %v, want the persisted %v", resp.BlockedAt, blockedAt)
		}
	})

	t.Run("empty reason is stored as nil", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 1, 1, 1), nil
		}
		var gotReason *string
		called := false
		f.slots.block = func(_ context.Context, _, _, _ string, _ uuid.UUID, reason *string) error {
			called = true
			gotReason = reason
			return nil
		}
		if _, err := f.service(t).BlockSlot(context.Background(), tDate, tTime, tRoute, adminID, "   "); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called || gotReason != nil {
			t.Errorf("blank reason must be nil, got %v (called=%v)", gotReason, called)
		}
	})
}

func TestUnblockSlot(t *testing.T) {
	t.Run("slot not found", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).UnblockSlot(context.Background(), tDate, tTime, tRoute); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 1, 1, 1), nil
		}
		unblocked := false
		f.slots.unblock = func(context.Context, string, string, string) error {
			unblocked = true
			return nil
		}
		resp, err := f.service(t).UnblockSlot(context.Background(), tDate, tTime, tRoute)
		if err != nil || !unblocked || resp.Blocked {
			t.Errorf("resp=%+v err=%v unblocked=%v", resp, err, unblocked)
		}
	})
}

// ----------------------------------------------------------------------------
// BlockDate / UnblockDate
// ----------------------------------------------------------------------------

func TestBlockDate(t *testing.T) {
	t.Run("already blocked conflicts", func(t *testing.T) {
		f := newAdminFixture()
		f.dateBlocks.find = func(_ context.Context, d string) (*model.DateBlock, error) {
			return &model.DateBlock{Date: d}, nil
		}
		if _, err := f.service(t).BlockDate(context.Background(), tDate, adminID, ""); !errors.Is(err, ErrAlreadyBlocked) {
			t.Errorf("want ErrAlreadyBlocked, got %v", err)
		}
	})

	t.Run("success snapshots clock and reason", func(t *testing.T) {
		f := newAdminFixture()
		f.dateBlocks = noDateBlocks()
		var saved *model.DateBlock
		f.dateBlocks.create = func(_ context.Context, db *model.DateBlock) error {
			saved = db
			return nil
		}
		resp, err := f.service(t).BlockDate(context.Background(), tDate, adminID, " storm ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if saved.Date != tDate || saved.BlockedBy != adminID || !saved.BlockedAt.Equal(testNow) {
			t.Errorf("saved = %+v", saved)
		}
		if saved.BlockReason == nil || *saved.BlockReason != "storm" {
			t.Errorf("reason = %v, want trimmed storm", saved.BlockReason)
		}
		if !resp.Blocked || resp.Date != tDate || !resp.BlockedAt.Equal(testNow) {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("create failure propagates", func(t *testing.T) {
		f := newAdminFixture()
		f.dateBlocks = noDateBlocks()
		boom := errors.New("insert failed")
		f.dateBlocks.create = func(context.Context, *model.DateBlock) error { return boom }
		if _, err := f.service(t).BlockDate(context.Background(), tDate, adminID, ""); !errors.Is(err, boom) {
			t.Errorf("want insert error, got %v", err)
		}
	})
}

func TestUnblockDate(t *testing.T) {
	t.Run("missing block maps to not found", func(t *testing.T) {
		f := newAdminFixture()
		f.dateBlocks.del = func(context.Context, string) error { return repository.ErrNotFound }
		if _, err := f.service(t).UnblockDate(context.Background(), tDate); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want not-found sentinel, got %v", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		f := newAdminFixture()
		f.dateBlocks.del = func(context.Context, string) error { return nil }
		resp, err := f.service(t).UnblockDate(context.Background(), tDate)
		if err != nil || resp.Blocked || resp.Date != tDate {
			t.Errorf("resp=%+v err=%v", resp, err)
		}
	})
}

// ----------------------------------------------------------------------------
// SetBookingsEnabled
// ----------------------------------------------------------------------------

func TestSetBookingsEnabled(t *testing.T) {
	t.Run("passes through and maps the row", func(t *testing.T) {
		f := newAdminFixture()
		var gotEnabled bool
		var gotReason *string
		f.system.setBookingsEnabled = func(_ context.Context, enabled bool, reason *string, admin uuid.UUID) (*model.SystemSettings, error) {
			gotEnabled = enabled
			gotReason = reason
			if admin != adminID {
				t.Errorf("admin = %v", admin)
			}
			return &model.SystemSettings{BookingsEnabled: enabled, Reason: reason, UpdatedAt: testNow}, nil
		}
		resp, err := f.service(t).SetBookingsEnabled(context.Background(), false, "  storm  ", adminID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotEnabled || gotReason == nil || *gotReason != "storm" {
			t.Errorf("written (%v, %v)", gotEnabled, gotReason)
		}
		if resp.BookingsEnabled || resp.Reason == nil || *resp.Reason != "storm" || !resp.UpdatedAt.Equal(testNow) {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("blank reason becomes nil", func(t *testing.T) {
		f := newAdminFixture()
		var gotReason *string
		f.system.setBookingsEnabled = func(_ context.Context, enabled bool, reason *string, _ uuid.UUID) (*model.SystemSettings, error) {
			gotReason = reason
			return &model.SystemSettings{BookingsEnabled: enabled}, nil
		}
		if _, err := f.service(t).SetBookingsEnabled(context.Background(), true, "  ", adminID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotReason != nil {
			t.Errorf("reason = %v, want nil", gotReason)
		}
	})
}

// ----------------------------------------------------------------------------
// CancelBooking
// ----------------------------------------------------------------------------

func TestCancelBooking(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		f := newAdminFixture()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).CancelBooking(context.Background(), uuid.New(), adminID, "no-show"); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("cancelling twice is idempotent and keeps the original audit", func(t *testing.T) {
		f := newAdminFixture()
		b := mkBooking()
		b.Status = model.StatusCancelled
		prior := testNow.Add(-24 * time.Hour)
		b.CancelledAt = &prior
		b.CancelReason = sptr("original reason")
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return b, nil }
		f.bookings.cancel = nil // must not be called again

		resp, err := f.service(t).CancelBooking(context.Background(), b.ID, adminID, "new reason")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Status != string(model.StatusCancelled) {
			t.Errorf("status = %s", resp.Status)
		}
		if !resp.CancelledAt.Equal(prior) || resp.Reason != "original reason" {
			t.Errorf("original audit must win: %+v", resp)
		}
	})

	t.Run("cancels an active booking", func(t *testing.T) {
		f := newAdminFixture()
		b := mkBooking()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return b, nil }
		var gotReason string
		f.bookings.cancel = func(_ context.Context, id uuid.UUID, admin uuid.UUID, reason string) error {
			if id != b.ID || admin != adminID {
				t.Errorf("cancel(%v, %v)", id, admin)
			}
			gotReason = reason
			return nil
		}
		resp, err := f.service(t).CancelBooking(context.Background(), b.ID, adminID, "weather")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotReason != "weather" || resp.Reason != "weather" {
			t.Errorf("reason = %q / %q", gotReason, resp.Reason)
		}
		if !resp.CancelledAt.Equal(testNow) {
			t.Errorf("cancelledAt = %v, want clock %v", resp.CancelledAt, testNow)
		}
	})

	t.Run("write failure propagates", func(t *testing.T) {
		f := newAdminFixture()
		boom := errors.New("update failed")
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return mkBooking(), nil }
		f.bookings.cancel = func(context.Context, uuid.UUID, uuid.UUID, string) error { return boom }
		if _, err := f.service(t).CancelBooking(context.Background(), uuid.New(), adminID, "x"); !errors.Is(err, boom) {
			t.Errorf("want write error, got %v", err)
		}
	})
}

// ----------------------------------------------------------------------------
// UpsertSlot / GetSlotBookings / ListBookings
// ----------------------------------------------------------------------------

func TestUpsertSlot(t *testing.T) {
	t.Run("invalid date rejected", func(t *testing.T) {
		f := newAdminFixture()
		if _, err := f.service(t).UpsertSlot(context.Background(), "bad-date", tTime, tRoute, 1, 1, 1, adminID); err == nil {
			t.Error("want date parse error")
		}
	})

	t.Run("passes capacities through and reports created", func(t *testing.T) {
		f := newAdminFixture()
		var gotRow model.Slot
		f.slots.upsert = func(_ context.Context, s model.Slot) (*model.Slot, bool, error) {
			gotRow = s
			return &s, true, nil
		}
		resp, err := f.service(t).UpsertSlot(context.Background(), tDate, tTime, tRoute, 3, 2, 1, adminID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotRow.CapacityBig != 3 || gotRow.CapacityMedium != 2 || gotRow.CapacitySmall != 1 {
			t.Errorf("upserted row = %+v", gotRow)
		}
		if !resp.Created || resp.Date != tDate || resp.Time != tTime || resp.CapacityBig != 3 {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("update path reports created=false", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.upsert = func(_ context.Context, s model.Slot) (*model.Slot, bool, error) {
			return &s, false, nil
		}
		resp, err := f.service(t).UpsertSlot(context.Background(), tDate, tTime, tRoute, 1, 1, 1, adminID)
		if err != nil || resp.Created {
			t.Errorf("resp=%+v err=%v", resp, err)
		}
	})
}

func TestGetSlotBookings(t *testing.T) {
	t.Run("slot must exist", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).GetSlotBookings(context.Background(), tDate, tTime, tRoute); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("maps bookings incl. effective amount", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 1, 1, 1), nil
		}
		b := mkBooking()
		b.PriceOverride = fptr(300)
		f.bookings.findBySlot = func(context.Context, string, string, string) ([]model.Booking, error) {
			return []model.Booking{*b}, nil
		}
		resp, err := f.service(t).GetSlotBookings(context.Background(), tDate, tTime, tRoute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Bookings) != 1 {
			t.Fatalf("bookings = %+v", resp.Bookings)
		}
		e := resp.Bookings[0]
		if e.ID != b.ID.String() || e.TotalAmount != 450 || e.EffectiveAmount != 300 {
			t.Errorf("entry = %+v, want override-aware amounts", e)
		}
		if e.Date != tDate || e.Time != tTime || e.Status != string(model.StatusPending) {
			t.Errorf("entry slot fields = %+v", e)
		}
	})
}

func TestListBookings(t *testing.T) {
	f := newAdminFixture()
	var gotFilter repository.BookingHistoryFilter
	f.bookings.findAllForAdmin = func(_ context.Context, filter repository.BookingHistoryFilter) ([]model.Booking, error) {
		gotFilter = filter
		return []model.Booking{*mkBooking()}, nil
	}
	resp, err := f.service(t).ListBookings(context.Background(), tDate, "pending", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := repository.BookingHistoryFilter{Date: tDate, Status: "pending", Limit: 10, Offset: 20}
	if gotFilter != want {
		t.Errorf("filter = %+v, want %+v", gotFilter, want)
	}
	if len(resp.Bookings) != 1 {
		t.Errorf("bookings = %+v", resp.Bookings)
	}
}

// ----------------------------------------------------------------------------
// CancelSlot / UncancelSlot (transactional)
// ----------------------------------------------------------------------------

func TestCancelSlot(t *testing.T) {
	t.Run("slot not found", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.lockForUpdate = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).CancelSlot(context.Background(), tDate, tTime, tRoute, adminID, ""); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("already cancelled conflicts", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 1, 1, 1)
			s.Cancelled = true
			return s, nil
		}
		if _, err := f.service(t).CancelSlot(context.Background(), tDate, tTime, tRoute, adminID, ""); !errors.Is(err, ErrAlreadyCancelled) {
			t.Errorf("want ErrAlreadyCancelled, got %v", err)
		}
	})

	t.Run("cancels the slot and cascades to bookings", func(t *testing.T) {
		f := newAdminFixture()
		cancelledAt := testNow.Add(-time.Minute)
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 1, 1, 1), nil
		}
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 1, 1, 1)
			s.Cancelled = true
			s.CancelReason = sptr("storm")
			s.CancelledAt = &cancelledAt
			return s, nil
		}
		var slotReason *string
		f.slots.cancel = func(_ context.Context, _, _, _ string, admin uuid.UUID, reason *string) error {
			if admin != adminID {
				t.Errorf("cancelled by %v", admin)
			}
			slotReason = reason
			return nil
		}
		var cascadeReason string
		f.bookings.cancelBySlot = func(_ context.Context, d, tm, r string, admin uuid.UUID, reason string) (int64, error) {
			if d != tDate || tm != tTime || r != tRoute || admin != adminID {
				t.Errorf("cascade args (%s %s %s %v)", d, tm, r, admin)
			}
			cascadeReason = reason
			return 4, nil
		}

		resp, err := f.service(t).CancelSlot(context.Background(), tDate, tTime, tRoute, adminID, " storm ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slotReason == nil || *slotReason != "storm" {
			t.Errorf("slot reason = %v", slotReason)
		}
		if cascadeReason != " storm " {
			// The cascade passes the raw reason through — assert current behaviour.
			t.Errorf("cascade reason = %q", cascadeReason)
		}
		if !resp.Cancelled || resp.CancelledBookings != 4 {
			t.Errorf("resp = %+v, want 4 cascaded cancellations", resp)
		}
		if !resp.CancelledAt.Equal(cancelledAt) || resp.Reason == nil || *resp.Reason != "storm" {
			t.Errorf("resp audit = %+v", resp)
		}
	})

	t.Run("cascade failure rolls up", func(t *testing.T) {
		f := newAdminFixture()
		boom := errors.New("cascade failed")
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 1, 1, 1), nil
		}
		f.slots.cancel = func(context.Context, string, string, string, uuid.UUID, *string) error { return nil }
		f.bookings.cancelBySlot = func(context.Context, string, string, string, uuid.UUID, string) (int64, error) {
			return 0, boom
		}
		if _, err := f.service(t).CancelSlot(context.Background(), tDate, tTime, tRoute, adminID, ""); !errors.Is(err, boom) {
			t.Errorf("want cascade error, got %v", err)
		}
	})
}

func TestUncancelSlot(t *testing.T) {
	t.Run("slot not found", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).UncancelSlot(context.Background(), tDate, tTime, tRoute); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("success re-opens the slot", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.findByDateTime = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 1, 1, 1)
			s.Cancelled = true
			return s, nil
		}
		uncancelled := false
		f.slots.uncancel = func(context.Context, string, string, string) error {
			uncancelled = true
			return nil
		}
		resp, err := f.service(t).UncancelSlot(context.Background(), tDate, tTime, tRoute)
		if err != nil || !uncancelled || resp.Cancelled {
			t.Errorf("resp=%+v err=%v uncancelled=%v", resp, err, uncancelled)
		}
	})
}

// ----------------------------------------------------------------------------
// MoveBooking (transactional)
// ----------------------------------------------------------------------------

func TestMoveBooking(t *testing.T) {
	// moveFixture: a 2/1/1 destination slot with 1/0/1 already booked, and a
	// pending 1-big booking currently on another slot.
	setup := func() (*adminFixture, *model.Booking) {
		f := newAdminFixture()
		b := mkBooking() // sits on 2026-08-01 07:00 Desna, QtyBig=1
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return b, nil }
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 2, 1, 1), nil
		}
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 1, 0, 1, nil
		}
		f.bookings.move = func(context.Context, uuid.UUID, string, string, string) error { return nil }
		return f, b
	}

	t.Run("input validation", func(t *testing.T) {
		f := newAdminFixture()
		svc := f.service(t)
		if _, err := svc.MoveBooking(context.Background(), uuid.New(), tDate, tTime, "Atlantis"); !errors.Is(err, ErrInvalidRoute) {
			t.Errorf("bad route: got %v", err)
		}
		if _, err := svc.MoveBooking(context.Background(), uuid.New(), tDate, "7am", tRoute); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("bad time: got %v", err)
		}
		if _, err := svc.MoveBooking(context.Background(), uuid.New(), "bad", tTime, tRoute); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("bad date: got %v", err)
		}
	})

	t.Run("booking not found", func(t *testing.T) {
		f, _ := setup()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).MoveBooking(context.Background(), uuid.New(), "2026-08-02", tTime, tRoute); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("terminal statuses are immutable", func(t *testing.T) {
		for _, st := range []model.BookingStatus{model.StatusCancelled, model.StatusExpired, model.StatusFailed} {
			f, b := setup()
			b.Status = st
			if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, ErrBookingNotPending) {
				t.Errorf("status %s: want ErrBookingNotPending, got %v", st, err)
			}
		}
	})

	t.Run("destination slot gates", func(t *testing.T) {
		f, b := setup()
		f.slots.lockForUpdate = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("missing dest: got %v", err)
		}

		f, b = setup()
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 2, 1, 1)
			s.Blocked = true
			return s, nil
		}
		if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, ErrSlotBlocked) {
			t.Errorf("blocked dest: got %v", err)
		}

		f, b = setup()
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 2, 1, 1)
			s.Cancelled = true
			return s, nil
		}
		if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, ErrSlotCancelled) {
			t.Errorf("cancelled dest: got %v", err)
		}
	})

	t.Run("destination without capacity conflicts", func(t *testing.T) {
		f, b := setup()
		b.QtyBig = 2 // 2 big > 2-1 remaining
		if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, ErrSlotTaken) {
			t.Errorf("want ErrSlotTaken, got %v", err)
		}
	})

	t.Run("moving within the same slot excludes its own quantities", func(t *testing.T) {
		f, b := setup()
		b.QtyBig = 2
		// Destination == current slot; booked sums (2/0/0) include this booking itself.
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 2, 0, 0, nil
		}
		moved := false
		f.bookings.move = func(context.Context, uuid.UUID, string, string, string) error {
			moved = true
			return nil
		}
		resp, err := f.service(t).MoveBooking(context.Background(), b.ID, tDate, tTime, tRoute)
		if err != nil {
			t.Fatalf("self-move must not count its own seats: %v", err)
		}
		if !moved || resp.BookingID != b.ID.String() {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("confirmed bookings can be moved", func(t *testing.T) {
		f, b := setup()
		b.Status = model.StatusConfirmed
		resp, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Status != string(model.StatusConfirmed) || resp.Date != "2026-08-02" {
			t.Errorf("resp = %+v", resp)
		}
	})

	t.Run("move failure rolls up", func(t *testing.T) {
		f, b := setup()
		boom := errors.New("move failed")
		f.bookings.move = func(context.Context, uuid.UUID, string, string, string) error { return boom }
		if _, err := f.service(t).MoveBooking(context.Background(), b.ID, "2026-08-02", tTime, tRoute); !errors.Is(err, boom) {
			t.Errorf("want move error, got %v", err)
		}
	})
}

// ----------------------------------------------------------------------------
// DeleteSlot (transactional)
// ----------------------------------------------------------------------------

func TestDeleteSlot(t *testing.T) {
	t.Run("slot not found", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.lockForUpdate = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
		if err := f.service(t).DeleteSlot(context.Background(), tDate, tTime, tRoute); !errors.Is(err, ErrSlotNotFound) {
			t.Errorf("want ErrSlotNotFound, got %v", err)
		}
	})

	t.Run("slot with active bookings refuses deletion", func(t *testing.T) {
		for _, booked := range [][3]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}} {
			f := newAdminFixture()
			f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
				return mkSlot(d, tm, r, 5, 5, 5), nil
			}
			b := booked
			f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
				return b[0], b[1], b[2], nil
			}
			f.slots.del = nil // must not be called
			if err := f.service(t).DeleteSlot(context.Background(), tDate, tTime, tRoute); !errors.Is(err, ErrSlotNotEmpty) {
				t.Errorf("booked %v: want ErrSlotNotEmpty, got %v", booked, err)
			}
		}
	})

	t.Run("empty slot deleted", func(t *testing.T) {
		f := newAdminFixture()
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 5, 5, 5), nil
		}
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 0, 0, 0, nil
		}
		deleted := false
		f.slots.del = func(context.Context, string, string, string) error {
			deleted = true
			return nil
		}
		if err := f.service(t).DeleteSlot(context.Background(), tDate, tTime, tRoute); err != nil || !deleted {
			t.Errorf("err=%v deleted=%v", err, deleted)
		}
	})
}
