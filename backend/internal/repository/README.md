# internal/repository

GORM persistence, one file per aggregate. Repositories are the only place `*gorm.DB` is touched directly — services depend on the interfaces here, never on GORM types.

## Files

| File | Interface | Notes |
|---|---|---|
| `booking_repository.go` | `BookingRepository` | The biggest one. Owns the capacity-counting queries (`SumActiveQuantitiesForSlot/ForDate/ForRange` — "active" = confirmed, or pending with `expires_at > NOW()`), idempotency lookup, admin history filtering (`FindAllForAdmin`, capped at `maxBookingHistoryLimit`), and the expiry sweep (`ExpirePending`) |
| `slot_repository.go` | `SlotRepository` | `LockForUpdate` issues `SELECT ... FOR UPDATE` — this is the row lock the whole booking-capacity invariant depends on. `Upsert` uses an `ON CONFLICT` clause keyed on `(date, time, route_name)` |
| `date_block_repository.go` | `DateBlockRepository` | Whole-day blocks; `FindManyInRange` backs the monthly availability query |
| `system_repository.go` | `SystemRepository` | Singleton `system_settings` row; `Get` lazily creates the row with `bookings_enabled: true` if it doesn't exist yet |
| `admin_repository.go` | `AdminRepository` | Single method, `IsAdmin` — existence check against the `admins` table |
| `promocode_repository.go` | `PromocodeRepository` | `IncrementUsage` is the one non-trivial method: an atomic `UPDATE ... WHERE times_used < max_uses`, returning whether a row was actually bumped (`RowsAffected == 1`) so the caller can tell "incremented" from "cap already reached" without a race |

## Conventions

- **`ErrNotFound`** (declared in `booking_repository.go`) is the shared sentinel every repository returns instead of `gorm.ErrRecordNotFound` — always translate with `errors.Is(err, gorm.ErrRecordNotFound)` at the repository boundary so callers in `service` never need to import `gorm`.
- **Transaction participation**: every repo has a private `tx(ctx) *gorm.DB` helper calling `platform.DBFromContext(ctx, r.db)`. Use it for every query/exec in new methods — this is what makes a repository method work both standalone and as part of a `platform.WithTx` transaction (see `internal/platform/README.md`).
- **Row locks live in the repository, not the service**: `SlotRepository.LockForUpdate` is the only place `clause.Locking{Strength: "UPDATE"}` appears. If you need a new locked read, add a method here rather than building the clause in `service`.
- Small conditional updates that could race (block/unblock, cancel/uncancel, `IncrementUsage`) check `res.RowsAffected` and surface `ErrNotFound` (or a bool) rather than assuming success.