# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Harbour & Wave backend: Go + Gin + GORM + PostgreSQL booking system. Authenticates Supabase JWTs, enforces slot capacity with pessimistic locks, integrates with a hosted payment gateway (LiqPay) and Poster POS, and exposes admin controls (price overrides, slot/date blocks, kill-switch, cancellations, promocodes).

## Commands

```bash
make run       # go run ./cmd/api
make build     # static binary -> bin/api
make test      # go test ./...
make fmt       # gofmt -w .
make vet       # go vet ./...
make migrate   # psql "$DATABASE_URL" -f migrations/0001_init.sql (requires DATABASE_URL)
```

Run a single test: `go test ./internal/service/... -run TestName -v` (there are currently no `*_test.go` files in the repo — write to the corresponding package alongside the code under test).

There is no linter config beyond `go vet`; `gofmt` is the formatting standard.

### Wire (DI) and migrations

- `cmd/api/wire.go` (build tag `wireinject`) declares the provider set; `cmd/api/wire_gen.go` is the generated real implementation used at build time. After editing `wire.go`'s provider set, regenerate with `go run -mod=mod github.com/google/wire/cmd/wire ./cmd/api` — or hand-edit `wire_gen.go` to match if `wire` isn't available, since both must stay in sync.
- New migrations live in `migrations/NNNN_description.sql`, numbered sequentially. `AutoMigrate` also runs at boot from the model list in `cmd/api/main.go` (`db.AutoMigrate(&model.X{}, ...)`) — new GORM models must be added there too. Prefer writing the `.sql` migration for anything AutoMigrate can't express (indexes with conditions, `ALTER COLUMN`, backfills, constraints).
- Swagger docs (`docs/docs.go`, `docs/swagger.json/yaml`) are swaggo-generated from `@Summary`/`@Router` annotations on handler methods — regenerate with `swag init -g cmd/api/main.go -o docs` after adding/changing annotated endpoints.

## Architecture

### Layering

```
cmd/api/          wiring (wire.go/wire_gen.go), router (server.go), main() boot/shutdown
internal/handler/  gin handlers + auth/admin middleware — HTTP concerns only, maps typed
                   service errors to httpx error codes (see mapBookingError-style switches)
internal/service/  business logic; interfaces + structs (e.g. BookingService/bookingService)
internal/repository/ GORM persistence, one file per aggregate
internal/model/    domain models (GORM tags) + request/response DTOs (dto.go)
internal/platform/ db, tx helper, logger, clock (clock is injected so time is fakeable in tests)
internal/provider/  PaymentGateway interface + LiqPay/stub adapters, Poster POS client
pkg/httpx/          response helpers + centralized error code constants
```

Each layer only talks to the one below via interfaces defined in `internal/service` / `internal/repository`; handlers never touch GORM directly.

### Transactions

`internal/platform/tx.go`: `platform.WithTx(ctx, db, fn)` runs `fn` inside a GORM transaction, stashing the `*gorm.DB` in the context. Repositories call `r.tx(ctx)` (a per-repo helper wrapping `platform.DBFromContext(ctx, r.db)`) so the same repository code transparently participates in an ongoing transaction or falls back to the plain db handle. When adding a repository method that must run inside a booking-creation transaction, follow this pattern rather than accepting a `*gorm.DB` directly.

### Auth

- `AuthMiddleware` validates the Supabase bearer token and stashes `*model.AuthUser` in the gin context (`handler.GetUserFromContext` / `GetUserIDFromContext`).
- `AdminMiddleware` (installed after `AuthMiddleware` on `/admin` routes) checks the `admins` DB table — admin status is not derived from the JWT alone.
- Supabase token lookups are cached in-process for 5 min (`patrickmn/go-cache`) to avoid hammering Supabase on every request.

### Booking flow (the core invariant to preserve)

`POST /bookings` (requires `X-Idempotency-Key`) creates a `pending` booking holding capacity for 15 minutes (`HoldDuration` in `booking_service.go`). Order of checks matters:

1. Field validation (date, time, quantities, contact — phone required).
2. Kill-switch: `system_settings.bookings_enabled` read fresh from DB on every request (never cached) → 503 `BOOKINGS_DISABLED`.
3. Date block → 422 `DATE_BLOCKED`.
4. Idempotency: same `(user_id, key)` returns the existing booking instead of creating a new one.
5. Inside a `platform.WithTx` transaction: `SELECT … FOR UPDATE` on the slot row, check slot-blocked, sum active quantities and check `big`/`medium` capacity **independently** (fleets aren't interchangeable; child seats don't count against capacity but require `big > 0`), then insert.
6. Optional promocode: validated (`PromocodeService.Validate`) and applied (`PricingService.ApplyDiscount`) before the total is frozen onto the booking row — the code, percent, and amount are snapshotted onto `Booking` so later edits/deletion of the promo don't retroactively change historical bookings.

A background goroutine (started in `main.go`) sweeps `pending` bookings whose `expires_at < NOW()` every minute, releasing held capacity.

`POST /checkout` creates a hosted LiqPay session and stores `payment_session_id`. `POST /payment/webhook` verifies HMAC against the **raw request body bytes** (`ParseWebhook` must receive raw bytes — re-encoding via unmarshal/marshal breaks the signature) and transitions status to `confirmed`/`failed`/`expired`. On `confirmed`: the Ukrainian confirmation email fires asynchronously (so it never delays the webhook's 200), any promocode's usage counter is incremented (`PromocodeRepository.IncrementUsage`, an atomic conditional `UPDATE` — logs a warning rather than failing if the cap was hit between hold and confirmation, honoring the already-given discount), and a Poster POS incoming order is created (`webhook_service.go`'s `buildPosterOrder`) with per-line kopeck pricing and the post-discount/post-override amount as the payment sum.

### Adding a payment provider

Implement `provider.PaymentGateway` (`internal/provider/payment_gateway.go`); HMAC-verify raw bytes inside `ParseWebhook` before any unmarshalling. Wire it into the gateway constructor in `cmd/api/main.go` and gate selection on the `PAYMENT_GATEWAY` env var.

### Error handling convention

Services return sentinel errors (`errors.New("SOME_CODE")`, e.g. `ErrPromoNotFound`, `ErrSlotBlocked`) declared near the service that raises them. Handlers `errors.Is`-switch on these in a `mapXError` function to pick the HTTP status and `httpx.Code*` constant (defined centrally in `pkg/httpx/response.go`). Follow this pattern for new error cases rather than returning raw strings or checking messages.

### Concurrency safety invariants

- One slot, one row, one `FOR UPDATE` lock per booking insert — don't add booking-capacity logic outside that transaction.
- Big and medium capacity are tracked and enforced independently.
- Idempotency is enforced by a unique index on `(user_id, idempotency_key)`, not just application logic.

## Environment

See `.env.example`. Required: `DATABASE_URL`, `SUPABASE_URL`, `SUPABASE_PROJECT_REFERENCE`, `SUPABASE_ANON_KEY`, `SUPABASE_SERVICE_ROLE_KEY`. Commonly set: `FRONTEND_URL` (CORS), `BACKEND_URL`, `PORT` (default 8080), `GIN_MODE`, `PAYMENT_GATEWAY` (default `stub`), `EMAIL_FROM`/`SMTP_*` (real SMTP only if all set, otherwise dry-run logging).
