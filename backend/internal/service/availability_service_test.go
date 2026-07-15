package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

func TestIsValidSlotTime(t *testing.T) {
	valid := []string{"00:00", "07:00", "12:30", "23:59"}
	for _, v := range valid {
		if !IsValidSlotTime(v) {
			t.Errorf("IsValidSlotTime(%q) = false, want true", v)
		}
	}
	invalid := []string{"", "7:00", "0700", "07-00", "24:00", "12:60", "ab:cd", "07:00:00", " 7:00"}
	for _, v := range invalid {
		if IsValidSlotTime(v) {
			t.Errorf("IsValidSlotTime(%q) = true, want false", v)
		}
	}
}

func TestParseMonth(t *testing.T) {
	start, end, err := ParseMonth("2026-08")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !start.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("start = %v", start)
	}
	if !end.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("end = %v (half-open upper bound)", end)
	}

	// December must roll over the year.
	_, end, err = ParseMonth("2026-12")
	if err != nil || !end.Equal(time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("december end = %v, err = %v", end, err)
	}

	for _, bad := range []string{"", "2026", "2026-13", "08-2026", "2026-8", "garbage"} {
		if _, _, err := ParseMonth(bad); err == nil {
			t.Errorf("ParseMonth(%q) should fail", bad)
		}
	}
}

func TestParseDate(t *testing.T) {
	d, err := ParseDate("2026-08-01")
	if err != nil || !d.Equal(time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("got (%v, %v)", d, err)
	}
	for _, bad := range []string{"", "2026-08", "01-08-2026", "2026-02-30", "garbage"} {
		if _, err := ParseDate(bad); err == nil {
			t.Errorf("ParseDate(%q) should fail", bad)
		}
	}
}

// availabilityFixture bundles the four repos with permissive defaults.
type availabilityFixture struct {
	slots      *mockSlotRepo
	bookings   *mockBookingRepo
	dateBlocks *mockDateBlockRepo
	system     *mockSystemRepo
}

func newAvailabilityFixture() *availabilityFixture {
	return &availabilityFixture{
		slots: &mockSlotRepo{
			findByMonth: func(context.Context, time.Time, time.Time, string) ([]model.Slot, error) { return nil, nil },
			findByDate:  func(context.Context, string, string) ([]model.Slot, error) { return nil, nil },
		},
		bookings: &mockBookingRepo{
			sumForRange: func(context.Context, string, string) (map[string]map[repository.SlotKey]repository.BookedQty, error) {
				return nil, nil
			},
			sumForDate: func(context.Context, string) (map[repository.SlotKey]repository.BookedQty, error) {
				return nil, nil
			},
		},
		dateBlocks: &mockDateBlockRepo{
			find: func(context.Context, string) (*model.DateBlock, error) { return nil, repository.ErrNotFound },
			findManyInRange: func(context.Context, string, string) ([]model.DateBlock, error) {
				return nil, nil
			},
		},
		system: bookingsEnabled(),
	}
}

func (f *availabilityFixture) service() AvailabilityService {
	return NewAvailabilityService(f.slots, f.bookings, f.dateBlocks, f.system)
}

func TestGetStatus(t *testing.T) {
	t.Run("enabled without reason", func(t *testing.T) {
		f := newAvailabilityFixture()
		got, err := f.service().GetStatus(context.Background())
		if err != nil || !got.BookingsEnabled || got.Reason != "" {
			t.Errorf("got (%+v, %v)", got, err)
		}
	})

	t.Run("disabled with reason", func(t *testing.T) {
		f := newAvailabilityFixture()
		f.system = &mockSystemRepo{
			get: func(context.Context) (*model.SystemSettings, error) {
				return &model.SystemSettings{BookingsEnabled: false, Reason: sptr("storm")}, nil
			},
		}
		got, err := f.service().GetStatus(context.Background())
		if err != nil || got.BookingsEnabled || got.Reason != "storm" {
			t.Errorf("got (%+v, %v)", got, err)
		}
	})

	t.Run("settings failure propagates", func(t *testing.T) {
		f := newAvailabilityFixture()
		boom := errors.New("db down")
		f.system = &mockSystemRepo{get: func(context.Context) (*model.SystemSettings, error) { return nil, boom }}
		if _, err := f.service().GetStatus(context.Background()); !errors.Is(err, boom) {
			t.Errorf("want settings error, got %v", err)
		}
	})
}

func TestGetMonth(t *testing.T) {
	const month = "2026-08" // 31 days

	t.Run("invalid month rejected", func(t *testing.T) {
		f := newAvailabilityFixture()
		if _, err := f.service().GetMonth(context.Background(), "2026-13", RouteDesna); err == nil {
			t.Error("want parse error")
		}
	})

	t.Run("fetch failures propagate", func(t *testing.T) {
		boom := errors.New("db down")

		f := newAvailabilityFixture()
		f.system = &mockSystemRepo{get: func(context.Context) (*model.SystemSettings, error) { return nil, boom }}
		if _, err := f.service().GetMonth(context.Background(), month, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("settings error: got %v", err)
		}

		f = newAvailabilityFixture()
		f.slots.findByMonth = func(context.Context, time.Time, time.Time, string) ([]model.Slot, error) { return nil, boom }
		if _, err := f.service().GetMonth(context.Background(), month, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("slots error: got %v", err)
		}

		f = newAvailabilityFixture()
		f.dateBlocks.findManyInRange = func(context.Context, string, string) ([]model.DateBlock, error) { return nil, boom }
		if _, err := f.service().GetMonth(context.Background(), month, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("date-block error: got %v", err)
		}

		f = newAvailabilityFixture()
		f.bookings.sumForRange = func(context.Context, string, string) (map[string]map[repository.SlotKey]repository.BookedQty, error) {
			return nil, boom
		}
		if _, err := f.service().GetMonth(context.Background(), month, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("booked-sum error: got %v", err)
		}
	})

	t.Run("kill-switch off zeroes every day and skips the booked query", func(t *testing.T) {
		f := newAvailabilityFixture()
		f.system = bookingsDisabled()
		f.bookings.sumForRange = nil // must not be called
		got, err := f.service().GetMonth(context.Background(), month, RouteDesna)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Month != month || len(got.Days) != 31 {
			t.Fatalf("month %q with %d days", got.Month, len(got.Days))
		}
		for _, d := range got.Days {
			if d.AvailableSlots != 0 || !d.FullyBlocked {
				t.Errorf("day %s: %+v, want zero availability", d.Date, d)
			}
		}
	})

	t.Run("full month shape", func(t *testing.T) {
		f := newAvailabilityFixture()
		f.dateBlocks.findManyInRange = func(context.Context, string, string) ([]model.DateBlock, error) {
			return []model.DateBlock{{Date: "2026-08-02"}}, nil
		}
		f.slots.findByMonth = func(context.Context, time.Time, time.Time, string) ([]model.Slot, error) {
			return []model.Slot{
				// 08-01: two open slots — availability is the MAX over slots, not the sum.
				*mkSlot("2026-08-01", "07:00", RouteDesna, 2, 1, 0),
				*mkSlot("2026-08-01", "10:00", RouteDesna, 1, 1, 1),
				// 08-02: has a slot but the date is blocked.
				*mkSlot("2026-08-02", "07:00", RouteDesna, 5, 5, 5),
				// 08-03: only blocked/cancelled slots.
				func() model.Slot { s := mkSlot("2026-08-03", "07:00", RouteDesna, 5, 5, 5); s.Blocked = true; return *s }(),
				func() model.Slot { s := mkSlot("2026-08-03", "10:00", RouteDesna, 5, 5, 5); s.Cancelled = true; return *s }(),
				// 08-04: fully booked → 0 available → fullyBlocked.
				*mkSlot("2026-08-04", "07:00", RouteDesna, 1, 0, 0),
				// 08-05: over-booked numbers must clamp at zero, not go negative.
				*mkSlot("2026-08-05", "07:00", RouteDesna, 1, 1, 0),
			}, nil
		}
		f.bookings.sumForRange = func(context.Context, string, string) (map[string]map[repository.SlotKey]repository.BookedQty, error) {
			key := repository.SlotKey{Time: "07:00", Route: RouteDesna}
			return map[string]map[repository.SlotKey]repository.BookedQty{
				"2026-08-01": {key: {Big: 1}},                       // 07:00 → 1+1+0=2 avail; 10:00 → 3
				"2026-08-04": {key: {Big: 1}},                       // full
				"2026-08-05": {key: {Big: 5, Medium: 5, Small: 5}},  // overbooked → clamp to 0
			}, nil
		}

		got, err := f.service().GetMonth(context.Background(), month, RouteDesna)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		byDate := map[string]model.AvailabilityDay{}
		for _, d := range got.Days {
			byDate[d.Date] = d
		}

		if d := byDate["2026-08-01"]; d.AvailableSlots != 3 || d.Blocked || d.FullyBlocked {
			t.Errorf("08-01 = %+v, want max(2,3)=3 available", d)
		}
		if d := byDate["2026-08-02"]; !d.Blocked || !d.FullyBlocked || d.AvailableSlots != 0 {
			t.Errorf("08-02 = %+v, want blocked date", d)
		}
		if d := byDate["2026-08-03"]; d.AvailableSlots != 0 || !d.FullyBlocked || d.Blocked {
			t.Errorf("08-03 = %+v, want zero (all slots blocked/cancelled) but not date-blocked", d)
		}
		if d := byDate["2026-08-04"]; d.AvailableSlots != 0 || !d.FullyBlocked {
			t.Errorf("08-04 = %+v, want fully booked", d)
		}
		if d := byDate["2026-08-05"]; d.AvailableSlots != 0 {
			t.Errorf("08-05 = %+v, negative availability must clamp to 0", d)
		}
		if d := byDate["2026-08-06"]; d.AvailableSlots != 0 || !d.FullyBlocked {
			t.Errorf("08-06 = %+v, day without slots must be fully blocked", d)
		}
	})
}

func TestGetDate(t *testing.T) {
	const date = "2026-08-01"

	t.Run("invalid date rejected", func(t *testing.T) {
		f := newAvailabilityFixture()
		if _, err := f.service().GetDate(context.Background(), "not-a-date", RouteDesna); err == nil {
			t.Error("want parse error")
		}
	})

	t.Run("fetch failures propagate", func(t *testing.T) {
		boom := errors.New("db down")
		f := newAvailabilityFixture()
		f.bookings.sumForDate = func(context.Context, string) (map[repository.SlotKey]repository.BookedQty, error) {
			return nil, boom
		}
		if _, err := f.service().GetDate(context.Background(), date, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("want booked-sum error, got %v", err)
		}

		f = newAvailabilityFixture()
		f.dateBlocks.find = func(context.Context, string) (*model.DateBlock, error) { return nil, boom }
		if _, err := f.service().GetDate(context.Background(), date, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("want date-block error, got %v", err)
		}
	})

	t.Run("maps slots with remaining capacity", func(t *testing.T) {
		f := newAvailabilityFixture()
		f.slots.findByDate = func(context.Context, string, string) ([]model.Slot, error) {
			blocked := mkSlot(date, "10:00", RouteDesna, 2, 2, 2)
			blocked.Blocked = true
			return []model.Slot{*mkSlot(date, "07:00", RouteDesna, 2, 1, 1), *blocked}, nil
		}
		f.bookings.sumForDate = func(context.Context, string) (map[repository.SlotKey]repository.BookedQty, error) {
			return map[repository.SlotKey]repository.BookedQty{
				{Time: "07:00", Route: RouteDesna}: {Big: 1, Medium: 5},
			}, nil
		}

		got, err := f.service().GetDate(context.Background(), date, RouteDesna)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Date != date || !got.BookingsEnabled || got.DateBlocked {
			t.Errorf("header wrong: %+v", got)
		}
		if got.FullyBlocked {
			t.Error("open capacity remains, must not be fully blocked")
		}
		if len(got.Slots) != 2 {
			t.Fatalf("slots = %+v", got.Slots)
		}
		s0 := got.Slots[0]
		if s0.AvailableBig != 1 || s0.TotalBig != 2 {
			t.Errorf("big availability = %d/%d, want 1/2", s0.AvailableBig, s0.TotalBig)
		}
		if s0.AvailableMedium != 0 {
			t.Errorf("overbooked medium must clamp to 0, got %d", s0.AvailableMedium)
		}
		if s0.AvailableSmall != 1 {
			t.Errorf("small availability = %d, want untouched 1", s0.AvailableSmall)
		}
		if !got.Slots[1].Blocked {
			t.Error("blocked slot must keep its flag")
		}
	})

	t.Run("fullyBlocked variants", func(t *testing.T) {
		// Bookings disabled.
		f := newAvailabilityFixture()
		f.system = bookingsDisabled()
		got, err := f.service().GetDate(context.Background(), date, RouteDesna)
		if err != nil || !got.FullyBlocked || got.BookingsEnabled {
			t.Errorf("disabled: (%+v, %v)", got, err)
		}

		// Date blocked.
		f = newAvailabilityFixture()
		f.dateBlocks.find = func(_ context.Context, d string) (*model.DateBlock, error) {
			return &model.DateBlock{Date: d}, nil
		}
		got, err = f.service().GetDate(context.Background(), date, RouteDesna)
		if err != nil || !got.FullyBlocked || !got.DateBlocked {
			t.Errorf("date-blocked: (%+v, %v)", got, err)
		}

		// No open slot: the only slot is cancelled.
		f = newAvailabilityFixture()
		f.slots.findByDate = func(context.Context, string, string) ([]model.Slot, error) {
			s := mkSlot(date, "07:00", RouteDesna, 2, 2, 2)
			s.Cancelled = true
			return []model.Slot{*s}, nil
		}
		got, err = f.service().GetDate(context.Background(), date, RouteDesna)
		if err != nil || !got.FullyBlocked {
			t.Errorf("cancelled-only: (%+v, %v)", got, err)
		}
	})
}

func TestGetSlotAvailability(t *testing.T) {
	const (
		date = "2026-08-01"
		tm   = "07:00"
	)

	t.Run("maps capacities, booked counts, and flags", func(t *testing.T) {
		f := newAvailabilityFixture()
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) {
			s := mkSlot(date, tm, RouteDesna, 3, 2, 1)
			s.Blocked = true
			return s, nil
		}
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 2, 1, 0, nil
		}
		f.dateBlocks.find = func(_ context.Context, d string) (*model.DateBlock, error) {
			return &model.DateBlock{Date: d}, nil
		}

		got, err := f.service().GetSlotAvailability(context.Background(), date, tm, RouteDesna)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.CapacityBig != 3 || got.CapacityMedium != 2 || got.CapacitySmall != 1 {
			t.Errorf("capacities wrong: %+v", got)
		}
		if got.BookedBig != 2 || got.BookedMedium != 1 || got.BookedSmall != 0 {
			t.Errorf("booked wrong: %+v", got)
		}
		if !got.SlotBlocked || got.SlotCancelled || !got.DateBlocked || !got.BookingsEnabled {
			t.Errorf("flags wrong: %+v", got)
		}
		if got.Date != date || got.Time != tm || got.RouteName != RouteDesna {
			t.Errorf("echo fields wrong: %+v", got)
		}
	})

	t.Run("slot lookup failure propagates", func(t *testing.T) {
		f := newAvailabilityFixture()
		boom := errors.New("db down")
		f.slots.findByDateTime = func(context.Context, string, string, string) (*model.Slot, error) { return nil, boom }
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 0, 0, 0, nil
		}
		if _, err := f.service().GetSlotAvailability(context.Background(), date, tm, RouteDesna); !errors.Is(err, boom) {
			t.Errorf("want slot error, got %v", err)
		}
	})
}
