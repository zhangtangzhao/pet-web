-- 009 回滚：站内已读列与未读索引

DROP INDEX IF EXISTS idx_notify_unread;
ALTER TABLE notification DROP COLUMN IF EXISTS read_at;
