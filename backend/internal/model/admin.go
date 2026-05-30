package model

import (
	"github.com/google/uuid"
)

// Admin is a row in the `admins` table.
// Adding a user here grants them access to all /admin/* endpoints.
// Removing the row immediately revokes access (no restart required).
type Admin struct {
	UserID uuid.UUID `gorm:"type:uuid;primaryKey"   json:"userId"`
}

func (Admin) TableName() string { return "admins" }
