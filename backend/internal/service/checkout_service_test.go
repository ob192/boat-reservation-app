package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// checkoutFixture wires a CheckoutService around a single stored booking.
type checkoutFixture struct {
	booking  *model.Booking
	bookings *mockBookingRepo
	gateway  *mockGateway
}

func newCheckoutFixture() *checkoutFixture {
	b := mkBooking()
	f := &checkoutFixture{booking: b}
	f.bookings = &mockBookingRepo{
		findByID: func(_ context.Context, id uuid.UUID) (*model.Booking, error) {
			if id == b.ID {
				return b, nil
			}
			return nil, repository.ErrNotFound
		},
		setPaymentSessionID: func(context.Context, uuid.UUID, string) error { return nil },
	}
	f.gateway = &mockGateway{
		createSession: func(context.Context, provider.CreateSessionRequest) (provider.CreateSessionResponse, error) {
			return provider.CreateSessionResponse{SessionID: "sess_1", CheckoutURL: "https://pay.example/1"}, nil
		},
	}
	return f
}

func (f *checkoutFixture) service(t *testing.T) CheckoutService {
	bookingSvc := NewBookingService(
		newTestDB(t), f.bookings, &mockSlotRepo{}, &mockDateBlockRepo{}, &mockSystemRepo{},
		NewPricingService(), &mockPromocodeService{}, fakeClock{t: testNow},
	)
	return NewCheckoutService(
		bookingSvc, f.bookings, NewPricingService(), f.gateway,
		&config.Config{}, fakeClock{t: testNow},
	)
}

func TestCreateCheckoutGates(t *testing.T) {
	t.Run("unknown booking", func(t *testing.T) {
		f := newCheckoutFixture()
		_, err := f.service(t).CreateCheckout(context.Background(), uuid.New(), f.booking.UserID, "https://app/result")
		if !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("stranger forbidden", func(t *testing.T) {
		f := newCheckoutFixture()
		_, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, uuid.New(), "https://app/result")
		if !errors.Is(err, ErrForbidden) {
			t.Errorf("want ErrForbidden, got %v", err)
		}
	})

	t.Run("non-pending statuses rejected", func(t *testing.T) {
		for _, st := range []model.BookingStatus{
			model.StatusConfirmed, model.StatusFailed, model.StatusExpired, model.StatusCancelled,
		} {
			f := newCheckoutFixture()
			f.booking.Status = st
			_, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result")
			if !errors.Is(err, ErrBookingNotPending) {
				t.Errorf("status %s: want ErrBookingNotPending, got %v", st, err)
			}
		}
	})

	t.Run("expired hold rejected, boundary inclusive", func(t *testing.T) {
		f := newCheckoutFixture()
		f.booking.ExpiresAt = testNow // exactly now → no longer After(now) → expired
		_, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result")
		if !errors.Is(err, ErrBookingExpired) {
			t.Errorf("want ErrBookingExpired at the boundary, got %v", err)
		}
	})
}

func TestCreateCheckoutSuccess(t *testing.T) {
	f := newCheckoutFixture()
	f.booking.QtyBig = 1
	f.booking.QtyChild = 1
	f.booking.TotalAmount = 675

	var gotReq provider.CreateSessionRequest
	f.gateway.createSession = func(_ context.Context, req provider.CreateSessionRequest) (provider.CreateSessionResponse, error) {
		gotReq = req
		return provider.CreateSessionResponse{SessionID: "sess_1", CheckoutURL: "https://pay.example/1"}, nil
	}
	var savedID uuid.UUID
	var savedSession string
	f.bookings.setPaymentSessionID = func(_ context.Context, id uuid.UUID, s string) error {
		savedID, savedSession = id, s
		return nil
	}

	resp, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.SessionID != "sess_1" || resp.CheckoutURL != "https://pay.example/1" {
		t.Errorf("response = %+v", resp)
	}
	if savedID != f.booking.ID || savedSession != "sess_1" {
		t.Errorf("session id persisted as (%v, %q)", savedID, savedSession)
	}

	if gotReq.BookingID != f.booking.ID.String() || gotReq.UserEmail != f.booking.UserEmail {
		t.Errorf("gateway identity fields wrong: %+v", gotReq)
	}
	if gotReq.AmountUAH != 675 {
		t.Errorf("amount = %v, want the booking total 675", gotReq.AmountUAH)
	}
	if gotReq.ResultURL != "https://app/result" {
		t.Errorf("result url = %q", gotReq.ResultURL)
	}
	if !gotReq.ExpiresAt.Equal(f.booking.ExpiresAt) {
		t.Errorf("session expiry %v must match the hold expiry %v", gotReq.ExpiresAt, f.booking.ExpiresAt)
	}
	if gotReq.Metadata["booking_id"] != f.booking.ID.String() || gotReq.Metadata["user_id"] != f.booking.UserID.String() {
		t.Errorf("metadata wrong: %v", gotReq.Metadata)
	}
	if !strings.Contains(gotReq.Description, "2026-08-01") ||
		!strings.Contains(gotReq.Description, "07:00") ||
		!strings.Contains(gotReq.Description, RouteDesna) {
		t.Errorf("description missing slot details: %q", gotReq.Description)
	}
	// Itemised line items: big + child.
	if len(gotReq.LineItems) != 2 {
		t.Fatalf("line items = %+v, want 2 rows", gotReq.LineItems)
	}
}

func TestCreateCheckoutOverrideAmount(t *testing.T) {
	f := newCheckoutFixture()
	f.booking.PriceOverride = fptr(300)

	var gotReq provider.CreateSessionRequest
	f.gateway.createSession = func(_ context.Context, req provider.CreateSessionRequest) (provider.CreateSessionResponse, error) {
		gotReq = req
		return provider.CreateSessionResponse{SessionID: "sess_1"}, nil
	}

	if _, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotReq.AmountUAH != 300 {
		t.Errorf("override must win: amount = %v, want 300", gotReq.AmountUAH)
	}
	if len(gotReq.LineItems) != 1 || gotReq.LineItems[0].AmountEUR != 300 || gotReq.LineItems[0].Quantity != 1 {
		t.Errorf("override must collapse to a single line: %+v", gotReq.LineItems)
	}
}

func TestCreateCheckoutFailures(t *testing.T) {
	t.Run("gateway failure propagates and session is not persisted", func(t *testing.T) {
		f := newCheckoutFixture()
		boom := errors.New("gateway 500")
		f.gateway.createSession = func(context.Context, provider.CreateSessionRequest) (provider.CreateSessionResponse, error) {
			return provider.CreateSessionResponse{}, boom
		}
		f.bookings.setPaymentSessionID = nil // must not be called
		if _, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result"); !errors.Is(err, boom) {
			t.Errorf("want gateway error, got %v", err)
		}
	})

	t.Run("session save failure propagates", func(t *testing.T) {
		f := newCheckoutFixture()
		boom := errors.New("save failed")
		f.bookings.setPaymentSessionID = func(context.Context, uuid.UUID, string) error { return boom }
		if _, err := f.service(t).CreateCheckout(context.Background(), f.booking.ID, f.booking.UserID, "https://app/result"); !errors.Is(err, boom) {
			t.Errorf("want save error, got %v", err)
		}
	})
}

func TestBuildLineItems(t *testing.T) {
	t.Run("itemises every non-zero quantity at list price", func(t *testing.T) {
		b := mkBooking()
		b.QtyBig, b.QtyMedium, b.QtySmall, b.QtyChild = 2, 1, 1, 3
		items := buildLineItems(b, 9999) // effective is irrelevant without an override
		if len(items) != 4 {
			t.Fatalf("got %d items, want 4: %+v", len(items), items)
		}
		wantQty := []int{2, 1, 1, 3}
		wantAmount := []float64{450, 450, 450, 225}
		for i, it := range items {
			if it.Quantity != wantQty[i] || it.AmountEUR != wantAmount[i] {
				t.Errorf("item %d = %+v, want qty %d amount %v", i, it, wantQty[i], wantAmount[i])
			}
		}
	})

	t.Run("zero quantities are omitted", func(t *testing.T) {
		b := mkBooking() // only QtyBig = 1
		items := buildLineItems(b, 450)
		if len(items) != 1 || items[0].Quantity != 1 {
			t.Errorf("items = %+v, want single big line", items)
		}
	})

	t.Run("price override collapses to one discounted line", func(t *testing.T) {
		b := mkBooking()
		b.QtyMedium = 2
		b.PriceOverride = fptr(500)
		items := buildLineItems(b, 500)
		if len(items) != 1 || items[0].AmountEUR != 500 || items[0].Quantity != 1 {
			t.Errorf("items = %+v, want single 500 line", items)
		}
	})

	t.Run("unknown route falls back to a single effective line", func(t *testing.T) {
		b := mkBooking()
		b.RouteName = "Atlantis"
		items := buildLineItems(b, 777)
		if len(items) != 1 || items[0].AmountEUR != 777 || items[0].Quantity != 1 {
			t.Errorf("items = %+v, want single fallback line", items)
		}
	})
}
