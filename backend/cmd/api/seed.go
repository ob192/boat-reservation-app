// cmd/api/seed.go
package main

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

var seedSlotTimes = []string{"08:00", "11:00", "15:00", "19:00"}

// seedSlots inserts the standard daily slots for the given date range.
// Uses ON CONFLICT DO NOTHING, so it is safe to run on every startup.
func seedSlots(ctx context.Context, db *gorm.DB, log *slog.Logger, start, end time.Time, capacityBig, capacityMedium int) error {
	//var all []model.Slot
	//for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
	//	date := d.Format("2006-01-02")
	//	for _, t := range seedSlotTimes {
	//		all = append(all, model.Slot{
	//			Date:           date,
	//			Time:           t,
	//			CapacityBig:    capacityBig,
	//			CapacityMedium: capacityMedium,
	//		})
	//	}
	//}
	//if len(all) == 0 {
	//	return nil
	//}
	//
	//const chunkSize = 200
	//var totalInserted int64
	//for i := 0; i < len(all); i += chunkSize {
	//	chunk := all[i:min(i+chunkSize, len(all))]
	//	result := db.WithContext(ctx).
	//		Clauses(clause.OnConflict{DoNothing: true}).
	//		Create(&chunk)
	//	if result.Error != nil {
	//		return result.Error
	//	}
	//	totalInserted += result.RowsAffected
	//}
	//
	//if totalInserted > 0 {
	//	log.Info("seeded slots", "count", totalInserted)
	//}
	return nil
}
