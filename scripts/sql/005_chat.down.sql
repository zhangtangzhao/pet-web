-- 005 回滚：人工客服会话与消息
DROP INDEX IF EXISTS idx_cs_message_session;
DROP INDEX IF EXISTS idx_cs_session_status;
DROP TABLE IF EXISTS cs_message;
DROP TABLE IF EXISTS cs_session;
