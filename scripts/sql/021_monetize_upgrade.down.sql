BEGIN;
DROP TABLE IF EXISTS home_config;
DROP TABLE IF EXISTS distributor_withdrawal;
DROP TABLE IF EXISTS distributor_commission;
DROP TABLE IF EXISTS distributor;
DROP TABLE IF EXISTS stud_service;
DROP TABLE IF EXISTS insurance_apply;
DROP TABLE IF EXISTS insurance_product;
DELETE FROM insurance_product WHERE id IN (8001, 8002);
ALTER TABLE pet_product
    DROP COLUMN IF EXISTS scheduled_off_sale_at,
    DROP COLUMN IF EXISTS chip_no,
    DROP COLUMN IF EXISTS cert_no,
    DROP COLUMN IF EXISTS cert_type;
ALTER TABLE orders
    DROP COLUMN IF EXISTS points_deduct,
    DROP COLUMN IF EXISTS points_used;
COMMIT;
