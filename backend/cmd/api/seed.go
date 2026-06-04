// cmd/api/seed.go
package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
)

var seedSlotTimes = []string{"08:00", "10:00", "12:00", "14:00", "16:00", "18:00", "20:00"}

// seedSlots inserts the standard daily slots (one per route) for the date range.
// ON CONFLICT DO NOTHING on the (date, time, route_name) PK, so it is safe to re-run.
func seedSlots(
	ctx context.Context,
	db *gorm.DB,
	log *slog.Logger,
	start, end time.Time,
	capacityBig, capacityMedium, capacitySmall int,
) error {
	var all []model.Slot
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		for _, t := range seedSlotTimes {
			for _, route := range service.AllRoutes() {
				all = append(all, model.Slot{
					Date:           pgtype.Date{Time: d, Valid: true},
					Time:           t,
					RouteName:      route,
					CapacityBig:    capacityBig,
					CapacityMedium: capacityMedium,
					CapacitySmall:  capacitySmall,
				})
			}
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
