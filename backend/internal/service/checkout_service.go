package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// CheckoutService creates payment sessions for pending bookings.
type CheckoutService interface {
	CreateCheckout(ctx context.Context, bookingID uuid.UUID, userID uuid.UUID, successURL string) (*model.CreateCheckoutResponse, error)
}

type checkoutService struct {
	bookingSvc BookingService
	bookings   repository.BookingRepository
	pricing    PricingService
	gateway    provider.PaymentGateway
	cfg        *config.Config
	clock      platform.Clock
}

func NewCheckoutService(
	bookingSvc BookingService,
	bookings repository.BookingRepository,
	pricing PricingService,
	gateway provider.PaymentGateway,
	cfg *config.Config,
	clock platform.Clock,
) CheckoutService {
	return &checkoutService{
		bookingSvc: bookingSvc,
		bookings:   bookings,
		pricing:    pricing,
		gateway:    gateway,
		cfg:        cfg,
		clock:      clock,
	}
}

func (s *checkoutService) CreateCheckout(
	ctx context.Context,
	bookingID uuid.UUID,
	userID uuid.UUID,
	resultUrl string,
) (*model.CreateCheckoutResponse, error) {
	b, err := s.bookingSvc.GetByIDForUser(ctx, bookingID, userID)
	if err != nil {
		return nil, err
	}

	if b.Status != model.StatusPending {
		return nil, ErrBookingNotPending
	}
	if !b.ExpiresAt.After(s.clock.Now()) {
		return nil, ErrBookingExpired
	}

	amount := s.pricing.EffectiveAmount(b.TotalAmount, b.PriceOverride)

	req := provider.CreateSessionRequest{
		BookingID:   b.ID.String(),
		UserEmail:   b.UserEmail,
		AmountUAH:   amount,
		Description: fmt.Sprintf("SUP Chernihiv — Бронювання %s %s (%s)", b.DateFormatted(), b.Time, b.RouteName),
		LineItems:   buildLineItems(b, amount),
		ResultURL:   resultUrl,
		ExpiresAt:   b.ExpiresAt,
		Metadata: map[string]string{
			"booking_id": b.ID.String(),
			"user_id":    b.UserID.String(),
		},
	}

	resp, err := s.gateway.CreateSession(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("create gateway session: %w", err)
	}

	if err := s.bookings.SetPaymentSessionID(ctx, b.ID, resp.SessionID); err != nil {
		return nil, fmt.Errorf("save session id: %w", err)
	}

	return &model.CreateCheckoutResponse{
		CheckoutURL: resp.CheckoutURL, //"https://www.privat24.ua/rd/send_qr/liqpay_static_qr/p/dhcK5x",
		SessionID:   resp.SessionID,
	}, nil
}

// buildLineItems chooses between itemised line items and a single discounted line
// per the spec's pricing rules.
func buildLineItems(b *model.Booking, effective float64) []provider.LineItem {
	if b.PriceOverride != nil {
		return []provider.LineItem{
			{
				Name:      "SUP Chernihiv — Бронювання (зі знижкою)",
				AmountEUR: effective,
				Quantity:  1,
			},
		}
	}

	p, ok := RoutePriceFor(b.RouteName)
	if !ok {
		// Unknown route — fall back to a single effective-amount line.
		return []provider.LineItem{{Name: "Бронювання", AmountEUR: effective, Quantity: 1}}
	}

	items := make([]provider.LineItem, 0, 4)
	if b.QtyBig > 0 {
		items = append(items, provider.LineItem{Name: "Великий борд", AmountEUR: p.Big, Quantity: b.QtyBig})
	}
	if b.QtyMedium > 0 {
		items = append(items, provider.LineItem{Name: "Середній борд", AmountEUR: p.Medium, Quantity: b.QtyMedium})
	}
	if b.QtySmall > 0 {
		items = append(items, provider.LineItem{Name: "Малий борд", AmountEUR: p.Small, Quantity: b.QtySmall})
	}
	if b.QtyChild > 0 {
		items = append(items, provider.LineItem{Name: "Дитяче місце", AmountEUR: p.Child, Quantity: b.QtyChild})
	}
	return items
}
