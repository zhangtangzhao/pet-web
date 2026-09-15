-- ============================================================
-- 008: 微信通知投递队列——客服回复（离线）+ 订单事件（退款/关单/售后结果）
-- 投递器每 10s 扫描待投递行；biz_key 全局唯一保证幂等
-- ============================================================

CREATE TABLE IF NOT EXISTS notification (
    id         BIGINT PRIMARY KEY,                             -- 雪花 ID（应用层生成）
    member_id  BIGINT NOT NULL,
    scene      SMALLINT NOT NULL,                              -- 1客服回复 2订单状态
    biz_key    VARCHAR(64) NOT NULL UNIQUE,                    -- 幂等键 cs:<msgID>/refund:<单号>/close:<单号>/aftersale:<售后单号>
    title      VARCHAR(64) NOT NULL DEFAULT '',
    content    VARCHAR(128) NOT NULL DEFAULT '',               -- 模板 thing 字段（≤20 字）
    order_no   VARCHAR(32) NOT NULL DEFAULT '',
    status     SMALLINT NOT NULL DEFAULT 0,                    -- 0待投递 1已投递 2失败 3降级（未配置模板）
    retry      INT NOT NULL DEFAULT 0,
    channel    SMALLINT NOT NULL DEFAULT 0,                    -- 0无 1小程序订阅消息 2公众号模板消息
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_notify_pending ON notification(id) WHERE status = 0;
CREATE INDEX IF NOT EXISTS idx_notify_member ON notification(member_id, id DESC);
