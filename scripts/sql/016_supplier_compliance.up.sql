-- 016: 供货商管理 + 商品合规/防疫扩展 + 多角色管理员种子
BEGIN;

CREATE TABLE IF NOT EXISTS supplier (
    id         BIGINT PRIMARY KEY,
    name       VARCHAR(64)  NOT NULL,
    contact    VARCHAR(32)  NOT NULL DEFAULT '',
    phone      VARCHAR(20)  NOT NULL DEFAULT '',
    address    VARCHAR(255) NOT NULL DEFAULT '',
    license_no VARCHAR(64)  NOT NULL DEFAULT '',
    remark     VARCHAR(255) NOT NULL DEFAULT '',
    status     SMALLINT NOT NULL DEFAULT 1,  -- 1启用 0停用
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_supplier_list ON supplier(status, id DESC);

ALTER TABLE pet_product
    ADD COLUMN IF NOT EXISTS supplier_id         BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS quarantine_cert_url VARCHAR(512) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS next_vaccine_date   DATE NULL,
    ADD COLUMN IF NOT EXISTS next_deworm_date    DATE NULL;

CREATE INDEX IF NOT EXISTS idx_product_vaccine_due ON pet_product(next_vaccine_date) WHERE next_vaccine_date IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_product_deworm_due  ON pet_product(next_deworm_date)  WHERE next_deworm_date  IS NOT NULL;

-- 多角色管理员种子：operator(运营) / support(客服)，初始口令即注释所示，生产上线后务必修改
INSERT INTO admin_user (id, username, password_hash, nickname, role, status) VALUES
 (2, 'operator', '$2a$10$uTjet.anMIlGkgpE4LWoG.Y.Wl1ekSTJxkp6P.LQXN22WJAeYFGJ2', '运营专员', 'operator', 1),
 (3, 'support',  '$2a$10$EJmbaYJ.zf/vaFM8jSd0ReRsqkOJq8E0yGZ4mo8/C7GgmZ4RHB3sy', '客服专员', 'support',  1)
ON CONFLICT (id) DO NOTHING;

COMMIT;
