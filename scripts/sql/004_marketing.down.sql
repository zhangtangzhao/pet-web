-- 004 回滚：营销表 + orders 优惠列（按依赖逆序）

DROP TABLE IF EXISTS member_coupon;
DROP TABLE IF EXISTS coupon_template;
DROP TABLE IF EXISTS service_item;

ALTER TABLE orders
    DROP COLUMN IF EXISTS discount_amount,
    DROP COLUMN IF EXISTS service_fee,
    DROP COLUMN IF EXISTS coupon_id,
    DROP COLUMN IF EXISTS coupon_info,
    DROP COLUMN IF EXISTS service_items;
