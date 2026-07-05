package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

type PromocodeRepository interface {
	Create(ctx context.Context, p *model.Promocode) error
	FindByCode(ctx context.Context, code string) (*model.Promocode, error)
	FindAll(ctx context.Context, createdBy *uuid.UUID) ([]model.Promocode, error)

	// IncrementUsage atomically bumps times_used only while under the cap.
	// Returns true if a row was updated, false if the cap was already reached.
	IncrementUsage(ctx context.Context, code string) (bool, error)
}

type promocodeRepo struct {
	db *gorm.DB
}

func NewPromocodeRepository(db *gorm.DB) PromocodeRepository {
	return &promocodeRepo{db: db}
}

func (r *promocodeRepo) tx(ctx context.Context) *gorm.DB {
	return platform.DBFromContext(ctx, r.db)
}

func (r *promocodeRepo) Create(ctx context.Context, p *model.Promocode) error {
	return r.tx(ctx).Create(p).Error
}

func (r *promocodeRepo) FindByCode(ctx context.Context, code string) (*model.Promocode, error) {
	var p model.Promocode
	err := r.tx(ctx).First(&p, "code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *promocodeRepo) FindAll(ctx context.Context, createdBy *uuid.UUID) ([]model.Promocode, error) {
	q := r.tx(ctx).Model(&model.Promocode{})
	if createdBy != nil {
		q = q.Where("created_by = ?", *createdBy)
	}
	var rows []model.Promocode
	err := q.Order("created_at DESC").Find(&rows).Error
	return rows, err
}

func (r *promocodeRepo) IncrementUsage(ctx context.Context, code string) (bool, error) {
	res := r.tx(ctx).
		Model(&model.Promocode{}).
		Where("code = ? AND times_used < max_uses", code).
		UpdateColumn("times_used", gorm.Expr("times_used + 1"))
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}
