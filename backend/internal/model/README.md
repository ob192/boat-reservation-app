# internal/model

Domain models (GORM-tagged, persisted) and request/response DTOs, in one package so services and handlers share a single vocabulary. There is no behavior here beyond small derived-value methods on the persisted structs — business rules live in `internal/service`.

## Files

| File | Contents |
|---|---|
| `booking.go` | `Booking` (the central aggregate), `BookingStatus` enum, `Quantities`. Methods: `Quantities()`, `EffectiveAmount()` (price override wins over total), `DateFormatted()`, `IsValidBookingStatus`, `BeforeUpdate` GORM hook (stamps `updated_at`) |
| `slot.go` | `Slot` (composite PK `date, time, route_name` — physical fleet capacity + block state), `DateBlock`, `SlotAvailability` (a query projection, not persisted, with `Available*`/`FullyUnavailable` helpers) |
| `promocode.go` | `Promocode` (admin-created discount code; PK is the code itself, stored upper-cased) |
| `system.go` | `SystemSettings` — singleton row (fixed PK `SystemSettingsRowID = 1`) holding the `bookings_enabled` kill-switch |
| `admin.go` | `Admin` — a row in this table grants `/admin/*` access; deleting the row revokes it immediately, no restart needed |
| `auth.go` | `AuthUser` — internal projection of a Supabase user, decoupled from the Supabase SDK's own type |
| `payment.go` | `PaymentStatus` — mirrors `provider.PaymentStatus` as a separate type so `service` doesn't need to import `provider` just for a status enum (avoids a cycle) |
| `dto.go` | All request/response shapes, one block per endpoint, grouped by comment headers (`POST /bookings`, `Admin DTOs`, `Promocode DTOs`, etc.) |

## Conventions

- **Nullable JSON fields** are pointers (`*string`, `*float64`, `*int`) with `omitempty` so "not set" is distinguishable from the zero value in both directions (DB and JSON). `QuantitiesRequest` uses `*int` for the same reason — to tell "missing" from "explicitly zero" during validation.
- **Snapshot fields**: anything an admin can later edit or delete but that must remain historically accurate on a booking (promo code/discount, price override) is copied onto `Booking` at creation/confirmation time rather than joined at read time. When adding a new "apply X to a booking" feature, prefer this snapshot pattern over a live foreign-key join.
- **`TableName()`** is defined explicitly on every persisted model so a future Go struct rename doesn't silently change the SQL table name GORM targets.
- DTOs are grouped by the endpoint that uses them (comment banners in `dto.go`), not by domain concept — when adding a new endpoint, add its request/response structs next to the others for that resource, and add a new `// ====` banner if it's a new resource.
