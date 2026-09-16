-- 017 回滚
BEGIN;
DELETE FROM coupon_template WHERE id IN (5111, 5112, 5113);
ALTER TABLE orders DROP COLUMN IF EXISTS level_discount;
ALTER TABLE member
    DROP COLUMN IF EXISTS level_reached,
    DROP COLUMN IF EXISTS growth_value;
COMMIT;
