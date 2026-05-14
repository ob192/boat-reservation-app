-- Seed the four standard daily slots for a given date range.
-- Adjust the date range and capacities to taste. Idempotent — safe to re-run.

WITH date_range AS (
    SELECT generate_series('2025-06-01'::date, '2025-09-30'::date, INTERVAL '1 day')::date AS d
),
slot_times(t) AS (
    VALUES ('08:00'), ('11:00'), ('15:00'), ('19:00')
)
INSERT INTO slots (date, time, capacity_big, capacity_medium, blocked)
SELECT d, t, 5, 10, FALSE
FROM date_range
CROSS JOIN slot_times
ON CONFLICT (date, time) DO NOTHING;
