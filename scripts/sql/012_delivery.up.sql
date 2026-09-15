CREATE TABLE IF NOT EXISTS ship_method (
    id          BIGINT PRIMARY KEY,
    name        VARCHAR(32)   NOT NULL,
    kind        SMALLINT      NOT NULL DEFAULT 2,  -- 1自提 2托运配送
    description VARCHAR(128)  NOT NULL DEFAULT '',
    fee         NUMERIC(10,2) NOT NULL DEFAULT 0,
    sort        INT           NOT NULL DEFAULT 0,
    status      SMALLINT      NOT NULL DEFAULT 1,  -- 1启用 0停用
    created_at  TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ship_method_list ON ship_method(status, sort ASC, id ASC);

INSERT INTO ship_method (id, name, kind, description, fee, sort, status) VALUES
 (7001, '到店自提', 1, '到店自提免运费，凭订单号取宠', 0,      1, 1),
 (7002, '专车配送', 2, '专业宠物专车，全程视频可查',   300.00, 2, 1),
 (7003, '航空托运', 2, '有氧舱托运，含航空箱与检疫协助', 500.00, 3, 1)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS ship_method_id   BIGINT        NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ship_method_name VARCHAR(32)   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ship_fee         NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS ship_address     VARCHAR(255)  NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ship_status      SMALLINT      NOT NULL DEFAULT 0, -- 0待配送 1配送中 2已送达
    ADD COLUMN IF NOT EXISTS ship_no          VARCHAR(64)   NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS shipped_at       TIMESTAMPTZ   NULL,
    ADD COLUMN IF NOT EXISTS delivered_at     TIMESTAMPTZ   NULL;
