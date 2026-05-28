-- Harbour & Wave — initial schema.
--
-- This file is the source-of-truth schema. GORM's AutoMigrate at boot will
-- create equivalent tables, but for production deploys you should run this
-- file via your migration tool of choice (golang-migrate, atlas, dbmate, etc.)
-- and disable AutoMigrate.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ---------------------------------------------------------------------------
-- updated_at trigger helper — used on bookings.
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ---------------------------------------------------------------------------
-- bookings
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS bookings (
    id                  UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID            NOT NULL,
    user_email          VARCHAR(255)    NOT NULL,

    date                DATE            NOT NULL,
    time                VARCHAR(5)      NOT NULL,
    qty_big             INTEGER         NOT NULL DEFAULT 0 CHECK (qty_big    >= 0),
    qty_medium          INTEGER         NOT NULL DEFAULT 0 CHECK (qty_medium >= 0),
    qty_child           INTEGER         NOT NULL DEFAULT 0 CHECK (qty_child  >= 0),

    first_name          VARCHAR(50)     NOT NULL,
    last_name           VARCHAR(50)     NOT NULL,
    phone               VARCHAR(50),

    total_amount        DECIMAL(10, 2)  NOT NULL CHECK (total_amount >= 0),
    price_override      DECIMAL(10, 2),
    override_reason     VARCHAR(255),
    overridden_by       UUID,
    overridden_at       TIMESTAMPTZ,

    status              VARCHAR(20)     NOT NULL,

    cancelled_by        UUID,
    cancelled_at        TIMESTAMPTZ,
    cancel_reason       VARCHAR(255),

    payment_session_id  VARCHAR(255),
    idempotency_key     VARCHAR(255)    NOT NULL,

    expires_at          TIMESTAMPTZ     NOT NULL,
    created_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ     NOT NULL DEFAULT NOW(),

    CONSTRAINT bookings_status_chk
        CHECK (status IN ('pending', 'confirmed', 'failed', 'expired', 'cancelled')),
    CONSTRAINT bookings_qty_total_min   CHECK (qty_big + qty_medium >= 1),
    CONSTRAINT bookings_child_needs_big CHECK (qty_child = 0 OR qty_big > 0)
);

-- Lookup hot paths.
CREATE INDEX        IF NOT EXISTS idx_bookings_user_id        ON bookings (user_id);
CREATE INDEX        IF NOT EXISTS idx_bookings_date_time      ON bookings (date, time);
CREATE INDEX        IF NOT EXISTS idx_bookings_status         ON bookings (status);
CREATE INDEX        IF NOT EXISTS idx_bookings_status_expires ON bookings (status, expires_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_session        ON bookings (payment_session_id) WHERE payment_session_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_idem_user      ON bookings (user_id, idempotency_key);

DROP TRIGGER IF EXISTS bookings_set_updated_at ON bookings;
CREATE TRIGGER bookings_set_updated_at
BEFORE UPDATE ON bookings
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ---------------------------------------------------------------------------
-- slots — physical fleet capacity per (date, time).
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS slots (
    date            DATE        NOT NULL,
    time            VARCHAR(5)  NOT NULL,
    capacity_big    INTEGER     NOT NULL DEFAULT 0 CHECK (capacity_big    >= 0),
    capacity_medium INTEGER     NOT NULL DEFAULT 0 CHECK (capacity_medium >= 0),
    blocked         BOOLEAN     NOT NULL DEFAULT FALSE,
    blocked_by      UUID,
    blocked_at      TIMESTAMPTZ,
    block_reason    VARCHAR(255),
    PRIMARY KEY (date, time));

-- ---------------------------------------------------------------------------
-- date_blocks — whole-day admin blocks.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS date_blocks (
    date         DATE         PRIMARY KEY,
    blocked_by   UUID         NOT NULL,
    blocked_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    block_reason VARCHAR(255)
);

-- ---------------------------------------------------------------------------
-- system_settings — single-row table with global flags.
-- Always exactly one row; id = 1.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS system_settings (
    id                INTEGER     PRIMARY KEY CHECK (id = 1),
    bookings_enabled  BOOLEAN     NOT NULL DEFAULT TRUE,
    reason            VARCHAR(255),
    updated_by        UUID,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed the singleton row if missing. Safe to re-run.
INSERT INTO system_settings (id, bookings_enabled, updated_at)
VALUES (1, TRUE, NOW())
ON CONFLICT (id) DO NOTHING;
