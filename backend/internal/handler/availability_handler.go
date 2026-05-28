package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	_ "github.com/harbour-wave/harbour-wave-backend/internal/model"

	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type AvailabilityHandler struct {
	svc service.AvailabilityService
	log *slog.Logger
}

func NewAvailabilityHandler(svc service.AvailabilityService, log *slog.Logger) *AvailabilityHandler {
	return &AvailabilityHandler{svc: svc, log: log}
}

// GetAvailability handles GET /availability/:month
//
// @Summary      Get monthly availability
// @Description  Returns daily availability for a given month
// @Tags         availability
// @Produce      json
// @Param        month path string true "Month in YYYY-MM format"
// @Success      200  {object}  model.AvailabilityMonthResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /availability/{month} [get]
func (h *AvailabilityHandler) GetAvailability(c *gin.Context) {
	month := c.Param("month")
	resp, err := h.svc.GetMonth(c.Request.Context(), month)
	if err != nil {
		// ParseMonth-style invalid input → 400 INVALID_MONTH; everything else is a 503.
		if isInvalidInput(err) {
			httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidMonth, err.Error())
			return
		}
		h.log.Error("availability/month", "err", err, "month", month)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// GetSlots handles GET /slots/:date
//
// @Summary      Get slots for a date
// @Description  Returns all available and blocked slots for a specific date
// @Tags         availability
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Success      200  {object}  model.SlotsForDateResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /slots/{date} [get]
func (h *AvailabilityHandler) GetSlots(c *gin.Context) {
	date := c.Param("date")
	resp, err := h.svc.GetDate(c.Request.Context(), date)
	if err != nil {
		if isInvalidInput(err) {
			httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, err.Error())
			return
		}
		h.log.Error("slots/date", "err", err, "date", date)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// GetStatus handles GET /status
//
// @Summary      Get booking status
// @Description  Returns whether bookings are globally available or fully blocked
// @Tags         availability
// @Produce      json
// @Success      200  {object}  model.BookingStatusResponse
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /status [get]
func (h *AvailabilityHandler) GetStatus(c *gin.Context) {
	resp, err := h.svc.GetStatus(c.Request.Context())
	if err != nil {
		h.log.Error("status", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// isInvalidInput recognises the validation errors thrown by service.ParseMonth / ParseDate.
// They surface as wrapped fmt.Errorf strings whose message starts with "invalid".
// Brittle by nature, but the surface area is tiny and well-known.
func isInvalidInput(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "invalid")
}
