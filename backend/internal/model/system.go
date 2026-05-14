package model

import (
	"time"

	"github.com/google/uuid"
)

// SystemSettingsRowID is the fixed primary-key value for the singleton settings row.
const SystemSettingsRowID = 1

// SystemSettings is a single-row table holding global feature flags.
type SystemSettings struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	BookingsEnabled bool       `gorm:"not null;default:true" json:"bookingsEnabled"`
	Reason          *string    `gorm:"type:varchar(255)" json:"reason,omitempty"`
	UpdatedBy       *uuid.UUID `gorm:"type:uuid" json:"updatedBy,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

func (SystemSettings) TableName() string { return "system_settings" }
