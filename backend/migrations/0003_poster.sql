ALTER TABLE bookings
    ADD COLUMN IF NOT EXISTS poster_incoming_order_id       BIGINT,
    ADD COLUMN IF NOT EXISTS poster_incoming_transaction_id BIGINT;