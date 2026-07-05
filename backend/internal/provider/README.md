# internal/provider

The seam between this app and external HTTP APIs: the payment gateway and Poster POS. Everything here is an interface + one or more concrete adapters, so swapping providers never touches `service`.

## Payment gateway

`payment_gateway.go` defines:
- `PaymentGateway` interface: `CreateSession` (start a hosted checkout) and `ParseWebhook` (turn a raw incoming webhook into a canonical `WebhookEvent`).
- **`ParseWebhook` must receive raw, unmodified request body bytes** — re-encoding via JSON unmarshal/marshal before verifying breaks HMAC signature checks. This is enforced by convention, not the type system, so watch for it in new adapters.
- `StubGateway` — the dev-mode adapter (`PAYMENT_GATEWAY=stub` or unset). Trusts the webhook body outright; never use in production.
- `PaymentStatus` canonical enum: `paid` / `failed` / `expired`.

`liqpay_gateway.go` — the real adapter (`PAYMENT_GATEWAY=liqpay`):
- `CreateSession` builds a Base64 JSON `data` payload, signs it (`SHA1(privateKey + data + privateKey)`, base64), and returns a redirect URL to LiqPay's hosted checkout. The booking's own UUID doubles as LiqPay's `order_id`/session ID (LiqPay has no separate session token).
- `ParseWebhook` expects `application/x-www-form-urlencoded` with `data`+`signature` fields (not JSON), re-derives the signature with `hmac.Equal` (constant-time), then decodes the Base64 JSON callback.
- `mapLiqPayStatus` treats in-flight statuses (`processing`, `wait_*`, `sandbox`, etc.) as errors deliberately — the webhook handler skips (200s) rather than acts on non-terminal statuses.
- To add a new provider: implement `PaymentGateway`, HMAC-verify raw bytes in `ParseWebhook` before any unmarshalling, and wire selection into `ProvidePaymentGateway` in `cmd/api/server.go` gated on `PAYMENT_GATEWAY`.

## Poster POS

`poster.go` — `PosterClient` interface (`CreateIncomingOrder`) and the real HTTP adapter hitting `joinposter.com/api`. `PosterOrder`/`PosterProduct`/`PosterPayment` are the request shapes; note `Price`/`Sum` are in **kopecks** (int64), not the app's usual float euros/hryvnia. `noop_poster.go` — `NoopPosterClient`, used when `POSTER_API_TOKEN` is unset (`ProvidePosterClient` in `cmd/api/server.go` picks between them).

The actual mapping from a confirmed `model.Booking` to a `PosterOrder` (pricing, comment, payment sum) lives in `service/webhook_service.go`'s `buildPosterOrder`, not here — this package only defines the wire format and the HTTP call.

## Misc

`json.go` — a one-line wrapper (`jsonUnmarshal`) around `encoding/json.Unmarshal`, kept as an indirection point in case the stub gateway ever needs strict-mode decoding.