# Frontend-facing API changes — promocodes

Everything below is additive and optional — no existing request/response field was removed or renamed, so no existing frontend code should break by ignoring this doc. The only backend change with zero frontend surface is the richer Poster POS order (server-to-Poster only; not documented here).

All paths below are relative to the API's `/api` mount (e.g. `POST /bookings` is actually `POST /api/bookings`).

## 1. Applying a promo code at booking time

`POST /bookings` gains an optional field:

```jsonc
{
  "date": "2026-07-10",
  "time": "10:00",
  "routeName": "Desna",
  "quantities": { "big": 1, "medium": 0, "small": 0, "child": 0 },
  "contact": { "firstName": "...", "lastName": "...", "phone": "..." },
  "promoCode": "SUMMER10"   // optional — omit or send "" for no promo
}
```

- Case-insensitive and trimmed server-side (`SUMMER10` == `summer10` == ` summer10 `) — send whatever the user typed as-is, no need to normalize client-side.
- If present, it's validated and the discount is applied **before** the total is computed/frozen. There is no separate "preview discount" endpoint — validation happens as part of booking creation.

**`CreateBookingResponse` gains 3 optional fields** (all `null`/absent if no promo was applied):

```jsonc
{
  "bookingId": "...",
  "totalAmount": 900.0,          // already net of discount
  "expiresAt": "...",
  "promoCode": "SUMMER10",
  "discountPercent": 10,
  "discountAmount": 100.0
}
```

`totalAmount` was already the field driving checkout amount — no change needed there, it's just smaller when a promo applied. Use `promoCode`/`discountPercent`/`discountAmount` purely for display (e.g. "10% off (SUMMER10): −₴100").

### New error cases on `POST /bookings`

All returned as HTTP 422 with the existing `{ "code": "...", ... }` error body shape, same as other booking errors (`SLOT_BLOCKED`, `DATE_BLOCKED`, etc.):

| Code | Meaning | Suggested UI handling |
|---|---|---|
| `PROMO_NOT_FOUND` | Code doesn't exist | "This promo code isn't valid." |
| `PROMO_INACTIVE` | Code exists but was deactivated by an admin | "This promo code is no longer active." |
| `PROMO_EXHAUSTED` | Code's global usage cap has been reached | "This promo code has reached its usage limit." |

These come back on the *booking creation* request itself (not a separate validation call) — if the rest of the booking is valid but the promo isn't, the whole booking fails with one of these codes and no booking is created. So a "check code" UX would need to just attempt the booking (or omit the promo and retry) rather than validate independently.

## 2. Previewing a promo code before booking

`GET /promocodes/:code` — authenticated (any logged-in user, not admin-only). Lets the UI show the discount before the user commits to creating a booking (e.g. in the cart, next to a "have a promo code?" field).

Response (`200`):
```jsonc
{ "discountPercent": 10 }
```

Deliberately minimal — only `discountPercent` is returned, never `maxUses`/`timesUsed`/`createdBy` (those stay admin-only via `/admin/promocodes`). Same validation as booking creation, so the errors match:

| Status | Code               | Meaning                        |
|--------|--------------------|---------------------------------|
| 422    | `PROMO_NOT_FOUND`  | Code doesn't exist              |
| 422    | `PROMO_INACTIVE`   | Code deactivated by an admin    |
| 422    | `PROMO_EXHAUSTED`  | Usage cap reached               |

Note this is a point-in-time check — nothing is reserved by calling it, so a code could still become exhausted between this call and the actual `POST /bookings` (in which case the booking call returns `PROMO_EXHAUSTED` itself; handle both places).

## 3. Admin: managing promocodes

Two new endpoints under `/admin/*` (same bearer-token + admin-role auth as other admin routes).

### `POST /admin/promocodes` — create a code

Request:
```jsonc
{
  "code": "SUMMER10",           // required, 1-64 chars, will be upper-cased/trimmed server-side
  "discountPercent": 10,        // required, 0-100
  "maxUses": 100                // required, >= 1 — global cap across all customers
}
```

Response (`201`):
```jsonc
{
  "code": "SUMMER10",
  "discountPercent": 10,
  "maxUses": 100,
  "timesUsed": 0,
  "active": true,
  "createdBy": "<admin-uuid>",
  "createdAt": "2026-07-05T12:00:00Z"
}
```

Errors:
- `400 INVALID_INPUT` — bad payload.
- `409 PROMO_ALREADY_EXISTS` — that code (case-insensitive) already exists; surface as "Code already in use."
- `503 SERVICE_UNAVAILABLE` — unexpected server error.

There is currently **no update/deactivate/delete endpoint** — promocodes can only be created and listed. If the admin UI needs to disable a code, that's not supported yet; flag this as a gap rather than assuming a PATCH exists.

### `GET /admin/promocodes` — list codes

Optional query param `?createdBy=<admin-uuid>` filters to codes created by one admin (useful for an affiliate view — "codes I created"). Omit it to get all codes.

Response:
```jsonc
{
  "promocodes": [
    { "code": "SUMMER10", "discountPercent": 10, "maxUses": 100, "timesUsed": 42, "active": true, "createdBy": "...", "createdAt": "..." },
    ...
  ]
}
```

Sorted newest-created first. `timesUsed` only increments once a booking using that code is **confirmed** (paid), not at hold-creation time — so a code can show pending bookings against it that haven't been counted yet. There's no live "reserved but unconfirmed" count exposed.

## 4. Admin: booking list/history now show promo info

`AdminBookingListEntry` (returned by `GET /admin/slots/:date/:time/:route/bookings` and `GET /admin/bookings`) gains the same 3 optional fields as the booking-creation response:

```jsonc
{
  "id": "...",
  ...
  "totalAmount": 900.0,
  "effectiveAmount": 900.0,
  "promoCode": "SUMMER10",       // null if no promo was used
  "discountPercent": 10,
  "discountAmount": 100.0,
  ...
}
```

These are a frozen snapshot taken at booking time — editing or later deleting the promocode does **not** change what's shown on past bookings.

## Summary of touched endpoints

| Endpoint | Change |
|----------|--------|
| `POST /bookings` | request: `+promoCode` (optional); response: `+promoCode`, `+discountPercent`, `+discountAmount`; new 422 error codes |
| `GET /promocodes/:code` | **new** |
| `POST /admin/promocodes` | **new** |
| `GET /admin/promocodes` | **new** |
| `GET /admin/slots/:date/:time/:route/bookings` | entries: `+promoCode`, `+discountPercent`, `+discountAmount` |
| `GET /admin/bookings` | entries: `+promoCode`, `+discountPercent`, `+discountAmount` |