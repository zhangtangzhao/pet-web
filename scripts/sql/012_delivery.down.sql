ALTER TABLE orders
    DROP COLUMN IF EXISTS delivered_at,
    DROP COLUMN IF EXISTS shipped_at,
    DROP COLUMN IF EXISTS ship_no,
    DROP COLUMN IF EXISTS ship_status,
    DROP COLUMN IF EXISTS ship_address,
    DROP COLUMN IF EXISTS ship_fee,
    DROP COLUMN IF EXISTS ship_method_name,
    DROP COLUMN IF EXISTS ship_method_id;

DROP TABLE IF EXISTS ship_method;
