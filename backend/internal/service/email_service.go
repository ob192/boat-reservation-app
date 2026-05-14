package service

import (
	"github.com/harbour-wave/harbour-wave-backend/internal/config"
	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"log/slog"
)

// EmailService sends booking confirmation emails.
//
// The send is fire-and-forget from the webhook path; the implementation must
// never block on SMTP for more than a few seconds.
type EmailService interface {
	SendConfirmation(b *model.Booking)
}

type emailService struct {
	cfg *config.Config
	log *slog.Logger
}

func NewEmailService(cfg *config.Config, log *slog.Logger) EmailService {
	return &emailService{cfg: cfg, log: log}
}

// SendConfirmation builds and sends the Ukrainian confirmation email.
// On failure we log and return — never panic, never block the caller's goroutine for long.
func (s *emailService) SendConfirmation(b *model.Booking) {
	if b == nil || b.UserEmail == "" {
		return
	}

	// "Ваше бронювання підтверджено — Harbour & Wave"
}
