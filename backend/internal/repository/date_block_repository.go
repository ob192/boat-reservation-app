package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

// DateBlockRepository persists whole-date admin blocks.
type DateBlockRepository interface {
	Find(ctx context.Context, date string) (*model.DateBlock, error)
	Create(ctx context.Context, db *model.DateBlock) error
	Delete(ctx context.Context, date string) error

	// FindManyInRange returns date blocks falling within [start, end) — used by the
	// monthly availability query.
	FindManyInRange(ctx context.Context, startDate, endDate string) ([]model.DateBlock, error)
}

type dateBlockRepo struct {
	db *gorm.DB
}

func NewDateBlockRepository(db *gorm.DB) DateBlockRepository {
	return &dateBlockRepo{db: db}
}

func (r *dateBlockRepo) tx(ctx context.Context) *gorm.DB {
	return platform.DBFromContext(ctx, r.db)
}

func (r *dateBlockRepo) Find(ctx context.Context, date string) (*model.DateBlock, error) {
	var db model.DateBlock
	err := r.tx(ctx).Where("date = ?", date).First(&db).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (r *dateBlockRepo) Create(ctx context.Context, db *model.DateBlock) error {
	// Spec returns 409 ALREADY_BLOCKED when re-creating; the service layer
	// detects "already blocked" *before* calling Create, so a duplicate-PK
	// error here would only happen on a race and bubbles up unchanged.
	return r.tx(ctx).Create(db).Error
}

func (r *dateBlockRepo) Delete(ctx context.Context, date string) error {
	res := r.tx(ctx).Where("date = ?", date).Delete(&model.DateBlock{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *dateBlockRepo) FindManyInRange(ctx context.Context, startDate, endDate string) ([]model.DateBlock, error) {
	var rows []model.DateBlock
	err := r.tx(ctx).
		Where("date >= ? AND date < ?", startDate, endDate).
		Order("date ASC").
		Find(&rows).Error
	return rows, err
}
