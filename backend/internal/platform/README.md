# internal/platform

Small, dependency-injected infrastructure primitives shared across the app. Nothing here knows about bookings, slots, or any domain concept — that's what makes it "platform."

## Files

- **`db.go`** — `NewDB(databaseURL, debug)` opens the GORM/Postgres connection pool with fixed pool settings (`MaxOpenConns: 25`, `MaxIdleConns: 5`, `ConnMaxLifetime: 30m`) and `PrepareStmt: true`, `SkipDefaultTransaction: true` for perf. Caller (`cmd/api/main.go`) is responsible for calling `AutoMigrate`.
- **`tx.go`** — the transaction pattern used by every repository:
  - `platform.WithTx(ctx, db, fn)` opens a GORM transaction and stashes the `*gorm.DB` on a derived context.
  - `platform.DBFromContext(ctx, defaultDB)` returns that stashed transaction if present, otherwise falls back to `defaultDB.WithContext(ctx)`.
  - Every repository has a private `tx(ctx)` helper that just calls `DBFromContext` — this is what lets the *same* repository method work standalone or as part of a larger multi-repository transaction (e.g. booking creation locks a slot row and inserts a booking in one transaction). When adding a repository, copy this `tx(ctx)` helper rather than accepting a `*gorm.DB` parameter.
- **`clock.go`** — `Clock` interface (`Now() time.Time`) wrapping `time.Now().UTC()`. Injected into services (`booking_service.go`, `checkout_service.go`, `promocode_service.go`) instead of calling `time.Now()` directly, so time-dependent logic (hold expiry, promo `createdAt`) is fakeable in tests even though none exist yet.
- **`logger.go`** — `NewLogger(debug)` returns a JSON `slog.Logger` to stdout; `debug` toggles `Debug` vs `Info` level.

## Convention

If you need "the current time" or "the current db handle" anywhere in `service`/`repository`, thread it through via `Clock`/`DBFromContext` rather than calling `time.Now()` or holding a bare `*gorm.DB` — that's the whole point of this package.