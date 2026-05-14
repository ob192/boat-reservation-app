package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

type WebhookHandler struct {
	gateway        provider.PaymentGateway
	webhookService service.WebhookService
	log            *slog.Logger
}

func NewWebhookHandler(
	gateway provider.PaymentGateway,
	webhookService service.WebhookService,
	log *slog.Logger,
) *WebhookHandler {
	return &WebhookHandler{
		gateway:        gateway,
		webhookService: webhookService,
		log:            log,
	}
}

// Handle is POST /payment/webhook.
//
// @Summary      Payment gateway webhook
// @Description  Receives payment status updates via HMAC-secured webhook.
// @Tags         webhooks
// @Accept       json
// @Produce      json
// @Success      200  {string}  string "OK"
// @Failure      400  {object}  httpx.ErrorBody "INVALID_SIGNATURE or CANNOT_READ_BODY"
// @Failure      500  {object}  httpx.ErrorBody "WEBHOOK_PROCESSING_FAILED"
// @Router       /payment/webhook [post]
func (h *WebhookHandler) Handle(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httpx.Err(c, http.StatusBadRequest, httpx.CodeCannotReadBody, "")
		return
	}

	headers := make(map[string]string, len(c.Request.Header))
	for k, v := range c.Request.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	event, err := h.gateway.ParseWebhook(c.Request.Context(), rawBody, headers)
	if err != nil {
		// Returning 400 INVALID_SIGNATURE on every parse failure (including
		// signature, malformed body, unknown status) is consistent with the spec —
		// gateways should retry, but only if their own signature is valid.
		h.log.Warn("webhook parse failed", "err", err)
		httpx.Err(c, http.StatusBadRequest, httpx.CodeInvalidSignature, "")
		return
	}

	if err := h.webhookService.Handle(c.Request.Context(), event); err != nil {
		h.log.Error("webhook process failed", "err", err, "session_id", event.SessionID)
		httpx.Err(c, http.StatusInternalServerError, httpx.CodeWebhookFailed, "")
		return
	}
	c.Status(http.StatusOK)
}
