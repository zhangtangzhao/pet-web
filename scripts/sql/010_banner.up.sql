-- ============================================================
-- 010: 首页运营位 banner——管理端可配，用户端 /api/home 数据驱动
-- jump_type: me/recommend/coupon/orders/notify/favorites/
--            product/category/search/custom
-- ============================================================

CREATE TABLE IF NOT EXISTS pet_banner (
    id         BIGINT PRIMARY KEY,                          -- 雪花 ID（应用层生成）
    title      VARCHAR(32) NOT NULL,
    sub_title  VARCHAR(64) NOT NULL DEFAULT '',
    icon       VARCHAR(255) NOT NULL DEFAULT '',            -- emoji 或图片 URL
    jump_type  VARCHAR(16) NOT NULL DEFAULT 'custom',       -- 跳转类型
    target     VARCHAR(255) NOT NULL DEFAULT '',            -- product→商品ID category→分类ID search→关键词 custom→页面路径
    sort       INT NOT NULL DEFAULT 0,
    status     SMALLINT NOT NULL DEFAULT 1,                 -- 1上架 0下架
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_banner_list ON pet_banner(status, sort ASC, id DESC);

-- 种子数据：迁移原首页硬编码入口 + 「我的」（幂等）
INSERT INTO pet_banner (id, title, sub_title, icon, jump_type, target, sort, status) VALUES
    (6001, '我的', '头像/订单/券一站式管理', '👤', 'me', '', 1, 1),
    (6002, '智能选宠', '说出你的条件，推荐 3 只', '🐾', 'recommend', '', 2, 1),
    (6003, '领券中心', '新人立省 30 元', '🎫', 'coupon', '', 3, 1),
    (6004, '我的订单', '支付/收货/评价/售后', '📦', 'orders', '', 4, 1),
    (6005, '消息中心', '订单/客服/优惠券提醒', '🔔', 'notify', '', 5, 1)
ON CONFLICT (id) DO NOTHING;
