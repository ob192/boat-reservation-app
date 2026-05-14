package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// ValidSlotTimes is the canonical, ordered list of bookable time slots.
// Any time received from the client outside this list is rejected.
var ValidSlotTimes = []string{"08:00", "11:00", "15:00", "19:00"}

// IsValidSlotTime reports whether t is one of the four allowed slot times.
func IsValidSlotTime(t string) bool {
	for _, v := range ValidSlotTimes {
		if v == t {
			return true
		}
	}
	return false
}

// AvailabilityService provides per-day and per-slot availability views.
type AvailabilityService interface {
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

func (s *availabilityService) GetMonth(ctx context.Context, month string) (*model.AvailabilityMonthResponse, error) {
	start, end, err := ParseMonth(month)
	if err != nil {
		return nil, err
	}

	// 1) Pull all the raw data for the range. Three small queries; combined in Go.
	settings, err := s.system.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	slots, err := s.slots.FindByMonth(ctx, start, end)
	if err != nil {
		return nil, fmt.Errorf("find slots: %w", err)
	}

	dateBlocks, err := s.dateBlocks.FindManyInRange(ctx,
		start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("find date blocks: %w", err)
	}
	blockedDates := make(map[string]bool, len(dateBlocks))
	for _, db := range dateBlocks {
		blockedDates[db.Date] = true
	}

	// 2) Group slots by date.
	slotsByDate := make(map[string][]model.Slot)
	for _, sl := range slots {
		slotsByDate[sl.DateString()] = append(slotsByDate[sl.DateString()], sl)
	}

	// 3) For each unique date in the result set, fetch booked sums per time slot.
	//    For a monthly view this is 1 query per date — fine for the ~30-day cardinality.
	//    If this ever becomes a hotspot, we can batch with one query per month.
	days := make([]model.AvailabilityDay, 0)
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		isBlocked := blockedDates[dateStr]

		// If global kill-switch is off, expose 0 available slots.
		if !settings.BookingsEnabled {
			days = append(days, model.AvailabilityDay{
				Date:           dateStr,
				AvailableSlots: 0,
				Blocked:        isBlocked,
			})
			continue
		}

		// If the date itself is blocked, no point summing up bookings.
		if isBlocked {
			days = append(days, model.AvailabilityDay{
				Date:           dateStr,
				AvailableSlots: 0,
				Blocked:        true,
			})
			continue
		}

		daySlots := slotsByDate[dateStr]
		if len(daySlots) == 0 {
			// No physical slots seeded for this date.
			days = append(days, model.AvailabilityDay{
				Date:           dateStr,
				AvailableSlots: 0,
				Blocked:        false,
			})
			continue
		}

		bookedByTime, err := s.bookings.SumActiveQuantitiesForDate(ctx, dateStr)
		if err != nil {
			return nil, fmt.Errorf("sum quantities for %s: %w", dateStr, err)
		}

		// Per spec: availableSlots = MAX over unblocked time slots of (avail_big + avail_medium).
		max := 0
		for _, sl := range daySlots {
			if sl.Blocked {
				continue
			}
			b := bookedByTime[sl.Time]
			availBig := sl.CapacityBig - b.Big
			if availBig < 0 {
				availBig = 0
			}
			availMedium := sl.CapacityMedium - b.Medium
			if availMedium < 0 {
				availMedium = 0
			}
			total := availBig + availMedium
			if total > max {
				max = total
			}
		}
		days = append(days, model.AvailabilityDay{
			Date:           dateStr,
			AvailableSlots: max,
			Blocked:        false,
		})
	}

	return &model.AvailabilityMonthResponse{
		Month: month,
		Days:  days,
	}, nil
}

func (s *availabilityService) GetDate(ctx context.Context, date string) (*model.SlotsForDateResponse, error) {
	if _, err := ParseDate(date); err != nil {
		return nil, err
	}

	settings, err := s.system.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	dateBlocked := false
	if _, err := s.dateBlocks.Find(ctx, date); err == nil {
		dateBlocked = true
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("find date block: %w", err)
	}

	slots, err := s.slots.FindByDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("find slots: %w", err)
	}

	// Pull aggregate booked quantities for the day in one query.
	bookedByTime, err := s.bookings.SumActiveQuantitiesForDate(ctx, date)
	if err != nil {
		return nil, fmt.Errorf("sum quantities: %w", err)
	}

	// Build a map from time → slot so we can return the canonical 4 even if some are missing.
	byTime := make(map[string]model.Slot, len(slots))
	for _, sl := range slots {
		byTime[sl.Time] = sl
	}

	out := make([]model.SlotForDate, 0, len(ValidSlotTimes))
	for _, t := range ValidSlotTimes {
		sl, ok := byTime[t]
		if !ok {
			// No physical slot row → expose zero capacity rather than 404 the date.
			out = append(out, model.SlotForDate{
				Time:    t,
				Blocked: false,
			})
			continue
		}
		b := bookedByTime[t]
		availBig := sl.CapacityBig - b.Big
		if availBig < 0 {
			availBig = 0
		}
		availMedium := sl.CapacityMedium - b.Medium
		if availMedium < 0 {
			availMedium = 0
		}
		out = append(out, model.SlotForDate{
			Time:            t,
			AvailableBig:    availBig,
			TotalBig:        sl.CapacityBig,
			AvailableMedium: availMedium,
			TotalMedium:     sl.CapacityMedium,
			Blocked:         sl.Blocked,
		})
	}

	return &model.SlotsForDateResponse{
		Date:            date,
		DateBlocked:     dateBlocked,
		BookingsEnabled: settings.BookingsEnabled,
		Slots:           out,
	}, nil
}

func (s *availabilityService) GetSlotAvailability(ctx context.Context, date, time string) (*model.SlotAvailability, error) {
	settings, err := s.system.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}

	dateBlocked := false
	if _, err := s.dateBlocks.Find(ctx, date); err == nil {
		dateBlocked = true
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, fmt.Errorf("find date block: %w", err)
	}

	slot, err := s.slots.FindByDateTime(ctx, date, time)
	if err != nil {
		return nil, err
	}

	bb, bm, err := s.bookings.SumActiveQuantitiesForSlot(ctx, date, time)
	if err != nil {
		return nil, fmt.Errorf("sum quantities: %w", err)
	}

	return &model.SlotAvailability{
		Date:            date,
		Time:            time,
		CapacityBig:     slot.CapacityBig,
		CapacityMedium:  slot.CapacityMedium,
		BookedBig:       bb,
		BookedMedium:    bm,
		SlotBlocked:     slot.Blocked,
		DateBlocked:     dateBlocked,
		BookingsEnabled: settings.BookingsEnabled,
	}, nil
}
