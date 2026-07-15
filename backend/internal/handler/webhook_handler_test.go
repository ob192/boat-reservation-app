package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func webhookRouter(gw provider.PaymentGateway, svc service.WebhookService) *gin.Engine {
	h := NewWebhookHandler(gw, svc, testLogger())
	r := gin.New()
	r.POST("/payment/webhook", h.Handle)
	return r
}

func TestWebhookHandler(t *testing.T) {
	t.Run("gateway receives the RAW body bytes and first header values", func(t *testing.T) {
		var gotBody []byte
		var gotHeaders map[string]string
		gw := &mockGateway{
			parseWebhook: func(_ context.Context, rawBody []byte, headers map[string]string) (provider.WebhookEvent, error) {
				gotBody = rawBody
				gotHeaders = headers
				return provider.WebhookEvent{SessionID: "sess_1", Status: provider.PaymentStatusPaid}, nil
			},
		}
		svc := &mockWebhookService{handle: func(context.Context, provider.WebhookEvent) error { return nil }}

		// Deliberately quirky JSON spacing — HMAC verification depends on the exact bytes.
		raw := `{ "sessionId":"sess_1",   "status":"paid" }`
		w := doRequest(t, webhookRouter(gw, svc), http.MethodPost, "/payment/webhook", raw, map[string]string{
			"X-Signature": "abc123",
		})
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
		}
		if string(gotBody) != raw {
			t.Errorf("raw body altered:\n got %q\nwant %q", gotBody, raw)
		}
		if gotHeaders["X-Signature"] != "abc123" {
			t.Errorf("headers = %v", gotHeaders)
		}
	})

	t.Run("parse failure is 400 INVALID_SIGNATURE", func(t *testing.T) {
		gw := &mockGateway{
			parseWebhook: func(context.Context, []byte, map[string]string) (provider.WebhookEvent, error) {
				return provider.WebhookEvent{}, errors.New("bad signature")
			},
		}
		// The webhook service must never see an unverified event.
		w := doRequest(t, webhookRouter(gw, &mockWebhookService{}), http.MethodPost, "/payment/webhook", `{}`, nil)
		wantError(t, w, http.StatusBadRequest, httpx.CodeInvalidSignature)
	})

	t.Run("processing failure is 500 so the gateway retries", func(t *testing.T) {
		gw := &mockGateway{
			parseWebhook: func(context.Context, []byte, map[string]string) (provider.WebhookEvent, error) {
				return provider.WebhookEvent{SessionID: "sess_1", Status: provider.PaymentStatusPaid}, nil
			},
		}
		svc := &mockWebhookService{
			handle: func(context.Context, provider.WebhookEvent) error { return errors.New("db down") },
		}
		w := doRequest(t, webhookRouter(gw, svc), http.MethodPost, "/payment/webhook", `{}`, nil)
		wantError(t, w, http.StatusInternalServerError, httpx.CodeWebhookFailed)
	})

	t.Run("parsed event is handed to the service", func(t *testing.T) {
		event := provider.WebhookEvent{SessionID: "sess_9", BookingID: "b-1", Status: provider.PaymentStatusFailed}
		gw := &mockGateway{
			parseWebhook: func(context.Context, []byte, map[string]string) (provider.WebhookEvent, error) {
				return event, nil
			},
		}
		var got provider.WebhookEvent
		svc := &mockWebhookService{
			handle: func(_ context.Context, ev provider.WebhookEvent) error {
				got = ev
				return nil
			},
		}
		w := doRequest(t, webhookRouter(gw, svc), http.MethodPost, "/payment/webhook", `{}`, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		if got.SessionID != "sess_9" || got.Status != provider.PaymentStatusFailed {
			t.Errorf("event = %+v", got)
		}
	})

	t.Run("unreadable body is 400 CANNOT_READ_BODY", func(t *testing.T) {
		gw := &mockGateway{}
		svc := &mockWebhookService{}
		r := webhookRouter(gw, svc)

		req := httptest.NewRequest(http.MethodPost, "/payment/webhook", failingReader{})
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		wantError(t, w, http.StatusBadRequest, httpx.CodeCannotReadBody)
	})
}

// failingReader errors on the first read, simulating a broken connection.
type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("connection reset") }
