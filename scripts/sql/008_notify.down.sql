-- 008 回滚：通知队列
DROP INDEX IF EXISTS idx_notify_member;
DROP INDEX IF EXISTS idx_notify_pending;
DROP TABLE IF EXISTS notification;
