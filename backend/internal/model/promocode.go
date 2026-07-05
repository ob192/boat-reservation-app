package model

import (
	"time"

	"github.com/google/uuid"
)

// Promocode is an admin-generated affiliate discount code.
// `code` is the primary key and is stored uppercased/trimmed for case-insensitive lookup.
type Promocode struct {
	Code            string    `gorm:"type:varchar(64);primaryKey" json:"code"`
	DiscountPercent int       `gorm:"not null" json:"discountPercent"` // 0..100
	MaxUses         int       `gorm:"not null" json:"maxUses"`         // global redemption cap
	TimesUsed       int       `gorm:"not null;default:0" json:"timesUsed"`
	Active          bool      `gorm:"not null;default:true" json:"active"`
	CreatedBy       uuid.UUID `gorm:"type:uuid;not null;index" json:"createdBy"`
	CreatedAt       time.Time `json:"createdAt"`
}

func (Promocode) TableName() string { return "promocodes" }
