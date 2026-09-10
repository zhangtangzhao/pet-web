# 贡献指南

感谢关注 Pet（宠物交易平台）。提交 Issue 或 PR 前，请先阅读本指南。

## 行为准则

参与本项目即表示同意 [行为准则](CODE_OF_CONDUCT.md)。

## 如何贡献

### 1. 提交 Issue

- Bug 反馈请使用 [Bug 报告模板](.github/ISSUE_TEMPLATE/bug_report.yml)，尽量附复现步骤、期望行为、日志片段；
- 新功能建议使用 [功能建议模板](.github/ISSUE_TEMPLATE/feature_request.yml)，说明使用场景而非仅描述实现；
- **安全漏洞请勿公开提交**，见 [SECURITY.md](SECURITY.md)。

### 2. 开发流程

```bash
# 前置：Go 1.25+ / Node 18+ / PostgreSQL 16 / Redis 7（或直接 docker compose up -d postgres redis）

# 后端
cd backend && go build ./... && go vet ./...

# 管理端
cd frontend/apps/admin && npm install && npm run build

# 移动端 H5
cd frontend/apps/mobile && npm install && npm run build:h5
```

数据库结构变更：在 `scripts/sql/` 新增版本化迁移（如 `003_xxx.up.sql` / `003_xxx.down.sql`），**禁止**修改已发布的历史迁移；同时同步更新 `docs/03-数据库设计.md`。

### 3. 提交规范

Commit message 采用 Conventional Commits：

```
<type>(<scope>): <subject>

type 取值：
  feat      新功能
  fix       缺陷修复
  docs      文档
  refactor  重构（不改行为）
  sql       数据库迁移
  chore     构建/工具/依赖
  test      测试
```

示例：`feat(trade): 支付回调增加幂等记录表`、`fix(favorite): 取消收藏后计数缓存未回退`。

### 4. PR 要求

- 标题遵循提交规范；一个 PR 聚焦一件事；
- 关联对应 Issue（`Fixes #123`）；
- PR 模板中的自检清单逐项确认；
- 涉及接口变更时同步更新 `docs/04-API接口设计.md`；
- 涉及部署 / 配置变更时同步更新 `scripts/README.md` 与 `.env.example`。

## 代码约定

| 范围 | 约定 |
| ---- | ---- |
| 后端 | go-zero 单体分层（handler → logic → model），业务错误统一走 `internal/common` 错误码，禁止在 handler 写业务逻辑 |
| 主键 | 一律雪花 ID（`common.NewID()`），由应用层显式赋值 |
| 敏感配置 | 只经环境变量注入（`${VAR}` 占位），禁止硬编码或提交明文密钥 |
| 前端 | TypeScript 严格模式；管理端 antd 5；移动端 Taro 4 + React，H5 与小程序共用一套页面代码 |
| 注释与文档 | 中文 |

## 本地验证清单

提 PR 前确认：

- [ ] `go build ./...` 与 `go vet ./...` 通过
- [ ] `npm run build`（admin）与 `npm run build:h5`（mobile）通过
- [ ] 数据库迁移 up/down 成对且可回滚
- [ ] 未提交任何密钥 / `.env` / `config/pet-api.yaml`
- [ ] 涉及的文档已同步更新
