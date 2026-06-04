-- 0004_routes_and_small.sql
-- Adds route support and a third "small" boat class.
-- Backfills existing rows with the 'classic' route. Change if your default differs.

-- ---- slots ----
ALTER TABLE slots
    ADD COLUMN IF NOT EXISTS route_name     VARCHAR(50),
    ADD COLUMN IF NOT EXISTS capacity_small INTEGER NOT NULL DEFAULT 0 CHECK (capacity_small >= 0);

UPDATE slots SET route_name = 'classic' WHERE route_name IS NULL;
ALTER TABLE slots ALTER COLUMN route_name SET NOT NULL;

-- Rebuild PK to include route_name.
ALTER TABLE slots DROP CONSTRAINT slots_pkey;
ALTER TABLE slots ADD PRIMARY KEY (date, time, route_name);

-- ---- bookings ----
ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS route_name VARCHAR(50),
    ADD COLUMN IF NOT EXISTS qty_small  INTEGER NOT NULL DEFAULT 0 CHECK (qty_small >= 0);

UPDATE bookings SET route_name = 'classic' WHERE route_name IS NULL;
ALTER TABLE bookings ALTER COLUMN route_name SET NOT NULL;

-- Count small boats in the "at least one boat" rule.
ALTER TABLE bookings DROP CONSTRAINT IF EXISTS bookings_qty_total_min;
ALTER TABLE bookings ADD CONSTRAINT bookings_qty_total_min
    CHECK (qty_big + qty_medium + qty_small >= 1);

-- Availability queries now group by route.
DROP INDEX IF EXISTS idx_bookings_date_time;
CREATE INDEX IF NOT EXISTS idx_bookings_date_time_route
    ON bookings (date, time, route_name);