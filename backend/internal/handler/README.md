# internal/handler

Gin HTTP handlers. This is the only layer that knows about HTTP — status codes, JSON binding, headers, gin.Context. It talks exclusively to `internal/service` interfaces and never touches GORM or the database directly.

## Files

| File | Routes | Notes |
|---|---|---|
| `auth_middleware.go` | — | `AuthMiddleware` (validates Supabase JWT, sets `model.AuthUser` in context), `AdminMiddleware` (checks `admins` table; must run after `AuthMiddleware`), `GetUserFromContext`/`GetUserIDFromContext` helpers |
| `availability_handler.go` | `GET /status`, `GET /availability/:month/:route`, `GET /slots/:date/:route` | Read-only, auth required |
| `booking_handler.go` | `POST /bookings`, `GET /bookings/by-session/:sessionId`, `GET /bookings` | Owns `mapBookingError`, the central switch from service sentinel errors to HTTP codes for the booking domain |
| `promocode_handler.go` | `GET /promocodes/:code` | User-facing preview — deliberately returns only `discountPercent`, reuses `PromocodeService.Validate` (the same call the booking flow makes) |
| `checkout_handler.go` | `POST /checkout` | Creates a hosted payment session for a pending booking |
| `webhook_handler.go` | `POST /payment/webhook` | Public, unauthenticated. Reads the raw body via `io.ReadAll` and passes it *unparsed* to `gateway.ParseWebhook` — never decode JSON before signature verification |
| `admin_handler.go` | everything under `/admin/*` | Largest file; one handler per admin action (slot/date blocking, price override, cancellations, moves, promocodes, kill-switch) |

## Conventions

- **Error mapping**: each handler with multiple failure modes has (or uses) a `mapXError(c, err, log)` function that `errors.Is`-switches on service-layer sentinel errors and picks the HTTP status + `httpx.Code*` constant. Follow this pattern instead of inlining status logic in the handler body — see `mapBookingError` in `booking_handler.go` for the canonical example.
- **Swagger annotations**: every public handler method has a `@Summary`/`@Router`/etc. comment block consumed by `swag init` to regenerate `docs/`. Keep these in sync when adding/changing routes — the frontend-facing contract lives partly in these comments.
- **Auth**: `GetUserFromContext(c)` / `GetUserIDFromContext(c)` are the only sanctioned way to read the authenticated user; never re-parse the Authorization header inside a handler.
- **Admin auth**: admin handlers assume `AdminMiddleware` already ran — they don't re-check admin status themselves, only extract the caller's ID (e.g. via `GetUserIDFromContext`) for audit fields like `createdBy`/`blockedBy`.
- Route params that need cross-cutting parsing (e.g. `route` validated against `service.IsValidRoute`) go through small shared helpers like `parseRoute` in `availability_handler.go` rather than being duplicated per handler.
