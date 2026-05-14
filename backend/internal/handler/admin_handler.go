package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type AdminHandler struct {
	adminSvc service.AdminService
	log      *slog.Logger
}

func NewAdminHandler(svc service.AdminService, log *slog.Logger) *AdminHandler {
	return &AdminHandler{adminSvc: svc, log: log}
}

// OverridePrice PATCH /admin/bookings/:bookingId/price
//
// @Summary      Override booking price
// @Description  Admin endpoint to manually override the price of a booking.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        bookingId path string true "Booking ID (UUID)"
// @Param        request body model.AdminPriceOverrideRequest true "Override details"
// @Success      200  {object}  model.AdminPriceOverrideResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      422  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/bookings/{bookingId}/price [patch]
func (h *AdminHandler) OverridePrice(c *gin.Context) {

	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}
	bookingID, err := uuid.Parse(c.Param("bookingId"))
	if err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "bookingId must be a UUID")
		return
	}

	var req model.AdminPriceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
		return
	}

	resp, err := h.adminSvc.OverridePrice(c.Request.Context(), bookingID, adminID, req.TotalAmount, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeBookingNotFound, "")
		case errors.Is(err, service.ErrAlreadyConfirmed):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeAlreadyConfirmed, "")
		case errors.Is(err, service.ErrInvalidInput):
			httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "")
		default:
			h.log.Error("admin override price", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// BlockSlot PUT /admin/slots/:date/:time/block
//
// @Summary      Block a specific slot
// @Description  Admin endpoint to block bookings for a specific date and time.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        request body model.AdminBlockSlotRequest false "Block reason"
// @Success      200  {object}  model.AdminBlockSlotResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/block [put]
func (h *AdminHandler) BlockSlot(c *gin.Context) {
	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}
	date := c.Param("date")
	timeParam := c.Param("time")
	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	if !service.IsValidSlotTime(timeParam) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidTime, "")
		return
	}

	var req model.AdminBlockSlotRequest
	// Body is optional — treat any binding error as empty.
	_ = c.ShouldBindJSON(&req)

	resp, err := h.adminSvc.BlockSlot(c.Request.Context(), date, timeParam, adminID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		case errors.Is(err, service.ErrAlreadyBlocked):
			httpx.Err(c, http.StatusConflict, httpx.CodeAlreadyBlocked, "")
		default:
			h.log.Error("admin block slot", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// UnblockSlot DELETE /admin/slots/:date/:time/block
//
// @Summary      Unblock a specific slot
// @Description  Admin endpoint to remove a block on a specific date and time.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Success      200  {object}  model.AdminUnblockSlotResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/block [delete]
func (h *AdminHandler) UnblockSlot(c *gin.Context) {
	date := c.Param("date")
	timeParam := c.Param("time")
	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	if !service.IsValidSlotTime(timeParam) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidTime, "")
		return
	}
	resp, err := h.adminSvc.UnblockSlot(c.Request.Context(), date, timeParam)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		default:
			h.log.Error("admin unblock slot", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// BlockDate PUT /admin/dates/:date/block
//
// @Summary      Block an entire date
// @Description  Admin endpoint to block bookings for an entire date.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        request body model.AdminBlockDateRequest false "Block reason"
// @Success      200  {object}  model.AdminBlockDateResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/dates/{date}/block [put]
func (h *AdminHandler) BlockDate(c *gin.Context) {
	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}
	date := c.Param("date")
	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	var req model.AdminBlockDateRequest
	_ = c.ShouldBindJSON(&req)

	resp, err := h.adminSvc.BlockDate(c.Request.Context(), date, adminID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAlreadyBlocked):
			httpx.Err(c, http.StatusConflict, httpx.CodeAlreadyBlocked, "")
		default:
			h.log.Error("admin block date", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// UnblockDate DELETE /admin/dates/:date/block
//
// @Summary      Unblock an entire date
// @Description  Admin endpoint to remove a block on an entire date.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Success      200  {object}  model.AdminUnblockDateResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/dates/{date}/block [delete]
func (h *AdminHandler) UnblockDate(c *gin.Context) {
	date := c.Param("date")
	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	resp, err := h.adminSvc.UnblockDate(c.Request.Context(), date)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			// UnblockDate maps "not found" → 404 NOT_FOUND per spec.
			httpx.Err(c, http.StatusNotFound, httpx.CodeNotFound, "")
		default:
			h.log.Error("admin unblock date", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// SetBookingsEnabled PUT /admin/system/bookings-enabled
//
// @Summary      Toggle global booking kill-switch
// @Description  Admin endpoint to enable or disable the creation of all new bookings.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        request body model.AdminSetBookingsEnabledRequest true "Kill-switch payload"
// @Success      200  {object}  model.AdminSetBookingsEnabledResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/system/bookings-enabled [put]
func (h *AdminHandler) SetBookingsEnabled(c *gin.Context) {
	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}
	var req model.AdminSetBookingsEnabledRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Enabled == nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "enabled is required")
		return
	}
	resp, err := h.adminSvc.SetBookingsEnabled(c.Request.Context(), *req.Enabled, req.Reason, adminID)
	if err != nil {
		h.log.Error("admin set bookings enabled", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// CancelBooking POST /admin/bookings/:bookingId/cancel
//
// @Summary      Cancel a booking
// @Description  Admin endpoint to force cancel a booking.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        bookingId path string true "Booking ID (UUID)"
// @Param        request body model.AdminCancelBookingRequest true "Cancellation reason"
// @Success      200  {object}  model.AdminCancelBookingResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/bookings/{bookingId}/cancel [post]
func (h *AdminHandler) CancelBooking(c *gin.Context) {
	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}
	bookingID, err := uuid.Parse(c.Param("bookingId"))
	if err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "bookingId must be a UUID")
		return
	}
	var req model.AdminCancelBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "reason is required")
		return
	}
	resp, err := h.adminSvc.CancelBooking(c.Request.Context(), bookingID, adminID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeBookingNotFound, "")
		default:
			h.log.Error("admin cancel booking", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// UpsertSlot PUT /admin/slots/:date/:time
//
// @Summary      Create or update a slot
// @Description  Admin endpoint to create a slot for a given date/time or update its capacity.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        request body model.AdminUpsertSlotRequest true "Capacity payload"
// @Success      200  {object}  model.AdminUpsertSlotResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time} [put]
func (h *AdminHandler) UpsertSlot(c *gin.Context) {
	date := c.Param("date")
	timeParam := c.Param("time")

	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	if !service.IsValidSlotTime(timeParam) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidTime, "")
		return
	}

	var req model.AdminUpsertSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
		return
	}

	adminID, err := GetUserIDFromContext(c)
	if err != nil {
		httpx.Err(c, http.StatusForbidden, httpx.CodeForbidden, "")
		return
	}

	resp, err := h.adminSvc.UpsertSlot(
		c.Request.Context(),
		date, timeParam,
		*req.CapacityBig, *req.CapacityMedium,
		adminID,
	)
	if err != nil {
		h.log.Error("admin upsert slot", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// GetSlotBookings GET /admin/slots/:date/:time/bookings
//
// @Summary      List bookings for a slot
// @Description  Admin endpoint returning all bookings (any status) for a specific date/time slot.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Success      200  {object}  model.AdminSlotBookingsResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/bookings [get]
func (h *AdminHandler) GetSlotBookings(c *gin.Context) {
	date := c.Param("date")
	timeParam := c.Param("time")

	if !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	if !service.IsValidSlotTime(timeParam) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidTime, "")
		return
	}

	resp, err := h.adminSvc.GetSlotBookings(c.Request.Context(), date, timeParam)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		default:
			h.log.Error("admin get slot bookings", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// isValidDate accepts YYYY-MM-DD strictly.
func isValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}
