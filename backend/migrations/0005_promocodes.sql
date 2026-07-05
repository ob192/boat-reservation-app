-- 0005_promocodes.sql
-- Promocodes (admin-generated affiliate discounts) + per-booking promo record.

CREATE TABLE IF NOT EXISTS promocodes (
    code             VARCHAR(64)  PRIMARY KEY,
    discount_percent INTEGER      NOT NULL CHECK (discount_percent BETWEEN 0 AND 100),
    max_uses         INTEGER      NOT NULL CHECK (max_uses >= 1),
    times_used       INTEGER      NOT NULL DEFAULT 0 CHECK (times_used >= 0),
    active           BOOLEAN      NOT NULL DEFAULT TRUE,
    created_by       UUID         NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_promocodes_created_by ON promocodes (created_by);

-- Booking keeps an immutable snapshot of the applied promo (survives promo edits/deletes).
ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS promo_code       VARCHAR(64),
    ADD COLUMN IF NOT EXISTS discount_percent INTEGER,
    ADD COLUMN IF NOT EXISTS discount_amount  DECIMAL(10, 2);

CREATE INDEX IF NOT EXISTS idx_bookings_promo_code
    ON bookings (promo_code) WHERE promo_code IS NOT NULL;