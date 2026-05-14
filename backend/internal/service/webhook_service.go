package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// WebhookService applies an incoming, already-verified payment event to the booking record.
type WebhookService interface {
	Handle(ctx context.Context, event provider.WebhookEvent) error
}

type webhookService struct {
	bookings repository.BookingRepository
	emails   EmailService
	log      *slog.Logger
}

func NewWebhookService(
	bookings repository.BookingRepository,
	emails EmailService,
	log *slog.Logger,
) WebhookService {
	return &webhookService{
		bookings: bookings,
		emails:   emails,
		log:      log,
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
		// Skip if already confirmed — webhook retries are safe.
		if booking.Status == model.StatusConfirmed {
			return nil
		}
		if err := s.bookings.SetStatus(ctx, booking.ID, model.StatusConfirmed); err != nil {
			return fmt.Errorf("set confirmed: %w", err)
		}
		// Fire-and-forget — we never want SMTP latency to make the gateway retry.
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
