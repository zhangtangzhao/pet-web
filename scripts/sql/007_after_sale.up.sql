-- ============================================================
-- 007: 用户侧售后——退款申请，管理端审核（默认全额，可调金额）
-- 状态 1待审核 2已同意(退款) 3已拒绝 4已撤销
-- ============================================================

CREATE TABLE IF NOT EXISTS after_sale (
    id                BIGINT PRIMARY KEY,                      -- 雪花 ID（应用层生成）
    after_sale_no     VARCHAR(32) NOT NULL UNIQUE,             -- 业务单号前缀 AS
    order_no          VARCHAR(32) NOT NULL,
    member_id         BIGINT NOT NULL,
    reason            VARCHAR(500) NOT NULL,
    refund_amount     NUMERIC(10,2) NOT NULL,                  -- 申请金额（默认全额，审核可调）
    status            SMALLINT NOT NULL DEFAULT 1,
    admin_note        VARCHAR(500) NOT NULL DEFAULT '',        -- 审核/调整说明
    refund_payment_no VARCHAR(64) NOT NULL DEFAULT '',         -- 退款流水号（RF+售后单号）
    audit_at          TIMESTAMPTZ NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- 同一订单同时仅允许一个进行中售后（部分唯一索引，GORM 层无法表达靠 23505 兜底）
CREATE UNIQUE INDEX IF NOT EXISTS uk_aftersale_active ON after_sale(order_no) WHERE status = 1;
CREATE INDEX IF NOT EXISTS idx_aftersale_member ON after_sale(member_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_aftersale_status ON after_sale(status, id DESC);
