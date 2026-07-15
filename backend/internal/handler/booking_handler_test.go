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
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

var handlerNow = time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)

func testBooking(userID uuid.UUID) *model.Booking {
	d, _ := time.Parse("2006-01-02", "2026-08-01")
	return &model.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		UserEmail:   "user@example.com",
		Date:        pgtype.Date{Time: d, Valid: true},
		Time:        "07:00",
		RouteName:   "Desna",
		QtyBig:      1,
		FirstName:   "Olena",
		LastName:    "Kovalenko",
		Phone:       sptr("+380501112233"),
		TotalAmount: 450,
		Status:      model.StatusPending,
		ExpiresAt:   handlerNow.Add(5 * time.Minute),
		CreatedAt:   handlerNow,
	}
}

func bookingRouter(svc service.BookingService, user *model.AuthUser) *gin.Engine {
	h := NewBookingHandler(svc, testLogger())
	r := gin.New()
	g := r.Group("/")
	if user != nil {
		g.Use(withUser(user))
	}
	g.POST("/bookings", h.CreateBooking)
	g.GET("/bookings", h.ListMyBookings)
	g.GET("/bookings/by-session/:sessionId", h.GetBySession)
	return r
}

const validCreateBody = `{
	"date": "2026-08-01",
	"time": "07:00",
	"routeName": "Desna",
	"quantities": {"big": 1, "child": 1},
	"contact": {"firstName": "Olena", "lastName": "Kovalenko", "phone": "+380501112233"},
	"promoCode": "SUMMER"
}`

func TestCreateBookingHandler(t *testing.T) {
	idemHeaders := map[string]string{"X-Idempotency-Key": "idem-1"}

	t.Run("unauthenticated", func(t *testing.T) {
		r := bookingRouter(&mockBookingService{}, nil)
		w := doRequest(t, r, http.MethodPost, "/bookings", validCreateBody, idemHeaders)
		wantError(t, w, http.StatusUnauthorized, httpx.CodeNotAuthenticated)
	})

	t.Run("missing idempotency key", func(t *testing.T) {
		r := bookingRouter(&mockBookingService{}, testUser())
		w := doRequest(t, r, http.MethodPost, "/bookings", validCreateBody, nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		r := bookingRouter(&mockBookingService{}, testUser())
		w := doRequest(t, r, http.MethodPost, "/bookings", `{"date": `, idemHeaders)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})

	t.Run("missing required binding fields", func(t *testing.T) {
		bodies := map[string]string{
			"no date":    `{"time":"07:00","routeName":"Desna","quantities":{"big":1},"contact":{"firstName":"A","lastName":"B","phone":"+3"}}`,
			"no contact": `{"date":"2026-08-01","time":"07:00","routeName":"Desna","quantities":{"big":1}}`,
			"no phone":   `{"date":"2026-08-01","time":"07:00","routeName":"Desna","quantities":{"big":1},"contact":{"firstName":"A","lastName":"B"}}`,
		}
		for name, body := range bodies {
			t.Run(name, func(t *testing.T) {
				r := bookingRouter(&mockBookingService{}, testUser())
				w := doRequest(t, r, http.MethodPost, "/bookings", body, idemHeaders)
				wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
			})
		}
	})

	t.Run("input assembled from JWT, header, and body", func(t *testing.T) {
		user := testUser()
		var gotIn service.CreateBookingInput
		svc := &mockBookingService{
			create: func(_ context.Context, in service.CreateBookingInput) (*model.Booking, error) {
				gotIn = in
				return testBooking(user.ID), nil
			},
		}
		r := bookingRouter(svc, user)
		w := doRequest(t, r, http.MethodPost, "/bookings", validCreateBody, idemHeaders)
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if gotIn.UserID != user.ID || gotIn.UserEmail != user.Email {
			t.Error("identity must come from the JWT context")
		}
		if gotIn.IdempotencyKey != "idem-1" {
			t.Errorf("idempotency key = %q", gotIn.IdempotencyKey)
		}
		if gotIn.Quantities != (model.Quantities{Big: 1, Child: 1}) {
			t.Errorf("quantities = %+v", gotIn.Quantities)
		}
		if gotIn.PromoCode != "SUMMER" || gotIn.Phone != "+380501112233" {
			t.Errorf("body fields lost: %+v", gotIn)
		}
	})

	t.Run("missing quantities default to zero", func(t *testing.T) {
		user := testUser()
		var gotIn service.CreateBookingInput
		svc := &mockBookingService{
			create: func(_ context.Context, in service.CreateBookingInput) (*model.Booking, error) {
				gotIn = in
				return testBooking(user.ID), nil
			},
		}
		body := `{"date":"2026-08-01","time":"07:00","routeName":"Desna","quantities":{"medium":2},"contact":{"firstName":"A","lastName":"B","phone":"+3"}}`
		doRequest(t, bookingRouter(svc, user), http.MethodPost, "/bookings", body, idemHeaders)
		if gotIn.Quantities != (model.Quantities{Medium: 2}) {
			t.Errorf("quantities = %+v, want only medium set", gotIn.Quantities)
		}
	})

	t.Run("201 body echoes the created booking", func(t *testing.T) {
		user := testUser()
		b := testBooking(user.ID)
		b.TotalAmount = 607.5
		b.PromoCode = sptr("SUMMER")
		b.DiscountPercent = iptr(10)
		b.DiscountAmount = fptr(67.5)
		svc := &mockBookingService{
			create: func(context.Context, service.CreateBookingInput) (*model.Booking, error) { return b, nil },
		}
		w := doRequest(t, bookingRouter(svc, user), http.MethodPost, "/bookings", validCreateBody, idemHeaders)
		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.CreateBookingResponse](t, w)
		if resp.BookingID != b.ID.String() || resp.TotalAmount != 607.5 {
			t.Errorf("resp = %+v", resp)
		}
		if resp.PromoCode == nil || *resp.PromoCode != "SUMMER" ||
			resp.DiscountPercent == nil || *resp.DiscountPercent != 10 ||
			resp.DiscountAmount == nil || *resp.DiscountAmount != 67.5 {
			t.Errorf("promo echo wrong: %+v", resp)
		}
		if !resp.ExpiresAt.Equal(b.ExpiresAt) {
			t.Errorf("expiresAt = %v", resp.ExpiresAt)
		}
	})

	t.Run("service errors map to the spec's HTTP codes", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrBookingsDisabled, http.StatusServiceUnavailable, httpx.CodeBookingsDisabled},
			{service.ErrDateBlocked, http.StatusUnprocessableEntity, httpx.CodeDateBlocked},
			{service.ErrSlotBlocked, http.StatusUnprocessableEntity, httpx.CodeSlotBlocked},
			{service.ErrSlotCancelled, http.StatusUnprocessableEntity, httpx.CodeSlotCancelled},
			{service.ErrSlotTaken, http.StatusConflict, httpx.CodeSlotTaken},
			{service.ErrSlotNotFound, http.StatusNotFound, httpx.CodeSlotNotFound},
			{service.ErrInvalidInput, http.StatusBadRequest, httpx.CodeInvalidInput},
			{service.ErrValidationFailed, http.StatusUnprocessableEntity, httpx.CodeValidationFailed},
			{service.ErrInvalidRoute, http.StatusBadRequest, httpx.CodeInvalidRoute},
			{service.ErrPromoNotFound, http.StatusUnprocessableEntity, httpx.CodePromoNotFound},
			{service.ErrPromoInactive, http.StatusUnprocessableEntity, httpx.CodePromoInactive},
			{service.ErrPromoExhausted, http.StatusUnprocessableEntity, httpx.CodePromoExhausted},
			{errors.New("connection reset"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			t.Run(tc.code, func(t *testing.T) {
				svc := &mockBookingService{
					create: func(context.Context, service.CreateBookingInput) (*model.Booking, error) {
						return nil, tc.err
					},
				}
				w := doRequest(t, bookingRouter(svc, testUser()), http.MethodPost, "/bookings", validCreateBody, idemHeaders)
				wantError(t, w, tc.status, tc.code)
			})
		}
	})

	t.Run("wrapped service errors still map", func(t *testing.T) {
		svc := &mockBookingService{
			create: func(context.Context, service.CreateBookingInput) (*model.Booking, error) {
				return nil, fmt.Errorf("%w: date is in the past", service.ErrInvalidInput)
			},
		}
		w := doRequest(t, bookingRouter(svc, testUser()), http.MethodPost, "/bookings", validCreateBody, idemHeaders)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidInput)
	})
}

func TestGetBySessionHandler(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		r := bookingRouter(&mockBookingService{}, nil)
		w := doRequest(t, r, http.MethodGet, "/bookings/by-session/sess_1", "", nil)
		wantError(t, w, http.StatusUnauthorized, httpx.CodeNotAuthenticated)
	})

	t.Run("error mapping", func(t *testing.T) {
		cases := []struct {
			err    error
			status int
			code   string
		}{
			{service.ErrBookingNotFound, http.StatusNotFound, httpx.CodeNotFound},
			{service.ErrForbidden, http.StatusForbidden, httpx.CodeForbidden},
			{errors.New("db down"), http.StatusServiceUnavailable, httpx.CodeServiceUnavailable},
		}
		for _, tc := range cases {
			t.Run(tc.code, func(t *testing.T) {
				svc := &mockBookingService{
					getBySession: func(context.Context, string, uuid.UUID) (*model.Booking, error) {
						return nil, tc.err
					},
				}
				w := doRequest(t, bookingRouter(svc, testUser()), http.MethodGet, "/bookings/by-session/sess_1", "", nil)
				wantError(t, w, tc.status, tc.code)
			})
		}
	})

	t.Run("pending booking exposes only the status", func(t *testing.T) {
		user := testUser()
		svc := &mockBookingService{
			getBySession: func(_ context.Context, sessionID string, userID uuid.UUID) (*model.Booking, error) {
				if sessionID != "sess_1" || userID != user.ID {
					t.Errorf("args (%q, %v)", sessionID, userID)
				}
				return testBooking(user.ID), nil
			},
		}
		w := doRequest(t, bookingRouter(svc, user), http.MethodGet, "/bookings/by-session/sess_1", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.BookingBySessionResponse](t, w)
		if resp.Status != "pending" || resp.Booking != nil {
			t.Errorf("pending view must hide details: %+v", resp)
		}
	})

	t.Run("confirmed booking exposes the public view with effective amount", func(t *testing.T) {
		user := testUser()
		b := testBooking(user.ID)
		b.Status = model.StatusConfirmed
		b.PriceOverride = fptr(300)
		svc := &mockBookingService{
			getBySession: func(context.Context, string, uuid.UUID) (*model.Booking, error) { return b, nil },
		}
		w := doRequest(t, bookingRouter(svc, user), http.MethodGet, "/bookings/by-session/sess_1", "", nil)
		resp := decodeJSON[model.BookingBySessionResponse](t, w)
		if resp.Status != "confirmed" || resp.Booking == nil {
			t.Fatalf("resp = %+v", resp)
		}
		v := resp.Booking
		if v.ID != b.ID.String() || v.Date != "2026-08-01" || v.Time != "07:00" || v.RouteName != "Desna" {
			t.Errorf("view = %+v", v)
		}
		if v.TotalAmount != 300 {
			t.Errorf("view total = %v, want the override 300", v.TotalAmount)
		}
		if v.Contact.FirstName != "Olena" || v.Contact.Email != b.UserEmail {
			t.Errorf("contact = %+v", v.Contact)
		}
	})
}

func TestListMyBookingsHandler(t *testing.T) {
	t.Run("unauthenticated", func(t *testing.T) {
		r := bookingRouter(&mockBookingService{}, nil)
		w := doRequest(t, r, http.MethodGet, "/bookings", "", nil)
		wantError(t, w, http.StatusUnauthorized, httpx.CodeNotAuthenticated)
	})

	t.Run("service failure is a 503", func(t *testing.T) {
		svc := &mockBookingService{
			getAllForUser: func(context.Context, uuid.UUID) ([]model.Booking, error) {
				return nil, errors.New("db down")
			},
		}
		w := doRequest(t, bookingRouter(svc, testUser()), http.MethodGet, "/bookings", "", nil)
		wantError(t, w, http.StatusServiceUnavailable, httpx.CodeServiceUnavailable)
	})

	t.Run("maps bookings with effective amounts", func(t *testing.T) {
		user := testUser()
		b1 := testBooking(user.ID)
		b2 := testBooking(user.ID)
		b2.PriceOverride = fptr(100)
		b2.Status = model.StatusConfirmed
		svc := &mockBookingService{
			getAllForUser: func(_ context.Context, userID uuid.UUID) ([]model.Booking, error) {
				if userID != user.ID {
					t.Errorf("queried user %v", userID)
				}
				return []model.Booking{*b1, *b2}, nil
			},
		}
		w := doRequest(t, bookingRouter(svc, user), http.MethodGet, "/bookings", "", nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		resp := decodeJSON[model.MyBookingsResponse](t, w)
		if len(resp.Bookings) != 2 {
			t.Fatalf("bookings = %+v", resp.Bookings)
		}
		if resp.Bookings[0].TotalAmount != 450 || resp.Bookings[1].TotalAmount != 100 {
			t.Errorf("amounts = %v / %v, want 450 / 100", resp.Bookings[0].TotalAmount, resp.Bookings[1].TotalAmount)
		}
		if resp.Bookings[1].Status != "confirmed" {
			t.Errorf("status = %s", resp.Bookings[1].Status)
		}
	})

	t.Run("empty list stays an empty array", func(t *testing.T) {
		svc := &mockBookingService{
			getAllForUser: func(context.Context, uuid.UUID) ([]model.Booking, error) { return nil, nil },
		}
		w := doRequest(t, bookingRouter(svc, testUser()), http.MethodGet, "/bookings", "", nil)
		resp := decodeJSON[model.MyBookingsResponse](t, w)
		if resp.Bookings == nil || len(resp.Bookings) != 0 {
			t.Errorf("bookings = %#v, want empty non-nil", resp.Bookings)
		}
	})
}
