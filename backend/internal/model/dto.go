package model

import "time"

// ============================================================================
// POST /bookings
// ============================================================================

// CreateBookingRequest is the JSON body for POST /bookings.
//
// Note: contact.email is accepted but ignored — the JWT-derived email is used.
// Phone is REQUIRED per project decision (overrides the original spec wording).
type CreateBookingRequest struct {
	Date       string            `json:"date"       binding:"required"`
	Time       string            `json:"time"       binding:"required"`
	Quantities QuantitiesRequest `json:"quantities" binding:"required"`
	Contact    ContactRequest    `json:"contact"    binding:"required"`
}

// QuantitiesRequest mirrors `quantities` in the request body.
// We use *int here so we can distinguish "missing" from "zero" during validation.
type QuantitiesRequest struct {
	Big    *int `json:"big"`
	Medium *int `json:"medium"`
	Child  *int `json:"child"`
}

// ContactRequest mirrors `contact` in the request body.
type ContactRequest struct {
	FirstName string `json:"firstName" binding:"required,min=1,max=50"`
	LastName  string `json:"lastName"  binding:"required,min=1,max=50"`
	Email     string `json:"email"`                        // ignored; JWT email used
	Phone     string `json:"phone"     binding:"required"` // REQUIRED
}

// CreateBookingResponse is returned by POST /bookings on success.
type CreateBookingResponse struct {
	BookingID   string    `json:"bookingId"`
	TotalAmount float64   `json:"totalAmount"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// ============================================================================
// POST /checkout
// ============================================================================

type CreateCheckoutRequest struct {
	BookingID string `json:"bookingId"  binding:"required"`
	ResultURL string `json:"resultUrl"  binding:"required,url"`
}

type CreateCheckoutResponse struct {
	CheckoutURL string `json:"checkoutUrl"`
	SessionID   string `json:"sessionId"`
}

// ============================================================================
// GET /bookings/by-session/:sessionId
// ============================================================================

type BookingBySessionResponse struct {
	Status  string             `json:"status"`
	Booking *BookingPublicView `json:"booking,omitempty"`
}

// BookingPublicView is the user-facing shape of a booking.
type BookingPublicView struct {
	ID          string            `json:"id"`
	Date        string            `json:"date"`
	Time        string            `json:"time"`
	Quantities  Quantities        `json:"quantities"`
	Contact     ContactPublicView `json:"contact"`
	TotalAmount float64           `json:"totalAmount"`
	Status      string            `json:"status"`
}

type ContactPublicView struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// ============================================================================
// GET /availability/:month
// ============================================================================

type AvailabilityMonthResponse struct {
	Month string            `json:"month"`
	Days  []AvailabilityDay `json:"days"`
}

type AvailabilityDay struct {
	Date           string `json:"date"`
	AvailableSlots int    `json:"availableSlots"`
	Blocked        bool   `json:"blocked"`
	FullyBlocked   bool   `json:"fullyBlocked"`
}

// ============================================================================
// GET /slots/:date
// ============================================================================

type SlotsForDateResponse struct {
	Date            string        `json:"date"`
	DateBlocked     bool          `json:"dateBlocked"`
	BookingsEnabled bool          `json:"bookingsEnabled"`
	FullyBlocked    bool          `json:"fullyBlocked"`
	Slots           []SlotForDate `json:"slots"`
}

type SlotForDate struct {
	Time            string `json:"time"`
	AvailableBig    int    `json:"availableBig"`
	TotalBig        int    `json:"totalBig"`
	AvailableMedium int    `json:"availableMedium"`
	TotalMedium     int    `json:"totalMedium"`
	Blocked         bool   `json:"blocked"`
}

// ============================================================================
// Admin DTOs
// ============================================================================

// PATCH /admin/bookings/:bookingId/price
type AdminPriceOverrideRequest struct {
	TotalAmount *float64 `json:"totalAmount"` // nil clears the override
	Reason      string   `json:"reason"      binding:"required,min=5"`
}

type AdminPriceOverrideResponse struct {
	BookingID       string   `json:"bookingId"`
	TotalAmount     float64  `json:"totalAmount"`
	PriceOverride   *float64 `json:"priceOverride"`
	OverrideReason  *string  `json:"overrideReason"`
	EffectiveAmount float64  `json:"effectiveAmount"`
}

// PUT /admin/slots/:date/:time/block
type AdminBlockSlotRequest struct {
	Reason string `json:"reason"`
}

type AdminBlockSlotResponse struct {
	Date      string    `json:"date"`
	Time      string    `json:"time"`
	Blocked   bool      `json:"blocked"`
	Reason    *string   `json:"reason,omitempty"`
	BlockedAt time.Time `json:"blockedAt"`
}

// DELETE /admin/slots/:date/:time/block
type AdminUnblockSlotResponse struct {
	Date    string `json:"date"`
	Time    string `json:"time"`
	Blocked bool   `json:"blocked"`
}

// PUT /admin/dates/:date/block
type AdminBlockDateRequest struct {
	Reason string `json:"reason"`
}

type AdminBlockDateResponse struct {
	Date      string    `json:"date"`
	Blocked   bool      `json:"blocked"`
	Reason    *string   `json:"reason,omitempty"`
	BlockedAt time.Time `json:"blockedAt"`
}

// DELETE /admin/dates/:date/block
type AdminUnblockDateResponse struct {
	Date    string `json:"date"`
	Blocked bool   `json:"blocked"`
}

// PUT /admin/system/bookings-enabled
type AdminSetBookingsEnabledRequest struct {
	Enabled *bool  `json:"enabled" binding:"required"`
	Reason  string `json:"reason"`
}

type AdminSetBookingsEnabledResponse struct {
	BookingsEnabled bool      `json:"bookingsEnabled"`
	Reason          *string   `json:"reason,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// POST /admin/bookings/:bookingId/cancel
type AdminCancelBookingRequest struct {
	Reason string `json:"reason" binding:"required,min=1"`
}

type AdminCancelBookingResponse struct {
	BookingID   string    `json:"bookingId"`
	Status      string    `json:"status"`
	CancelledAt time.Time `json:"cancelledAt"`
	Reason      string    `json:"reason"`
}

// PUT /admin/slots/:date/:time
type AdminUpsertSlotRequest struct {
	CapacityBig    *int `json:"capacityBig"    binding:"required,gte=0"`
	CapacityMedium *int `json:"capacityMedium" binding:"required,gte=0"`
}

type AdminUpsertSlotResponse struct {
	Date           string `json:"date"`
	Time           string `json:"time"`
	CapacityBig    int    `json:"capacityBig"`
	CapacityMedium int    `json:"capacityMedium"`
	Blocked        bool   `json:"blocked"`
	Created        bool   `json:"created"` // true = new row, false = updated
}

// PATCH /admin/slots/:date/:time/capacity
type AdminUpdateSlotCapacityRequest struct {
	CapacityBig    *int `json:"capacityBig"    binding:"required,gte=0"`
	CapacityMedium *int `json:"capacityMedium" binding:"required,gte=0"`
}

type AdminUpdateSlotCapacityResponse struct {
	Date           string `json:"date"`
	Time           string `json:"time"`
	CapacityBig    int    `json:"capacityBig"`
	CapacityMedium int    `json:"capacityMedium"`
	Blocked        bool   `json:"blocked"`
}

// GET /admin/slots/:date/:time/bookings
type AdminSlotBookingsResponse struct {
	Date     string                  `json:"date"`
	Time     string                  `json:"time"`
	Bookings []AdminBookingListEntry `json:"bookings"`
}

type AdminBookingListEntry struct {
	ID                          string     `json:"id"`
	UserEmail                   string     `json:"userEmail"`
	FirstName                   string     `json:"firstName"`
	LastName                    string     `json:"lastName"`
	Phone                       *string    `json:"phone,omitempty"`
	Quantities                  Quantities `json:"quantities"`
	TotalAmount                 float64    `json:"totalAmount"`
	EffectiveAmount             float64    `json:"effectiveAmount"`
	Status                      string     `json:"status"`
	CreatedAt                   time.Time  `json:"createdAt"`
	PosterIncomingOrderID       *int64     `json:"posterIncomingOrderId,omitempty"`
	PosterIncomingTransactionID *int64     `json:"posterIncomingTransactionId,omitempty"`
}

type BookingStatusResponse struct {
	BookingsEnabled bool   `json:"bookingsEnabled"`
	Reason          string `json:"reason,omitempty"`
}
