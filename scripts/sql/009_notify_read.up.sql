-- ============================================================
-- 009: 站内消息中心——notification 增加 read_at（站内已读，与微信投递状态解耦）
-- 未读部分索引支撑消息中心角标/计数查询
-- ============================================================

ALTER TABLE notification ADD COLUMN IF NOT EXISTS read_at TIMESTAMPTZ NULL;
CREATE INDEX IF NOT EXISTS idx_notify_unread ON notification(member_id) WHERE read_at IS NULL;
