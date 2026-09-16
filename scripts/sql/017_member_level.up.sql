-- 017: 会员等级（成长值 + 等级折扣 + 升级礼包）
BEGIN;

ALTER TABLE member
    ADD COLUMN IF NOT EXISTS growth_value  BIGINT   NOT NULL DEFAULT 0, -- 成长值 = 累计实付金额（元取整），只增不减
    ADD COLUMN IF NOT EXISTS level_reached SMALLINT NOT NULL DEFAULT 0; -- 已发放升级礼包的最高等级

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS level_discount NUMERIC(10,2) NOT NULL DEFAULT 0; -- 等级折扣优惠金额快照

-- 回填历史成长值：累计实付（已支付及之后状态）
UPDATE member SET growth_value = COALESCE((
    SELECT FLOOR(SUM(o.pay_amount))::BIGINT FROM orders o
    WHERE o.member_id = member.id AND o.status IN (20, 30, 60)
), 0);

-- 升级礼包券模板（首达 V1/V2/V3 时由后端自动发放；有效期发放日起算由 member_coupon 记录）
INSERT INTO coupon_template (id, name, type, threshold_amount, discount_amount, total_count, per_limit, valid_start, valid_end, status) VALUES
 (5111, 'V1升级礼·满500减30',  1, 500.00,  30.00, 99999, 1, '2026-01-01', '2027-12-31', 1),
 (5112, 'V2升级礼·满1000减80', 1, 1000.00, 80.00, 99999, 1, '2026-01-01', '2027-12-31', 1),
 (5113, 'V3升级礼·满2000减200',1, 2000.00, 200.00, 99999, 1, '2026-01-01', '2027-12-31', 1)
ON CONFLICT (id) DO NOTHING;

COMMIT;
