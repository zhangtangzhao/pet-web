-- 007 回滚：售后单
DROP INDEX IF EXISTS idx_aftersale_status;
DROP INDEX IF EXISTS idx_aftersale_member;
DROP INDEX IF EXISTS uk_aftersale_active;
DROP TABLE IF EXISTS after_sale;
