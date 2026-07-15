package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PosterProduct is one line on the incoming order.
type PosterProduct struct {
	ProductID int `json:"product_id"`
	Count     int `json:"count"`
}

// PosterPayment records prepayment info (Poster expects kopecks).
type PosterPayment struct {
	Type     int    `json:"type"`     // 1 = prepaid 0 = not paid default
	Sum      int64  `json:"sum"`      // kopecks
	Currency string `json:"currency"` // must match account currency, e.g. "UAH"
}

// PosterOrder is the createIncomingOrder request body.
type PosterOrder struct {
	SpotID      int             `json:"spot_id"`
	ServiceMode int             `json:"service_mode"`
	FirstName   string          `json:"first_name,omitempty"`
	LastName    string          `json:"last_name,omitempty"`
	Phone       string          `json:"phone"`
	Email       string          `json:"email,omitempty"`
	Comment     string          `json:"comment,omitempty"`
	Products    []PosterProduct `json:"products"`
	Payment     *PosterPayment  `json:"payment,omitempty"`
}

// PosterOrderResult is the subset of the response we persist.
type PosterOrderResult struct {
	IncomingOrderID       int64
	IncomingTransactionID int64
}

// PosterClient creates incoming orders in Poster.
type PosterClient interface {
	CreateIncomingOrder(ctx context.Context, order PosterOrder) (PosterOrderResult, error)
}

type posterClient struct {
	token   string
	baseURL string
	http    *http.Client
}

func NewPosterClient(token string) PosterClient {
	return &posterClient{
		token:   token,
		baseURL: "https://joinposter.com/api",
		http:    &http.Client{Timeout: 15 * time.Second},
	}
}

// posterResponse mirrors Poster's envelope. `response` shape varies, so decode
// loosely and pull the two ids we need.
type posterResponse struct {
	Response *struct {
		IncomingOrderID json.Number `json:"incoming_order_id"`
		TransactionID   json.Number `json:"transaction_id"`
	} `json:"response"`
	Error any `json:"error"`
}

func (c *posterClient) CreateIncomingOrder(ctx context.Context, order PosterOrder) (PosterOrderResult, error) {

	body, err := json.Marshal(order)
	if err != nil {
		return PosterOrderResult{}, fmt.Errorf("marshal poster order: %w", err)
	}

	url := fmt.Sprintf("%s/incomingOrders.createIncomingOrder?token=%s", c.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return PosterOrderResult{}, fmt.Errorf("build poster request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return PosterOrderResult{}, fmt.Errorf("poster request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PosterOrderResult{}, fmt.Errorf("poster returned status %d", resp.StatusCode)
	}

	var pr posterResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return PosterOrderResult{}, fmt.Errorf("decode poster response: %w", err)
	}
	if pr.Error != nil && !isEmptyError(pr.Error) {
		return PosterOrderResult{}, fmt.Errorf("poster error: %v", pr.Error)
	}
	if pr.Response == nil {
		return PosterOrderResult{}, fmt.Errorf("poster: empty response body")
	}

	orderID, err := pr.Response.IncomingOrderID.Int64()
	if err != nil {
		return PosterOrderResult{}, fmt.Errorf("parse incoming_order_id: %w", err)
	}
	txID, err := pr.Response.TransactionID.Int64()
	if err != nil {
		return PosterOrderResult{}, fmt.Errorf("parse transaction_id: %w", err)
	}

	return PosterOrderResult{IncomingOrderID: orderID, IncomingTransactionID: txID}, nil
}

func isEmptyError(v any) bool {

	switch e := v.(type) {
	case string:
		return strings.TrimSpace(e) == ""
	case float64:
		return e == 0
	default:
		return false
	}
}
