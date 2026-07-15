package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func availabilityRouter(svc service.AvailabilityService) *gin.Engine {
	h := NewAvailabilityHandler(svc, testLogger())
	r := gin.New()
	r.GET("/availability/:month/:route", h.GetAvailability)
	r.GET("/slots/:date/:route", h.GetSlots)
	r.GET("/status", h.GetStatus)
	return r
}

func TestGetAvailabilityHandler(t *testing.T) {
	t.Run("unknown route rejected before the service runs", func(t *testing.T) {
		w := doRequest(t, availabilityRouter(&mockAvailabilityService{}), http.MethodGet, "/availability/2026-08/Atlantis", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidRoute)
	})

	t.Run("invalid month surfaces as INVALID_MONTH", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getMonth: func(_ context.Context, month, _ string) (*model.AvailabilityMonthResponse, error) {
				return nil, fmt.Errorf("invalid month %q: parse fail", month)
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/availability/2026-13/Desna", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidMonth)
	})

	t.Run("other failures are 503", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getMonth: func(context.Context, string, string) (*model.AvailabilityMonthResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/availability/2026-08/Desna", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})

	t.Run("success passes params and returns the month", func(t *testing.T) {
		var gotMonth, gotRoute string
		svc := &mockAvailabilityService{
			getMonth: func(_ context.Context, month, route string) (*model.AvailabilityMonthResponse, error) {
				gotMonth, gotRoute = month, route
				return &model.AvailabilityMonthResponse{
					Month: month,
					Days:  []model.AvailabilityDay{{Date: "2026-08-01", AvailableSlots: 3}},
				}, nil
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/availability/2026-08/Desna", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if gotMonth != "2026-08" || gotRoute != "Desna" {
			t.Errorf("service args (%q, %q)", gotMonth, gotRoute)
		}
		resp := decodeJSON[model.AvailabilityMonthResponse](t, w)
		if resp.Month != "2026-08" || len(resp.Days) != 1 || resp.Days[0].AvailableSlots != 3 {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestGetSlotsHandler(t *testing.T) {
	t.Run("unknown route rejected", func(t *testing.T) {
		w := doRequest(t, availabilityRouter(&mockAvailabilityService{}), http.MethodGet, "/slots/2026-08-01/Atlantis", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidRoute)
	})

	t.Run("invalid date surfaces as INVALID_DATE", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getDate: func(_ context.Context, date, _ string) (*model.SlotsForDateResponse, error) {
				return nil, fmt.Errorf("invalid date %q: parse fail", date)
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/slots/garbage/Desna", "", nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidDate)
	})

	t.Run("other failures are 503", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getDate: func(context.Context, string, string) (*model.SlotsForDateResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/slots/2026-08-01/Desna", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})

	t.Run("success returns the day view", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getDate: func(_ context.Context, date, route string) (*model.SlotsForDateResponse, error) {
				return &model.SlotsForDateResponse{
					Date:            date,
					BookingsEnabled: true,
					Slots: []model.SlotForDate{
						{Time: "07:00", RouteName: route, AvailableBig: 1, TotalBig: 2},
					},
				}, nil
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/slots/2026-08-01/Desna", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.SlotsForDateResponse](t, w)
		if resp.Date != "2026-08-01" || len(resp.Slots) != 1 || resp.Slots[0].AvailableBig != 1 {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestGetStatusHandler(t *testing.T) {
	t.Run("failure is 503", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getStatus: func(context.Context) (*model.BookingStatusResponse, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/status", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})

	t.Run("success", func(t *testing.T) {
		svc := &mockAvailabilityService{
			getStatus: func(context.Context) (*model.BookingStatusResponse, error) {
				return &model.BookingStatusResponse{BookingsEnabled: false, Reason: "storm"}, nil
			},
		}
		w := doRequest(t, availabilityRouter(svc), http.MethodGet, "/status", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.BookingStatusResponse](t, w)
		if resp.BookingsEnabled || resp.Reason != "storm" {
			t.Errorf("resp = %+v", resp)
		}
	})
}

func TestIsInvalidInputHelper(t *testing.T) {
	if !isInvalidInput(errors.New(`invalid month "x": boom`)) {
		t.Error("errors prefixed with 'invalid' must be recognised")
	}
	if isInvalidInput(errors.New("db down")) {
		t.Error("other errors must not be treated as validation failures")
	}
	if isInvalidInput(nil) {
		t.Error("nil is not invalid input")
	}
}
