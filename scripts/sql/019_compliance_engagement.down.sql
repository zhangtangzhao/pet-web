BEGIN;
DELETE FROM points_product WHERE id IN (700101, 700102);
DELETE FROM coupon_template WHERE id = 5130;
DROP TABLE IF EXISTS community_post_like;
DROP TABLE IF EXISTS community_post;
DROP TABLE IF EXISTS points_order;
DROP TABLE IF EXISTS points_product;
DROP TABLE IF EXISTS store;
DROP TABLE IF EXISTS order_trace;
ALTER TABLE orders
    DROP COLUMN IF EXISTS store_id,
    DROP COLUMN IF EXISTS agreement_version,
    DROP COLUMN IF EXISTS agreement_signed_at;
ALTER TABLE after_sale
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS exchange_product_id,
    DROP COLUMN IF EXISTS price_diff,
    DROP COLUMN IF EXISTS return_ship_no,
    DROP COLUMN IF EXISTS exchange_ship_no;
ALTER TABLE member
    DROP COLUMN IF EXISTS free_ship_cards,
    DROP COLUMN IF EXISTS birthday,
    DROP COLUMN IF EXISTS delete_cooldown_until,
    DROP COLUMN IF EXISTS delete_requested_at;
COMMIT;
