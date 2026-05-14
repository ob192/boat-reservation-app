# Harbour & Wave — Backend

Go + Gin + GORM + PostgreSQL backend implementing the Harbour & Wave booking system. Authenticates Supabase JWTs, enforces slot capacity with pessimistic locks, integrates with a hosted payment gateway, and exposes admin controls (price overrides, slot/date blocks, kill-switch, cancellations).

## Stack

- **Go 1.22**, [Gin](https://github.com/gin-gonic/gin) HTTP framework
- **GORM** + PostgreSQL
- **Supabase** auth via [`supabase-community/auth-go`](https://github.com/supabase-community/auth-go)
- **In-process cache** ([`patrickmn/go-cache`](https://github.com/patrickmn/go-cache)) — 5 min TTL on Supabase token lookups
- Provider-agnostic `PaymentGateway` interface (stub adapter shipped; LiqPay/Stripe/etc. plug in)

## Quickstart

```bash
# 1. Clone & enter
cd harbour-wave-backend

# 2. Configure
cp .env.example .env
# edit .env — at minimum DATABASE_URL, SUPABASE_*, FRONTEND_URL must be set

# 3. Bring up Postgres (any way you like)
#    e.g. docker run -d --name hw-pg -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:16

# 4. Apply migrations (optional — AutoMigrate runs at boot too)
psql "$DATABASE_URL" -f migrations/0001_init.sql
psql "$DATABASE_URL" -f migrations/0002_seed_slots.sql

# 5. Build & run
go mod tidy
go run ./cmd/api
```

The server listens on `:8080` by default. `GET /healthz` returns `{"status":"ok"}` with no auth required.

## Project layout

```
cmd/api/main.go                       wiring, router, graceful shutdown
internal/config/                      env-var loader
internal/handler/                     gin handlers + auth/admin middleware
internal/service/                     business logic (booking, checkout, admin, …)
internal/repository/                  GORM persistence
internal/model/                       domain models + DTOs
internal/platform/                    db, tx, logger, clock
internal/provider/                    PaymentGateway interface + stub adapter
pkg/httpx/                            response helpers + error code constants
migrations/                           SQL DDL + slot-seed example
```

## Environment variables

See [`.env.example`](./.env.example). Required:

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | Postgres DSN |
| `SUPABASE_URL` | `https://<ref>.supabase.co` |
| `SUPABASE_PROJECT_REFERENCE` | The `<ref>` portion |
| `SUPABASE_ANON_KEY` | For user token validation |
| `SUPABASE_SERVICE_ROLE_KEY` | For admin user lookups |

Optional but commonly set:

| Variable | Default | Purpose |
|---|---|---|
| `FRONTEND_URL` | `http://localhost:3000` | CORS allow-list |
| `BACKEND_URL` | `http://localhost:8080` | Used by stub gateway URLs |
| `PORT` | `8080` | Listen port |
| `GIN_MODE` | `debug` | `debug` / `release` |
| `PAYMENT_GATEWAY` | `stub` | Adapter selector |
| `EMAIL_FROM`, `SMTP_*` | unset | Real SMTP if all set, otherwise dry-run |

## Routes

```
GET  /healthz                                      -- public
POST /payment/webhook                              -- public, HMAC-verified

GET  /availability/:month                          -- auth
GET  /slots/:date                                  -- auth
POST /bookings                                     -- auth, X-Idempotency-Key required
POST /checkout                                     -- auth
GET  /bookings/by-session/:sessionId               -- auth

PATCH  /admin/bookings/:bookingId/price            -- admin
PUT    /admin/slots/:date/:time/block              -- admin
DELETE /admin/slots/:date/:time/block              -- admin
PUT    /admin/dates/:date/block                    -- admin
DELETE /admin/dates/:date/block                    -- admin
PUT    /admin/system/bookings-enabled              -- admin
POST   /admin/bookings/:bookingId/cancel           -- admin
```

Admin role is detected via Supabase JWT `app_metadata.role == "admin"`.

## Booking flow

`POST /bookings` (with `X-Idempotency-Key`) creates a `pending` booking holding capacity for **15 minutes**. Validation order:

1. Field validation (date, time, quantities, contact, **phone required**).
2. Kill-switch: `system_settings.bookings_enabled` → 503 `BOOKINGS_DISABLED`.
3. Date block → 422 `DATE_BLOCKED`.
4. Idempotency: same `(user_id, key)` returns the existing booking.
5. Inside a transaction:
   - `SELECT … FOR UPDATE` on the slot row.
   - Slot blocked → 422 `SLOT_BLOCKED`.
   - Sum active quantities; check `big` and `medium` capacity independently → 409 `SLOT_TAKEN`.
   - Insert.

`POST /checkout` creates a hosted-checkout session with the gateway and stores `payment_session_id`. The gateway's webhook (`POST /payment/webhook`) verifies HMAC against the **raw request body** (never re-encode!) and transitions status to `confirmed` / `failed` / `expired`. On `confirmed` we fire the Ukrainian confirmation email asynchronously so we never delay the webhook 200.

A background goroutine sweeps pending bookings whose `expires_at < NOW()` every minute.

## Adding a real payment provider

1. Implement `provider.PaymentGateway` (see `internal/provider/payment_gateway.go`).
2. **HMAC-verify** the raw bytes in `ParseWebhook` — do not unmarshal first.
3. Wire it into `buildPaymentGateway` in `cmd/api/main.go` and gate it on `PAYMENT_GATEWAY=<your-name>`.

## Admin notes

- Price override accepts `null` to clear, or a positive number with a non-empty `reason` (≥5 chars).
- Slot block / date block are admin-controlled; existing confirmed bookings are unaffected.
- `bookings-enabled = false` is a hard kill-switch — read directly from DB on every booking request to avoid stale-cache leakage.

## Concurrency safety

- One slot, one row, one `FOR UPDATE` lock per booking insert.
- Big and medium capacity are independent (fleets aren't interchangeable).
- Child seats don't count against capacity but require `big > 0`.
- Idempotency key is enforced by a unique index on `(user_id, idempotency_key)`.
