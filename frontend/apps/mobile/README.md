# mobile · 用户端（微信小程序 / H5）

Taro 4 + React 18，**一套代码编译微信小程序与 H5 双端**。品牌主色 `#ff7a45`，纯原生 CSS，无重型 UI 库依赖（控制小程序包体积）。

## 架构

### 页面

| 页面 | 路径 | 说明 |
| ---- | ---- | ---- |
| 首页 | `pages/index` | 分类导航、精选商品、领券中心入口 |
| 详情 | `pages/detail` | 图片轮播 / 视频、宠物档案、AI 问宠、立即购买 |
| 确认下单 | `pages/checkout` | 增值服务多选、优惠券选择、金额明细实时计算 |
| 领券中心 | `pages/coupon-center` | 领券中心 + 我的优惠券（可用 / 已用 / 已过期） |
| AI 问宠 | `pages/ai-chat` | 基于知识库 + 宠物档案的智能客服会话 |
| 在线客服 | `pages/service-chat` | 人工客服实时聊天：文本 / 图片，弱网自动重连 + 断线消息补拉（`src/ws.ts`） |
| 收藏 | `pages/favorites` | 收藏列表 |
| 登录 | `pages/login` | 手机号验证码 / 微信授权，支持 `redirect` 回跳 |

### 目录结构

```
src/
├── app.config.ts    # 页面注册、窗口配置（新增页面必须在此注册）
├── request.ts       # 统一请求封装：携带 token、401 处理、双 token 存储
├── types.ts         # 与 backend/internal/types 对齐的接口定义
├── pages/           # 页面（Taro 三件套：index.tsx / config.ts / index.css）
└── app.ts / app.css
config/index.ts      # Taro 编译配置：H5 publicPath、router(browser)、dev 代理
```

### 双端差异处理

- **路由**：小程序原生页面栈；H5 为 browser 模式（`config/index.ts` → `h5.router.mode`），nginx `try_files` 兜底 `index.html`
- **登录**：小程序走 `code2session`；H5 走公众号 OAuth2。未登录访问受限页统一 `redirectTo` 登录页并带 `redirect` 参数
- **支付**：下单后小程序用返回的 `payParams` 调 `Taro.requestPayment`；H5 用 `h5PayUrl` 跳转

## 本地开发

```bash
npm install
npm run dev:h5      # H5 → http://127.0.0.1:10086（devServer 已代理 /api → 127.0.0.1:8888）
npm run dev:weapp   # 小程序 watch 构建 → 微信开发者工具导入项目根目录的 dist/
```

前置条件：后端已启动（见 [backend/README.md](../../../backend/README.md)）且数据库已初始化。

## 编译与打包

```bash
npm run build:h5      # H5 产物 → dist/（部署到 nginx）
npm run build:weapp   # 小程序产物 → dist/（微信开发者工具上传）
```

- **H5 发布**：构建后 `dist/` 由 [scripts/deploy/docker-compose.yml](../../../scripts/deploy/docker-compose.yml) 只读挂载进 nginx 容器 `/usr/share/nginx/html/h5`
- **小程序发布**：微信开发者工具打开 `dist/`（appid 在后台上配置），上传代码 → 提交审核。可后续接入 `miniprogram-ci` 做流水线自动上传（预留于 `scripts/dev`）

## Docker 部署与编排

本应用**不单独打镜像**：H5 为纯静态产物，由编排中的 nginx 容器直接托管，随 `scripts/deploy` 一键部署。

```bash
# 在本目录构建产物
npm run build:h5

# 启动 / 更新编排（nginx 挂载 dist）
cd ../../../scripts/deploy && docker compose up -d --build
# 仅更新 H5 静态资源时，重启 nginx 即可
docker compose restart nginx
```

nginx 侧规则（[scripts/deploy/nginx.conf](../../../scripts/deploy/nginx.conf)）：`location /` 托管 H5 并 `try_files` 回退 `index.html`；`/api/` 反代 `api:8888`——生产为同源访问，无跨域问题。

## 注意事项

- **dist 双端共用，会互相覆盖**：`build:weapp` 与 `build:h5` 输出同一 `dist/`。部署 H5 前务必最后执行的是 `build:h5`（或构建后另行拷贝保存），否则 nginx 托管的目录会被小程序产物覆盖
- **API 地址**：H5 dev 走 `config/index.ts` 的 devServer 代理；生产走 nginx 同源 `/api`。请求封装 `request.ts` 不写死域名，若单端独立部署需自行调整 baseURL
- **小程序合法域名**：正式发布前需在微信公众平台配置 request 合法域名（API 域名），否则真机请求失败；开发工具可勾选「不校验合法域名」联调
- **页面注册**：新增页面必须同时建三件套并在 `app.config.ts` 的 `pages` 注册，否则编译报「页面不存在」
- **登录态**：双 token 存 `Taro.getStorageSync`（`request.ts` 统一注入 `Authorization`）；后端返回 40100/40101 时引导重新登录
- **类型对齐**：`types.ts` 与后端 `internal/types` 为手工同步，后端接口字段变更时需同步更新
