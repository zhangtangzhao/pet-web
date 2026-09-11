-- ============================================================
-- 004: 营销——增值服务项 + 优惠券（模板/用户券）+ orders 优惠列
-- 金额统一 NUMERIC(10,2)；券状态 1可用 2已锁定(待支付) 3已使用 4已过期 5已作废
-- ============================================================

-- 增值服务项（订单附加，管理端维护）
CREATE TABLE IF NOT EXISTS service_item (
    id             BIGINT PRIMARY KEY,                        -- 雪花 ID（应用层生成）
    name           VARCHAR(64) NOT NULL,
    description    VARCHAR(255) NOT NULL DEFAULT '',
    original_price NUMERIC(10,2) NOT NULL DEFAULT 0,          -- 原价（划线展示）
    price          NUMERIC(10,2) NOT NULL,                    -- 服务价（下单按此计费）
    sort           INT NOT NULL DEFAULT 0,
    status         SMALLINT NOT NULL DEFAULT 1,               -- 1启用 0停用
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 优惠券模板
CREATE TABLE IF NOT EXISTS coupon_template (
    id                  BIGINT PRIMARY KEY,
    name                VARCHAR(64) NOT NULL,
    type                SMALLINT NOT NULL,                    -- 1满减 2折扣 3无门槛立减
    threshold_amount    NUMERIC(10,2) NOT NULL DEFAULT 0,     -- 满减门槛（type=1）
    discount_amount     NUMERIC(10,2) NOT NULL DEFAULT 0,     -- 满减/立减金额
    discount_percent    INT NOT NULL DEFAULT 0,               -- 折扣（type=2，90=9折）
    max_discount_amount NUMERIC(10,2) NOT NULL DEFAULT 0,     -- 折扣上限（type=2，0=不封顶）
    total_count         INT NOT NULL DEFAULT 0,               -- 发放总量，0=不限
    issued_count        INT NOT NULL DEFAULT 0,               -- 已发放
    per_limit           INT NOT NULL DEFAULT 1,               -- 每人限领
    new_user_only       SMALLINT NOT NULL DEFAULT 0,          -- 1=注册自动赠送
    pickup_start        TIMESTAMPTZ NULL,                     -- 可领时段（NULL 不限）
    pickup_end          TIMESTAMPTZ NULL,
    valid_start         TIMESTAMPTZ NULL,                     -- 可用时段（NULL 不限）
    valid_end           TIMESTAMPTZ NULL,
    status              SMALLINT NOT NULL DEFAULT 1,          -- 1启用 0停用
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 用户券
CREATE TABLE IF NOT EXISTS member_coupon (
    id          BIGINT PRIMARY KEY,
    member_id   BIGINT NOT NULL REFERENCES member(id),
    template_id BIGINT NOT NULL REFERENCES coupon_template(id),
    status      SMALLINT NOT NULL DEFAULT 1,                  -- 1可用 2锁定(待支付) 3已使用 4已过期 5已作废
    order_id    BIGINT NULL,                                  -- 锁定/核销时的订单
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    used_at     TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_member_coupon_member ON member_coupon(member_id, status);
CREATE INDEX IF NOT EXISTS idx_member_coupon_order ON member_coupon(order_id);
CREATE INDEX IF NOT EXISTS idx_coupon_template_status ON coupon_template(status);

-- orders 增加优惠列
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS service_fee     NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS coupon_id       BIGINT NULL,
    ADD COLUMN IF NOT EXISTS coupon_info     VARCHAR(128) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS service_items   TEXT NOT NULL DEFAULT '';

-- ── 种子数据 ──

-- 增值服务项
INSERT INTO service_item (id, name, description, original_price, price, sort) VALUES
    (5001, '专业托运', '全国范围专业活体托运，航空箱+全程温控，到家无忧', 300.00, 200.00, 1),
    (5002, '半年健康保障', '半年期延长健康保障，非人为疾病诊疗费用最高赔付 3000 元', 199.00, 99.00, 2)
ON CONFLICT (id) DO NOTHING;

-- 优惠券模板（其中一个注册赠送）
INSERT INTO coupon_template
    (id, name, type, threshold_amount, discount_amount, discount_percent, max_discount_amount,
     total_count, issued_count, per_limit, new_user_only, valid_start, valid_end) VALUES
    (5101, '新人立减券', 3, 0, 30.00, 0, 0, 0, 0, 1, 1, NULL, now() + interval '90 days'),
    (5102, '满1000减100券', 1, 1000.00, 100.00, 0, 0, 500, 0, 1, 0, NULL, now() + interval '60 days'),
    (5103, '全场9折券', 2, 0, 0, 90, 200.00, 200, 0, 1, 0, NULL, now() + interval '30 days')
ON CONFLICT (id) DO NOTHING;
