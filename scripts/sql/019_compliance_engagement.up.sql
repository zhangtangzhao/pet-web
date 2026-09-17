-- 019: 合规与留存套件——账号注销/购买协议/物流轨迹/售后换货/多门店自提/
--       签到日历补签/积分商城/生日礼包/晒单广场（百科复用 ai_knowledge）
BEGIN;

-- ───────── 账号注销（冷静期后匿名化，交易记录保留）─────────
ALTER TABLE member
    ADD COLUMN IF NOT EXISTS delete_requested_at  TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS delete_cooldown_until TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS birthday DATE NULL,
    ADD COLUMN IF NOT EXISTS free_ship_cards INT NOT NULL DEFAULT 0; -- 积分商城免运费卡

-- ───────── 电子购买协议（下单签署快照）─────────
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS agreement_version   VARCHAR(16) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS agreement_signed_at TIMESTAMPTZ NULL;

-- ───────── 物流轨迹（管理端录入 / 预留第三方订阅适配）─────────
CREATE TABLE IF NOT EXISTS order_trace (
    id          BIGINT PRIMARY KEY,
    order_no    VARCHAR(32) NOT NULL,
    happened_at TIMESTAMPTZ NOT NULL,
    status_desc VARCHAR(64) NOT NULL,
    detail      VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_trace_order ON order_trace(order_no, happened_at);

-- ───────── 售后换货 ─────────
ALTER TABLE after_sale
    ADD COLUMN IF NOT EXISTS type SMALLINT NOT NULL DEFAULT 1,        -- 1退款 2换货
    ADD COLUMN IF NOT EXISTS exchange_product_id BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS price_diff NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS return_ship_no VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS exchange_ship_no VARCHAR(32) NOT NULL DEFAULT '';

-- ───────── 多门店自提 ─────────
CREATE TABLE IF NOT EXISTS store (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    address VARCHAR(255) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    business_hours VARCHAR(64) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE orders ADD COLUMN IF NOT EXISTS store_id BIGINT NOT NULL DEFAULT 0;

-- ───────── 积分商城 ─────────
CREATE TABLE IF NOT EXISTS points_product (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    image VARCHAR(512) NOT NULL DEFAULT '',
    points_cost INT NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    type SMALLINT NOT NULL DEFAULT 1, -- 1优惠券 2实物 3免运费卡
    coupon_template_id BIGINT NOT NULL DEFAULT 0,
    description VARCHAR(255) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    sort INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_points_shop ON points_product(status, sort, id DESC);

CREATE TABLE IF NOT EXISTS points_order (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    product_name VARCHAR(64) NOT NULL DEFAULT '',
    image VARCHAR(512) NOT NULL DEFAULT '',
    points_cost INT NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0待发货 1已发货 2已完成
    ship_no VARCHAR(32) NOT NULL DEFAULT '',
    contact VARCHAR(32) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    address VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_points_order_member ON points_order(member_id, id DESC);

-- ───────── 晒单广场 ─────────
CREATE TABLE IF NOT EXISTS community_post (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    content VARCHAR(500) NOT NULL DEFAULT '',
    images JSONB NOT NULL DEFAULT '[]',
    status SMALLINT NOT NULL DEFAULT 0, -- 0待审核 1显示 2已隐藏
    like_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_post_list ON community_post(status, id DESC);
CREATE TABLE IF NOT EXISTS community_post_like (
    id BIGINT PRIMARY KEY,
    post_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_post_like ON community_post_like(post_id, member_id);

-- ───────── 种子 ─────────
INSERT INTO coupon_template (id, name, type, threshold_amount, discount_amount, total_count, per_limit, valid_start, valid_end, status) VALUES
 (5130, '生日礼·满300减40', 1, 300.00, 40.00, 99999, 1, '2026-01-01', '2027-12-31', 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO points_product (id, name, image, points_cost, stock, type, coupon_template_id, description, status, sort) VALUES
 (700101, '免运费卡', '', 200, 9999, 3, 0, '下单自动抵扣一次运费', 1, 1),
 (700102, '积分兑券·满100减10', '', 150, 9999, 1, 5111, '兑换后到"我的优惠券"查看', 1, 2)
ON CONFLICT (id) DO NOTHING;

COMMIT;
