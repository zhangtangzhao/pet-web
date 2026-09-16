ALTER TABLE coupon_template DROP COLUMN IF EXISTS points_cost;
ALTER TABLE service_item DROP COLUMN IF EXISTS guarantee_days;
DROP INDEX IF EXISTS uk_flash_sale_active;
DROP INDEX IF EXISTS idx_flash_sale_product;
DROP TABLE IF EXISTS flash_sale;
ALTER TABLE orders
    DROP COLUMN IF EXISTS guarantee_days,
    DROP COLUMN IF EXISTS flash_sale_id,
    DROP COLUMN IF EXISTS tail_expire_at,
    DROP COLUMN IF EXISTS deposit_amount;
