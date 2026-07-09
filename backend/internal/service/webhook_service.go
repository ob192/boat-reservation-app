package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// WebhookService applies an incoming, already-verified payment event to the booking record.
type WebhookService interface {
	Handle(ctx context.Context, event provider.WebhookEvent) error
}

type webhookService struct {
	bookings   repository.BookingRepository
	promocodes repository.PromocodeRepository
	emails     EmailService
	poster     provider.PosterClient
	pricing    PricingService
	log        *slog.Logger
}

func NewWebhookService(
	bookings repository.BookingRepository,
	promocodes repository.PromocodeRepository,
	emails EmailService,
	poster provider.PosterClient,
	pricing PricingService,
	log *slog.Logger,
) WebhookService {
	return &webhookService{
		bookings:   bookings,
		promocodes: promocodes,
		emails:     emails,
		poster:     poster,
		pricing:    pricing,
		log:        log,
	}
}

func (s *webhookService) Handle(ctx context.Context, event provider.WebhookEvent) error {
	booking, err := s.bookings.FindByPaymentSessionID(ctx, event.SessionID)
	if errors.Is(err, repository.ErrNotFound) {
		// Common cases: replayed webhook with old session, or session belonging
		// to another tenant. Log and accept (200) so the gateway doesn't keep retrying.
		s.log.Warn("webhook for unknown session", "session_id", event.SessionID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("find booking by session %s: %w", event.SessionID, err)
	}

	switch event.Status {
	case provider.PaymentStatusPaid:
		if booking.Status == model.StatusConfirmed {
			return nil
		}
		if err := s.bookings.SetStatus(ctx, booking.ID, model.StatusConfirmed); err != nil {
			return fmt.Errorf("set confirmed: %w", err)
		}

		if booking.PromoCode != nil && *booking.PromoCode != "" {
			ok, err := s.promocodes.IncrementUsage(ctx, *booking.PromoCode)
			if err != nil {
				s.log.Error("promo increment failed", "err", err, "code", *booking.PromoCode, "booking_id", booking.ID)
			} else if !ok {
				s.log.Warn("promo cap already reached at confirmation; discount was honored",
					"code", *booking.PromoCode, "booking_id", booking.ID)
			}
		}

		if err := s.createPosterOrder(ctx, booking); err != nil {
			s.log.Error("poster order failed", "err", err, "booking_id", booking.ID)
		}

		bookingSnapshot := *booking
		bookingSnapshot.Status = model.StatusConfirmed
		go s.emails.SendConfirmation(&bookingSnapshot)

	case provider.PaymentStatusFailed:
		// Don't downgrade an already-confirmed booking from a stale webhook.
		if booking.Status == model.StatusConfirmed {
			return nil
		}
		if err := s.bookings.SetStatus(ctx, booking.ID, model.StatusFailed); err != nil {
			return fmt.Errorf("set failed: %w", err)
		}

	case provider.PaymentStatusExpired:
		if booking.Status == model.StatusConfirmed {
			return nil
		}
		if err := s.bookings.SetStatus(ctx, booking.ID, model.StatusExpired); err != nil {
			return fmt.Errorf("set expired: %w", err)
		}

	default:
		s.log.Warn("unknown webhook status", "status", string(event.Status))
	}
	return nil
}

func (s *webhookService) createPosterOrder(ctx context.Context, booking *model.Booking) error {
	if booking.Phone == nil || *booking.Phone == "" {
		return fmt.Errorf("booking %s has no phone, skipping poster order", booking.ID)
	}

	order := s.buildPosterOrder(booking)

	go func() {

		ctx := context.WithoutCancel(ctx)
		ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		res, err := s.poster.CreateIncomingOrder(ctx, order)
		if err != nil {
			slog.Error("create poster incoming order",
				"error", err,
				"booking_id", booking.ID,
				"spot_id", order.SpotID,
				"phone", order.Phone,
			)
			return
		}

		slog.Info("poster incoming order created",
			"booking_id", booking.ID,
			"incoming_order_id", res.IncomingOrderID,
			"incoming_transaction_id", res.IncomingTransactionID,
		)

		if err := s.bookings.SetPosterIDs(ctx, booking.ID, res.IncomingOrderID, res.IncomingTransactionID); err != nil {
			slog.Error("save poster ids",
				"error", err,
				"booking_id", booking.ID,
				"incoming_order_id", res.IncomingOrderID,
				"incoming_transaction_id", res.IncomingTransactionID,
			)
			return
		}
	}()

	return nil
}

// buildPosterOrder maps a confirmed booking to a Poster incoming order:
//   - full client details (name, phone, email),
//   - per-line catalog prices in kopecks,
//   - payment.sum = the amount actually charged (post-discount / post-override).
func (s *webhookService) buildPosterOrder(booking *model.Booking) provider.PosterOrder {
	effective := s.pricing.EffectiveAmount(booking.TotalAmount, booking.PriceOverride)
	price, hasPrice := s.pricing.RoutePrice(booking.RouteName)

	products := make([]provider.PosterProduct, 0, 2)

	boardCount := booking.QtyBig + booking.QtyMedium + booking.QtySmall
	if boardCount > 0 {
		p := provider.PosterProduct{ProductID: 1, Count: boardCount}
		if hasPrice {
			// Boards are lumped into one product line; use the weighted list unit price.
			listBoards := float64(booking.QtyBig)*price.Big +
				float64(booking.QtyMedium)*price.Medium +
				float64(booking.QtySmall)*price.Small
			unit := toKopecks(listBoards / float64(boardCount))
			p.Price = &unit
		}
		products = append(products, p)
	}
	if booking.QtyChild > 0 {
		p := provider.PosterProduct{ProductID: 6, Count: booking.QtyChild}
		if hasPrice {
			unit := toKopecks(price.Child)
			p.Price = &unit
		}
		products = append(products, p)
	}

	comment := fmt.Sprintf("Бронювання %s %s (%s)", booking.DateFormatted(), booking.Time, booking.RouteName)
	if booking.PromoCode != nil && *booking.PromoCode != "" {
		pct := 0
		if booking.DiscountPercent != nil {
			pct = *booking.DiscountPercent
		}
		comment += fmt.Sprintf(" | Промокод %s (-%d%%)", *booking.PromoCode, pct)
	}

	phone := ""
	if booking.Phone != nil {
		phone = *booking.Phone
	}

	sum := toKopecks(effective)
	return provider.PosterOrder{
		SpotID:      1,
		ServiceMode: 1,
		FirstName:   booking.FirstName,
		LastName:    booking.LastName,
		Phone:       phone,
		Email:       booking.UserEmail,
		Comment:     comment,
		Products:    products,
		Payment: &provider.PosterPayment{
			Type:     1,
			Sum:      sum,
			Currency: "UAH",
		},
	}
}

func toKopecks(amount float64) int64 {
	return int64(math.Round(amount * 100))
}
