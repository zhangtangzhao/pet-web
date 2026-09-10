-- ============================================================
-- 宠物交易平台 · 初始建表（PostgreSQL 16）
-- 来源：docs/03-数据库设计.md §3
-- ============================================================

-- 3.1 会员表 member（用户端账号）
CREATE TABLE member (
    id            BIGINT PRIMARY KEY,
    nickname      VARCHAR(64)  NOT NULL DEFAULT '',
    avatar        VARCHAR(512) NOT NULL DEFAULT '',
    phone         VARCHAR(20)  NOT NULL DEFAULT '',
    gender        SMALLINT     NOT NULL DEFAULT 0,
    status        SMALLINT     NOT NULL DEFAULT 1,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX uk_member_phone ON member (phone) WHERE phone <> '';

-- 3.2 微信授权表 wechat_auth（多端 openid 归并）
CREATE TABLE wechat_auth (
    id          BIGINT PRIMARY KEY,
    member_id   BIGINT       NOT NULL,
    app_type    SMALLINT     NOT NULL,
    openid      VARCHAR(64)  NOT NULL,
    unionid     VARCHAR(64)  NOT NULL DEFAULT '',
    session_key VARCHAR(128) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (app_type, openid)
);
CREATE INDEX idx_wechat_auth_unionid ON wechat_auth (unionid) WHERE unionid <> '';
CREATE INDEX idx_wechat_auth_member  ON wechat_auth (member_id);

-- 3.3 平台用户表 admin_user
CREATE TABLE admin_user (
    id            BIGINT PRIMARY KEY,
    username      VARCHAR(32)  NOT NULL UNIQUE,
    password_hash VARCHAR(128) NOT NULL,
    nickname      VARCHAR(64)  NOT NULL DEFAULT '',
    role          VARCHAR(20)  NOT NULL DEFAULT 'operator',
    status        SMALLINT     NOT NULL DEFAULT 1,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- 3.4 宠物分类 category
CREATE TABLE category (
    id         BIGINT PRIMARY KEY,
    name       VARCHAR(32)  NOT NULL,
    icon       VARCHAR(512) NOT NULL DEFAULT '',
    sort       INT          NOT NULL DEFAULT 0,
    status     SMALLINT     NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- 3.5 宠物品种 breed
CREATE TABLE breed (
    id          BIGINT PRIMARY KEY,
    category_id BIGINT       NOT NULL,
    name        VARCHAR(64)  NOT NULL,
    intro       TEXT         NOT NULL DEFAULT '',
    cover       VARCHAR(512) NOT NULL DEFAULT '',
    sort        INT          NOT NULL DEFAULT 0,
    status      SMALLINT     NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    FOREIGN KEY (category_id) REFERENCES category (id)
);
CREATE INDEX idx_breed_category ON breed (category_id, sort);

-- 3.6 宠物商品 pet_product（SPU = 一只在售宠物）
CREATE TABLE pet_product (
    id             BIGINT PRIMARY KEY,
    spu_no         VARCHAR(32)   NOT NULL UNIQUE,
    title          VARCHAR(128)  NOT NULL,
    category_id    BIGINT        NOT NULL,
    breed_id       BIGINT        NOT NULL,
    price          NUMERIC(10,2) NOT NULL,
    original_price NUMERIC(10,2) NOT NULL DEFAULT 0,
    status         SMALLINT      NOT NULL DEFAULT 0,
    pet_gender     SMALLINT      NOT NULL DEFAULT 0,
    birth_date     DATE,
    vaccine_desc   VARCHAR(255)  NOT NULL DEFAULT '',
    deworm_desc    VARCHAR(255)  NOT NULL DEFAULT '',
    body_type      VARCHAR(20)   NOT NULL DEFAULT '',
    coat_color     VARCHAR(32)   NOT NULL DEFAULT '',
    personality    VARCHAR(255)  NOT NULL DEFAULT '',
    health_desc    TEXT          NOT NULL DEFAULT '',
    main_image     VARCHAR(512)  NOT NULL DEFAULT '',
    video_url      VARCHAR(512)  NOT NULL DEFAULT '',
    video_cover    VARCHAR(512)  NOT NULL DEFAULT '',
    detail_html    TEXT          NOT NULL DEFAULT '',
    stock          INT           NOT NULL DEFAULT 1,
    sales          INT           NOT NULL DEFAULT 0,
    view_count     INT           NOT NULL DEFAULT 0,
    favorite_count INT           NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_product_category ON pet_product (category_id, status);
CREATE INDEX idx_product_breed    ON pet_product (breed_id, status);
CREATE INDEX idx_product_list     ON pet_product (status, created_at DESC);
CREATE INDEX idx_product_hot      ON pet_product (status, sales DESC);

-- 3.7 商品图集 product_image
CREATE TABLE product_image (
    id         BIGINT PRIMARY KEY,
    product_id BIGINT       NOT NULL,
    url        VARCHAR(512) NOT NULL,
    sort       INT          NOT NULL DEFAULT 0,
    FOREIGN KEY (product_id) REFERENCES pet_product (id)
);
CREATE INDEX idx_product_image_product ON product_image (product_id, sort);

-- 3.8 收藏表 member_favorite
CREATE TABLE member_favorite (
    id         BIGINT PRIMARY KEY,
    member_id  BIGINT      NOT NULL,
    product_id BIGINT      NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (member_id, product_id)
);
CREATE INDEX idx_favorite_member ON member_favorite (member_id, created_at DESC);

-- 3.9 订单表 orders
CREATE TABLE orders (
    id            BIGINT PRIMARY KEY,
    order_no      VARCHAR(32)   NOT NULL UNIQUE,
    member_id     BIGINT        NOT NULL,
    total_amount  NUMERIC(10,2) NOT NULL,
    pay_amount    NUMERIC(10,2) NOT NULL,
    status        SMALLINT      NOT NULL DEFAULT 10,
    contact_name  VARCHAR(32)   NOT NULL DEFAULT '',
    contact_phone VARCHAR(20)   NOT NULL DEFAULT '',
    remark        VARCHAR(255)  NOT NULL DEFAULT '',
    paid_at       TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    canceled_at   TIMESTAMPTZ,
    cancel_reason VARCHAR(255)  NOT NULL DEFAULT '',
    expire_at     TIMESTAMPTZ   NOT NULL,
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_member ON orders (member_id, status, created_at DESC);
CREATE INDEX idx_order_expire ON orders (status, expire_at) WHERE status = 10;

-- 3.10 订单明细 order_item（成交快照）
CREATE TABLE order_item (
    id            BIGINT PRIMARY KEY,
    order_id      BIGINT        NOT NULL,
    product_id    BIGINT        NOT NULL,
    product_title VARCHAR(128)  NOT NULL,
    product_image VARCHAR(512)  NOT NULL DEFAULT '',
    breed_name    VARCHAR(64)   NOT NULL DEFAULT '',
    price         NUMERIC(10,2) NOT NULL,
    quantity      INT           NOT NULL DEFAULT 1,
    FOREIGN KEY (order_id) REFERENCES orders (id)
);
CREATE INDEX idx_order_item_order   ON order_item (order_id);
CREATE INDEX idx_order_item_product ON order_item (product_id);

-- 3.11 支付流水 payment
CREATE TABLE payment (
    id             BIGINT PRIMARY KEY,
    payment_no     VARCHAR(32)   NOT NULL UNIQUE,
    order_id       BIGINT        NOT NULL,
    order_no       VARCHAR(32)   NOT NULL,
    member_id      BIGINT        NOT NULL,
    amount         NUMERIC(10,2) NOT NULL,
    channel        SMALLINT      NOT NULL,
    pay_type       SMALLINT      NOT NULL DEFAULT 1,
    transaction_id VARCHAR(64)   NOT NULL DEFAULT '',
    status         SMALLINT      NOT NULL DEFAULT 0,
    callback_at    TIMESTAMPTZ,
    raw_notify     TEXT          NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT now(),
    UNIQUE (order_no, pay_type)
);
CREATE INDEX idx_payment_transaction ON payment (transaction_id);
CREATE INDEX idx_payment_member      ON payment (member_id, created_at DESC);
