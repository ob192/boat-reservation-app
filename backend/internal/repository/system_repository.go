package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

// SystemRepository accesses the singleton system_settings row.
type SystemRepository interface {
	Get(ctx context.Context) (*model.SystemSettings, error)
	SetBookingsEnabled(ctx context.Context, enabled bool, reason *string, adminID uuid.UUID) (*model.SystemSettings, error)
}

type systemRepo struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) SystemRepository {
	return &systemRepo{db: db}
}

func (r *systemRepo) tx(ctx context.Context) *gorm.DB {
	return platform.DBFromContext(ctx, r.db)
}

// Get returns the singleton settings row. If it doesn't exist, it's created
// with sensible defaults (bookings_enabled = true) and returned.
func (r *systemRepo) Get(ctx context.Context) (*model.SystemSettings, error) {
	var s model.SystemSettings
	err := r.tx(ctx).First(&s, "id = ?", model.SystemSettingsRowID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s = model.SystemSettings{
			ID:              model.SystemSettingsRowID,
			BookingsEnabled: true,
			UpdatedAt:       time.Now().UTC(),
		}
		if err := r.tx(ctx).Create(&s).Error; err != nil {
			return nil, err
		}
		return &s, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *systemRepo) SetBookingsEnabled(
	ctx context.Context,
	enabled bool,
	reason *string,
	adminID uuid.UUID,
) (*model.SystemSettings, error) {
	// Make sure the row exists first.
	if _, err := r.Get(ctx); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	err := r.tx(ctx).
		Model(&model.SystemSettings{}).
		Where("id = ?", model.SystemSettingsRowID).
		Updates(map[string]any{
			"bookings_enabled": enabled,
			"reason":           reason,
			"updated_by":       adminID,
			"updated_at":       now,
		}).Error
	if err != nil {
		return nil, err
	}
	return r.Get(ctx)
}
