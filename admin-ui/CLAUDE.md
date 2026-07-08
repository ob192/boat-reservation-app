# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Harbour & Wave **admin dashboard** — a Next.js 15 (App Router) SPA for staff to manage the boat-reservation system: browse/edit slots and capacity, block/unblock slots and whole dates, cancel & move bookings, flip a global bookings kill-switch, and view day reports / booking history. It is a thin client over the Go backend (`../backend`, which has its own `CLAUDE.md` documenting the booking invariants). This app holds no business logic — the backend enforces everything; the UI is convenience + UX.

This lives in a monorepo alongside `../backend` (Go API) and `../frontend` (the public customer-facing booking site).

## Commands

Package manager is **pnpm** (`pnpm-lock.yaml`), despite what the README says.

```bash
pnpm install
pnpm dev      # next dev — runs on PORT from .env (3001), not the default 3000
pnpm build    # next build
pnpm start    # next start (prod)
pnpm lint     # next lint (eslint-config-next; .eslintrc.json only extends "next")
```

There is no test setup (no test runner, no `*.test.tsx`). "Verify" here means running `pnpm build` / `pnpm lint` and exercising the flow against a running backend.

## Architecture

### Request path (the important part)

The browser never calls the backend directly. `adminFetch` (`src/lib/api.ts`) issues **relative** `/api/...` requests, and `next.config.ts` `rewrites()` proxies `/api/:path*` → `${BACKEND_URL}/api/:path*`. So:

- `BACKEND_URL` (server-only env) is the real proxy target used by Next's rewrite.
- `NEXT_PUBLIC_API_URL` is declared but effectively unused by `adminFetch` — don't rely on it for routing.
- `adminFetch` pulls the current Supabase session and attaches `Authorization: Bearer <access_token>`; no session → throws `ApiError(401)`. Every backend call must go through `adminFetch` so the token and error handling stay consistent.

### Auth

- `src/lib/supabase.ts` — single browser Supabase client (Google OAuth). Session lives client-side.
- `src/components/AdminGuard.tsx` wraps protected pages: checks for a session (→ `/signin`) and an `admin` role (→ `/forbidden`), and re-checks on `onAuthStateChange`.
- **Gotcha:** `decodeJwtRole` currently `return "admin"` unconditionally (the real `payload.app_metadata.role` read is commented out), so the client-side admin check is effectively disabled. Real authorization is enforced by the backend's `AdminMiddleware`; this guard is UX-only. If you touch auth, restore the real role read rather than assuming the guard protects anything.

### Data layer

TanStack Query v5. One hook file per resource in `src/hooks/` (`useSlots`, `useSlotBookings`, `useDayBookings`, `useBookingHistory`, `useAvailability`, `useSystem`, `useMoveBooking`). Conventions to follow when adding hooks:

- Query keys are arrays like `['slots', date, route]`, `['slot-bookings', ...]`, `['booking-history']`.
- Mutations `invalidateQueries` the affected keys in `onSuccess`. Note the **partial-match** invalidation: `invalidateQueries({ queryKey: ['slots', date] })` clears every `['slots', date, <route>]` entry. A mutation that changes bookings should also invalidate `slot-bookings` and `booking-history` (see `useCancelSlot`).
- Backend response/DTO shapes are hand-mirrored in `src/lib/types.ts` — keep it in sync with the backend's `internal/model/dto.go`.

### Domain conventions

- **Fleets are independent and never interchangeable:** `big`, `medium`, `small` capacities are tracked separately; `child` seats don't consume capacity. This mirrors the backend booking invariant.
- **Routes are backend-owned string IDs** (`Desna`, `Klochkov` in `src/lib/routes.ts`), not localized on the wire. `routeLabel()` maps them to Ukrainian display strings and falls back to the raw ID, so the list can grow backend-side without breaking the UI.
- Slot identity is `(date, time, route)` — most admin endpoints are keyed on that triple.

### UI / styling

- **No Tailwind, no CSS framework.** Styling is a single global stylesheet (`src/app/globals.css`) built around CSS custom properties (`--navy`, `--teal`, `--coral`, `--cream`, `--radius`, `--shadow`, etc.) plus inline `style={{...}}` on components. Reuse the existing CSS variables and class names rather than introducing a styling library.
- The UI is **Ukrainian** (`<html lang="uk">`); user-facing copy is in Ukrainian. Fonts (Playfair Display / DM Sans) are pulled from Google Fonts via `@import` in `globals.css`.
- `src/components/` holds shared primitives (`Modal`, `Drawer`, `Toast`, `StatusBadge`, `CapacityBar`, `ConfirmInline`, `Calendar`, `RouteSelector`); `src/components/slots/` holds slot-specific composites. `Toast` is driven by the `useToast` hook and mounted globally in `providers.tsx`.
- Path alias `@/*` → `src/*`.

### Pages (App Router, `src/app/`)

`/slots` (main: browse by date+route, edit capacity, view/cancel/move bookings), `/blocks` (block slots or whole dates), `/system` (bookings kill-switch), `/history` (booking history), plus `/signin`, `/forbidden`, `/auth/callback`. All authed pages are wrapped in `<AdminGuard>` and share `<Providers>` (QueryClient + Toast + devtools) from the root layout.

## Environment

`.env.local` (see `.env.local.example`): `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY`, `NEXT_PUBLIC_API_URL`, and the server-only `BACKEND_URL` (proxy target) + `PORT`. Run the backend (`cd ../backend && make run`, default `:8080`) alongside `pnpm dev` for a working local stack.