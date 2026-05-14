package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // LiqPay mandates SHA-1
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// liqpayAPIURL is the hosted-checkout endpoint.
const liqpayAPIURL = "https://www.liqpay.ua/api/3/checkout"

// liqpayWebhookSignatureHeader is not a real HTTP header — LiqPay sends
// `data` and `signature` as POST form fields in the webhook body.
// We keep this constant only for documentation clarity.
const liqpayWebhookDataField = "data"
const liqpayWebhookSignatureField = "signature"

// LiqPayGateway implements PaymentGateway using the LiqPay v3 API.
//
// CreateSession builds a signed `data` payload and returns the hosted
// checkout URL that the front-end redirects the user to.
//
// ParseWebhook verifies the HMAC-SHA1 signature over the raw `data` field
// and maps LiqPay's status string to our canonical PaymentStatus.
type LiqPayGateway struct {
	publicKey  string
	privateKey string
	backendURL string
}

// NewLiqPayGateway constructs a LiqPayGateway.
// backendURL is used to build the server_url (webhook callback) for each session.
func NewLiqPayGateway(publicKey, privateKey, backendURL string) *LiqPayGateway {
	return &LiqPayGateway{
		publicKey:  publicKey,
		privateKey: privateKey,
		backendURL: strings.TrimRight(backendURL, "/"),
	}
}

// ─── CreateSession ────────────────────────────────────────────────────────────

// liqpayRequest is the v3 payment params object that gets Base64-encoded as `data`.
type liqpayRequest struct {
	Version     int    `json:"version"`
	PublicKey   string `json:"public_key"`
	Action      string `json:"action"`
	Amount      string `json:"amount"` // UAH, no decimals for kopecks needed — ceil to hryvnia
	Currency    string `json:"currency"`
	Description string `json:"description"`
	OrderID     string `json:"order_id"`
	ServerURL   string `json:"server_url"`
	ResultURL   string `json:"result_url"`
	ExpiredDate string `json:"expired_date"` // "YYYY-MM-DD HH:MM:SS"
	Language    string `json:"language"`
}

func (g *LiqPayGateway) CreateSession(ctx context.Context, req CreateSessionRequest) (CreateSessionResponse, error) {
	// LiqPay works in UAH; the rest of the app uses EUR.
	// The caller passes AmountEUR — we treat it as the display currency and pass
	// currency=EUR so LiqPay converts at its live rate.  Change to UAH + conversion
	// if you prefer a fixed-rate approach.
	amount := math.Ceil(req.AmountEUR*100) / 100 // round to 2 dp

	webhookURL := fmt.Sprintf("%s/payment/webhook?booking_id=%s", g.backendURL, req.BookingID)

	params := liqpayRequest{
		Version:     3,
		PublicKey:   g.publicKey,
		Action:      "pay",
		Amount:      fmt.Sprintf("%.2f", amount),
		Currency:    "EUR",
		Description: req.Description,
		OrderID:     req.BookingID, // use booking UUID as LiqPay order_id
		ServerURL:   webhookURL,
		ResultURL:   req.ResultURL,
		ExpiredDate: req.ExpiresAt.UTC().Format("2006-01-02 15:04:05"),
		Language:    "uk",
	}

	data, err := g.encodeData(params)
	if err != nil {
		return CreateSessionResponse{}, fmt.Errorf("liqpay encode data: %w", err)
	}
	signature := g.sign(data)

	// Build the redirect URL with data + signature as query params.
	checkoutURL := fmt.Sprintf("%s?data=%s&signature=%s",
		liqpayAPIURL,
		url.QueryEscape(data),
		url.QueryEscape(signature),
	)

	// LiqPay does not return a separate session token; we use the booking UUID
	// so our webhook handler can look up the booking from order_id.
	return CreateSessionResponse{
		SessionID:   req.BookingID,
		CheckoutURL: checkoutURL,
	}, nil
}

// ─── ParseWebhook ─────────────────────────────────────────────────────────────

// liqpayCallback is the decoded webhook payload sent by LiqPay.
type liqpayCallback struct {
	OrderID        string  `json:"order_id"`
	PaymentID      int64   `json:"payment_id"`
	Status         string  `json:"status"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	ErrCode        string  `json:"err_code,omitempty"`
	ErrDescription string  `json:"err_description,omitempty"`
}

// ParseWebhook expects a raw application/x-www-form-urlencoded body with
// `data` and `signature` fields — exactly what LiqPay POSTs to server_url.
//
// Verification: SHA1( privateKey + data + privateKey ), then Base64-encode,
// compare to the `signature` field. This is LiqPay's documented algorithm.
func (g *LiqPayGateway) ParseWebhook(
	ctx context.Context,
	rawBody []byte,
	headers map[string]string,
) (WebhookEvent, error) {
	// 1. Parse form body.
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return WebhookEvent{}, fmt.Errorf("liqpay: parse webhook body: %w", err)
	}

	data := values.Get(liqpayWebhookDataField)
	signature := values.Get(liqpayWebhookSignatureField)
	if data == "" || signature == "" {
		return WebhookEvent{}, fmt.Errorf("liqpay: missing data or signature field")
	}

	// 2. Verify HMAC-SHA1.
	expected := g.sign(data)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return WebhookEvent{}, fmt.Errorf("liqpay: invalid signature")
	}

	// 3. Decode the Base64 JSON payload.
	jsonBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return WebhookEvent{}, fmt.Errorf("liqpay: decode data: %w", err)
	}
	var cb liqpayCallback
	if err := json.Unmarshal(jsonBytes, &cb); err != nil {
		return WebhookEvent{}, fmt.Errorf("liqpay: unmarshal callback: %w", err)
	}

	// 4. Map LiqPay status to our canonical status.
	//    https://www.liqpay.ua/documentation/api/information/status/doc
	status, err := mapLiqPayStatus(cb.Status)
	if err != nil {
		return WebhookEvent{}, err
	}

	return WebhookEvent{
		SessionID: cb.OrderID, // we set order_id = booking UUID in CreateSession
		BookingID: cb.OrderID,
		Status:    status,
		RawBody:   rawBody,
	}, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// encodeData JSON-encodes v and returns the Base64-standard-encoded string.
func (g *LiqPayGateway) encodeData(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// sign computes Base64( SHA1( privateKey + data + privateKey ) ).
// This is the exact algorithm from LiqPay's v3 documentation.
func (g *LiqPayGateway) sign(data string) string {
	h := sha1.New() //nolint:gosec // required by LiqPay
	h.Write([]byte(g.privateKey + data + g.privateKey))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// mapLiqPayStatus translates a LiqPay status string to our PaymentStatus.
// Statuses that indicate the payment is still in-flight are treated as errors
// so the webhook handler ignores them (returns nil → 200 to LiqPay).
func mapLiqPayStatus(s string) (PaymentStatus, error) {
	switch s {
	// Terminal success
	case "success", "subscribed":
		return PaymentStatusPaid, nil

	// Terminal failures
	case "error", "failure", "reversed":
		return PaymentStatusFailed, nil

	// Expiry
	case "expired":
		return PaymentStatusExpired, nil

	// In-flight / informational — return an error so the caller can skip
	case "processing", "prepared", "hold_wait", "wait_secure", "wait_accept",
		"wait_liqpay", "sandbox":
		return "", fmt.Errorf("liqpay: non-terminal status %q — skip", s)

	default:
		return "", fmt.Errorf("liqpay: unknown status %q", s)
	}
}

// ─── Unused but satisfies the interface ───────────────────────────────────────
var _ PaymentGateway = (*LiqPayGateway)(nil)

// httpClient is package-level so it can be overridden in tests.
var httpClient = &http.Client{Timeout: 15 * time.Second}
