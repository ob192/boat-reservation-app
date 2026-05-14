package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// CreateSessionRequest is everything the gateway needs to start a hosted checkout.
type CreateSessionRequest struct {
	BookingID   string
	UserEmail   string
	AmountEUR   float64
	Description string
	LineItems   []LineItem
	SuccessURL  string
	CancelURL   string
	ExpiresAt   time.Time
	Metadata    map[string]string
}

// LineItem is one row on the gateway's hosted checkout.
type LineItem struct {
	Name      string
	AmountEUR float64
	Quantity  int
}

// CreateSessionResponse is what we hand back to the caller after creating a session.
type CreateSessionResponse struct {
	SessionID   string
	CheckoutURL string
}

// WebhookEvent is the normalised view of an incoming gateway webhook.
type WebhookEvent struct {
	SessionID string
	BookingID string
	Status    PaymentStatus
	RawBody   []byte
}

// PaymentStatus is the canonical payment outcome.
type PaymentStatus string

const (
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusExpired PaymentStatus = "expired"
)

// PaymentGateway is the single seam between this app and any payment provider.
// Add a new provider by writing a new struct implementing this interface.
type PaymentGateway interface {
	CreateSession(ctx context.Context, req CreateSessionRequest) (CreateSessionResponse, error)
	// ParseWebhook MUST receive the raw, unmodified request body bytes.
	// Re-encoding (e.g. via JSON unmarshal/marshal) breaks HMAC verification.
	ParseWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (WebhookEvent, error)
}

// ============================================================================
// Stub adapter — used in local dev when PAYMENT_GATEWAY=stub.
// Generates a fake session id, returns a placeholder URL.
// ParseWebhook expects a JSON body with sessionId/bookingId/status fields.
// Real providers should replace this with HMAC-verified parsing.
// ============================================================================

type StubGateway struct {
	BackendURL string
}

func NewStubGateway(backendURL string) *StubGateway {
	return &StubGateway{BackendURL: strings.TrimRight(backendURL, "/")}
}

func (g *StubGateway) CreateSession(ctx context.Context, req CreateSessionRequest) (CreateSessionResponse, error) {
	id, err := randomHex(12)
	if err != nil {
		return CreateSessionResponse{}, fmt.Errorf("generate session id: %w", err)
	}
	sessionID := "sess_" + id
	return CreateSessionResponse{
		SessionID:   sessionID,
		CheckoutURL: fmt.Sprintf("%s/payment/stub/%s", g.BackendURL, sessionID),
	}, nil
}

// ParseWebhook for the stub trusts the body. DO NOT use this in production.
func (g *StubGateway) ParseWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (WebhookEvent, error) {
	// Body shape: {"sessionId":"sess_xxx","bookingId":"...","status":"paid|failed|expired"}
	var p struct {
		SessionID string `json:"sessionId"`
		BookingID string `json:"bookingId"`
		Status    string `json:"status"`
	}
	if err := jsonUnmarshal(rawBody, &p); err != nil {
		return WebhookEvent{}, fmt.Errorf("decode stub webhook: %w", err)
	}
	if p.SessionID == "" {
		return WebhookEvent{}, fmt.Errorf("missing sessionId")
	}
	var status PaymentStatus
	switch strings.ToLower(p.Status) {
	case "paid":
		status = PaymentStatusPaid
	case "failed":
		status = PaymentStatusFailed
	case "expired":
		status = PaymentStatusExpired
	default:
		return WebhookEvent{}, fmt.Errorf("unknown status %q", p.Status)
	}
	return WebhookEvent{
		SessionID: p.SessionID,
		BookingID: p.BookingID,
		Status:    status,
		RawBody:   rawBody,
	}, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
