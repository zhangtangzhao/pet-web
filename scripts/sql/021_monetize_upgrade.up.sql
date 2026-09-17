-- 021: 变现与运营升级——积分抵现/保险货架/血统芯片/配种/分销/定时上下架/首页装修/会员标签
BEGIN;

-- 积分抵现
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS points_deduct NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS points_used INT NOT NULL DEFAULT 0;

-- 血统证书/芯片号 + 定时下架
ALTER TABLE pet_product
    ADD COLUMN IF NOT EXISTS cert_type VARCHAR(16) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS cert_no VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS chip_no VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS scheduled_off_sale_at TIMESTAMPTZ NULL;

-- 宠物保险
CREATE TABLE IF NOT EXISTS insurance_product (
    id BIGINT PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    company VARCHAR(64) NOT NULL DEFAULT '',
    cover_desc TEXT NOT NULL DEFAULT '',
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS insurance_apply (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    product_name VARCHAR(64) NOT NULL DEFAULT '',
    contact VARCHAR(32) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0, -- 0待处理 1已联系
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 配种服务
CREATE TABLE IF NOT EXISTS stud_service (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    breed_name VARCHAR(32) NOT NULL DEFAULT '',
    pet_name VARCHAR(32) NOT NULL DEFAULT '',
    health_certs VARCHAR(255) NOT NULL DEFAULT '',
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    description VARCHAR(500) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 分销
CREATE TABLE IF NOT EXISTS distributor (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL UNIQUE,
    level SMALLINT NOT NULL DEFAULT 1, -- 1普通 2高级
    commission_rate NUMERIC(5,2) NOT NULL DEFAULT 5.00, -- %
    total_commission NUMERIC(10,2) NOT NULL DEFAULT 0,
    balance NUMERIC(10,2) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS distributor_commission (
    id BIGINT PRIMARY KEY,
    distributor_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL DEFAULT '',
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0, -- 0待结算 1已结算
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS distributor_withdrawal (
    id BIGINT PRIMARY KEY,
    distributor_id BIGINT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0待审核 1已打款 2已拒绝
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 首页装修配置
CREATE TABLE IF NOT EXISTS home_config (
    id INT PRIMARY KEY DEFAULT 1,
    config JSONB NOT NULL DEFAULT '[]',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 种子
INSERT INTO insurance_product (id, name, company, cover_desc, price, status) VALUES
 (8001, '宠物意外医疗险', '平安产险', '意外伤害+门诊医疗，最高赔付5000元/年', 199.00, 1),
 (8002, '宠物重大疾病险', '众安保险', '犬瘟/细小等重大疾病，最高赔付20000元', 399.00, 1)
ON CONFLICT (id) DO NOTHING;

INSERT INTO home_config (id, config) VALUES (1, '[]') ON CONFLICT (id) DO NOTHING;

COMMIT;
