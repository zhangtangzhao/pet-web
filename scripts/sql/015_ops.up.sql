-- 015 运营合规：管理端审计日志 + 敏感词
CREATE TABLE IF NOT EXISTS admin_audit_log (
    id         BIGINT PRIMARY KEY,
    admin_id   BIGINT NOT NULL,
    admin_name VARCHAR(64)  NOT NULL DEFAULT '',
    method     VARCHAR(8)   NOT NULL,
    path       VARCHAR(128) NOT NULL,
    ip         VARCHAR(64)  NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_audit_admin ON admin_audit_log(admin_id, id DESC);
CREATE INDEX IF NOT EXISTS idx_audit_time ON admin_audit_log(id DESC);

CREATE TABLE IF NOT EXISTS sensitive_word (
    id         BIGINT PRIMARY KEY,
    word       VARCHAR(64) NOT NULL,
    status     SMALLINT NOT NULL DEFAULT 1,  -- 1启用 0停用
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_sensitive_word_status ON sensitive_word(status);
INSERT INTO sensitive_word (id, word, status) VALUES
 (8001,'线下交易',1),
 (8002,'加微信',1),
 (8003,'私下转账',1)
ON CONFLICT (id) DO NOTHING;
