# Promocodes + richer Poster orders

Uncommitted working-tree changes as of 2026-07-05.

## 1. Promocodes (new feature)

Affiliate-style discount codes that admins create, with a global usage cap and percentage discount, applied at booking time.

**New files**
- `internal/model/promocode.go` — `Promocode` GORM model (`code` PK, `discountPercent`, `maxUses`, `timesUsed`, `active`, `createdBy`, `createdAt`).
- `internal/repository/promocode_repository.go` — `PromocodeRepository`: `Create`, `FindByCode`, `FindAll(createdBy?)`, and `IncrementUsage` (atomic `UPDATE ... WHERE times_used < max_uses`, returns whether a row was actually bumped).
- `internal/service/promocode_service.go` — `PromocodeService`: `Create`, `List`, `Validate`. Codes are normalized (upper-cased, trimmed) via `NormalizeCode`. New typed errors: `ErrPromoNotFound`, `ErrPromoInactive`, `ErrPromoExhausted`, `ErrPromoAlreadyExists`.
- `migrations/0005_promocodes.sql` — creates `promocodes` table + index on `created_by`; adds `promo_code`, `discount_percent`, `discount_amount` columns to `bookings` (a frozen snapshot, so it survives later edits/deletes of the promo) plus a partial index on `bookings.promo_code`.

**Wiring**
- `model.Promocode` added to auto-migrate list (`cmd/api/main.go`).
- `repository.NewPromocodeRepository` / `service.NewPromocodeService` added to the wire provider set (`cmd/api/wire.go`, regenerated `wire_gen.go`).
- New admin routes (`cmd/api/server.go`): `POST /admin/promocodes`, `GET /admin/promocodes`.

**Admin API** (`internal/handler/admin_handler.go`)
- `CreatePromocode` — validates `code` (required), `discountPercent` (0–100), `maxUses` (≥1); maps `ErrPromoAlreadyExists` → 409 `PROMO_ALREADY_EXISTS`, `ErrInvalidInput` → 400.
- `ListPromocodes` — optional `?createdBy=<uuid>` filter.
- New DTOs in `internal/model/dto.go`: `AdminCreatePromocodeRequest`, `AdminPromocodeResponse`, `AdminPromocodeListResponse`.
- New error codes in `pkg/httpx/response.go`: `PROMO_NOT_FOUND`, `PROMO_INACTIVE`, `PROMO_EXHAUSTED`, `PROMO_ALREADY_EXISTS`.

**Booking flow** (`internal/service/booking_service.go`, `internal/handler/booking_handler.go`)
- `CreateBookingRequest.promoCode` (optional) → `CreateBookingInput.PromoCode`.
- On create, if a promo code is present it's validated (`PromocodeService.Validate`), the discount is applied to the computed total via `PricingService.ApplyDiscount`, and the code/percent/amount are frozen onto the `Booking` row (`PromoCode`, `DiscountPercent`, `DiscountAmount`).
- `PricingService.ApplyDiscount(total, discountPercent)` (new, `internal/service/pricing_service.go`) — rounds to 2dp, clamps 0–100%.
- Promo validation errors surface through `mapBookingError` as 422s (`PROMO_NOT_FOUND` / `PROMO_INACTIVE` / `PROMO_EXHAUSTED`).
- `CreateBookingResponse` now echoes `promoCode` / `discountPercent` / `discountAmount`.
- Redemption is only counted once payment is confirmed: `webhook_service.go`'s `Handle` calls `PromocodeRepository.IncrementUsage` after marking a booking confirmed (logs a warning, doesn't fail the webhook, if the cap was hit between hold and confirmation — the discount already given is honored).

**Admin booking list**
- `AdminBookingListEntry` and `admin_service.go`'s `toAdminEntry` now include `promoCode` / `discountPercent` / `discountAmount`.

## 2. Richer Poster orders (`internal/provider/poster.go`, `internal/service/webhook_service.go`)

`createPosterOrder` was a bare-bones stub; it's replaced with `buildPosterOrder`, which sends Poster the full picture of a confirmed booking:
- Customer `first_name` / `last_name` / `email` (previously only phone).
- Per-line prices in kopecks (`PosterProduct.Price`), derived from route pricing:
  - one line for all board sizes combined, using a **quantity-weighted average unit price** across big/medium/small (Poster only has one "boards" product line),
  - a separate line for children if any.
- A `payment` block (`type: 1` = prepaid) whose `sum` is the **effective amount actually charged** (post price-override, post-discount) — not the list price.
- A human-readable `comment`: date/time/route, plus `Промокод <code> (-N%)` when a promo was applied.
- `PosterOrder` / `PosterProduct` gained the new optional fields; `PosterPayment` is a new type.

## 3. Other

- `pricing_service.go` gained a shared `round2` helper used by `ApplyDiscount`.
- `migrations/0002_seed_slots.sql` deleted (no longer needed — being tracked as a deletion, not investigated further here).

## Not yet done / to double check
- No corresponding tests added for promocode service/repo or the new Poster order builder.
- `internal/repository/promocode_repository.go`, `internal/service/promocode_service.go`, `internal/model/promocode.go`, and `migrations/0005_promocodes.sql` are currently **untracked** (`git add` needed before committing).
