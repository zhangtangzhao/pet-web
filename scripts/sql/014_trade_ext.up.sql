-- 014 交易扩展：定金锁宠 + 秒杀 + 健康保障 + 积分兑券
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS deposit_amount NUMERIC(10,2) NOT NULL DEFAULT 0,  -- >0 即定金单
    ADD COLUMN IF NOT EXISTS tail_expire_at TIMESTAMPTZ NULL,                  -- 尾款补款截止
    ADD COLUMN IF NOT EXISTS flash_sale_id  BIGINT      NOT NULL DEFAULT 0,    -- 命中的秒杀活动
    ADD COLUMN IF NOT EXISTS guarantee_days INT         NOT NULL DEFAULT 0;    -- 健康保障天数快照

CREATE TABLE IF NOT EXISTS flash_sale (
    id         BIGINT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    sale_price NUMERIC(10,2) NOT NULL,
    stock      INT NOT NULL DEFAULT 1,
    sold       INT NOT NULL DEFAULT 0,
    start_at   TIMESTAMPTZ NOT NULL,
    end_at     TIMESTAMPTZ NOT NULL,
    status     SMALLINT NOT NULL DEFAULT 1,  -- 1启用 0停用
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_flash_sale_product ON flash_sale(product_id, status);
-- 每个商品同时至多一个启用中的活动
CREATE UNIQUE INDEX IF NOT EXISTS uk_flash_sale_active ON flash_sale(product_id) WHERE status = 1;

ALTER TABLE service_item ADD COLUMN IF NOT EXISTS guarantee_days INT NOT NULL DEFAULT 0; -- 保障卡天数，0=普通服务
ALTER TABLE coupon_template ADD COLUMN IF NOT EXISTS points_cost INT NOT NULL DEFAULT 0; -- >0 需积分兑换
