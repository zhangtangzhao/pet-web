# admin · 平台管理端（PC）

React 18 + Vite 6 + Ant Design 5 + React Router 6，仅 PC 浏览器访问，全部接口走 `/api/admin/*`（用户端接口不可调）。品牌主色 `#ff7a45`。

## 架构

### 页面

| 路由 | 页面 | 说明 |
| ---- | ---- | ---- |
| `/login` | 登录 | 用户名密码 + 图形验证码 |
| `/` | Dashboard | 在售 / 今日订单 / 今日 GMV / 会员数 + **库存预警清单**（在售库存 ≤ 阈值红色高亮） |
| `/products` | 商品管理 | 商品 CRUD、上下架、锁定 / 售出状态；表单含供货商 Select / 检疫证明 URL / 疫苗与驱虫到期日 / **详情长图 / 库存预警阈值 / SKU 规格编辑器** |
| `/catalog` | 分类与品种 | 分类 / 品种 CRUD |
| `/orders` | 订单管理 | 详情（含优惠 / 会员折扣 / 定金明细 + 配送信息 + 健康保障天数 + **自提核销码**）、退款、托运发货（运单号必填）/ 标记送达 / 打印托运单 / **自提核销（6 位码 20→30）**、导出 CSV |
| `/banner` | Banner 管理 | 首页运营位 CRUD：跳转类型（10 种白名单）+ 参数联动输入、上架 Switch、删除 |
| `/shipping` | 配送方式 | 到店自提 / 托运配送 CRUD：运费、排序、启用停用（停用后用户端不可见，历史订单靠快照） |
| `/aftersale` | 售后管理 | 状态筛选；审核：同意（金额默认全额可调）/ 拒绝（备注必填），同意即触发原路退款 |
| `/reviews` | 评价管理 | 全量评价（游标分页）；显示 / 隐藏 Switch、删除违规评价、官方回复（可覆盖更新，用户端评价区展示）、**三维评分列** |
| `/members` | 会员管理 | 搜索、等级列（等级 Tag + 成长值）、启用 / 禁用、**黑名单拉黑 / 解除**（可登录禁交易）、会员导出 CSV |
| `/finance` | 财务对账 | 交易流水（订单号搜索 + 分页）/ 每日净收入汇总（7-90 天切换）；导出订单 / 会员 / 积分 CSV（UTF-8 BOM） |
| `/suppliers` | 供货商管理 | 供货商 CRUD（关键词搜索、启用停用、在售商品计数；被商品引用禁删 41901） |
| `/knowledge` | AI 知识库 | 平台通用 + 品种专属养护知识维护 |
| `/marketing` | 营销管理 | 优惠券模板（新建 / 编辑 / 定向发放 / 删除，积分兑换券）、增值服务（含健康保障天数） |
| `/flash-sales` | 秒杀活动 | 秒杀活动 CRUD（商品远程搜索、秒杀价 / 名额 / 时间窗、每商品至多一个启用中） |
| `/group-buys` | 拼团活动 | 拼团活动 CRUD（商品远程搜索、拼团价 / 成团人数 2-10 / 时限 1-72h；超时未成团自动退款） |
| `/report` | 经营报表 | 今日 / 本月 GMV 与订单、近 N 天成交趋势、商品销量 Top10、品类分布 |
| `/ops` | 审计 / 敏感词 | 管理端操作审计日志（非读操作自动落库）、敏感词库管理（聊天拦截 + 评价/售后打码）、**风控拦截记录**（下单频控 / 评价联系方式 / 黑名单） |
| `/support` | 人工客服 | 公共会话池工作台：实时收发、未读角标、图片回复、结束 / 自动重开会话 |

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
- **权限**：多角色 RBAC 以服务端为准（super_admin 全部 / operator 禁退款与账号管理 / support 客服白名单 + 只读），前端仅做登录态守卫，越权 403 统一 toast

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
