# Promocodes — frontend integration

Adds customer-facing promo-code support to the booking wizard: capture a code
from the URL, apply it at booking, and show the discount before and after
purchase. A code may be redeemed as many times as the backend allows — the
client never blocks a repeat redemption.

Backend contract this implements: `POST /bookings` `+promoCode` (optional) with
`+promoCode/+discountPercent/+discountAmount` on the response, new
`GET /promocodes/:code` preview endpoint, and new 422 codes
(`PROMO_NOT_FOUND` / `PROMO_INACTIVE` / `PROMO_EXHAUSTED`).

## User-facing flow

1. **Capture** — landing on any URL with `?promo=CODE` stashes the (trimmed,
   upper-cased) code. Previously redeemed codes are captured again.
2. **Preview** — on the details step the code is validated via
   `GET /promocodes/:code`; the order summary shows a discount row, the running
   total net of the discount, and a status line (applied / checking / error).
3. **Apply** — `POST /bookings` sends `promoCode`; the authoritative
   `totalAmount` (already net) drives checkout.
4. **Confirm** — the success page shows the discount row and *keeps* the code
   in the store, so "make another booking" carries the same discount.

## Files changed

### New

| File | Purpose |
|------|---------|
| `src/features/booking/promo.ts` | localStorage helpers: normalize + per-booking discount receipt (`savePromoReceipt` / `getPromoReceipt`). Deliberately has no "already redeemed" list. |
| `src/features/booking/components/client/PromoCapture.tsx` | Reads `?promo=` app-wide (Suspense-wrapped `useSearchParams`) and stores the code. |
| `src/app/api/promocodes/[code]/route.ts` | BFF proxy for the preview endpoint; forwards `PROMO_*` error codes. |

### Modified

| File | Change |
|------|--------|
| `src/app/layout.tsx` | Mounts `<PromoCapture />` so `?promo=` works on any page (including before login — the code persists and applies after auth). |
| `src/features/booking/store/bookingStore.ts` | `promoCode` state + `setPromoCode`; persisted; **independent** of the route/date/time cascade and **preserved across `reset()`** so a repeat booking reuses it. Only the × control or a sign-out drops it. |
| `src/shared/lib/api/types.ts` | `promoCode?` on `CreateBookingBody`; `promoCode/discountPercent/discountAmount` on `CreateBookingResponse` and `BookingDetail`; new `PromoPreviewResponse`. |
| `src/features/booking/schema/booking.schema.ts` | `promoCode: z.string().max(64).optional()` so the BFF forwards it (Zod strips unknown keys). |
| `src/features/booking/hooks/index.ts` | `usePromoPreview(code)` query (`retry: false`, surfaces `PROMO_*` errors). |
| `src/app/api/bookings/route.ts` | Forwards `PROMO_NOT_FOUND / PROMO_INACTIVE / PROMO_EXHAUSTED` 422s (reads `code` or `message`). |
| `src/app/book/details/page.tsx` | Preview + discount row + status line + discounted total; sends `promoCode` (only when preview didn't fail); saves the receipt from the response; maps promo errors in the error banner. |
| `src/features/booking/components/client/ProcessingScreen.tsx` | Confirmation card shows the discount row (from receipt, falling back to `BookingDetail` fields); the code stays in the store for the next booking. |
| `src/features/booking/messages.ts` | Ukrainian promo strings + promo error messages. |
| `src/shared/lib/idempotency.ts` | Folds `promoCode` into the fingerprint so applying/changing a code re-mints the key instead of reusing an earlier discount-free hold. |

## Design decisions

- **Receipt in localStorage.** `GET /bookings/by-session/:id` is not documented
  to echo promo fields, so the applied discount from the creation response is
  snapshotted keyed by `bookingId` and read back on the success page (with a
  fallback to any promo fields the status response does include).
- **Only send a validated code.** If the preview call failed, the code is
  dropped from the booking request so a stale/invalid promo can't fail an
  otherwise-valid booking (backend stays authoritative — a race that exhausts
  the code between preview and booking is still handled via the 422 mapping).
- **No client-side reuse limit.** The backend is the only thing that can
  exhaust a code (`maxUses = 0` means never). The client used to keep a
  `harbour-wave-promo-used` list and refuse a second redemption on the same
  device — that key is gone, and any stale copy in a customer's localStorage is
  simply ignored.
- **Promo is cascade-independent and survives a reset.** Changing route/date/
  time does not clear the captured code, and neither does starting a new booking
  — a reusable code shouldn't force the customer back through their `?promo=`
  link.

## Out of scope

- Admin promocode management (create/list) — belongs to the `admin-ui` package.
- Backend has **no** deactivate/update/delete promocode endpoint yet (flagged in
  the API doc), so there's nothing to build against here.

## Verification

- `pnpm type-check` — clean.
- `pnpm build` — succeeds; `/api/promocodes/[code]` registered.
- `pnpm lint` — not run: ESLint is unconfigured in this repo and `next lint`
  drops into an interactive setup prompt (pre-existing, unrelated).
- Not driven end-to-end (needs backend + a live promo code + payment redirect).
