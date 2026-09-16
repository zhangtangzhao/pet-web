# mobile · 用户端（微信小程序 / H5）

Taro 4 + React 18，**一套代码编译微信小程序与 H5 双端**。品牌主色 `#ff7a45`，纯原生 CSS，无重型 UI 库依赖（控制小程序包体积）。

## 架构

### 页面

| 页面 | 路径 | 说明 |
| ---- | ---- | ---- |
| 首页 | `pages/index` | 假搜索框跳列表、banner 运营位（`/api/home` 数据驱动按 jump_type 跳转）、分类导航跳列表、限时秒杀专区（秒杀价 + 剩余名额）、最近看过横滑（浏览历史 ≤10）、精选商品；分享卡片带邀请码参数 |
| 商品列表 | `pages/list` | 搜索框聚焦弹搜索面板（联想下拉 + 搜索历史 chips + 热搜榜 Top10）+ 分类 chips + 排序（默认/销量/价格升降），触底翻页 + 下拉刷新 |
| 我的 | `pages/me` | 头像昵称 / 手机号打码 / 注册时间 + 会员等级卡（等级徽章 / 成长值进度条 / 折扣）+ 入口列表（订单/评价/售后/优惠券/收藏/消息/智能选宠/客服）+ 退出登录 |
| 我的评价 | `pages/my-reviews` | 游标分页；含商家回复与平台隐藏标记 |
| 我的售后 | `pages/my-aftersales` | 分页列表（状态胶囊 / 金额 / 原因 / 平台备注），点击进售后进度页 |
| 详情 | `pages/detail` | 图片轮播 / 视频、宠物档案、AI 问宠、用户评价区（评分摘要 + 全部评价半屏弹层 + 商家回复）、看了又看横滑推荐、分享卡片带邀请码、立即购买 |
| 确认下单 | `pages/checkout` | 增值服务多选、配送方式单选（自提免填地址 / 托运选地址簿 chips 或手填）、优惠券、定金锁宠开关（禁用券，展示定金 + 尾款）、秒杀价自动生效、金额明细（含运费与会员折扣行，与后端公式一致） |
| 我的订单 | `pages/orders` | 状态 chips（含待补尾款）；待支付剩余时间小字；卡片点击进详情；去支付 / 补尾款 / 取消并退定金 / 确认收货 / 申请售后 / 撤销售后 / 去评价（动作收敛在 `src/orderActions.ts`，小程序端点击内同步请求订阅消息授权） |
| 订单详情 | `pages/order-detail` | 状态时间线（下单→支付→发货→送达→完成，自提单跳过发货/送达）、待支付 1s 倒计时、待补尾款紫色状态卡、定金明细、会员优惠行、配送信息与售后进度卡、健康证书入口、全套订单操作（演示环境无支付参数自动走 mock-pay） |
| 电子健康证书 | `pages/certificate` | 已完成订单的精美证书卡（证书编号 / 保障期 / 检疫证明预览），截图保存 |
| 订单评价 | `pages/order-review` | 星级 + 文字 + 图片（COS 直传，≤9 张），一单一评 |
| 售后服务 | `pages/after-sale` | 发起退款申请（默认全额原路退回）/ 状态进度与撤销 |
| 消息中心 | `pages/notify` | 站内消息列表（未读红点 / 单条已读乐观更新 / 已读置灰 / 全部已读 / 下拉刷新），场景图标区分客服 / 订单 / 优惠券提醒 |
| 领券中心 | `pages/coupon-center` | 领券中心 + 我的优惠券（可用 / 已用 / 已过期）；积分兑换券需确认消耗积分 |
| 积分签到 | `pages/points` | 积分余额 + 每日签到 + 积分明细流水，积分兑好礼跳领券中心 |
| 邀请有礼 | `pages/invite` | 邀请码复制 + 已邀人数 / 累计积分统计 + 玩法说明（双方得积分）；首页 / 详情分享卡片自动带邀请码参数，新用户落地缓存、登录时归因（仅新注册生效） |
| 地址簿 | `pages/address` | 收货地址卡列表（默认标记 / 设默认 / 删除）+ 内联新增编辑，checkout 托运地址与之同源 |
| AI 问宠 | `pages/ai-chat` | 基于知识库 + 宠物档案的智能客服会话 |
| 智能选宠 | `pages/recommend` | 条件描述 + 快捷标签 → 推荐 3 只（推荐分 + 理由，跳详情） |
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
- **订阅消息**：小程序端模板 ID 进页时经 `GET /api/notify/tmpl` 预取，必须在用户**点击回调内同步**调 `Taro.requestSubscribeMessage`；H5 公众号模板消息无需订阅动作

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
