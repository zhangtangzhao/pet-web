-- 020: 玩法与留存套件——砍价/竞拍/任务中心/付费会员/服务预约/圈子标签/发票
BEGIN;

-- ───────── 砍价 ─────────
CREATE TABLE IF NOT EXISTS bargain_activity (
    id BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    bottom_price NUMERIC(10,2) NOT NULL, -- 底价
    duration_hours INT NOT NULL DEFAULT 24,
    max_helpers INT NOT NULL DEFAULT 5,  -- 需要的助力人数
    status SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS bargain_launch (
    id BIGINT PRIMARY KEY,
    activity_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,
    current_price NUMERIC(10,2) NOT NULL,
    helper_count INT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0, -- 0进行中 1已到底可购 2已购买 3已过期
    expire_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_bargain_launch_member ON bargain_launch(member_id, id DESC);
CREATE TABLE IF NOT EXISTS bargain_help (
    id BIGINT PRIMARY KEY,
    launch_id BIGINT NOT NULL,
    helper_member_id BIGINT NOT NULL,
    amount NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_bargain_help ON bargain_help(launch_id, helper_member_id);

-- ───────── 竞拍 ─────────
CREATE TABLE IF NOT EXISTS auction (
    id BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    start_price NUMERIC(10,2) NOT NULL,
    step_price NUMERIC(10,2) NOT NULL DEFAULT 100,
    deposit_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    status SMALLINT NOT NULL DEFAULT 0, -- 0未开始 1进行中 2已成交 3流拍
    highest_member_id BIGINT NOT NULL DEFAULT 0,
    highest_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_auction_status ON auction(status, end_at);
CREATE TABLE IF NOT EXISTS auction_deposit (
    id BIGINT PRIMARY KEY,
    auction_id BIGINT NOT NULL,
    member_id BIGINT NOT NULL,
    payment_no VARCHAR(32) NOT NULL DEFAULT '',
    status SMALLINT NOT NULL DEFAULT 0, -- 0待支付 1已支付 2已退回
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_auction_deposit ON auction_deposit(auction_id, member_id);

-- ───────── 任务中心 ─────────
CREATE TABLE IF NOT EXISTS member_task (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    task_key VARCHAR(32) NOT NULL,
    reward_points INT NOT NULL,
    task_date DATE NOT NULL DEFAULT CURRENT_DATE, -- 成长型任务固定 '2000-01-01'
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_member_task ON member_task(member_id, task_key, task_date);

-- ───────── 付费会员 / 预售 / 圈子标签 / 发票 ─────────
ALTER TABLE member ADD COLUMN IF NOT EXISTS vip_expire_at TIMESTAMPTZ NULL;
ALTER TABLE pet_product ADD COLUMN IF NOT EXISTS pre_sale SMALLINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pre_sale_price NUMERIC(10,2) NULL,
    ADD COLUMN IF NOT EXISTS pre_sale_eta DATE NULL;
ALTER TABLE community_post ADD COLUMN IF NOT EXISTS breed_id BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS service_booking (
    id BIGINT PRIMARY KEY,
    booking_no VARCHAR(32) NOT NULL,
    member_id BIGINT NOT NULL,
    service_id BIGINT NOT NULL,
    service_name VARCHAR(64) NOT NULL DEFAULT '',
    store_id BIGINT NOT NULL DEFAULT 0,
    booking_date DATE NOT NULL,
    time_slot VARCHAR(16) NOT NULL,
    contact VARCHAR(32) NOT NULL DEFAULT '',
    phone VARCHAR(20) NOT NULL DEFAULT '',
    pet_name VARCHAR(32) NOT NULL DEFAULT '',
    price NUMERIC(10,2) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0, -- 0待到店 1已完成 2已取消
    verify_code VARCHAR(8) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_booking_member ON service_booking(member_id, id DESC);
CREATE UNIQUE INDEX IF NOT EXISTS uk_booking_no ON service_booking(booking_no);

CREATE TABLE IF NOT EXISTS invoice (
    id BIGINT PRIMARY KEY,
    member_id BIGINT NOT NULL,
    order_no VARCHAR(32) NOT NULL,
    title_type SMALLINT NOT NULL DEFAULT 1, -- 1个人 2企业
    title VARCHAR(128) NOT NULL,
    tax_no VARCHAR(32) NOT NULL DEFAULT '',
    amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0, -- 0待开 1已开
    link VARCHAR(512) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_invoice_order ON invoice(order_no);

-- ───────── 种子 ─────────
INSERT INTO coupon_template (id, name, type, threshold_amount, discount_amount, total_count, per_limit, valid_start, valid_end, status) VALUES
 (5131, 'VIP月礼·满100减15', 1, 100.00, 15.00, 99999, 1, '2026-01-01', '2027-12-31', 1)
ON CONFLICT (id) DO NOTHING;

COMMIT;
