package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// webhookFixture wires a WebhookService around one stored booking and records
// the async side effects (poster order, poster ids, email) on channels so
// tests can await the goroutines Handle spawns.
type webhookFixture struct {
	booking *model.Booking

	bookings *mockBookingRepo
	promos   *mockPromoRepo
	poster   *mockPosterClient
	emails   *mockEmailService

	statusCh    chan model.BookingStatus
	orderCh     chan provider.PosterOrder
	posterIDsCh chan [2]int64
	emailCh     chan *model.Booking
}

func newWebhookFixture() *webhookFixture {
	f := &webhookFixture{
		booking:     mkBooking(),
		statusCh:    make(chan model.BookingStatus, 4),
		orderCh:     make(chan provider.PosterOrder, 4),
		posterIDsCh: make(chan [2]int64, 4),
		emailCh:     make(chan *model.Booking, 4),
	}
	f.bookings = &mockBookingRepo{
		findByPaymentSessionID: func(_ context.Context, sessionID string) (*model.Booking, error) {
			if f.booking.PaymentSessionID != nil && sessionID == *f.booking.PaymentSessionID {
				return f.booking, nil
			}
			return nil, repository.ErrNotFound
		},
		setStatus: func(_ context.Context, _ uuid.UUID, st model.BookingStatus) error {
			f.statusCh <- st
			return nil
		},
		setPosterIDs: func(_ context.Context, _ uuid.UUID, orderID, txID int64) error {
			f.posterIDsCh <- [2]int64{orderID, txID}
			return nil
		},
	}
	f.booking.PaymentSessionID = sptr("sess_1")
	f.promos = &mockPromoRepo{
		incrementUsage: func(context.Context, string) (bool, error) { return true, nil },
	}
	f.poster = &mockPosterClient{
		createIncomingOrder: func(_ context.Context, order provider.PosterOrder) (provider.PosterOrderResult, error) {
			f.orderCh <- order
			return provider.PosterOrderResult{IncomingOrderID: 101, IncomingTransactionID: 202}, nil
		},
	}
	f.emails = &mockEmailService{
		sendConfirmation: func(b *model.Booking) { f.emailCh <- b },
	}
	return f
}

func (f *webhookFixture) service() WebhookService {
	return NewWebhookService(f.bookings, f.promos, f.emails, f.poster, NewPricingService(), testLogger())
}

func paidEvent() provider.WebhookEvent {
	return provider.WebhookEvent{SessionID: "sess_1", Status: provider.PaymentStatusPaid}
}

// await pulls one value off ch or fails the test after a grace period —
// Handle fires poster/email work on goroutines.
func await[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		panic("unreachable")
	}
}

func assertNoSignal[T any](t *testing.T, ch <-chan T, what string) {
	t.Helper()
	select {
	case <-ch:
		t.Fatalf("unexpected %s", what)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestHandleUnknownSession(t *testing.T) {
	f := newWebhookFixture()
	// Unknown session is accepted quietly so the gateway stops retrying.
	err := f.service().Handle(context.Background(), provider.WebhookEvent{SessionID: "sess_unknown", Status: provider.PaymentStatusPaid})
	if err != nil {
		t.Errorf("want nil for unknown session, got %v", err)
	}
	assertNoSignal(t, f.statusCh, "status change")
}

func TestHandleLookupFailure(t *testing.T) {
	f := newWebhookFixture()
	boom := errors.New("db down")
	f.bookings.findByPaymentSessionID = func(context.Context, string) (*model.Booking, error) { return nil, boom }
	if err := f.service().Handle(context.Background(), paidEvent()); !errors.Is(err, boom) {
		t.Errorf("want lookup error, got %v", err)
	}
}

func TestHandlePaid(t *testing.T) {
	t.Run("already confirmed is a no-op", func(t *testing.T) {
		f := newWebhookFixture()
		f.booking.Status = model.StatusConfirmed
		f.bookings.setStatus = nil // must not be called
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Errorf("want nil, got %v", err)
		}
		assertNoSignal(t, f.emailCh, "confirmation email")
	})

	t.Run("set-status failure aborts", func(t *testing.T) {
		f := newWebhookFixture()
		boom := errors.New("update failed")
		f.bookings.setStatus = func(context.Context, uuid.UUID, model.BookingStatus) error { return boom }
		if err := f.service().Handle(context.Background(), paidEvent()); !errors.Is(err, boom) {
			t.Errorf("want set-status error, got %v", err)
		}
		assertNoSignal(t, f.emailCh, "confirmation email after failed status update")
	})

	t.Run("full happy path: status, promo, poster, ids, email", func(t *testing.T) {
		f := newWebhookFixture()
		f.booking.PromoCode = sptr("SUMMER")
		f.booking.DiscountPercent = iptr(10)
		f.booking.DiscountAmount = fptr(45)
		var promoBumped string
		f.promos.incrementUsage = func(_ context.Context, code string) (bool, error) {
			promoBumped = code
			return true, nil
		}

		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if st := await(t, f.statusCh, "status update"); st != model.StatusConfirmed {
			t.Errorf("status set to %s, want confirmed", st)
		}
		if promoBumped != "SUMMER" {
			t.Errorf("promo usage bumped for %q, want SUMMER", promoBumped)
		}

		order := await(t, f.orderCh, "poster order")
		if order.Phone != *f.booking.Phone || order.Email != f.booking.UserEmail {
			t.Errorf("poster contact fields wrong: %+v", order)
		}

		ids := await(t, f.posterIDsCh, "poster ids persisted")
		if ids != [2]int64{101, 202} {
			t.Errorf("poster ids = %v, want [101 202]", ids)
		}

		email := await(t, f.emailCh, "confirmation email")
		if email.Status != model.StatusConfirmed {
			t.Errorf("email snapshot status = %s, want confirmed", email.Status)
		}
		if email == f.booking {
			t.Error("email must receive a snapshot copy, not the shared booking pointer")
		}
	})

	t.Run("promo increment failure does not block confirmation", func(t *testing.T) {
		f := newWebhookFixture()
		f.booking.PromoCode = sptr("SUMMER")
		f.promos.incrementUsage = func(context.Context, string) (bool, error) {
			return false, errors.New("db down")
		}
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Errorf("promo failure must be swallowed, got %v", err)
		}
		await(t, f.emailCh, "confirmation email")
	})

	t.Run("promo cap reached at confirmation is only a warning", func(t *testing.T) {
		f := newWebhookFixture()
		f.booking.PromoCode = sptr("SUMMER")
		f.promos.incrementUsage = func(context.Context, string) (bool, error) { return false, nil }
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Errorf("cap-hit must be swallowed, got %v", err)
		}
		await(t, f.emailCh, "confirmation email")
	})

	t.Run("no promo skips the increment", func(t *testing.T) {
		f := newWebhookFixture()
		f.promos.incrementUsage = nil // must not be called
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		await(t, f.emailCh, "confirmation email")
	})

	t.Run("missing phone skips poster but still emails", func(t *testing.T) {
		f := newWebhookFixture()
		f.booking.Phone = nil
		f.poster.createIncomingOrder = nil // must not be called
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Errorf("poster skip must not fail the webhook, got %v", err)
		}
		await(t, f.emailCh, "confirmation email")
	})

	t.Run("poster failure is swallowed and ids are not saved", func(t *testing.T) {
		f := newWebhookFixture()
		posterCalled := make(chan struct{}, 1)
		f.poster.createIncomingOrder = func(context.Context, provider.PosterOrder) (provider.PosterOrderResult, error) {
			posterCalled <- struct{}{}
			return provider.PosterOrderResult{}, errors.New("poster 500")
		}
		f.bookings.setPosterIDs = nil // must not be called
		if err := f.service().Handle(context.Background(), paidEvent()); err != nil {
			t.Errorf("poster failure must be swallowed, got %v", err)
		}
		await(t, posterCalled, "poster attempt")
		await(t, f.emailCh, "confirmation email")
	})
}

func TestHandleFailedAndExpired(t *testing.T) {
	cases := []struct {
		status provider.PaymentStatus
		want   model.BookingStatus
	}{
		{provider.PaymentStatusFailed, model.StatusFailed},
		{provider.PaymentStatusExpired, model.StatusExpired},
	}
	for _, tc := range cases {
		t.Run(string(tc.status), func(t *testing.T) {
			f := newWebhookFixture()
			ev := paidEvent()
			ev.Status = tc.status
			if err := f.service().Handle(context.Background(), ev); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st := await(t, f.statusCh, "status update"); st != tc.want {
				t.Errorf("status = %s, want %s", st, tc.want)
			}
			assertNoSignal(t, f.emailCh, "email on non-paid webhook")
		})

		t.Run(string(tc.status)+" never downgrades a confirmed booking", func(t *testing.T) {
			f := newWebhookFixture()
			f.booking.Status = model.StatusConfirmed
			f.bookings.setStatus = nil // must not be called
			ev := paidEvent()
			ev.Status = tc.status
			if err := f.service().Handle(context.Background(), ev); err != nil {
				t.Errorf("stale webhook must be ignored, got %v", err)
			}
		})

		t.Run(string(tc.status)+" set-status failure propagates", func(t *testing.T) {
			f := newWebhookFixture()
			boom := errors.New("update failed")
			f.bookings.setStatus = func(context.Context, uuid.UUID, model.BookingStatus) error { return boom }
			ev := paidEvent()
			ev.Status = tc.status
			if err := f.service().Handle(context.Background(), ev); !errors.Is(err, boom) {
				t.Errorf("want error, got %v", err)
			}
		})
	}
}

func TestHandleUnknownStatus(t *testing.T) {
	f := newWebhookFixture()
	f.bookings.setStatus = nil // must not be called
	ev := paidEvent()
	ev.Status = provider.PaymentStatus("chargeback")
	if err := f.service().Handle(context.Background(), ev); err != nil {
		t.Errorf("unknown status must be logged and accepted, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// buildPosterOrder
// ----------------------------------------------------------------------------

func buildOrderFor(b *model.Booking) provider.PosterOrder {
	s := &webhookService{pricing: NewPricingService(), log: testLogger()}
	return s.buildPosterOrder(b)
}

func TestBuildPosterOrderLines(t *testing.T) {
	t.Run("boards lumped into one line, child separate, list prices in kopecks", func(t *testing.T) {
		b := mkBooking()
		b.QtyBig, b.QtyMedium, b.QtySmall, b.QtyChild = 1, 1, 0, 2

		order := buildOrderFor(b)
		if len(order.Products) != 2 {
			t.Fatalf("products = %+v, want boards + child", order.Products)
		}

		boards := order.Products[0]
		if boards.ProductID != 1 || boards.Count != 2 {
			t.Errorf("boards line = %+v, want product 1 × 2", boards)
		}
		// Desna boards are 450 each → weighted unit 450 → 45000 kopecks.
		if boards.Price == nil || *boards.Price != 45000 {
			t.Errorf("boards price = %v, want 45000", boards.Price)
		}

		child := order.Products[1]
		if child.ProductID != 6 || child.Count != 2 {
			t.Errorf("child line = %+v, want product 6 × 2", child)
		}
		if child.Price == nil || *child.Price != 22500 {
			t.Errorf("child price = %v, want 22500 (225 EUR)", child.Price)
		}
	})

	t.Run("promocode discount is applied per line", func(t *testing.T) {
		b := mkBooking()
		b.QtyBig, b.QtyChild = 2, 1
		b.PromoCode = sptr("SUMMER")
		b.DiscountPercent = iptr(10)
		b.DiscountAmount = fptr(112.5)

		order := buildOrderFor(b)
		if len(order.Products) != 2 {
			t.Fatalf("products = %+v", order.Products)
		}
		// 450 × 0.9 = 405 → 40500; 225 × 0.9 = 202.5 → 20250.
		if *order.Products[0].Price != 40500 {
			t.Errorf("discounted board price = %d, want 40500", *order.Products[0].Price)
		}
		if *order.Products[1].Price != 20250 {
			t.Errorf("discounted child price = %d, want 20250", *order.Products[1].Price)
		}

		// Line totals must reconstruct the charged (net) amount.
		total := *order.Products[0].Price*int64(order.Products[0].Count) +
			*order.Products[1].Price*int64(order.Products[1].Count)
		if total != 101250 { // 1012.50 UAH in kopecks
			t.Errorf("net line total = %d kopecks, want 101250", total)
		}
	})

	t.Run("unknown route omits prices so Poster uses its catalog", func(t *testing.T) {
		b := mkBooking()
		b.RouteName = "Atlantis"
		b.QtyChild = 1
		order := buildOrderFor(b)
		for i, p := range order.Products {
			if p.Price != nil {
				t.Errorf("product %d has price %d, want nil for unknown route", i, *p.Price)
			}
		}
	})

	t.Run("no boards yields only the child line", func(t *testing.T) {
		b := mkBooking()
		b.QtyBig, b.QtyChild = 0, 2
		order := buildOrderFor(b)
		if len(order.Products) != 1 || order.Products[0].ProductID != 6 {
			t.Errorf("products = %+v, want single child line", order.Products)
		}
	})

	t.Run("client details and constants", func(t *testing.T) {
		order := buildOrderFor(mkBooking())
		if order.SpotID != 1 || order.ServiceMode != 1 {
			t.Errorf("spot/mode = %d/%d, want 1/1", order.SpotID, order.ServiceMode)
		}
		if order.FirstName != "Olena" || order.LastName != "Kovalenko" ||
			order.Phone != "+380501112233" || order.Email != "user@example.com" {
			t.Errorf("client fields wrong: %+v", order)
		}
	})

	t.Run("nil phone becomes empty string", func(t *testing.T) {
		b := mkBooking()
		b.Phone = nil
		if got := buildOrderFor(b).Phone; got != "" {
			t.Errorf("phone = %q, want empty", got)
		}
	})
}

func TestPosterProductLine(t *testing.T) {
	t.Run("no price when route pricing unknown", func(t *testing.T) {
		p := posterProductLine(1, 2, 450, false, 10)
		if p.Price != nil {
			t.Errorf("price = %v, want nil", p.Price)
		}
		if p.ProductID != 1 || p.Count != 2 {
			t.Errorf("line = %+v", p)
		}
	})

	t.Run("zero discount keeps list price", func(t *testing.T) {
		p := posterProductLine(1, 1, 450, true, 0)
		if p.Price == nil || *p.Price != 45000 {
			t.Errorf("price = %v, want 45000", p.Price)
		}
	})

	t.Run("100%% discount zeroes the line", func(t *testing.T) {
		p := posterProductLine(1, 1, 450, true, 100)
		if p.Price == nil || *p.Price != 0 {
			t.Errorf("price = %v, want 0", p.Price)
		}
	})

	t.Run("fractional net rounds to nearest kopeck", func(t *testing.T) {
		// 333 × (100−15)/100 = 283.05 → 28305 kopecks.
		p := posterProductLine(1, 1, 333, true, 15)
		if p.Price == nil || *p.Price != 28305 {
			t.Errorf("price = %v, want 28305", p.Price)
		}
	})
}

func TestPosterComment(t *testing.T) {
	s := &webhookService{pricing: NewPricingService(), log: testLogger()}

	t.Run("plain booking", func(t *testing.T) {
		got := s.posterComment(mkBooking())
		want := fmt.Sprintf("Бронювання 2026-08-01 07:00 (%s)", RouteDesna)
		if got != want {
			t.Errorf("comment = %q, want %q", got, want)
		}
	})

	t.Run("promocode appended with percent", func(t *testing.T) {
		b := mkBooking()
		b.PromoCode = sptr("SUMMER")
		b.DiscountPercent = iptr(10)
		got := s.posterComment(b)
		if !strings.HasSuffix(got, "| Промокод SUMMER (-10%)") {
			t.Errorf("comment = %q, want promo suffix", got)
		}
	})

	t.Run("promocode with nil percent defaults to 0", func(t *testing.T) {
		b := mkBooking()
		b.PromoCode = sptr("SUMMER")
		got := s.posterComment(b)
		if !strings.Contains(got, "(-0%)") {
			t.Errorf("comment = %q, want -0%% fallback", got)
		}
	})
}

func TestToKopecks(t *testing.T) {
	tests := []struct {
		in   float64
		want int64
	}{
		{0, 0},
		{1, 100},
		{450, 45000},
		{202.5, 20250},
		{0.01, 1},
		{10.004, 1000}, // rounds down
		{10.006, 1001}, // rounds up
	}
	for _, tt := range tests {
		if got := toKopecks(tt.in); got != tt.want {
			t.Errorf("toKopecks(%v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestDerefString(t *testing.T) {
	if derefString(nil) != "" {
		t.Error("nil must deref to empty string")
	}
	if derefString(sptr("x")) != "x" {
		t.Error("pointer must deref to its value")
	}
}
