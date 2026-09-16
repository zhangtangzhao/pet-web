-- 018 down
BEGIN;

DROP TABLE IF EXISTS risk_log;
DROP TABLE IF EXISTS pet_profile;
DROP TABLE IF EXISTS group_team;
DROP TABLE IF EXISTS group_buy;
DROP TABLE IF EXISTS product_sku;
DROP TABLE IF EXISTS cart;

ALTER TABLE notification DROP CONSTRAINT IF EXISTS uk_notify_member_biz;
ALTER TABLE member
    DROP COLUMN IF EXISTS blacklist;
ALTER TABLE order_review
    DROP COLUMN IF EXISTS health_score,
    DROP COLUMN IF EXISTS look_score,
    DROP COLUMN IF EXISTS service_score;
ALTER TABLE orders
    DROP COLUMN IF EXISTS group_team_id,
    DROP COLUMN IF EXISTS pickup_code;
ALTER TABLE order_item
    DROP COLUMN IF EXISTS sku_specs;
ALTER TABLE pet_product
    DROP COLUMN IF EXISTS has_sku,
    DROP COLUMN IF EXISTS detail_images,
    DROP COLUMN IF EXISTS stock_warn_threshold;

DELETE FROM coupon_template WHERE id = 5120;

COMMIT;
