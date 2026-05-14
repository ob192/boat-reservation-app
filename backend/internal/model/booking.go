package model

import (
	"time"

	"github.com/google/uuid"
)

// BookingStatus is the lifecycle state of a booking.
type BookingStatus string

const (
	StatusPending   BookingStatus = "pending"
	StatusConfirmed BookingStatus = "confirmed"
	StatusFailed    BookingStatus = "failed"
	StatusExpired   BookingStatus = "expired"
	StatusCancelled BookingStatus = "cancelled"
)

// Quantities mirrors the request body shape and is also used internally for pricing.
type Quantities struct {
	Big    int `json:"big"`
	Medium int `json:"medium"`
	Child  int `json:"child"`
}

// Booking is the persisted booking record.
//
// Notes:
//   - `user_id` and `user_email` always come from the JWT, never the request body.
//   - `total_amount` is server-computed; `price_override` (if set) wins for the gateway.
//   - `idempotency_key` is unique together with `user_id` to make POST /bookings safe to retry.
type Booking struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_bookings_idem_user,priority:1" json:"userId"`
	UserEmail string    `gorm:"type:varchar(255);not null" json:"userEmail"`

	Date      string `gorm:"type:date;not null;index:idx_bookings_date_time" json:"date"`       // YYYY-MM-DD
	Time      string `gorm:"type:varchar(5);not null;index:idx_bookings_date_time" json:"time"` // HH:MM
	QtyBig    int    `gorm:"not null;default:0" json:"qtyBig"`
	QtyMedium int    `gorm:"not null;default:0" json:"qtyMedium"`
	QtyChild  int    `gorm:"not null;default:0" json:"qtyChild"`

	FirstName string  `gorm:"type:varchar(50);not null" json:"firstName"`
	LastName  string  `gorm:"type:varchar(50);not null" json:"lastName"`
	Phone     *string `gorm:"type:varchar(50)" json:"phone,omitempty"`

	TotalAmount float64 `gorm:"type:decimal(10,2);not null" json:"totalAmount"`

	PriceOverride  *float64   `gorm:"type:decimal(10,2)" json:"priceOverride,omitempty"`
	OverrideReason *string    `gorm:"type:varchar(255)" json:"overrideReason,omitempty"`
	OverriddenBy   *uuid.UUID `gorm:"type:uuid" json:"overriddenBy,omitempty"`
	OverriddenAt   *time.Time `json:"overriddenAt,omitempty"`

	Status BookingStatus `gorm:"type:varchar(20);not null;index" json:"status"`

	CancelledBy  *uuid.UUID `gorm:"type:uuid" json:"cancelledBy,omitempty"`
	CancelledAt  *time.Time `json:"cancelledAt,omitempty"`
	CancelReason *string    `gorm:"type:varchar(255)" json:"cancelReason,omitempty"`

	PaymentSessionID *string `gorm:"type:varchar(255);uniqueIndex" json:"paymentSessionId,omitempty"`

	IdempotencyKey string `gorm:"type:varchar(255);not null;uniqueIndex:idx_bookings_idem_user,priority:2" json:"-"`

	ExpiresAt time.Time `gorm:"not null" json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TableName makes the table name explicit so a future rename of the struct doesn't surprise anyone.
func (Booking) TableName() string { return "bookings" }

// Quantities returns the request-shaped Quantities for this booking.
func (b *Booking) Quantities() Quantities {
	return Quantities{Big: b.QtyBig, Medium: b.QtyMedium, Child: b.QtyChild}
}

// EffectiveAmount returns price_override if set, otherwise total_amount.
func (b *Booking) EffectiveAmount() float64 {
	if b.PriceOverride != nil {
		return *b.PriceOverride
	}
	return b.TotalAmount
}
