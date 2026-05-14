package model

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"

	"github.com/google/uuid"
)

// Slot defines physical fleet capacity and block state for a (date, time) pair.
// Primary key is composite: (date, time).
type Slot struct {
	Date pgtype.Date `gorm:"type:date;primaryKey" json:"date"`       // YYYY-MM-DD
	Time string      `gorm:"type:varchar(5);primaryKey" json:"time"` // HH:MM, one of 08:00/11:00/15:00/19:00

	CapacityBig    int `gorm:"not null;default:0" json:"capacityBig"`
	CapacityMedium int `gorm:"not null;default:0" json:"capacityMedium"`

	Blocked     bool       `gorm:"not null;default:false" json:"blocked"`
	BlockedBy   *uuid.UUID `gorm:"type:uuid" json:"blockedBy,omitempty"`
	BlockedAt   *time.Time `json:"blockedAt,omitempty"`
	BlockReason *string    `gorm:"type:varchar(255)" json:"blockReason,omitempty"`
}

func (Slot) TableName() string { return "slots" }

func (s Slot) DateString() string {
	return s.Date.Time.UTC().Format("2006-01-02")
}

// DateBlock marks a whole date as admin-blocked. Stored separately from `slots`
// so that dates without pre-seeded slot rows can still be blocked.
type DateBlock struct {
	Date        string    `gorm:"type:date;primaryKey" json:"date"`
	BlockedBy   uuid.UUID `gorm:"type:uuid;not null" json:"blockedBy"`
	BlockedAt   time.Time `gorm:"not null" json:"blockedAt"`
	BlockReason *string   `gorm:"type:varchar(255)" json:"blockReason,omitempty"`
}

func (DateBlock) TableName() string { return "date_blocks" }

// SlotAvailability is the projection used by the availability endpoints.
// Built in-memory from a slot, the booking sums for that slot, and block flags.
type SlotAvailability struct {
	Date            string
	Time            string
	CapacityBig     int
	CapacityMedium  int
	BookedBig       int
	BookedMedium    int
	SlotBlocked     bool
	DateBlocked     bool
	BookingsEnabled bool
}

// AvailableBig returns remaining big capacity (never negative).
func (s SlotAvailability) AvailableBig() int {
	if s.CapacityBig-s.BookedBig < 0 {
		return 0
	}
	return s.CapacityBig - s.BookedBig
}

// AvailableMedium returns remaining medium capacity (never negative).
func (s SlotAvailability) AvailableMedium() int {
	if s.CapacityMedium-s.BookedMedium < 0 {
		return 0
	}
	return s.CapacityMedium - s.BookedMedium
}

// FullyUnavailable reports whether users should be unable to book this slot.
// Useful for the calendar view; not used at the booking-creation step
// (which performs the same checks individually with specific error codes).
func (s SlotAvailability) FullyUnavailable() bool {
	if !s.BookingsEnabled || s.DateBlocked || s.SlotBlocked {
		return true
	}
	return s.AvailableBig() == 0 && s.AvailableMedium() == 0
}
