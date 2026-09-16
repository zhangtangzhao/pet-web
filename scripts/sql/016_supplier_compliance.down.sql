-- 016 回滚
BEGIN;
DROP INDEX IF EXISTS idx_product_deworm_due;
DROP INDEX IF EXISTS idx_product_vaccine_due;
ALTER TABLE pet_product
    DROP COLUMN IF EXISTS next_deworm_date,
    DROP COLUMN IF EXISTS next_vaccine_date,
    DROP COLUMN IF EXISTS quarantine_cert_url,
    DROP COLUMN IF EXISTS supplier_id;
DELETE FROM admin_user WHERE id IN (2, 3);
DROP TABLE IF EXISTS supplier;
COMMIT;
