package handler

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func promocodeRouter(svc service.PromocodeService) *gin.Engine {
	h := NewPromocodeHandler(svc, testLogger())
	r := gin.New()
	r.GET("/promocodes/:code", h.GetPromoCode)
	return r
}

func TestGetPromoCodeHandler(t *testing.T) {
	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrPromoNotFound, http.StatusUnprocessableEntity, httpx.CodePromoNotFound},
			{service.ErrPromoInactive, http.StatusUnprocessableEntity, httpx.CodePromoInactive},
			{service.ErrPromoExhausted, http.StatusUnprocessableEntity, httpx.CodePromoExhausted},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			t.Run(tc.code, func(t *testing.T) {
				svc := &mockPromocodeService{
					validate: func(context.Context, string) (*model.Promocode, error) { return nil, tc.err },
				}
				w := doRequest(t, promocodeRouter(svc), http.MethodGet, "/promocodes/NOPE", "", nil)
				wantError(t, w, tc.status, tc.code)
			})
		}
	})

	t.Run("success exposes ONLY the discount percent", func(t *testing.T) {
		var asked string
		svc := &mockPromocodeService{
			validate: func(_ context.Context, code string) (*model.Promocode, error) {
				asked = code
				return &model.Promocode{
					Code: "SUMMER", DiscountPercent: 15, MaxUses: 100, TimesUsed: 42, Active: true,
				}, nil
			},
		}
		w := doRequest(t, promocodeRouter(svc), http.MethodGet, "/promocodes/summer", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if asked != "summer" {
			t.Errorf("raw path param must reach the service, got %q", asked)
		}
		resp := decodeJSON[map[string]any](t, w)
		if resp["discountPercent"] != float64(15) {
			t.Errorf("discountPercent = %v", resp["discountPercent"])
		}
		// Usage counts and ownership are admin-only — must never leak here.
		for _, forbidden := range []string{"maxUses", "timesUsed", "createdBy", "code", "active"} {
			if _, ok := resp[forbidden]; ok {
				t.Errorf("field %q leaked into the public preview: %v", forbidden, resp)
			}
		}
	})
}
