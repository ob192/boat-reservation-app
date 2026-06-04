package model

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"

	"github.com/google/uuid"
)

// Slot defines physical fleet capacity and block state for a (date, time, route) tuple.
// Primary key is composite: (date, time, route_name).
type Slot struct {
	Date      pgtype.Date `gorm:"type:date;primaryKey" json:"date"`       // YYYY-MM-DD
	Time      string      `gorm:"type:varchar(5);primaryKey" json:"time"` // HH:MM
	RouteName string      `gorm:"type:varchar(50);primaryKey" json:"routeName"`

	CapacityBig    int `gorm:"not null;default:0" json:"capacityBig"`
	CapacityMedium int `gorm:"not null;default:0" json:"capacityMedium"`
	CapacitySmall  int `gorm:"not null;default:0" json:"capacitySmall"`

	Blocked     bool       `gorm:"not null;default:false" json:"blocked"`
	BlockedBy   *uuid.UUID `gorm:"type:uuid" json:"blockedBy,omitempty"`
	BlockedAt   *time.Time `json:"blockedAt,omitempty"`
	BlockReason *string    `gorm:"type:varchar(255)" json:"blockReason,omitempty"`
}

func (Slot) TableName() string { return "slots" }

func (s Slot) DateString() string {
	return s.Date.Time.UTC().Format("2006-01-02")
}

// DateBlock marks a whole date as admin-blocked.
type DateBlock struct {
	Date        string    `gorm:"type:date;primaryKey" json:"date"`
	BlockedBy   uuid.UUID `gorm:"type:uuid;not null" json:"blockedBy"`
	BlockedAt   time.Time `gorm:"not null" json:"blockedAt"`
	BlockReason *string   `gorm:"type:varchar(255)" json:"blockReason,omitempty"`
}

func (DateBlock) TableName() string { return "date_blocks" }

// SlotAvailability is the projection used by the availability endpoints.
type SlotAvailability struct {
	Date            string
	Time            string
	RouteName       string
	CapacityBig     int
	CapacityMedium  int
	CapacitySmall   int
	BookedBig       int
	BookedMedium    int
	BookedSmall     int
	SlotBlocked     bool
	DateBlocked     bool
	BookingsEnabled bool
}

func (s SlotAvailability) AvailableBig() int {
	if s.CapacityBig-s.BookedBig < 0 {
		return 0
	}
	return s.CapacityBig - s.BookedBig
}

func (s SlotAvailability) AvailableMedium() int {
	if s.CapacityMedium-s.BookedMedium < 0 {
		return 0
	}
	return s.CapacityMedium - s.BookedMedium
}

func (s SlotAvailability) AvailableSmall() int {
	if s.CapacitySmall-s.BookedSmall < 0 {
		return 0
	}
	return s.CapacitySmall - s.BookedSmall
}

func (s SlotAvailability) FullyUnavailable() bool {
	if !s.BookingsEnabled || s.DateBlocked || s.SlotBlocked {
		return true
	}
	return s.AvailableBig() == 0 && s.AvailableMedium() == 0 && s.AvailableSmall() == 0
}
