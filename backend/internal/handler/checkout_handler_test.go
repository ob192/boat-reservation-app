package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func checkoutRouter(svc service.CheckoutService, user *model.AuthUser) *gin.Engine {
	h := NewCheckoutHandler(svc, testLogger())
	r := gin.New()
	g := r.Group("/")
	if user != nil {
		g.Use(withUser(user))
	}
	g.POST("/checkout", h.CreateCheckout)
	return r
}

func checkoutBody(bookingID string) string {
	return fmt.Sprintf(`{"bookingId": %q, "resultUrl": "https://app.example/result"}`, bookingID)
}

func TestCreateCheckoutHandler(t *testing.T) {
	bookingID := uuid.New()

	t.Run("unauthenticated", func(t *testing.T) {
		w := doRequest(t, checkoutRouter(&mockCheckoutService{}, nil), http.MethodPost, "/checkout", checkoutBody(bookingID.String()), nil)
		wantError(t, w, http.StatusUnauthorized, httpx.CodeNotAuthenticated)
	})

	t.Run("body validation", func(t *testing.T) {
		bodies := map[string]string{
			"malformed JSON":      `{"bookingId":`,
			"missing bookingId":   `{"resultUrl": "https://app.example/result"}`,
			"missing resultUrl":   fmt.Sprintf(`{"bookingId": %q}`, bookingID),
			"resultUrl not a url": fmt.Sprintf(`{"bookingId": %q, "resultUrl": "not-a-url"}`, bookingID),
		}
		for name, body := range bodies {
			t.Run(name, func(t *testing.T) {
				w := doRequest(t, checkoutRouter(&mockCheckoutService{}, testUser()), http.MethodPost, "/checkout", body, nil)
				wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
			})
		}
	})

	t.Run("non-uuid booking id", func(t *testing.T) {
		w := doRequest(t, checkoutRouter(&mockCheckoutService{}, testUser()), http.MethodPost, "/checkout", checkoutBody("not-a-uuid"), nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("service error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrBookingNotFound, http.StatusNotFound, httpx.CodeBookingNotFound},
			{service.ErrForbidden, http.StatusForbidden, httpx.CodeForbidden},
			{service.ErrBookingNotPending, http.StatusUnprocessableEntity, httpx.CodeBookingNotPending},
			{service.ErrBookingExpired, http.StatusGone, httpx.CodeBookingExpired},
			{errors.New("gateway down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			t.Run(tc.code, func(t *testing.T) {
				svc := &mockCheckoutService{
					createCheckout: func(context.Context, uuid.UUID, uuid.UUID, string) (*model.CreateCheckoutResponse, error) {
						return nil, tc.err
					},
				}
				w := doRequest(t, checkoutRouter(svc, testUser()), http.MethodPost, "/checkout", checkoutBody(bookingID.String()), nil)
				wantError(t, w, tc.status, tc.code)
			})
		}
	})

	t.Run("success passes identity and echoes the session", func(t *testing.T) {
		user := testUser()
		var gotBooking, gotUser uuid.UUID
		var gotURL string
		svc := &mockCheckoutService{
			createCheckout: func(_ context.Context, b, u uuid.UUID, url string) (*model.CreateCheckoutResponse, error) {
				gotBooking, gotUser, gotURL = b, u, url
				return &model.CreateCheckoutResponse{CheckoutURL: "https://pay.example/1", SessionID: "sess_1"}, nil
			},
		}
		w := doRequest(t, checkoutRouter(svc, user), http.MethodPost, "/checkout", checkoutBody(bookingID.String()), nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotBooking != bookingID || gotUser != user.ID || gotURL != "https://app.example/result" {
			t.Errorf("service args (%v, %v, %q)", gotBooking, gotUser, gotURL)
		}
		resp := decodeJSON[model.CreateCheckoutResponse](t, w)
		if resp.SessionID != "sess_1" || resp.CheckoutURL != "https://pay.example/1" {
			t.Errorf("resp = %+v", resp)
		}
	})
}
