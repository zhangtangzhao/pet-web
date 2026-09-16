-- 013 用户增长：收货地址簿 + 积分 + 邀请归因
CREATE TABLE IF NOT EXISTS member_address (
    id         BIGINT PRIMARY KEY,
    member_id  BIGINT NOT NULL,
    name       VARCHAR(32)  NOT NULL,
    phone      VARCHAR(20)  NOT NULL,
    address    VARCHAR(255) NOT NULL,
    is_default SMALLINT NOT NULL DEFAULT 0,  -- 1默认地址
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_member_address_member ON member_address(member_id, is_default DESC, updated_at DESC);

ALTER TABLE member
    ADD COLUMN IF NOT EXISTS points      BIGINT      NOT NULL DEFAULT 0,  -- 积分余额
    ADD COLUMN IF NOT EXISTS invite_code VARCHAR(16) NOT NULL DEFAULT '', -- 本人邀请码
    ADD COLUMN IF NOT EXISTS invited_by  BIGINT      NOT NULL DEFAULT 0;  -- 邀请人 member.id
CREATE UNIQUE INDEX IF NOT EXISTS uk_member_invite_code ON member(invite_code) WHERE invite_code <> '';

CREATE TABLE IF NOT EXISTS points_log (
    id            BIGINT PRIMARY KEY,
    member_id     BIGINT NOT NULL,
    change        BIGINT NOT NULL,               -- 正负变动
    balance_after BIGINT NOT NULL,
    reason        VARCHAR(32) NOT NULL,          -- sign/review/order/invite/invite_reward/exchange
    ref           VARCHAR(64) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_points_log_member ON points_log(member_id, id DESC);
