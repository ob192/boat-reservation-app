// cmd/api/seed.go
package main

import (
	"context"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm/clause"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

var seedSlotTimes = []string{"08:00", "10:00", "12:00", "14:00", "16:00", "18:00", "20:00"}

// seedSlots inserts the standard daily slots for the given date range.
// Uses ON CONFLICT DO NOTHING, so it is safe to run on every startup.
func seedSlots(ctx context.Context, db *gorm.DB, log *slog.Logger, start, end time.Time, capacityBig, capacityMedium int) error {
	var all []model.Slot
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		for _, t := range seedSlotTimes {
			all = append(all, model.Slot{
				Date:           pgtype.Date{Time: d, Valid: true},
				Time:           t,
				CapacityBig:    capacityBig,
				CapacityMedium: capacityMedium,
			})
		}
	}
	if len(all) == 0 {
		return nil
	}

	const chunkSize = 200
	var totalInserted int64
	for i := 0; i < len(all); i += chunkSize {
		chunk := all[i:min(i+chunkSize, len(all))]
		result := db.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&chunk)
		if result.Error != nil {
			return result.Error
		}
		totalInserted += result.RowsAffected
	}

	if totalInserted > 0 {
		log.Info("seeded slots", "count", totalInserted)
	}
	return nil
}
