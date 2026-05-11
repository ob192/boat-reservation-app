# Harbour & Wave — Next.js 15 Frontend

Production-ready boat booking frontend. Next.js is the **presentation layer only** — all business logic, persistence, slot locking, payment session creation, and Stripe webhooks live on a separate backend service.

---

## Architecture Overview

```
Browser
  └─ Next.js 15 (Vercel)
       ├─ Public pages (SC): /, /signin, /auth/callback
       ├─ Protected wizard (CC): /book/*  ← AuthGuard (client-side JWT check)
       └─ BFF /api/* (Route Handlers)  ← dumb proxy, forwards JWT
            └─ Backend Service (your API)
                 ├─ JWT validation (Supabase JWKS or SUPABASE_JWT_SECRET)
                 ├─ Slot locking + booking persistence
                 ├─ Stripe Checkout Session creation
                 └─ Stripe webhook handler
```

**Auth model:** Pure JWT via `@supabase/supabase-js`. No cookies. No `@supabase/ssr`. JWT lives in `localStorage`, auto-refreshed by supabase-js before expiry. All `/book/*` pages are Client Components protected by `AuthGuard`.

---

## Tech Stack

| Concern | Library |
|---|---|
| Framework | Next.js 15 + App Router |
| Language | TypeScript (strict) |
| Auth | @supabase/supabase-js (implicit OAuth flow) |
| Server state | TanStack Query v5 |
| Form validation | React Hook Form + Zod |
| Wizard state | Zustand with persist |
| Styling | Tailwind CSS v4 + custom CSS vars |
| Animations | Framer Motion |
| Fonts | next/font (Playfair Display + DM Sans) |

---

## Quick Start

```bash
git clone <repo>
cd harbour-wave
npm install

cp .env.local.example .env.local
# Edit .env.local with your Supabase + backend values

npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

---

## Environment Variables

### Public (exposed to browser via `NEXT_PUBLIC_`)

| Variable | Description |
|---|---|
| `NEXT_PUBLIC_SUPABASE_URL` | Your Supabase project URL (`https://xxx.supabase.co`) |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Supabase anonymous/public key |
| `NEXT_PUBLIC_APP_URL` | Your deployed app URL (e.g. `https://harbour-wave.vercel.app`) |

### Server-only (BFF Route Handlers only — never sent to browser)

| Variable | Description |
|---|---|
| `BACKEND_URL` | Your backend API base URL (e.g. `https://api.harbour-wave.com`) |

**What is NOT in Next.js env:**
- No Stripe keys
- No Supabase service role key
- No database credentials
- No JWT secret

The Next.js app is genuinely secret-light — it's a rendering shell.

---

## Vercel Deployment

1. Push to GitHub
2. Import to Vercel
3. Set environment variables in **Vercel Dashboard → Settings → Environment Variables**:
   - `NEXT_PUBLIC_SUPABASE_URL`
   - `NEXT_PUBLIC_SUPABASE_ANON_KEY`
   - `NEXT_PUBLIC_APP_URL` → your Vercel URL
   - `BACKEND_URL` → your backend API URL
4. Deploy

For preview deployments, set `NEXT_PUBLIC_APP_URL` per-environment or use Vercel's `VERCEL_URL` system variable.

---

## Supabase Google OAuth Setup

### Step 1 — Google Cloud Console

1. Go to [console.cloud.google.com](https://console.cloud.google.com)
2. Create OAuth 2.0 Client ID (Web application)
3. Add Authorized redirect URIs:
   ```
   https://YOUR_PROJECT.supabase.co/auth/v1/callback
   ```
4. Copy the **Client ID** and **Client Secret**

### Step 2 — Supabase Dashboard

1. Go to **Authentication → Providers → Google**
2. Enable Google, paste Client ID and Client Secret
3. Go to **Authentication → URL Configuration**
4. Set **Site URL**:
   ```
   https://your-app.vercel.app
   ```
5. Add **Redirect URLs** (all of these):
   ```
   https://your-app.vercel.app/auth/callback
   https://*-your-team.vercel.app/auth/callback
   http://localhost:3000/auth/callback
   ```
6. Go to **Authentication → Settings**
   - JWT expiry: `3600` (1 hour — auto-refresh handles it)

### Step 3 — How the OAuth flow works

```
User clicks "Continue with Google"
  → supabase.auth.signInWithOAuth({ provider: 'google', redirectTo: '/auth/callback?next=...' })
  → Redirects to Google
  → Google redirects to Supabase callback URL
  → Supabase redirects to /auth/callback#access_token=...&refresh_token=...
  → supabase-js parses the URL fragment (detectSessionInUrl: true)
  → JWT stored in localStorage under key 'harbour-wave-auth'
  → /auth/callback page detects session via useSession(), redirects to /book
```

No server-side code exchange. No cookies. The JWT never touches a cookie jar.

---

## Backend API Contract

All endpoints require `Authorization: Bearer <supabase-jwt>`. The backend validates the JWT using either:

- **Option A (simple):** Call `supabase.auth.getUser(token)` — one extra Supabase HTTP hop per request
- **Option B (fast):** Verify locally with `SUPABASE_JWT_SECRET` (HS256) — fetch the secret from Supabase Dashboard → Settings → API → JWT Secret

Either way, extract `user.id` (the `sub` claim) and `user.email` and bind them to the booking.

### `GET /availability/:month` (YYYY-MM)

Returns slot availability summary per day for calendar display.

```json
// Response 200
{
  "month": "2025-07",
  "days": [
    { "date": "2025-07-15", "availableSlots": 12 },
    { "date": "2025-07-16", "availableSlots": 0 }
  ]
}
```

### `GET /slots/:date` (YYYY-MM-DD)

Returns per-time-slot availability for a specific date.

```json
// Response 200
{
  "date": "2025-07-15",
  "slots": [
    { "time": "08:00", "available": 12, "total": 15 },
    { "time": "11:00", "available": 3,  "total": 15 },
    { "time": "15:00", "available": 0,  "total": 15 },
    { "time": "19:00", "available": 15, "total": 15 }
  ]
}
```

### `POST /bookings`

Creates a pending booking and locks the slot. The backend binds `user_id` from the JWT — the body's `contact.email` is overridden by the JWT's email to prevent impersonation.

```json
// Request headers
Authorization: Bearer <supabase-jwt>
X-Idempotency-Key: <uuid>

// Request body
{
  "date": "2025-07-15",
  "time": "11:00",
  "quantities": { "big": 2, "medium": 1, "child": 1 },
  "contact": {
    "firstName": "Іван",
    "lastName": "Петренко",
    "email": "ivan@example.com",
    "phone": "+380671234567"
  }
}

// Response 201
{
  "bookingId": "bk_abc123",
  "totalAmount": 107.50,
  "expiresAt": "2025-07-15T10:25:00Z"
}

// Error responses
// 409 SLOT_TAKEN — slot was taken between selection and submission
// 422 VALIDATION_FAILED — business validation failed
// 401 SESSION_EXPIRED
```

**Important:** The backend MUST recompute `totalAmount` from `(date, time, quantities, user_id)` when creating the Stripe Checkout Session. Never trust the amount from the request body.

### `POST /checkout`

Creates a Stripe Checkout Session for a pending booking. Backend verifies `bookingId` belongs to the authenticated user.

```json
// Request headers
Authorization: Bearer <supabase-jwt>

// Request body
{ "bookingId": "bk_abc123" }

// Response 200
{
  "checkoutUrl": "https://checkout.stripe.com/pay/cs_...",
  "sessionId": "cs_abc123"
}

// Error responses
// 403 FORBIDDEN — booking doesn't belong to this user
// 404 BOOKING_NOT_FOUND
```

Stripe `success_url` should be: `https://your-app.vercel.app/book/success?session_id={CHECKOUT_SESSION_ID}`  
Stripe `cancel_url` should be: `https://your-app.vercel.app/book/cancelled`

### `GET /bookings/by-session/:sessionId`

Polls booking confirmation status. Backend verifies the session belongs to the authenticated user.

```json
// Response 200 — pending
{ "status": "pending" }

// Response 200 — confirmed
{
  "status": "confirmed",
  "booking": {
    "id": "bk_abc123",
    "date": "2025-07-15",
    "time": "11:00",
    "quantities": { "big": 2, "medium": 1, "child": 1 },
    "contact": { "firstName": "Іван", "lastName": "Петренко", "email": "ivan@example.com" },
    "totalAmount": 107.50,
    "status": "confirmed"
  }
}

// Response 200 — failed or expired
{ "status": "failed" }
{ "status": "expired" }
```

### Stripe Webhook → Backend (NOT Next.js)

The Stripe webhook endpoint lives on your backend, not on Next.js. On `checkout.session.completed`:
1. Look up booking by Stripe session ID
2. Set `booking.status = 'confirmed'`
3. Send confirmation email
4. Respond `200` to Stripe within 30s

---

## Project Structure

```
src/
├── app/
│   ├── layout.tsx                   # root: fonts, providers
│   ├── globals.css                  # CSS variables + design system
│   ├── page.tsx                     # Landing page (Server Component)
│   ├── signin/page.tsx              # Google OAuth entry
│   ├── auth/callback/page.tsx       # Parses #access_token fragment
│   ├── book/
│   │   ├── layout.tsx               # AuthGuard + wizard shell
│   │   ├── date/page.tsx            # Step 1: Calendar
│   │   ├── time/page.tsx            # Step 2: Time slots
│   │   ├── boats/page.tsx           # Step 3: Boat quantities
│   │   ├── details/page.tsx         # Step 4: Contact form + payment
│   │   ├── processing/page.tsx      # Stripe redirect spinner
│   │   ├── success/page.tsx         # Polls booking status
│   │   └── cancelled/page.tsx       # Stripe cancel return
│   └── api/                         # BFF proxy routes (forward JWT)
│       ├── availability/[month]/route.ts
│       ├── slots/[date]/route.ts
│       ├── bookings/route.ts
│       ├── bookings/[sessionId]/route.ts
│       └── checkout/route.ts
├── features/
│   ├── auth/
│   │   ├── components/
│   │   │   ├── AuthGuard.tsx        # Client-side auth gate
│   │   │   ├── SignInButton.tsx     # Google OAuth button
│   │   │   └── UserMenu.tsx         # Header avatar + sign out
│   │   ├── hooks/
│   │   │   ├── useSession.ts        # onAuthStateChange subscription
│   │   │   └── useUser.ts           # convenience hook
│   │   └── lib/
│   │       └── supabase.ts          # createClient (implicit flow, localStorage)
│   └── booking/
│       ├── components/client/
│       │   ├── Calendar.tsx         # Monday-first, availability dots
│       │   ├── TimeSlots.tsx        # Slot bars + selection
│       │   ├── BoatSelector.tsx     # Qty controls + capacity check
│       │   ├── DetailsForm.tsx      # RHF + Zod, readonly email
│       │   ├── StepIndicator.tsx    # Stable layout, no jumping
│       │   └── ProcessingScreen.tsx # Spinner + SuccessPoller
│       ├── hooks/index.ts           # TanStack Query hooks
│       ├── hooks/useStepGuard.ts    # Redirects to first incomplete step
│       ├── messages.ts              # All Ukrainian UI strings
│       ├── schema/booking.schema.ts # Zod schemas
│       └── store/bookingStore.ts   # Zustand (UI state only)
└── shared/
    ├── lib/
    │   ├── api/
    │   │   ├── client.ts            # apiFetch — attaches Bearer JWT
    │   │   └── types.ts             # API response types
    │   ├── currency.ts
    │   └── env.ts                   # Zod env validation
    └── providers/
        ├── QueryProvider.tsx        # TanStack Query
        └── SupabaseProvider.tsx     # Session detection on mount
```

---

## Security Notes

### JWT in localStorage

LocalStorage is XSS-vulnerable. Mitigations applied:
- **CSP header** in `next.config.ts` — restricts script sources
- React sanitizes all rendered content by default
- No `dangerouslySetInnerHTML` with user data anywhere
- No third-party analytics scripts on auth-touching pages

### Backend responsibilities (critical)

1. **Re-validate JWT on every request.** The BFF is a dumb proxy — it could be bypassed.
2. **Verify booking ownership.** On `/checkout` and `/bookings/by-session/:id`, confirm `user_id` from JWT matches `booking.user_id`. Otherwise user A could pay for user B's booking.
3. **Recompute the total.** When creating the Stripe Checkout Session, compute `totalAmount` from `(date, time, quantities)` server-side. Never trust the amount from the request body.
4. **Email from JWT wins.** On `POST /bookings`, override `contact.email` with the email from the validated JWT to prevent impersonation.

---

## Design System

All original CSS variables preserved:

```css
--navy: #0f2333      /* Primary dark background */
--deep: #081a27      /* Deepest background */
--teal: #1b7a8a      /* Primary interactive */
--seafoam: #4ab5c4   /* Active/selected states */
--sand: #f2ead8      /* Card headers, nav backgrounds */
--cream: #faf6ee     /* Card body backgrounds */
--coral: #e8603c     /* Errors, full slots */
--gold: #c9a84c      /* Limited availability */
--mist: #cde0e8      /* Borders, subtle backgrounds */
--subtle: #6a8a99    /* Secondary text */
```

Mobile-first breakpoints: `<360px` (very narrow), base (360px+), `600px+` (tablet/desktop).

---

## No-cookie Auth: Why This Matters

Because we use `flowType: 'implicit'` with `detectSessionInUrl: true`:

- The OAuth callback returns `#access_token=...` in the URL **hash** (fragment)
- The hash is never sent to the server — only `supabase-js` in the browser reads it
- This means **no server-side code exchange**, no `/api/auth/callback` route needed
- The JWT is stored in `localStorage` under `harbour-wave-auth`
- `autoRefreshToken: true` means supabase-js silently refreshes before expiry

This is why there is no `middleware.ts` — middleware runs on the server edge and cannot read localStorage.

---

## What's NOT in This Repo

As per spec, the following live on the **separate backend service**:

- Stripe SDK and webhook handlers
- Database schema, ORM, migrations
- Slot locking logic
- Email sending
- JWT validation (backend validates every inbound token)
- Authoritative price computation
