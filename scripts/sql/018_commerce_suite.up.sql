-- 018: 商城深化套件——购物车 / SKU 规格 / 拼团 / 自提核销 / 评价三维评分 /
--       宠物档案 / 详情富媒体 / 库存预警 / 营销自动化幂等键 / 风控
BEGIN;

-- ───────── 购物车 ─────────
CREATE TABLE IF NOT EXISTS cart (
    id         BIGINT PRIMARY KEY,
    member_id  BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    sku_id     BIGINT NOT NULL DEFAULT 0, -- 0=无规格商品
    checked    SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_cart_member_product ON cart(member_id, product_id);
CREATE INDEX IF NOT EXISTS idx_cart_member ON cart(member_id, checked, updated_at DESC);

-- ───────── SKU 规格（宠物单只库存不拆，规格只影响价格/权益套餐）─────────
CREATE TABLE IF NOT EXISTS product_sku (
    id         BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    specs      VARCHAR(128) NOT NULL,       -- 如 "3个月|含三针疫苗"
    price      NUMERIC(10,2) NOT NULL,
    sort       INT NOT NULL DEFAULT 0,
    status     SMALLINT NOT NULL DEFAULT 1, -- 1启用 0停用
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sku_product ON product_sku(product_id, sort);

ALTER TABLE pet_product
    ADD COLUMN IF NOT EXISTS has_sku              SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS detail_images        JSONB    NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS stock_warn_threshold INT      NOT NULL DEFAULT 1;

ALTER TABLE order_item
    ADD COLUMN IF NOT EXISTS sku_specs VARCHAR(128) NOT NULL DEFAULT '';

-- ───────── 拼团 ─────────
CREATE TABLE IF NOT EXISTS group_buy (
    id         BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    price      NUMERIC(10,2) NOT NULL, -- 拼团价
    size       INT NOT NULL DEFAULT 2, -- 成团人数
    hours      INT NOT NULL DEFAULT 24,-- 成团时限（小时）
    status     SMALLINT NOT NULL DEFAULT 1, -- 1启用 0停用
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_group_buy_product ON group_buy(product_id, status);

CREATE TABLE IF NOT EXISTS group_team (
    id               BIGINT PRIMARY KEY,
    group_buy_id     BIGINT NOT NULL,
    leader_member_id BIGINT NOT NULL,
    member_count     INT NOT NULL DEFAULT 0,
    status           SMALLINT NOT NULL DEFAULT 0, -- 0进行中 1成功 2失败
    expire_at        TIMESTAMPTZ NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_team_open ON group_team(group_buy_id, status, expire_at);

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS group_team_id BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pickup_code   VARCHAR(8) NOT NULL DEFAULT '';

-- ───────── 评价三维评分 ─────────
ALTER TABLE order_review
    ADD COLUMN IF NOT EXISTS health_score  SMALLINT NOT NULL DEFAULT 5,
    ADD COLUMN IF NOT EXISTS look_score    SMALLINT NOT NULL DEFAULT 5,
    ADD COLUMN IF NOT EXISTS service_score SMALLINT NOT NULL DEFAULT 5;

-- ───────── 宠物档案 ─────────
CREATE TABLE IF NOT EXISTS pet_profile (
    id               BIGINT PRIMARY KEY,
    member_id        BIGINT NOT NULL,
    name             VARCHAR(32) NOT NULL,
    breed_name       VARCHAR(32) NOT NULL DEFAULT '',
    gender           SMALLINT NOT NULL DEFAULT 0, -- 0未知 1公 2母
    birthday         DATE NULL,
    weight           NUMERIC(5,2) NULL,
    avatar           VARCHAR(512) NOT NULL DEFAULT '',
    vaccine_at       DATE NULL,                   -- 最近疫苗日期
    next_vaccine_date DATE NULL,
    next_deworm_date  DATE NULL,
    remark           VARCHAR(255) NOT NULL DEFAULT '',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_pet_member ON pet_profile(member_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_pet_vaccine_due ON pet_profile(next_vaccine_date) WHERE next_vaccine_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_pet_deworm_due ON pet_profile(next_deworm_date) WHERE next_deworm_date IS NOT NULL;

-- ───────── 风控 ─────────
CREATE TABLE IF NOT EXISTS risk_log (
    id         BIGINT PRIMARY KEY,
    member_id  BIGINT NOT NULL DEFAULT 0,
    rule       VARCHAR(32) NOT NULL,  -- order_freq / review_contact / member_blacklist
    detail     VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_risk_time ON risk_log(created_at DESC);

ALTER TABLE member
    ADD COLUMN IF NOT EXISTS blacklist SMALLINT NOT NULL DEFAULT 0; -- 1=黑名单（可登录，禁交易/评价/领券）

-- ───────── 营销自动化幂等键（notify 去重）─────────
CREATE UNIQUE INDEX IF NOT EXISTS uk_notify_member_biz ON notification(member_id, biz_key);

-- ───────── 沉睡召回券模板 ─────────
INSERT INTO coupon_template (id, name, type, threshold_amount, discount_amount, total_count, per_limit, valid_start, valid_end, status) VALUES
 (5120, '沉睡召回·满200减20', 1, 200.00, 20.00, 99999, 1, '2026-01-01', '2027-12-31', 1)
ON CONFLICT (id) DO NOTHING;

COMMIT;
