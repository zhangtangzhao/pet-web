-- ============================================================
-- 006: 订单评价——一单一评，仅已完成订单可评，管理端可隐藏/删除
-- ============================================================

CREATE TABLE IF NOT EXISTS order_review (
    id            BIGINT PRIMARY KEY,                          -- 雪花 ID（应用层生成）
    order_no      VARCHAR(32) NOT NULL UNIQUE,
    member_id     BIGINT NOT NULL,
    product_id    BIGINT NOT NULL,
    product_title VARCHAR(128) NOT NULL DEFAULT '',            -- 冗余快照，防商品改名
    rating        SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    content       VARCHAR(500) NOT NULL DEFAULT '',
    images        TEXT NOT NULL DEFAULT '[]',                  -- JSON 数组（COS 图片 URL）
    status        SMALLINT NOT NULL DEFAULT 1,                 -- 1显示 0隐藏（管理端管控）
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_review_product ON order_review(product_id, status, id DESC);
CREATE INDEX IF NOT EXISTS idx_review_member ON order_review(member_id, id DESC);
