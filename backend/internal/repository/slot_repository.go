package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

// SlotRepository persists fleet capacity and per-slot block state.
type SlotRepository interface {
	FindByDateTime(ctx context.Context, date, time, route string) (*model.Slot, error)
	FindByDate(ctx context.Context, date, route string) ([]model.Slot, error)
	FindByMonth(ctx context.Context, monthStart, monthEnd time.Time, route string) ([]model.Slot, error)

	LockForUpdate(ctx context.Context, date, time, route string) (*model.Slot, error)

	Block(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error
	Unblock(ctx context.Context, date, time, route string) error

	Cancel(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error
	Uncancel(ctx context.Context, date, time, route string) error

	Upsert(ctx context.Context, s model.Slot) (result *model.Slot, created bool, err error)
}

type slotRepo struct {
	db *gorm.DB
}

func NewSlotRepository(db *gorm.DB) SlotRepository {
	return &slotRepo{db: db}
}

func (r *slotRepo) tx(ctx context.Context) *gorm.DB {
	return platform.DBFromContext(ctx, r.db)
}

func (r *slotRepo) FindByDateTime(ctx context.Context, date, time, route string) (*model.Slot, error) {
	var s model.Slot
	err := r.tx(ctx).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *slotRepo) FindByDate(ctx context.Context, date, route string) ([]model.Slot, error) {
	var slots []model.Slot
	err := r.tx(ctx).
		Where("date = ? AND route_name = ?", date, route).
		Order("time ASC").
		Find(&slots).Error
	if err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *slotRepo) FindByMonth(ctx context.Context, monthStart, monthEnd time.Time, route string) ([]model.Slot, error) {
	var slots []model.Slot
	err := r.tx(ctx).
		Where("date >= ? AND date < ? AND route_name = ?",
			monthStart.Format("2006-01-02"), monthEnd.Format("2006-01-02"), route).
		Order("date ASC, time ASC").
		Find(&slots).Error
	return slots, err
}

func (r *slotRepo) LockForUpdate(ctx context.Context, date, time, route string) (*model.Slot, error) {
	var s model.Slot
	err := r.tx(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *slotRepo) Block(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error {
	now := timeNow()
	res := r.tx(ctx).
		Model(&model.Slot{}).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		Updates(map[string]any{
			"blocked":      true,
			"blocked_by":   adminID,
			"blocked_at":   now,
			"block_reason": reason,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *slotRepo) Unblock(ctx context.Context, date, time, route string) error {
	res := r.tx(ctx).
		Model(&model.Slot{}).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		Updates(map[string]any{
			"blocked":      false,
			"blocked_by":   nil,
			"blocked_at":   nil,
			"block_reason": nil,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *slotRepo) Cancel(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error {
	now := timeNow()
	res := r.tx(ctx).
		Model(&model.Slot{}).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		Updates(map[string]any{
			"cancelled":     true,
			"cancelled_by":  adminID,
			"cancelled_at":  now,
			"cancel_reason": reason,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *slotRepo) Uncancel(ctx context.Context, date, time, route string) error {
	res := r.tx(ctx).
		Model(&model.Slot{}).
		Where("date = ? AND time = ? AND route_name = ?", date, time, route).
		Updates(map[string]any{
			"cancelled":     false,
			"cancelled_by":  nil,
			"cancelled_at":  nil,
			"cancel_reason": nil,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *slotRepo) Upsert(ctx context.Context, s model.Slot) (*model.Slot, bool, error) {
	res := r.tx(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}, {Name: "time"}, {Name: "route_name"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"capacity_big",
				"capacity_medium",
				"capacity_small",
			}),
		}).
		Create(&s)
	if res.Error != nil {
		return nil, false, res.Error
	}
	wasCreated := res.RowsAffected == 1
	final, err := r.FindByDateTime(ctx, s.DateString(), s.Time, s.RouteName)
	if err != nil {
		return nil, wasCreated, err
	}
	return final, wasCreated, nil
}

// timeNow is a small indirection to allow injecting a clock in future tests.
var timeNow = func() time.Time { return time.Now().UTC() }
