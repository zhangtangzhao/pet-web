-- ============================================================
-- 005: 人工客服——会话 + 消息（WebSocket 实时聊天）
-- 发送方 1会员 2客服；消息类型 1文本 2图片
-- ============================================================

-- 客服会话：每位会员一个会话，结束后再来新消息自动重开
CREATE TABLE IF NOT EXISTS cs_session (
    id                BIGINT PRIMARY KEY,                      -- 雪花 ID（应用层生成）
    member_id         BIGINT NOT NULL UNIQUE REFERENCES member(id),
    status            SMALLINT NOT NULL DEFAULT 1,             -- 1进行中 2已结束
    unread_admin      INT NOT NULL DEFAULT 0,                  -- 客服侧未读数
    unread_member     INT NOT NULL DEFAULT 0,                  -- 会员侧未读数
    last_message_text VARCHAR(128) NOT NULL DEFAULT '',        -- 会话列表预览（图片为 [图片]）
    last_message_at   TIMESTAMPTZ NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cs_session_status ON cs_session(status, last_message_at);

-- 客服消息
CREATE TABLE IF NOT EXISTS cs_message (
    id           BIGINT PRIMARY KEY,                           -- 雪花 ID（应用层生成）
    session_id   BIGINT NOT NULL REFERENCES cs_session(id),
    sender_role  SMALLINT NOT NULL,                            -- 1会员 2客服
    sender_id    BIGINT NOT NULL,
    msg_type     SMALLINT NOT NULL DEFAULT 1,                  -- 1文本 2图片
    content      TEXT NOT NULL DEFAULT '',                     -- 文本内容 / 图片 URL
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_cs_message_session ON cs_message(session_id, id);
