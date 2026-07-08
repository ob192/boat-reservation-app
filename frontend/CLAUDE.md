# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Scope

This is the **frontend** package (`harbour-wave`) of the `boat-reservation-app` monorepo. Siblings are `../backend` (Go API — has its own CLAUDE.md) and `../admin-ui`. This app is the **presentation layer only**: no business logic, persistence, slot locking, pricing authority, Stripe, or webhooks live here. Those are the backend's job.

Package manager is **pnpm** (see `pnpm-lock.yaml`) though `package.json` scripts reference `npm`.

## Commands

```bash
pnpm dev          # next dev — http://localhost:3000
pnpm build        # next build
pnpm start        # serve production build
pnpm lint         # next lint (eslint 9 + eslint-config-next)
pnpm type-check   # tsc --noEmit (strict mode)
```

There is **no test suite** — do not assume `pnpm test` exists. Verify changes with `pnpm type-check && pnpm lint` and by driving the flow in `pnpm dev`.

## Architecture

Next.js 15 App Router. Two layers on the client, one thin proxy layer on the server:

1. **Client pages/components** (`'use client'`) — the booking wizard and auth.
2. **BFF route handlers** (`src/app/api/**/route.ts`) — dumb proxies. Each reads the incoming `Authorization: Bearer` header, forwards to `${env.BACKEND_URL}/api/...`, and passes the response through. They add zero logic and MUST NOT be trusted for auth (the backend re-validates every JWT). Pattern: return 401 if no Bearer header, 503 on fetch failure, otherwise passthrough status + body.

### Auth (cookieless, localStorage JWT)

Supabase JS with `flowType: 'implicit'` + `detectSessionInUrl: true`. The OAuth callback returns the token in the URL **hash**, which never reaches the server — so there is **no `middleware.ts`, no server code exchange, no cookies**. JWT lives in `localStorage` (`harbour-wave-auth`), auto-refreshed by supabase-js. `/book/*` pages are gated client-side by `<AuthGuard>` (`src/features/booking/*` layout wraps everything in it). `src/shared/lib/api/client.ts` `apiFetch` attaches the Bearer token and, on a 401, signs out and redirects to `/signin?error=session_expired`.

Note: the codebase also has anonymous auth (see recent `feat(ui): add anon auth` commit) in addition to Google OAuth described in the README.

### The booking wizard

State lives in a Zustand store (`src/features/booking/store/bookingStore.ts`), persisted to `localStorage` (`harbour-wave-booking`). **Only** `selectedRoute/selectedDate/selectedTime/quantities` are persisted — contact info, `bookingId`, and `sessionId` are intentionally NOT.

Step order (see `src/app/book/layout.tsx` `STEP_MAP`):

`/book/route` (1) → `/book/date` (2) → `/book/time` (3) → `/book/boats` (4) → `/book/details` (5) → `/book/processing` → `/book/success` (6). `/book/cancelled` returns to step 1.

**Route selection is step 1 and cascades.** Setting the route (or date, or time) resets all downstream state, because availability and prices are route-scoped. Route IDs (`Desna`, `Klochkov`, in `src/features/booking/routes.ts`) are **case-sensitive and sent verbatim** to the backend — labels are Ukrainian and never localized server-side.

`src/features/booking/hooks/useStepGuard.ts` redirects to the first incomplete step, so pages assume prior steps are filled.

### Server-state hooks & API contract

TanStack Query hooks in `src/features/booking/hooks/index.ts` are the only callers of `apiFetch`. Endpoints (all via BFF, note the route segment):
- `GET /availability/:month/:route`, `GET /slots/:date/:route`
- `POST /bookings`, `POST /checkout`, `GET /bookings/:sessionId`
- `GET /status` (system status), `GET /weather` (see below)

### Idempotency (important, non-obvious)

`POST /bookings` sends `X-Idempotency-Key` from `getStableIdempotencyKey` (`src/shared/lib/idempotency.ts`). The key is fingerprinted on **route + date + time + quantities only** — contact edits (e.g. fixing a phone typo) must NOT mint a new key or you double-hold the slot. Keys are reused for ≤5 min. `store.reset()` calls `clearIdempotencyKey()`.

### Pricing

`src/features/booking/pricing.ts` computes a **client-side display estimate only**. The backend's `totalAmount` is always authoritative for the charge — keep this table roughly in sync with backend route/price config but never rely on it for correctness.

### Weather

`src/app/api/weather/` is the one BFF route with real logic: an in-memory 7-day cache (`cache.ts`) warmed on module load. `GET /api/weather?date=YYYY-MM-DD` returns one day, no param returns all 7.

### Env validation

`src/shared/lib/env.ts` validates env with Zod at import time and throws on invalid config. Client vars: `NEXT_PUBLIC_SUPABASE_URL`, `NEXT_PUBLIC_SUPABASE_ANON_KEY`, `NEXT_PUBLIC_APP_URL` (optional). Server-only: `BACKEND_URL`. Import `env` (server) for `BACKEND_URL`; `clientEnv` for public vars. This app is deliberately secret-light — no Stripe keys, no service role key, no JWT secret.

## Conventions

- Path alias `@/*` → `src/*`.
- Feature-first layout under `src/features/{auth,booking}` (components/hooks/lib/store/schema); cross-cutting code in `src/shared`.
- Styling is **plain CSS** (design tokens in `src/app/styles/variables.css` and `globals.css`, split per-concern files under `src/app/styles/`) despite Tailwind v4 being installed. Match the existing CSS-variable approach rather than adding Tailwind utility classes.
- UI strings are Ukrainian, centralized in `src/features/booking/messages.ts` and `consent-text.ts`.
- Forms use React Hook Form + Zod (`src/features/booking/schema/booking.schema.ts`).
- Security headers (CSP etc.) are set in `next.config.ts`. Adding an external script/style/frame/connect origin requires editing the CSP there.

## README drift

`README.md` is detailed but predates several changes — trust the code when they conflict. Known drift: README omits the **route** step (shows date as step 1), lists boat types as `big/medium/child` (code also has `small`), and shows API paths without the `/:route` segment. `quantities` in the store is `{ big, medium, small, child }`.