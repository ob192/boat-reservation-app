package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func adminHandlerRouter(adminSvc service.AdminService, promoSvc service.PromocodeService, user *model.AuthUser) *gin.Engine {
	h := NewAdminHandler(adminSvc, promoSvc, testLogger())
	r := gin.New()
	g := r.Group("/admin")
	if user != nil {
		g.Use(withUser(user))
	}
	g.PATCH("/bookings/:bookingId/price", h.OverridePrice)
	g.POST("/bookings/:bookingId/cancel", h.CancelBooking)
	g.POST("/bookings/:bookingId/move", h.MoveBooking)
	g.GET("/bookings", h.ListBookings)
	g.PUT("/slots/:date/:time/:route/block", h.BlockSlot)
	g.DELETE("/slots/:date/:time/:route/block", h.UnblockSlot)
	g.PUT("/slots/:date/:time/:route/cancel", h.CancelSlot)
	g.DELETE("/slots/:date/:time/:route/cancel", h.UncancelSlot)
	g.PUT("/slots/:date/:time/:route", h.UpsertSlot)
	g.DELETE("/slots/:date/:time/:route", h.DeleteSlot)
	g.GET("/slots/:date/:time/:route/bookings", h.GetSlotBookings)
	g.PUT("/dates/:date/block", h.BlockDate)
	g.DELETE("/dates/:date/block", h.UnblockDate)
	g.PUT("/system/bookings-enabled", h.SetBookingsEnabled)
	g.POST("/promocodes", h.CreatePromocode)
	g.GET("/promocodes", h.ListPromocodes)
	return r
}

// slotParamCases exercises the shared date/time/route validation of slot endpoints.
func slotParamCases(t *testing.T, method, pathFmt string, body string) {
	t.Helper()
	r := adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser())

	w := doRequest(t, r, method, fmt.Sprintf(pathFmt, "bad-date", "07:00", "Desna"), body, nil)
	wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidDate)

	w = doRequest(t, r, method, fmt.Sprintf(pathFmt, "2026-08-01", "7am", "Desna"), body, nil)
	wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidTime)

	w = doRequest(t, r, method, fmt.Sprintf(pathFmt, "2026-08-01", "07:00", "Atlantis"), body, nil)
	wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidRoute)
}

// ----------------------------------------------------------------------------
// OverridePrice
// ----------------------------------------------------------------------------

func TestOverridePriceHandler(t *testing.T) {
	bookingID := uuid.New()
	path := "/admin/bookings/" + bookingID.String() + "/price"
	validBody := `{"totalAmount": 300, "reason": "vip guest"}`

	t.Run("no admin in context is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPatch, path, validBody, nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("non-uuid booking id", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPatch, "/admin/bookings/nope/price", validBody, nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("reason is required with min length", func(t *testing.T) {
		for _, body := range []string{`{"totalAmount": 300}`, `{"totalAmount": 300, "reason": "abc"}`} {
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPatch, path, body, nil)
			wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
		}
	})

	t.Run("service error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrBookingNotFound, http.StatusNotFound, httpx.CodeBookingNotFound},
			{service.ErrAlreadyConfirmed, http.StatusUnprocessableEntity, httpx.CodeAlreadyConfirmed},
			{service.ErrInvalidInput, http.StatusBadRequest, httpx.CodeInvalidInput},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			svc := &mockAdminService{
				overridePrice: func(context.Context, uuid.UUID, uuid.UUID, *float64, string) (*model.AdminPriceOverrideResponse, error) {
					return nil, tc.err
				},
			}
			w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPatch, path, validBody, nil)
			wantError(t, w, tc.status, tc.code)
		}
	})

	t.Run("success forwards amount, admin, and reason", func(t *testing.T) {
		admin := testUser()
		var gotBooking, gotAdmin uuid.UUID
		var gotAmount *float64
		var gotReason string
		svc := &mockAdminService{
			overridePrice: func(_ context.Context, b, a uuid.UUID, amount *float64, reason string) (*model.AdminPriceOverrideResponse, error) {
				gotBooking, gotAdmin, gotAmount, gotReason = b, a, amount, reason
				return &model.AdminPriceOverrideResponse{BookingID: b.String(), EffectiveAmount: 300}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, admin), http.MethodPatch, path, validBody, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotBooking != bookingID || gotAdmin != admin.ID {
			t.Errorf("ids (%v, %v)", gotBooking, gotAdmin)
		}
		if gotAmount == nil || *gotAmount != 300 || gotReason != "vip guest" {
			t.Errorf("args (%v, %q)", gotAmount, gotReason)
		}
	})

	t.Run("null amount clears the override", func(t *testing.T) {
		var gotAmount *float64
		svc := &mockAdminService{
			overridePrice: func(_ context.Context, b, _ uuid.UUID, amount *float64, _ string) (*model.AdminPriceOverrideResponse, error) {
				gotAmount = amount
				return &model.AdminPriceOverrideResponse{BookingID: b.String()}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPatch, path, `{"totalAmount": null, "reason": "clear it"}`, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotAmount != nil {
			t.Errorf("amount = %v, want nil passthrough", gotAmount)
		}
	})
}

// ----------------------------------------------------------------------------
// BlockSlot / UnblockSlot
// ----------------------------------------------------------------------------

func TestBlockSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna/block"

	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPut, path, "", nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodPut, "/admin/slots/%s/%s/%s/block", "")
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrSlotNotFound, http.StatusNotFound, httpx.CodeSlotNotFound},
			{service.ErrAlreadyBlocked, http.StatusConflict, httpx.CodeAlreadyBlocked},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			svc := &mockAdminService{
				blockSlot: func(context.Context, string, string, string, uuid.UUID, string) (*model.AdminBlockSlotResponse, error) {
					return nil, tc.err
				},
			}
			w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, "", nil)
			wantError(t, w, tc.status, tc.code)
		}
	})

	t.Run("success with optional reason body", func(t *testing.T) {
		var gotReason string
		svc := &mockAdminService{
			blockSlot: func(_ context.Context, date, tm, route string, _ uuid.UUID, reason string) (*model.AdminBlockSlotResponse, error) {
				gotReason = reason
				return &model.AdminBlockSlotResponse{Date: date, Time: tm, RouteName: route, Blocked: true, BlockedAt: time.Now()}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, `{"reason": "maintenance"}`, nil)
		if w.Code != http.StatusOK || gotReason != "maintenance" {
			t.Errorf("status=%d reason=%q", w.Code, gotReason)
		}

		// Body entirely optional.
		w = doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, "", nil)
		if w.Code != http.StatusOK {
			t.Errorf("bodyless block: status = %d", w.Code)
		}
	})
}

func TestUnblockSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna/block"

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodDelete, "/admin/slots/%s/%s/%s/block", "")
	})

	t.Run("slot not found", func(t *testing.T) {
		svc := &mockAdminService{
			unblockSlot: func(context.Context, string, string, string) (*model.AdminUnblockSlotResponse, error) {
				return nil, service.ErrSlotNotFound
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
		wantError(t, w, http.StatusNotFound, httpx.CodeSlotNotFound)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminService{
			unblockSlot: func(_ context.Context, date, tm, route string) (*model.AdminUnblockSlotResponse, error) {
				return &model.AdminUnblockSlotResponse{Date: date, Time: tm, RouteName: route}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
	})
}

// ----------------------------------------------------------------------------
// BlockDate / UnblockDate
// ----------------------------------------------------------------------------

func TestBlockDateHandler(t *testing.T) {
	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPut, "/admin/dates/2026-08-01/block", "", nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("invalid date", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPut, "/admin/dates/garbage/block", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidDate)
	})

	t.Run("already blocked is 409", func(t *testing.T) {
		svc := &mockAdminService{
			blockDate: func(context.Context, string, uuid.UUID, string) (*model.AdminBlockDateResponse, error) {
				return nil, service.ErrAlreadyBlocked
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, "/admin/dates/2026-08-01/block", "", nil)
		wantError(t, w, http.StatusConflict, httpx.CodeAlreadyBlocked)
	})

	t.Run("success", func(t *testing.T) {
		var gotDate, gotReason string
		svc := &mockAdminService{
			blockDate: func(_ context.Context, date string, _ uuid.UUID, reason string) (*model.AdminBlockDateResponse, error) {
				gotDate, gotReason = date, reason
				return &model.AdminBlockDateResponse{Date: date, Blocked: true, BlockedAt: time.Now()}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, "/admin/dates/2026-08-01/block", `{"reason":"storm"}`, nil)
		if w.Code != http.StatusOK || gotDate != "2026-08-01" || gotReason != "storm" {
			t.Errorf("status=%d date=%q reason=%q", w.Code, gotDate, gotReason)
		}
	})
}

func TestUnblockDateHandler(t *testing.T) {
	t.Run("invalid date", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodDelete, "/admin/dates/garbage/block", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidDate)
	})

	t.Run("missing block maps to plain NOT_FOUND", func(t *testing.T) {
		svc := &mockAdminService{
			unblockDate: func(context.Context, string) (*model.AdminUnblockDateResponse, error) {
				return nil, service.ErrBookingNotFound
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, "/admin/dates/2026-08-01/block", "", nil)
		wantError(t, w, http.StatusNotFound, httpx.CodeNotFound)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminService{
			unblockDate: func(_ context.Context, date string) (*model.AdminUnblockDateResponse, error) {
				return &model.AdminUnblockDateResponse{Date: date}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, "/admin/dates/2026-08-01/block", "", nil)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
	})
}

// ----------------------------------------------------------------------------
// SetBookingsEnabled
// ----------------------------------------------------------------------------

func TestSetBookingsEnabledHandler(t *testing.T) {
	path := "/admin/system/bookings-enabled"

	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPut, path, `{"enabled": false}`, nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("enabled flag is required — absent or null rejected", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"enabled": null}`, `{"reason": "x"}`, `not json`} {
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPut, path, body, nil)
			wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
		}
	})

	t.Run("false is a valid value, not a missing one", func(t *testing.T) {
		var gotEnabled bool
		var gotReason string
		svc := &mockAdminService{
			setBookingsEnabled: func(_ context.Context, enabled bool, reason string, _ uuid.UUID) (*model.AdminSetBookingsEnabledResponse, error) {
				gotEnabled, gotReason = enabled, reason
				return &model.AdminSetBookingsEnabledResponse{BookingsEnabled: enabled}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, `{"enabled": false, "reason": "storm"}`, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotEnabled || gotReason != "storm" {
			t.Errorf("args (%v, %q)", gotEnabled, gotReason)
		}
	})

	t.Run("service failure is 503", func(t *testing.T) {
		svc := &mockAdminService{
			setBookingsEnabled: func(context.Context, bool, string, uuid.UUID) (*model.AdminSetBookingsEnabledResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, `{"enabled": true}`, nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})
}

// ----------------------------------------------------------------------------
// CancelBooking / MoveBooking
// ----------------------------------------------------------------------------

func TestCancelBookingHandler(t *testing.T) {
	bookingID := uuid.New()
	path := "/admin/bookings/" + bookingID.String() + "/cancel"

	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPost, path, `{"reason":"x"}`, nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("non-uuid id", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPost, "/admin/bookings/nope/cancel", `{"reason":"x"}`, nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("reason required", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"reason": ""}`} {
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPost, path, body, nil)
			wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := &mockAdminService{
			cancelBooking: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.AdminCancelBookingResponse, error) {
				return nil, service.ErrBookingNotFound
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPost, path, `{"reason":"no-show"}`, nil)
		wantError(t, w, http.StatusNotFound, httpx.CodeBookingNotFound)
	})

	t.Run("success", func(t *testing.T) {
		admin := testUser()
		var gotID, gotAdmin uuid.UUID
		var gotReason string
		svc := &mockAdminService{
			cancelBooking: func(_ context.Context, id, a uuid.UUID, reason string) (*model.AdminCancelBookingResponse, error) {
				gotID, gotAdmin, gotReason = id, a, reason
				return &model.AdminCancelBookingResponse{BookingID: id.String(), Status: "cancelled", Reason: reason}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, admin), http.MethodPost, path, `{"reason":"no-show"}`, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotID != bookingID || gotAdmin != admin.ID || gotReason != "no-show" {
			t.Errorf("args (%v, %v, %q)", gotID, gotAdmin, gotReason)
		}
	})
}

func TestMoveBookingHandler(t *testing.T) {
	bookingID := uuid.New()
	path := "/admin/bookings/" + bookingID.String() + "/move"
	validBody := `{"date":"2026-08-02","time":"10:00","routeName":"Desna"}`

	t.Run("non-uuid id", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPost, "/admin/bookings/nope/move", validBody, nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("body validation", func(t *testing.T) {
		cases := map[string]struct {
			body string
			code string
		}{
			"missing fields": {`{}`, httpx.CodeInvalidInput},
			"bad date":       {`{"date":"garbage","time":"10:00","routeName":"Desna"}`, httpx.CodeInvalidDate},
			"bad time":       {`{"date":"2026-08-02","time":"10am","routeName":"Desna"}`, httpx.CodeInvalidTime},
			"unknown route":  {`{"date":"2026-08-02","time":"10:00","routeName":"Atlantis"}`, httpx.CodeInvalidRoute},
		}
		for name, tc := range cases {
			t.Run(name, func(t *testing.T) {
				w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPost, path, tc.body, nil)
				wantError(t, w, http.StatusBadRequest, tc.code)
			})
		}
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrBookingNotFound, http.StatusNotFound, httpx.CodeBookingNotFound},
			{service.ErrBookingNotPending, http.StatusUnprocessableEntity, httpx.CodeBookingNotPending},
			{service.ErrSlotNotFound, http.StatusNotFound, httpx.CodeSlotNotFound},
			{service.ErrSlotBlocked, http.StatusUnprocessableEntity, httpx.CodeSlotBlocked},
			{service.ErrSlotCancelled, http.StatusUnprocessableEntity, httpx.CodeSlotCancelled},
			{service.ErrSlotTaken, http.StatusConflict, httpx.CodeSlotTaken},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			t.Run(tc.code, func(t *testing.T) {
				svc := &mockAdminService{
					moveBooking: func(context.Context, uuid.UUID, string, string, string) (*model.AdminMoveBookingResponse, error) {
						return nil, tc.err
					},
				}
				w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPost, path, validBody, nil)
				wantError(t, w, tc.status, tc.code)
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		var gotID uuid.UUID
		var gotDate, gotTime, gotRoute string
		svc := &mockAdminService{
			moveBooking: func(_ context.Context, id uuid.UUID, date, tm, route string) (*model.AdminMoveBookingResponse, error) {
				gotID, gotDate, gotTime, gotRoute = id, date, tm, route
				return &model.AdminMoveBookingResponse{BookingID: id.String(), Date: date, Time: tm, RouteName: route}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPost, path, validBody, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotID != bookingID || gotDate != "2026-08-02" || gotTime != "10:00" || gotRoute != "Desna" {
			t.Errorf("args (%v, %q, %q, %q)", gotID, gotDate, gotTime, gotRoute)
		}
	})
}

// ----------------------------------------------------------------------------
// UpsertSlot / DeleteSlot / GetSlotBookings / CancelSlot / UncancelSlot
// ----------------------------------------------------------------------------

func TestUpsertSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna"
	validBody := `{"capacityBig": 3, "capacityMedium": 2, "capacitySmall": 1}`

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodPut, "/admin/slots/%s/%s/%s", validBody)
	})

	t.Run("capacities are required", func(t *testing.T) {
		for _, body := range []string{`{}`, `{"capacityBig": 1}`, `{"capacityBig": -1, "capacityMedium": 0, "capacitySmall": 0}`} {
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPut, path, body, nil)
			wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
		}
	})

	t.Run("success forwards capacities", func(t *testing.T) {
		var gotBig, gotMedium, gotSmall int
		svc := &mockAdminService{
			upsertSlot: func(_ context.Context, date, tm, route string, big, medium, small int, _ uuid.UUID) (*model.AdminUpsertSlotResponse, error) {
				gotBig, gotMedium, gotSmall = big, medium, small
				return &model.AdminUpsertSlotResponse{Date: date, Time: tm, RouteName: route, Created: true}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, validBody, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotBig != 3 || gotMedium != 2 || gotSmall != 1 {
			t.Errorf("capacities (%d, %d, %d)", gotBig, gotMedium, gotSmall)
		}
	})

	t.Run("zero capacities are valid", func(t *testing.T) {
		svc := &mockAdminService{
			upsertSlot: func(_ context.Context, date, tm, route string, _, _, _ int, _ uuid.UUID) (*model.AdminUpsertSlotResponse, error) {
				return &model.AdminUpsertSlotResponse{Date: date, Time: tm, RouteName: route}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, `{"capacityBig": 0, "capacityMedium": 0, "capacitySmall": 0}`, nil)
		if w.Code != http.StatusOK {
			t.Errorf("zero capacities must bind, status = %d (%s)", w.Code, w.Body.String())
		}
	})
}

func TestDeleteSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna"

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodDelete, "/admin/slots/%s/%s/%s", "")
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrSlotNotFound, http.StatusNotFound, httpx.CodeSlotNotFound},
			{service.ErrSlotNotEmpty, http.StatusConflict, httpx.CodeSlotNotEmpty},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			svc := &mockAdminService{
				deleteSlot: func(context.Context, string, string, string) error { return tc.err },
			}
			w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
			wantError(t, w, tc.status, tc.code)
		}
	})

	t.Run("success is 204 with no body", func(t *testing.T) {
		svc := &mockAdminService{
			deleteSlot: func(context.Context, string, string, string) error { return nil },
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
		if w.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", w.Code)
		}
		if w.Body.Len() != 0 {
			t.Errorf("body = %q, want empty", w.Body.String())
		}
	})
}

func TestGetSlotBookingsHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna/bookings"

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodGet, "/admin/slots/%s/%s/%s/bookings", "")
	})

	t.Run("slot not found", func(t *testing.T) {
		svc := &mockAdminService{
			getSlotBookings: func(context.Context, string, string, string) (*model.AdminSlotBookingsResponse, error) {
				return nil, service.ErrSlotNotFound
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodGet, path, "", nil)
		wantError(t, w, http.StatusNotFound, httpx.CodeSlotNotFound)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminService{
			getSlotBookings: func(_ context.Context, date, tm, route string) (*model.AdminSlotBookingsResponse, error) {
				return &model.AdminSlotBookingsResponse{
					Date: date, Time: tm, RouteName: route,
					Bookings: []model.AdminBookingListEntry{{ID: "b-1", Status: "pending"}},
				}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodGet, path, "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.AdminSlotBookingsResponse](t, w)
		if len(resp.Bookings) != 1 || resp.Bookings[0].ID != "b-1" {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestCancelSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna/cancel"

	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPut, path, "", nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodPut, "/admin/slots/%s/%s/%s/cancel", "")
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrSlotNotFound, http.StatusNotFound, httpx.CodeSlotNotFound},
			{service.ErrAlreadyCancelled, http.StatusConflict, httpx.CodeAlreadyCancelled},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			svc := &mockAdminService{
				cancelSlot: func(context.Context, string, string, string, uuid.UUID, string) (*model.AdminCancelSlotResponse, error) {
					return nil, tc.err
				},
			}
			w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, "", nil)
			wantError(t, w, tc.status, tc.code)
		}
	})

	t.Run("success reports cascaded cancellations", func(t *testing.T) {
		svc := &mockAdminService{
			cancelSlot: func(_ context.Context, date, tm, route string, _ uuid.UUID, reason string) (*model.AdminCancelSlotResponse, error) {
				return &model.AdminCancelSlotResponse{
					Date: date, Time: tm, RouteName: route, Cancelled: true,
					CancelledBookings: 4, CancelledAt: time.Now(),
				}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodPut, path, `{"reason":"storm"}`, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.AdminCancelSlotResponse](t, w)
		if !resp.Cancelled || resp.CancelledBookings != 4 {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestUncancelSlotHandler(t *testing.T) {
	path := "/admin/slots/2026-08-01/07:00/Desna/cancel"

	t.Run("param validation", func(t *testing.T) {
		slotParamCases(t, http.MethodDelete, "/admin/slots/%s/%s/%s/cancel", "")
	})

	t.Run("slot not found", func(t *testing.T) {
		svc := &mockAdminService{
			uncancelSlot: func(context.Context, string, string, string) (*model.AdminUncancelSlotResponse, error) {
				return nil, service.ErrSlotNotFound
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
		wantError(t, w, http.StatusNotFound, httpx.CodeSlotNotFound)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockAdminService{
			uncancelSlot: func(_ context.Context, date, tm, route string) (*model.AdminUncancelSlotResponse, error) {
				return &model.AdminUncancelSlotResponse{Date: date, Time: tm, RouteName: route}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodDelete, path, "", nil)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d", w.Code)
		}
	})
}

// ----------------------------------------------------------------------------
// ListBookings
// ----------------------------------------------------------------------------

func TestListBookingsHandler(t *testing.T) {
	t.Run("query validation", func(t *testing.T) {
		r := adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser())
		w := doRequest(t, r, http.MethodGet, "/admin/bookings?date=garbage", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidDate)

		w = doRequest(t, r, http.MethodGet, "/admin/bookings?status=paid", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("filters and paging forwarded, bad numbers fall back to defaults", func(t *testing.T) {
		var gotDate, gotStatus string
		var gotLimit, gotOffset int
		svc := &mockAdminService{
			listBookings: func(_ context.Context, date, status string, limit, offset int) (*model.AdminBookingHistoryResponse, error) {
				gotDate, gotStatus, gotLimit, gotOffset = date, status, limit, offset
				return &model.AdminBookingHistoryResponse{Bookings: []model.AdminBookingListEntry{}}, nil
			},
		}
		r := adminHandlerRouter(svc, &mockPromocodeService{}, testUser())

		w := doRequest(t, r, http.MethodGet, "/admin/bookings?date=2026-08-01&status=confirmed&limit=10&offset=5", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotDate != "2026-08-01" || gotStatus != "confirmed" || gotLimit != 10 || gotOffset != 5 {
			t.Errorf("args (%q, %q, %d, %d)", gotDate, gotStatus, gotLimit, gotOffset)
		}

		// Non-numeric and negative values quietly fall back to the zero default.
		doRequest(t, r, http.MethodGet, "/admin/bookings?limit=abc&offset=-4", "", nil)
		if gotLimit != 0 || gotOffset != 0 {
			t.Errorf("fallback args (%d, %d), want zeros", gotLimit, gotOffset)
		}
	})

	t.Run("service failure is 503", func(t *testing.T) {
		svc := &mockAdminService{
			listBookings: func(context.Context, string, string, int, int) (*model.AdminBookingHistoryResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, adminHandlerRouter(svc, &mockPromocodeService{}, testUser()), http.MethodGet, "/admin/bookings", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})
}

// ----------------------------------------------------------------------------
// Promocodes (admin)
// ----------------------------------------------------------------------------

func TestCreatePromocodeHandler(t *testing.T) {
	path := "/admin/promocodes"
	validBody := `{"code": "SUMMER", "discountPercent": 10, "maxUses": 100}`

	t.Run("no admin is 403", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, nil), http.MethodPost, path, validBody, nil)
		wantError(t, w, http.StatusForbidden, httpx.CodeForbidden)
	})

	t.Run("body validation", func(t *testing.T) {
		bad := []string{
			`{}`,
			`{"code": "X"}`, // missing pct + maxUses
			`{"code": "X", "discountPercent": 101, "maxUses": 1}`, // pct > 100
			`{"code": "X", "discountPercent": -1, "maxUses": 1}`,  // pct < 0
			`{"code": "X", "discountPercent": 10, "maxUses": 0}`,  // maxUses < 1
			`{"discountPercent": 10, "maxUses": 5}`,               // missing code
		}
		for _, body := range bad {
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodPost, path, body, nil)
			wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
		}
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrPromoAlreadyExists, http.StatusConflict, httpx.CodePromoAlreadyExists},
			{service.ErrInvalidInput, http.StatusBadRequest, httpx.CodeInvalidInput},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			promo := &mockPromocodeService{
				create: func(context.Context, string, int, int, uuid.UUID) (*model.AdminPromocodeResponse, error) {
					return nil, tc.err
				},
			}
			w := doRequest(t, adminHandlerRouter(&mockAdminService{}, promo, testUser()), http.MethodPost, path, validBody, nil)
			wantError(t, w, tc.status, tc.code)
		}
	})

	t.Run("success is 201 with args forwarded", func(t *testing.T) {
		admin := testUser()
		var gotCode string
		var gotPct, gotMax int
		var gotAdmin uuid.UUID
		promo := &mockPromocodeService{
			create: func(_ context.Context, code string, pct, max int, adminID uuid.UUID) (*model.AdminPromocodeResponse, error) {
				gotCode, gotPct, gotMax, gotAdmin = code, pct, max, adminID
				return &model.AdminPromocodeResponse{Code: "SUMMER", DiscountPercent: pct, MaxUses: max, Active: true}, nil
			},
		}
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, promo, admin), http.MethodPost, path, validBody, nil)
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotCode != "SUMMER" || gotPct != 10 || gotMax != 100 || gotAdmin != admin.ID {
			t.Errorf("args (%q, %d, %d, %v)", gotCode, gotPct, gotMax, gotAdmin)
		}
		resp := decodeJSON[model.AdminPromocodeResponse](t, w)
		if resp.Code != "SUMMER" || !resp.Active {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestListPromocodesHandler(t *testing.T) {
	t.Run("createdBy must be a uuid", func(t *testing.T) {
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, &mockPromocodeService{}, testUser()), http.MethodGet, "/admin/promocodes?createdBy=nope", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("filter forwarded when present, nil when absent", func(t *testing.T) {
		adminID := uuid.New()
		var gotFilter *uuid.UUID
		promo := &mockPromocodeService{
			list: func(_ context.Context, createdBy *uuid.UUID) (*model.AdminPromocodeListResponse, error) {
				gotFilter = createdBy
				return &model.AdminPromocodeListResponse{Promocodes: []model.AdminPromocodeResponse{}}, nil
			},
		}
		r := adminHandlerRouter(&mockAdminService{}, promo, testUser())

		w := doRequest(t, r, http.MethodGet, "/admin/promocodes?createdBy="+adminID.String(), "", nil)
		if w.Code != http.StatusOK || gotFilter == nil || *gotFilter != adminID {
			t.Errorf("status=%d filter=%v", w.Code, gotFilter)
		}

		doRequest(t, r, http.MethodGet, "/admin/promocodes", "", nil)
		if gotFilter != nil {
			t.Errorf("filter = %v, want nil without the query param", gotFilter)
		}
	})

	t.Run("service failure is 503", func(t *testing.T) {
		promo := &mockPromocodeService{
			list: func(context.Context, *uuid.UUID) (*model.AdminPromocodeListResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, adminHandlerRouter(&mockAdminService{}, promo, testUser()), http.MethodGet, "/admin/promocodes", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})
}

// ----------------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------------

func TestParseIntDefault(t *testing.T) {
	tests := []struct {
		in   string
		def  int
		want int
	}{
		{"", 7, 7},
		{"42", 0, 42},
		{"0", 7, 0},
		{"-3", 7, 7},  // negatives fall back
		{"abc", 7, 7}, // garbage falls back
		{"1.5", 7, 7}, // non-integer falls back
	}
	for _, tt := range tests {
		if got := parseIntDefault(tt.in, tt.def); got != tt.want {
			t.Errorf("parseIntDefault(%q, %d) = %d, want %d", tt.in, tt.def, got, tt.want)
		}
	}
}

func TestIsValidDate(t *testing.T) {
	for _, ok := range []string{"2026-08-01", "2000-01-01", "2026-02-28"} {
		if !isValidDate(ok) {
			t.Errorf("isValidDate(%q) = false", ok)
		}
	}
	for _, bad := range []string{"", "garbage", "2026-8-1", "01-08-2026", "2026-02-30", "2026-13-01"} {
		if isValidDate(bad) {
			t.Errorf("isValidDate(%q) = true", bad)
		}
	}
}
