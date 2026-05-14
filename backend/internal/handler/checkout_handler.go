package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type CheckoutHandler struct {
	checkout service.CheckoutService
	log      *slog.Logger
}

func NewCheckoutHandler(checkout service.CheckoutService, log *slog.Logger) *CheckoutHandler {
	return &CheckoutHandler{checkout: checkout, log: log}
}

// CreateCheckout handles POST /checkout.
//
// @Summary      Create checkout session
// @Description  Generates a payment gateway checkout URL for a pending booking.
// @Tags         checkout
// @Accept       json
// @Produce      json
// @Param        request body model.CreateCheckoutRequest true "Checkout Request"
// @Success      200  {object}  model.CreateCheckoutResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      401  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      410  {object}  httpx.ErrorBody "BOOKING_EXPIRED"
// @Failure      422  {object}  httpx.ErrorBody "BOOKING_NOT_PENDING"
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /checkout [post]
func (h *CheckoutHandler) CreateCheckout(c *gin.Context) {
	user, err := GetUserFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusUnauthorized, httpx.CodeNotAuthenticated, "")
		return
	}

	var req model.CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
		return
	}
	bookingID, err := uuid.Parse(req.BookingID)
	if err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "bookingId must be a UUID")
		return
	}

	resp, err := h.checkout.CreateCheckout(c.Request.Context(), bookingID, user.ID, req.SuccessURL, req.CancelURL)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeBookingNotFound, "")
		case errors.Is(err, service.ErrForbidden):
			httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		case errors.Is(err, service.ErrBookingNotPending):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeBookingNotPending, "")
		case errors.Is(err, service.ErrBookingExpired):
			httpx.Err(c, http.StatusGone, httpx.CodeBookingExpired, "")
		default:
			h.log.Error("checkout", "err", err, "booking_id", bookingID)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}
