-- 006 回滚：订单评价
DROP INDEX IF EXISTS idx_review_member;
DROP INDEX IF EXISTS idx_review_product;
DROP TABLE IF EXISTS order_review;
