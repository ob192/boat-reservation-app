# Harbour & Wave — Admin Dashboard

Admin control panel for the Harbour & Wave boat reservation system.

## Stack

- **Next.js 15** (App Router)
- **Supabase** — auth (Google OAuth) + session management
- **TanStack Query v5** — server state
- **React Hook Form + Zod** — forms & validation
- **Zustand** — minimal UI state
- **DM Sans + Playfair Display** — typography

## Setup

1. **Clone and install**

```bash
npm install
```

2. **Configure environment**

```bash
cp .env.local.example .env.local
```

Fill in:

```env
NEXT_PUBLIC_SUPABASE_URL=https://your-project.supabase.co
NEXT_PUBLIC_SUPABASE_ANON_KEY=your-anon-key
NEXT_PUBLIC_API_URL=http://localhost:8080
```

3. **Supabase setup**

- Enable Google OAuth provider in your Supabase project.
- Add `http://localhost:3000/auth/callback` to your allowed redirect URLs.
- Set `app_metadata.role = "admin"` for admin users via the Supabase dashboard or a server function.

4. **Run**

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Authorization

The app enforces admin-only access client-side via `<AdminGuard>`. The backend also enforces `app_metadata.role == "admin"` via `AdminMiddleware` — the frontend check is for UX only.

To make a user an admin, run in the Supabase SQL editor:

```sql
UPDATE auth.users
SET raw_app_meta_data = raw_app_meta_data || '{"role": "admin"}'::jsonb
WHERE email = 'admin@example.com';
```

## Sections

| Route | Purpose |
|---|---|
| `/slots` | Browse slots by date, edit capacity, view/cancel bookings |
| `/blocks` | Block/unblock individual slots or entire dates |
| `/system` | Global kill-switch for all bookings |

## Project Structure

```
src/
├── app/               # Next.js App Router pages
├── components/        # Shared UI components
│   └── slots/         # Slot-specific components
├── hooks/             # TanStack Query hooks
└── lib/               # API client, Supabase, types
```
