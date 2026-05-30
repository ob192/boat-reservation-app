package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

// AdminRepository checks whether a user is in the admins table.
type AdminRepository interface {
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

type adminRepo struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepo{db: db}
}

func (r *adminRepo) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	var row model.Admin
	err := platform.DBFromContext(ctx, r.db).
		Where("user_id = ?", userID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
