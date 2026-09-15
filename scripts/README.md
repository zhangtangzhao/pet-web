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

说明：

- 静态站点：构建后 `frontend/apps/mobile/dist`（H5）与 `frontend/apps/admin/dist`（管理端，构建已内置 `base=/admin/`）由 nginx 直接托管，本地开发阶段目录不存在时对应路由不可用属正常；
- API 调试：api/postgres/redis 端口均只绑定 `127.0.0.1`，生产对外仅暴露 nginx 的 80/443；
- HTTPS：放开 compose 中 443 映射并挂载证书目录，或在云上由 LB/CDN 终结 TLS。

状态：deploy/ 与 sql/ 已就绪（SQL 已在 PostgreSQL 16 实库验证）；dev/ 辅助脚本按需补充。
