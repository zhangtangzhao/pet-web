-- ============================================================
-- 011: 评价官方回复——管理端可回复（可覆盖更新），用户端展示
-- ============================================================

ALTER TABLE order_review
    ADD COLUMN IF NOT EXISTS reply      TEXT NULL,
    ADD COLUMN IF NOT EXISTS replied_at TIMESTAMPTZ NULL;
