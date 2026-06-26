package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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

// BlockSlot PUT /admin/slots/:date/:time/:route/block
//
// @Summary      Block a specific slot
// @Description  Admin endpoint to block bookings for a specific date, time, and route.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Param        request body model.AdminBlockSlotRequest false "Block reason"
// @Success      200  {object}  model.AdminBlockSlotResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE, INVALID_TIME, or INVALID_ROUTE"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route}/block [put]
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}

	var req model.AdminBlockSlotRequest
	_ = c.ShouldBindJSON(&req)

	resp, err := h.adminSvc.BlockSlot(c.Request.Context(), date, timeParam, route, adminID, req.Reason)
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

// UnblockSlot DELETE /admin/slots/:date/:time/:route/block
//
// @Summary      Unblock a specific slot
// @Description  Admin endpoint to remove a block on a specific date, time, and route.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Success      200  {object}  model.AdminUnblockSlotResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE, INVALID_TIME, or INVALID_ROUTE"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route}/block [delete]
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}
	resp, err := h.adminSvc.UnblockSlot(c.Request.Context(), date, timeParam, route)
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

// UpsertSlot PUT /admin/slots/:date/:time/:route
//
// @Summary      Create or update a slot
// @Description  Admin endpoint to create a slot for a given date/time/route or update its capacities.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Param        request body model.AdminUpsertSlotRequest true "Capacity payload"
// @Success      200  {object}  model.AdminUpsertSlotResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE, INVALID_TIME, INVALID_ROUTE, or INVALID_INPUT"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route} [put]
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
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
		date, timeParam, route,
		*req.CapacityBig, *req.CapacityMedium, *req.CapacitySmall,
		adminID,
	)
	if err != nil {
		h.log.Error("admin upsert slot", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// GetSlotBookings GET /admin/slots/:date/:time/:route/bookings
//
// @Summary      List bookings for a slot
// @Description  Admin endpoint returning all bookings (any status) for a specific date/time/route slot.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Success      200  {object}  model.AdminSlotBookingsResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE, INVALID_TIME, or INVALID_ROUTE"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route}/bookings [get]
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}

	resp, err := h.adminSvc.GetSlotBookings(c.Request.Context(), date, timeParam, route)
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

// ListBookings GET /admin/bookings
//
// @Summary      List booking history
// @Description  Admin endpoint returning bookings of any status, newest first (created_at DESC). Optional filters.
// @Tags         admin
// @Produce      json
// @Param        date   query string false "Filter by slot date (YYYY-MM-DD)"
// @Param        status query string false "Filter by status: pending, confirmed, failed, expired, cancelled"
// @Param        limit  query int    false "Max rows (default 50, max 200)"
// @Param        offset query int    false "Rows to skip (default 0)"
// @Success      200  {object}  model.AdminBookingHistoryResponse
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE or INVALID_INPUT"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/bookings [get]
func (h *AdminHandler) ListBookings(c *gin.Context) {
	date := c.Query("date")
	if date != "" && !isValidDate(date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}

	status := c.Query("status")
	if status != "" && !model.IsValidBookingStatus(status) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "unknown status")
		return
	}

	limit := parseIntDefault(c.Query("limit"), 0) // 0 → service default
	offset := parseIntDefault(c.Query("offset"), 0)

	resp, err := h.adminSvc.ListBookings(c.Request.Context(), date, status, limit, offset)
	if err != nil {
		h.log.Error("admin list bookings", "err", err)
		httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		return
	}
	httpx.OK(c, resp)
}

// CancelSlot PUT /admin/slots/:date/:time/:route/cancel
//
// @Summary      Cancel a specific slot
// @Description  Admin endpoint to cancel a slot and cascade-cancel all its bookings.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Param        request body model.AdminCancelSlotRequest false "Cancel reason"
// @Success      200  {object}  model.AdminCancelSlotResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route}/cancel [put]
func (h *AdminHandler) CancelSlot(c *gin.Context) {
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}

	var req model.AdminCancelSlotRequest
	_ = c.ShouldBindJSON(&req)

	resp, err := h.adminSvc.CancelSlot(c.Request.Context(), date, timeParam, route, adminID, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		case errors.Is(err, service.ErrAlreadyCancelled):
			httpx.Err(c, http.StatusConflict, httpx.CodeAlreadyCancelled, "")
		default:
			h.log.Error("admin cancel slot", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// UncancelSlot DELETE /admin/slots/:date/:time/:route/cancel
//
// @Summary      Un-cancel a specific slot
// @Description  Re-opens a cancelled slot for booking. Does not reinstate already-cancelled bookings.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Success      200  {object}  model.AdminUncancelSlotResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route}/cancel [delete]
func (h *AdminHandler) UncancelSlot(c *gin.Context) {
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}
	resp, err := h.adminSvc.UncancelSlot(c.Request.Context(), date, timeParam, route)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		default:
			h.log.Error("admin uncancel slot", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// MoveBooking POST /admin/bookings/:bookingId/move
//
// @Summary      Move a booking to another slot
// @Description  Admin endpoint to relocate a booking to a different date/time/route.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        bookingId path string true "Booking ID (UUID)"
// @Param        request body model.AdminMoveBookingRequest true "Destination slot"
// @Success      200  {object}  model.AdminMoveBookingResponse
// @Failure      400  {object}  httpx.ErrorBody
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody
// @Failure      422  {object}  httpx.ErrorBody
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/bookings/{bookingId}/move [post]
func (h *AdminHandler) MoveBooking(c *gin.Context) {
	bookingID, err := uuid.Parse(c.Param("bookingId"))
	if err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, "bookingId must be a UUID")
		return
	}
	var req model.AdminMoveBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidInput, err.Error())
		return
	}
	if !isValidDate(req.Date) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidDate, "")
		return
	}
	if !service.IsValidSlotTime(req.Time) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidTime, "")
		return
	}
	if !service.IsValidRoute(req.RouteName) {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}

	resp, err := h.adminSvc.MoveBooking(c.Request.Context(), bookingID, req.Date, req.Time, req.RouteName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBookingNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeBookingNotFound, "")
		case errors.Is(err, service.ErrBookingNotPending):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeBookingNotPending, "")
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		case errors.Is(err, service.ErrSlotBlocked):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeSlotBlocked, "")
		case errors.Is(err, service.ErrSlotCancelled):
			httpx.Err(c, http.StatusUnprocessableEntity, httpx.CodeSlotCancelled, "")
		case errors.Is(err, service.ErrSlotTaken):
			httpx.Err(c, http.StatusConflict, httpx.CodeSlotTaken, "")
		default:
			h.log.Error("admin move booking", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	httpx.OK(c, resp)
}

// DeleteSlot DELETE /admin/slots/:date/:time/:route
//
// @Summary      Delete a slot
// @Description  Admin endpoint to delete a slot, only if it has no active bookings.
// @Tags         admin
// @Produce      json
// @Param        date path string true "Date in YYYY-MM-DD format"
// @Param        time path string true "Time in HH:MM format"
// @Param        route path string true "Route name"
// @Success      204
// @Failure      400  {object}  httpx.ErrorBody "INVALID_DATE, INVALID_TIME, or INVALID_ROUTE"
// @Failure      403  {object}  httpx.ErrorBody
// @Failure      404  {object}  httpx.ErrorBody
// @Failure      409  {object}  httpx.ErrorBody "SLOT_NOT_EMPTY"
// @Failure      503  {object}  httpx.ErrorBody
// @Security     BearerAuth
// @Router       /admin/slots/{date}/{time}/{route} [delete]
func (h *AdminHandler) DeleteSlot(c *gin.Context) {
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
	route, ok := parseRoute(c)
	if !ok {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidRoute, "")
		return
	}

	err := h.adminSvc.DeleteSlot(c.Request.Context(), date, timeParam, route)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlotNotFound):
			httpx.Err(c, http.StatusNotFound, httpx.CodeSlotNotFound, "")
		case errors.Is(err, service.ErrSlotNotEmpty):
			httpx.Err(c, http.StatusConflict, httpx.CodeSlotNotEmpty, "")
		default:
			h.log.Error("admin delete slot", "err", err)
			httpx.Err(c, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable, "")
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return def
	}
	return n
}

// isValidDate accepts YYYY-MM-DD strictly.
func isValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

func parseRoute(c *gin.Context) (string, bool) {
	route := c.Param("route")
	if !service.IsValidRoute(route) {
		return "", false
	}
	return route, true
}
