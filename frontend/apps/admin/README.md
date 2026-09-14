# admin · 平台管理端（PC）

React 18 + Vite 6 + Ant Design 5 + React Router 6，仅 PC 浏览器访问，全部接口走 `/api/admin/*`（用户端接口不可调）。品牌主色 `#ff7a45`。

## 架构

### 页面

| 路由 | 页面 | 说明 |
| ---- | ---- | ---- |
| `/login` | 登录 | 用户名密码 + 图形验证码 |
| `/` | Dashboard | 在售 / 今日订单 / 今日 GMV / 会员数 |
| `/products` | 商品管理 | 商品 CRUD、上下架、锁定 / 售出状态 |
| `/catalog` | 分类与品种 | 分类 / 品种 CRUD |
| `/orders` | 订单管理 | 详情（含优惠明细）、退款 |
| `/members` | 会员管理 | 搜索、启用 / 禁用 |
| `/knowledge` | AI 知识库 | 平台通用 + 品种专属养护知识维护 |
| `/marketing` | 营销管理 | 优惠券模板（新建 / 编辑 / 定向发放 / 删除）、增值服务 |

### 目录结构

```
src/
├── main.tsx            # 入口：antd ConfigProvider（中文 + 主题）+ BrowserRouter
├── App.tsx             # 路由表与登录守卫
├── api/
│   ├── client.ts       # axios 实例：请求注入 token、401 拦截处理
│   └── admin.ts        # 全量管理端 API 封装
├── pages/              # 页面组件（Layout 为侧边菜单框架）
└── components/         # 通用组件
vite.config.ts          # 构建配置（生产 base=/admin/ + /api 代理）
```

### 关键机制

- **双 token**：accessToken 短期有效、refreshToken 长期，过期 / 401 时自动换取并重放请求（`client.ts` 拦截器统一处理）
- **生产子路径部署**：vite 生产构建 `base=/admin/`（`vite.config.ts` 按 mode 切换），`BrowserRouter` 取 `import.meta.env.BASE_URL` 作为 basename——dev（`/`）与容器部署（`/admin`）开箱即用
- **权限**：单管理员角色，前端仅做登录态守卫，鉴权以服务端 JWT 为准

## 本地开发

```bash
npm install
npm run dev        # → http://127.0.0.1:5173（dev 已代理 /api → 127.0.0.1:8888）
```

前置条件：后端已启动（见 [backend/README.md](../../../backend/README.md)）。种子管理员 `admin / admin123456`。

## 编译与打包

```bash
npm run build      # tsc -b 类型检查 + vite build → dist/
npm run preview    # 本地预览生产产物
```

生产产物资源路径自带 `/admin/` 前缀（构建时按 mode 自动注入，无需手工传 `--base`），直接可被 nginx 挂载。

## Docker 部署与编排

本应用**不单独打镜像**：纯静态产物由编排中的 nginx 容器托管在 `/admin` 子路径，随 `scripts/deploy` 一键部署。

```bash
# 在本目录构建产物
npm run build

# 启动 / 更新编排
cd ../../../scripts/deploy && docker compose up -d --build
# 仅更新管理端静态资源时
docker compose restart nginx
```

nginx 侧规则（[scripts/deploy/nginx.conf](../../../scripts/deploy/nginx.conf)）：

- `location /admin/api/` → 反代 `api:8888`（管理端接口）
- `location /admin` → 托管 `dist/` 并 `try_files` 回退 `/admin/index.html`（SPA 刷新不 404）

## 注意事项

- **路径前缀是硬约定**：产物 base（vite.config）、路由 basename（main.tsx）、nginx `location /admin` 三处必须一致；若要改为根路径部署，三处同步修改
- **接口同源**：生产依赖 nginx 将 `/admin/api/` 反代到后端，前端不配置跨域；本地 dev 依赖 vite proxy
- **图形验证码**：登录需先取 captcha 接口；验证码有过期时间，过期需刷新重输
- **优惠券表单**：类型（满减 / 折扣 / 立减）联动不同字段显隐，折扣类型可设封顶金额；已发放的模板 `issued_count` 不可回退编辑，有领取记录的模板禁止删除（后端 41206）
- **构建体积**：antd 全量引入使单 chunk 偏大（>500kB 警告），如需优化可加 `manualChunks` 拆包，不影响功能
