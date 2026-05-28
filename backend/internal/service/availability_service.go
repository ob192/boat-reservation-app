package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

func IsValidSlotTime(t string) bool {
	if len(t) != 5 || t[2] != ':' {
		return false
	}
	_, err := time.Parse("15:04", t)
	return err == nil
}

// AvailabilityService provides per-day and per-slot availability views.
type AvailabilityService interface {
	GetStatus(ctx context.Context) (*model.BookingStatusResponse, error)
	GetMonth(ctx context.Context, month string) (*model.AvailabilityMonthResponse, error)
	GetDate(ctx context.Context, date string) (*model.SlotsForDateResponse, error)

	// GetSlotAvailability returns a fully-populated SlotAvailability for a single (date, time).
	// Used by the booking flow to surface specific block reasons.
	GetSlotAvailability(ctx context.Context, date, time string) (*model.SlotAvailability, error)
}

type availabilityService struct {
	slots      repository.SlotRepository
	bookings   repository.BookingRepository
	dateBlocks repository.DateBlockRepository
	system     repository.SystemRepository
}

func NewAvailabilityService(
	slots repository.SlotRepository,
	bookings repository.BookingRepository,
	dateBlocks repository.DateBlockRepository,
	system repository.SystemRepository,
) AvailabilityService {
	return &availabilityService{
		slots:      slots,
		bookings:   bookings,
		dateBlocks: dateBlocks,
		system:     system,
	}
}

// ParseMonth validates `YYYY-MM` and returns the [start, end) range bounding the month.
func ParseMonth(s string) (time.Time, time.Time, error) {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid month %q: %w", s, err)
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

// ParseDate validates `YYYY-MM-DD` and returns it normalised.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: %w", s, err)
	}
	return t, nil
}

// GetMonth returns daily availability for an entire month.
//
// Performance notes:
//   - All independent reads (settings, slots, date_blocks, optionally booked sums)
//     run concurrently via errgroup — wall time ≈ slowest single query.
//   - Booked quantities are aggregated for the whole month in ONE query
//     (SumActiveQuantitiesForRange), not N queries-per-day.
//   - When the global kill-switch is OFF we skip the booked-sum query entirely
//     since every day will report zero anyway.
func (s *availabilityService) GetMonth(ctx context.Context, month string) (*model.AvailabilityMonthResponse, error) {
	start, end, err := ParseMonth(month)
	if err != nil {
		return nil, err
	}
	startStr := start.Format("2006-01-02")
	endStr := end.Format("2006-01-02")

	// Phase 1: concurrent fetch of settings + date blocks + slots.
	// We can't decide whether to fetch booked sums until we know whether the
	// kill-switch is on, so it's a second phase.
	var (
		settings   *model.SystemSettings
		slots      []model.Slot
		dateBlocks []model.DateBlock
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := s.system.Get(gctx)
		if err != nil {
			return fmt.Errorf("get settings: %w", err)
		}
		settings = v
		return nil
	})
	g.Go(func() error {
		v, err := s.slots.FindByMonth(gctx, start, end)
		if err != nil {
			return fmt.Errorf("find slots: %w", err)
		}
		slots = v
		return nil
	})
	g.Go(func() error {
		v, err := s.dateBlocks.FindManyInRange(gctx, startStr, endStr)
		if err != nil {
			return fmt.Errorf("find date blocks: %w", err)
		}
		dateBlocks = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	blockedDates := make(map[string]bool, len(dateBlocks))
	for _, db := range dateBlocks {
		blockedDates[db.Date] = true
	}

	// Phase 2: only query bookings if the kill-switch lets days have availability.
	// (If bookings are globally disabled we still expose date-block state, but every
	// day reports 0 available slots regardless of bookings.)
	var bookedByDateTime map[string]map[string]repository.BookedQty
	if settings.BookingsEnabled {
		bookedByDateTime, err = s.bookings.SumActiveQuantitiesForRange(ctx, startStr, endStr)
		if err != nil {
			return nil, fmt.Errorf("sum quantities for range %s..%s: %w", startStr, endStr, err)
		}
	}

	// Group slots by date.
	slotsByDate := make(map[string][]model.Slot, 31)
	for _, sl := range slots {
		d := sl.DateString()
		slotsByDate[d] = append(slotsByDate[d], sl)
	}

	// Assemble per-day output.
	days := make([]model.AvailabilityDay, 0, 31)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		isBlocked := blockedDates[dateStr]

		// Kill-switch OFF or date blocked → zero available.
		if !settings.BookingsEnabled || isBlocked {
			days = append(days, model.AvailabilityDay{
				Date:           dateStr,
				AvailableSlots: 0,
				Blocked:        isBlocked,
				FullyBlocked:   true,
			})
			continue
		}

		daySlots := slotsByDate[dateStr]
		if len(daySlots) == 0 {
			days = append(days, model.AvailabilityDay{
				Date:           dateStr,
				AvailableSlots: 0,
				Blocked:        false,
				FullyBlocked:   true,
			})
			continue
		}

		bookedForDay := bookedByDateTime[dateStr] // nil-safe lookup below

		// Per spec: availableSlots = MAX over unblocked time slots of (avail_big + avail_medium).
		maxAvail := 0
		for _, sl := range daySlots {
			if sl.Blocked {
				continue
			}
			b := bookedForDay[sl.Time] // zero value if missing — fine
			availBig := sl.CapacityBig - b.Big
			if availBig < 0 {
				availBig = 0
			}
			availMedium := sl.CapacityMedium - b.Medium
			if availMedium < 0 {
				availMedium = 0
			}
			total := availBig + availMedium
			if total > maxAvail {
				maxAvail = total
			}
		}
		days = append(days, model.AvailabilityDay{
			Date:           dateStr,
			AvailableSlots: maxAvail,
			Blocked:        false,
			FullyBlocked:   maxAvail == 0,
		})
	}

	return &model.AvailabilityMonthResponse{
		Month: month,
		Days:  days,
	}, nil
}

func (s *availabilityService) GetStatus(ctx context.Context) (*model.BookingStatusResponse, error) {
	settings, err := s.system.Get(ctx)
	if err != nil {
		return nil, err
	}
	reason := ""
	if settings.Reason != nil {
		reason = *settings.Reason
	}
	return &model.BookingStatusResponse{
		BookingsEnabled: settings.BookingsEnabled,
		Reason:          reason,
	}, nil
}

// GetDate runs settings, date-block, slots, and booked-sum reads concurrently.
// This endpoint fires every time the user picks a day on the calendar, so
// parallelizing the 4 independent queries trims real latency.
func (s *availabilityService) GetDate(ctx context.Context, date string) (*model.SlotsForDateResponse, error) {
	if _, err := ParseDate(date); err != nil {
		return nil, err
	}

	var (
		settings     *model.SystemSettings
		slots        []model.Slot
		bookedByTime map[string]struct{ Big, Medium int }
		dateBlocked  bool
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := s.system.Get(gctx)
		if err != nil {
			return fmt.Errorf("get settings: %w", err)
		}
		settings = v
		return nil
	})
	g.Go(func() error {
		if _, err := s.dateBlocks.Find(gctx, date); err == nil {
			dateBlocked = true
			return nil
		} else if errors.Is(err, repository.ErrNotFound) {
			return nil
		} else {
			return fmt.Errorf("find date block: %w", err)
		}
	})
	g.Go(func() error {
		v, err := s.slots.FindByDate(gctx, date)
		if err != nil {
			return fmt.Errorf("find slots: %w", err)
		}
		slots = v
		return nil
	})
	g.Go(func() error {
		v, err := s.bookings.SumActiveQuantitiesForDate(gctx, date)
		if err != nil {
			return fmt.Errorf("sum quantities: %w", err)
		}
		bookedByTime = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Build a map from time → slot so we can return the canonical 4 even if some are missing.
	byTime := make(map[string]model.Slot, len(slots))
	for _, sl := range slots {
		byTime[sl.Time] = sl
	}

	out := make([]model.SlotForDate, 0, len(slots))
	for _, sl := range slots {
		b := bookedByTime[sl.Time]
		availBig := sl.CapacityBig - b.Big
		if availBig < 0 {
			availBig = 0
		}
		availMedium := sl.CapacityMedium - b.Medium
		if availMedium < 0 {
			availMedium = 0
		}
		out = append(out, model.SlotForDate{
			Time:            sl.Time,
			AvailableBig:    availBig,
			TotalBig:        sl.CapacityBig,
			AvailableMedium: availMedium,
			TotalMedium:     sl.CapacityMedium,
			Blocked:         sl.Blocked,
		})
	}

	fullyBlocked := !settings.BookingsEnabled || dateBlocked
	if !fullyBlocked {
		anyOpen := false
		for _, sl := range out {
			if !sl.Blocked && (sl.AvailableBig > 0 || sl.AvailableMedium > 0) {
				anyOpen = true
				break
			}
		}
		fullyBlocked = !anyOpen
	}

	return &model.SlotsForDateResponse{
		Date:            date,
		DateBlocked:     dateBlocked,
		BookingsEnabled: settings.BookingsEnabled,
		FullyBlocked:    fullyBlocked,
		Slots:           out,
	}, nil
}

// GetSlotAvailability — same parallel-fetch trick. Called inside the booking
// path, so every millisecond shaved off helps under concurrent load.
func (s *availabilityService) GetSlotAvailability(ctx context.Context, date, time string) (*model.SlotAvailability, error) {
	var (
		settings    *model.SystemSettings
		slot        *model.Slot
		bookedBig   int
		bookedMed   int
		dateBlocked bool
	)

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := s.system.Get(gctx)
		if err != nil {
			return fmt.Errorf("get settings: %w", err)
		}
		settings = v
		return nil
	})
	g.Go(func() error {
		if _, err := s.dateBlocks.Find(gctx, date); err == nil {
			dateBlocked = true
			return nil
		} else if errors.Is(err, repository.ErrNotFound) {
			return nil
		} else {
			return fmt.Errorf("find date block: %w", err)
		}
	})
	g.Go(func() error {
		v, err := s.slots.FindByDateTime(gctx, date, time)
		if err != nil {
			return err
		}
		slot = v
		return nil
	})
	g.Go(func() error {
		bb, bm, err := s.bookings.SumActiveQuantitiesForSlot(gctx, date, time)
		if err != nil {
			return fmt.Errorf("sum quantities: %w", err)
		}
		bookedBig, bookedMed = bb, bm
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &model.SlotAvailability{
		Date:            date,
		Time:            time,
		CapacityBig:     slot.CapacityBig,
		CapacityMedium:  slot.CapacityMedium,
		BookedBig:       bookedBig,
		BookedMedium:    bookedMed,
		SlotBlocked:     slot.Blocked,
		DateBlocked:     dateBlocked,
		BookingsEnabled: settings.BookingsEnabled,
	}, nil
}
