<!-- 标题遵循 Conventional Commits，如 feat(trade): xxx -->

## 变更说明

<!-- 做了什么、为什么做；关联 Issue 用 Fixes #123 -->

## 变更类型

- [ ] 新功能（feat）
- [ ] 缺陷修复（fix）
- [ ] 文档（docs）
- [ ] 重构（refactor，不改行为）
- [ ] 数据库迁移（sql）
- [ ] 构建 / 依赖（chore）

## 自检清单

- [ ] `go build ./...` 与 `go vet ./...` 通过（后端改动）
- [ ] `npm run build`（admin）/ `npm run build:h5`（mobile）通过（前端改动）
- [ ] 数据库迁移 up / down 成对且已在本地验证（如涉及）
- [ ] 同步更新了对应文档（`docs/04-API接口设计.md`、`docs/03-数据库设计.md`、`scripts/README.md` 等）
- [ ] 未提交任何密钥 / `.env` / `config/pet-api.yaml`
- [ ] 无敏感信息输出到日志（手机号已脱敏等）
