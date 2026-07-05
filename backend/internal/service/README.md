# internal/service

Business logic. Each service is an interface + a struct implementation, taking its dependencies (repositories, other services, `platform.Clock`) via constructor injection — see `cmd/api/wire.go`/`wire_gen.go` for the assembled graph.

## Files

| File | Service | Responsibility |
|---|---|---|
| `booking_service.go` | `BookingService` | The core orchestrator — `Create` implements the full booking-creation flow (kill-switch → date-block → idempotency → locked capacity check → promo → insert). Declares most of the domain's sentinel errors (`ErrSlotTaken`, `ErrBookingsDisabled`, etc.) |
| `pricing_service.go` | `PricingService` | Route price table (`routePrices`, the single source of truth for per-boat-class prices), `ComputeTotal`, `EffectiveAmount` (override-wins-over-total), `ApplyDiscount` (percentage discount, 2dp-rounded, clamped 0–100) |
| `promocode_service.go` | `PromocodeService` | Admin create/list, and `Validate` (exists + active + under usage cap) — the single validation path shared by booking creation and the user-facing preview endpoint. `NormalizeCode` (upper-case + trim) is exported for reuse |
| `availability_service.go` | `AvailabilityService` | Read-only projections for the calendar/slot UI. Fans out independent reads (settings, slots, date-blocks, booked sums) concurrently via `errgroup` — see the perf notes in `GetMonth`'s doc comment before changing its query shape |
| `checkout_service.go` | `CheckoutService` | Builds the gateway `CreateSessionRequest` (itemised or single discounted line, per `PriceOverride`) and persists the returned session ID |
| `webhook_service.go` | `WebhookService` | Applies an already-verified payment event: flips booking status, increments promo usage on confirm (best-effort, logs rather than fails), fires the confirmation email async, and builds/pushes the Poster order (`buildPosterOrder`) |
| `email_service.go` | `EmailService` | Fire-and-forget confirmation email; must never block the webhook path for long |
| `auth_service.go` | `AuthService` | Wraps the Supabase auth client with a 5-minute in-process cache (`go-cache`), keyed by raw JWT or user ID |
| `admin_service.go` | `AdminService` | All `/admin/*` business logic — slot/date blocking, price overrides, cancellations, moves, kill-switch, and `toAdminEntry` (the `Booking` → `AdminBookingListEntry` projection) |

## Conventions

- **Sentinel errors**: every service that can fail in a caller-distinguishable way declares `var Err... = errors.New("SOME_CODE")` near its interface (see the block at the top of `booking_service.go` or `promocode_service.go`). Handlers `errors.Is`-switch on these — never return ad hoc error strings that a handler would need to string-match.
- **Clock injection**: anything that reads "now" for business logic (hold expiry, promo `createdAt`) takes a `platform.Clock`, not `time.Now()` directly.
- **Cross-service dependencies flow one way**: `booking_service` depends on `pricing_service` and `promocode_service`; `checkout_service` depends on `booking_service`; `webhook_service` depends on `pricing_service`. Don't introduce a cycle back the other way — if two services both need shared logic, it usually belongs in a third service (like `pricing_service`) both can depend on.
- **Snapshot-then-freeze pattern**: when a service applies something mutable (a price override, a promo discount) to a booking, it resolves the value once and writes it onto the `Booking` row rather than storing a foreign key to re-resolve later. Follow this for any new "apply admin-controlled X to a booking" feature.