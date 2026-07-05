# internal/config

Single file, `config.go`. `Config` is a flat struct of everything the app needs at runtime, populated once by `Load()` from environment variables (via `godotenv`-loaded `.env` in `main.go`, or real env vars in prod).

- `getenv(key, fallback)` — reads `os.LookupEnv`; an env var that's unset **or** set to an empty string both fall through to `fallback` (it checks `v != ""`, not just `ok`).
- `validate()` enforces that `DATABASE_URL` and all four `SUPABASE_*` vars are non-empty; `Load()` returns an error (not a panic) if any are missing, which `cmd/api/main.go` treats as fatal at boot.
- `PosterEnabled` is derived (`POSTER_API_TOKEN != ""`), not read directly from an env var — don't add a separate `POSTER_ENABLED` flag, just check the token.
- LiqPay defaults (`LiqpayPublicKey`/`LiqpayPrivateKey`) fall back to LiqPay's public sandbox credentials if unset, so local dev works without any LiqPay-specific env setup — real deployments must override both.

When adding a new required integration, add its field here, read it in `Load()`, and add it to the `required` map in `validate()` only if the app genuinely cannot run without it (optional integrations like Poster should stay soft-optional, gated by an `*Enabled` derived bool instead).