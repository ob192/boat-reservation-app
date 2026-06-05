package httpx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard error codes returned in the `message` field of error responses.
// These are spec-defined and must remain stable for client compatibility.
const (
	CodeNotAuthenticated   = "NOT_AUTHENTICATED"
	CodeSessionExpired     = "SESSION_EXPIRED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeBookingNotFound    = "BOOKING_NOT_FOUND"
	CodeSlotNotFound       = "SLOT_NOT_FOUND"
	CodeSlotTaken          = "SLOT_TAKEN"
	CodeAlreadyBlocked     = "ALREADY_BLOCKED"
	CodeAlreadyCancelled   = "ALREADY_CANCELLED"
	CodeBookingExpired     = "BOOKING_EXPIRED"
	CodeInvalidInput       = "INVALID_INPUT"
	CodeInvalidJSON        = "INVALID_JSON"
	CodeInvalidDate        = "INVALID_DATE"
	CodeInvalidTime        = "INVALID_TIME"
	CodeInvalidMonth       = "INVALID_MONTH"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeSlotCancelled      = "SLOT_CANCELLED"
	CodeSlotBlocked        = "SLOT_BLOCKED"
	CodeDateBlocked        = "DATE_BLOCKED"
	CodeBookingNotPending  = "BOOKING_NOT_PENDING"
	CodeAlreadyConfirmed   = "ALREADY_CONFIRMED"
	CodeInvalidSignature   = "INVALID_SIGNATURE"
	CodeBackendError       = "BACKEND_ERROR"
	CodeBookingsDisabled   = "BOOKINGS_DISABLED"
	CodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	CodeCannotReadBody     = "CANNOT_READ_BODY"
	CodeWebhookFailed      = "WEBHOOK_PROCESSING_FAILED"
	CodeInvalidRoute       = "INVALID_ROUTE"
)

// Issue describes a single field-level validation problem.
type Issue struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Message string  `json:"message"`
	Detail  string  `json:"detail,omitempty"`
	Issues  []Issue `json:"issues,omitempty"`
}

// OK writes a 200 JSON response.
func OK(c *gin.Context, payload any) {
	c.JSON(http.StatusOK, payload)
}

// Created writes a 201 JSON response.
func Created(c *gin.Context, payload any) {
	c.JSON(http.StatusCreated, payload)
}

func Err(c *gin.Context, status int, code string, detail string) {
	c.AbortWithStatusJSON(status, ErrorBody{ // <-- Capital E here
		Message: code,
		Detail:  detail,
	})
}

func ErrValidation(c *gin.Context, issues []Issue) {
	c.AbortWithStatusJSON(http.StatusBadRequest, ErrorBody{ // <-- Capital E here
		Message: CodeInvalidInput,
		Issues:  issues,
	})
}
