package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type BookingHandler struct {
	bookingSvc service.BookingService
	log        *slog.Logger
}

func NewBookingHandler(bookingSvc service.BookingService, log *slog.Logger) *BookingHandler {
	return &BookingHandler{bookingSvc: bookingSvc, log: log}
}

// CreateBooking handles POST /bookings.
//
// @Summary      Create a new booking
// @Description  Creates a pending booking and holds slot capacity. Requires idempotency key.
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        X-Idempotency-Key header string true "Idempotency Key"
// @Param        request body model.CreateBookingRequest true "Booking payload"
// @Success      201  {object}  model.CreateBookingResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_INPUT"
// @Failure      401  {object}  httpx.ErrorBody "NOT_AUTHENTICATED"
// @Failure      409  {object}  httpx.ErrorBody "SLOT_TAKEN"
// @Failure      422  {object}  httpx.ErrorBody "VALIDATION_FAILED, DATE_BLOCKED, or SLOT_BLOCKED"
// @Failure      503  {object}  httpx.ErrorBody "BOOKINGS_DISABLED or SERVICE_UNAVAILABLE"
// @Security     BearerAuth
// @Router       /bookings [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusUnauthorized, httpx.CodeNotAuthenticated, "")
		return
	}

	idemKey := c.GetHeader("X-Idempotency-Key")
	if idemKey == "" {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "missing X-Idempotency-Key header")
		return
	}

	var req model.CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Distinguish malformed JSON from missing required fields.
		// Validator surfaces both with errors but it's hard to tell apart safely;
		// we expose INVALID_INPUT for both since the body shape is broken either way.
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
		return
	}

	// Quantities use *int — coerce to plain ints, defaulting to 0 for missing.
	q := model.Quantities{
		Big:    deref(req.Quantities.Big),
		Medium: deref(req.Quantities.Medium),
		Child:  deref(req.Quantities.Child),
	}

	in := service.CreateBookingInput{
		UserID:         user.ID,
		UserEmail:      user.Email,
		IdempotencyKey: idemKey,
		Date:           req.Date,
		Time:           req.Time,
		Quantities:     q,
		FirstName:      req.Contact.FirstName,
		LastName:       req.Contact.LastName,
		Phone:          req.Contact.Phone,
	}

	b, err := h.bookingSvc.Create(c.Request.Context(), in)
	if err != nil {
		mapBookingError(c, err, h.log)
		return
	}

	httpx.Created(c, model.CreateBookingResponse{
		BookingID:   b.ID.String(),
		TotalAmount: b.TotalAmount,
		ExpiresAt:   b.ExpiresAt,
	})
}

// GetBySession handles GET /bookings/by-session/:sessionId.
//
// @Summary      Get booking by session
// @Description  Returns minimal status info while pending, or the full booking once confirmed.
// @Tags         bookings
// @Produce      json
// @Param        sessionId path string true "Payment Session ID"
// @Success      200  {object}  model.BookingBySessionResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      401  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /bookings/by-session/{sessionId} [get]
func (h *BookingHandler) GetBySession(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusUnauthorized, httpx.CodeNotAuthenticated, "")
		return
	}
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "missing session id")
		return
	}

	b, err := h.bookingSvc.GetBySession(c.Request.Context(), sessionID, user.ID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeNotFound, "")
		case errors.Is(err, service.ErrForbidden):
			httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		default:
			h.log.Error("get by session", "err", err, "session_id", sessionID)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}

	resp := model.BookingBySessionResponse{Status: string(b.Status)}
	if b.Status == model.StatusConfirmed {
		resp.Booking = &model.BookingPublicView{
			ID:         b.ID.String(),
			Date:       b.DateFormatted(),
			Time:       b.Time,
			Quantities: b.Quantities(),
			Contact: model.ContactPublicView{
				FirstName: b.FirstName,
				LastName:  b.LastName,
				Email:     b.UserEmail,
			},
			TotalAmount: b.EffectiveAmount(),
			Status:      string(b.Status),
		}
	}
	httpx.OK(c, resp)
}

// mapBookingError translates a service-layer error into the spec's HTTP shape.
func mapBookingError(c *gin.Context, err error, log *slog.Logger) {
	switch {
	case errors.Is(err, service.ErrBookingsDisabled):
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeBookingsDisabled, "")
	case errors.Is(err, service.ErrDateBlocked):
		httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeDateBlocked, "")
	case errors.Is(err, service.ErrSlotBlocked):
		httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeSlotBlocked, "")
	case errors.Is(err, service.ErrSlotTaken):
		httpx.Err(c, http.StatusConflict, httpx.CodeSlotTaken, "")
	case errors.Is(err, service.ErrSlotNotFound):
		httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
	case errors.Is(err, service.ErrInvalidInput):
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
	case errors.Is(err, service.ErrValidationFailed):
		httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeValidationFailed, err.Error())
	case errors.Is(err, service.ErrBookingNotPending):
		httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeBookingNotPending, "")
	case errors.Is(err, service.ErrBookingExpired):
		httpx.Err(c, http.StatusGone, httpx.CodeBookingExpired, "")
	case errors.Is(err, service.ErrAlreadyConfirmed):
		httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeAlreadyConfirmed, "")
	case errors.Is(err, service.ErrForbidden):
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
	case errors.Is(err, service.ErrBookingNotFound):
		httpx.Err(c, http.StatusNotFound, httpx.CodeBookingNotFound, "")
	default:
		log.Error("booking error", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
	}
}

func deref(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
