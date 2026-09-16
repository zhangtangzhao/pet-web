# scripts · 脚本目录

与业务代码分离，所有非业务脚本统一收口，保证「换环境只动 scripts」。

| 子目录 | 职责 |
| ------ | ---- |
| [sql/](sql/) | PostgreSQL 版本化建表 SQL + 种子数据（表结构见 [docs/03-数据库设计](../docs/03-数据库设计.md)） |
| [deploy/](deploy/) | 部署编排（已就绪，见下） |
| [dev/](dev/) | 本地开发辅助：启动依赖容器、mock 数据、小程序 ci 上传 |

## deploy/ 文件说明

| 文件 | 说明 |
| ---- | ---- |
| [Dockerfile.api](deploy/Dockerfile.api) | 后端多阶段构建：golang:1.25-alpine 构建 → alpine 运行时（非 root 用户，含时区/CA） |
| [docker-compose.yml](deploy/docker-compose.yml) | 一期单机编排：nginx + api + postgres16 + redis7，含健康检查与数据卷 |
| [nginx.conf](deploy/nginx.conf) | H5 / 管理端静态托管 + `/api`、`/admin/api` 反向代理（含客服 WebSocket `/api/ws/` 升级配置） |
| [config/pet-api.yaml.example](deploy/config/pet-api.yaml.example) | 服务配置模板，敏感项全部为 `${VAR}` 占位（JWT/PG/Redis/微信登录/支付/短信/COS） |
| [.env.example](deploy/.env.example) | 环境变量模板（全部密钥的唯一填写入口，勿提交） |
| [monitoring/](deploy/monitoring/) | Prometheus 抓取配置与告警规则（接口错误率 / 时延 / 关单激增 / 支付回调失败 / 限流命中），compose 一键附带 Prometheus(:9090) + Grafana(:3001) |

### 环境变量注入流程

```text
.env  →  docker compose environment 映射  →  api 容器环境变量  →  conf.UseEnv() 展开 pet-api.yaml 中的 ${VAR}
```

- `pet-api.yaml` 中只写 `${VAR}` 占位（go-zero 的 env 展开不支持 `${VAR:-默认值}`，默认值统一写在 compose 的 `environment:` 里）；
- 服务启动时执行生产校验：`APP_MODE=pro` 下 JWT 双 secret / 数据库 / Redis 缺失会拒绝启动，`SMS_TEST_CODE` 非空也拒绝启动；
- 本地开发用 `backend/etc/pet-api.yaml`（明文 dev 值，不占位），两个配置文件互不影响。

## 快速开始（单机）

```bash
cd scripts/deploy

cp config/pet-api.yaml.example config/pet-api.yaml   # 填入真实配置（密钥勿提交）
cp .env.example .env                                  # 生产必须修改默认密码

docker compose up -d --build

# 建表 + 种子数据（容器内 psql，或本机 psql 均可）
docker exec -i $(docker compose ps -q postgres) \
  psql -U pet -d pet -v ON_ERROR_STOP=1 < ../sql/001_init.up.sql
docker exec -i $(docker compose ps -q postgres) \
  psql -U pet -d pet -v ON_ERROR_STOP=1 < ../sql/002_seed.up.sql
```

## sql/ 文件说明

| 文件 | 说明 |
| ---- | ---- |
| [sql/001_init.up.sql](sql/001_init.up.sql) | 全量建表（11 张表 + 索引），幂等建议空库执行 |
| [sql/001_init.down.sql](sql/001_init.down.sql) | 回滚（按依赖逆序 DROP） |
| [sql/002_seed.up.sql](sql/002_seed.up.sql) | 种子数据：管理员 `admin/admin123456`（bcrypt）+ 示例分类/品种/商品 |
| [sql/002_seed.down.sql](sql/002_seed.down.sql) | 回滚种子数据 |
| [sql/003_ai_knowledge.up.sql](sql/003_ai_knowledge.up.sql) | AI 客服知识库表 + 种子条目（平台通用 + 品种专属养护知识） |
| [sql/003_ai_knowledge.down.sql](sql/003_ai_knowledge.down.sql) | 回滚知识库表 |
| [sql/004_marketing.up.sql](sql/004_marketing.up.sql) | 营销表：增值服务 / 优惠券模板 / 用户券 + orders 优惠列 + 种子（服务项与 3 张券模板） |
| [sql/004_marketing.down.sql](sql/004_marketing.down.sql) | 回滚营销表与 orders 优惠列 |
| [sql/005_chat.up.sql](sql/005_chat.up.sql) | 人工客服表：cs_session（每会员一会话）/ cs_message（文本 / 图片消息）+ 索引 |
| [sql/005_chat.down.sql](sql/005_chat.down.sql) | 回滚客服两表 |
| [sql/006_review.up.sql](sql/006_review.up.sql) | 订单评价表 order_review（一单一评唯一约束 + 评分 CHECK + 商品/会员索引） |
| [sql/006_review.down.sql](sql/006_review.down.sql) | 回滚评价表 |
| [sql/007_after_sale.up.sql](sql/007_after_sale.up.sql) | 售后表 after_sale（进行中部分唯一索引 + 状态/会员索引） |
| [sql/007_after_sale.down.sql](sql/007_after_sale.down.sql) | 回滚售后表 |
| [sql/008_notify.up.sql](sql/008_notify.up.sql) | 通知投递队列 notification（biz_key 幂等唯一 + 待投递部分索引） |
| [sql/008_notify.down.sql](sql/008_notify.down.sql) | 回滚通知表 |
| [sql/009_notify_read.up.sql](sql/009_notify_read.up.sql) | 站内消息中心：notification 加 read_at 已读列 + 未读部分索引 |
| [sql/009_notify_read.down.sql](sql/009_notify_read.down.sql) | 回滚已读列与未读索引 |
| [sql/010_banner.up.sql](sql/010_banner.up.sql) | 首页 Banner 运营位表 home_banner（jump_type 白名单 + 上架部分索引）+ 种子 |
| [sql/010_banner.down.sql](sql/010_banner.down.sql) | 回滚 Banner 表 |
| [sql/011_review_reply.up.sql](sql/011_review_reply.up.sql) | order_review 加官方回复列（reply / replied_at） |
| [sql/011_review_reply.down.sql](sql/011_review_reply.down.sql) | 回滚官方回复列 |
| [sql/012_delivery.up.sql](sql/012_delivery.up.sql) | 配送：ship_method 表（自提/托运 + 种子 3 条）+ orders 配送快照 8 列（方式/运费/地址/子状态/运单号/时间） |
| [sql/012_delivery.down.sql](sql/012_delivery.down.sql) | 回滚 ship_method 与 orders 配送列 |
| [sql/013_member_growth.up.sql](sql/013_member_growth.up.sql) | 用户增长：member_address 地址簿 + points_log 积分流水（append-only）+ member 加积分/邀请码列 + 种子 |
| [sql/013_member_growth.down.sql](sql/013_member_growth.down.sql) | 回滚增长表与 member 列 |
| [sql/014_trade_ext.up.sql](sql/014_trade_ext.up.sql) | 交易扩展：flash_sale 秒杀表（每商品至多一个启用中部分唯一索引）+ orders 定金/保障卡列 + service_item 保障天数 + coupon_template 积分兑换 |
| [sql/014_trade_ext.down.sql](sql/014_trade_ext.down.sql) | 回滚秒杀表与交易扩展列 |
| [sql/015_ops.up.sql](sql/015_ops.up.sql) | 运营保障：admin_audit_log 管理端操作审计 + sensitive_word 敏感词库（种子 3 条） |
| [sql/015_ops.down.sql](sql/015_ops.down.sql) | 回滚审计与敏感词表 |
| [sql/016_supplier_compliance.up.sql](sql/016_supplier_compliance.up.sql) | 供货商 supplier 表 + pet_product 合规/防疫列（供货商/检疫证明/疫苗驱虫到期日）+ 多角色管理员种子（operator/support） |
| [sql/016_supplier_compliance.down.sql](sql/016_supplier_compliance.down.sql) | 回滚供货商表、商品扩展列与角色种子 |
| [sql/017_member_level.up.sql](sql/017_member_level.up.sql) | 会员等级：member 成长值/level_reached 列 + orders 等级折扣快照 + 历史成长值回填 + 升级礼包券种子（5111-5113） |
| [sql/017_member_level.down.sql](sql/017_member_level.down.sql) | 回滚等级列与升级礼包券种子 |

说明：

- 静态站点：构建后 `frontend/apps/mobile/dist`（H5）与 `frontend/apps/admin/dist`（管理端，构建已内置 `base=/admin/`）由 nginx 直接托管，本地开发阶段目录不存在时对应路由不可用属正常；
- API 调试：api/postgres/redis 端口均只绑定 `127.0.0.1`，生产对外仅暴露 nginx 的 80/443；
- HTTPS：放开 compose 中 443 映射并挂载证书目录，或在云上由 LB/CDN 终结 TLS。

状态：deploy/ 与 sql/ 已就绪（SQL 已在 PostgreSQL 16 实库验证）；dev/ 辅助脚本按需补充。
