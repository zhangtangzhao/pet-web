# 04 - API 接口设计

> 宠物交易平台 · 接口设计文档 · REST

## 1. 接口规范

### 1.1 路由域

| 域 | 前缀 | 鉴权 | 说明 |
| -- | ---- | ---- | ---- |
| 用户端公开接口 | `/api/v1/open/*` | 无 | 首页、商品列表/详情（游客可浏览） |
| 用户端私有接口 | `/api/v1/*` | 用户 JWT | 收藏、订单、支付、个人信息 |
| 平台端接口 | `/admin/api/v1/*` | 管理 JWT + 角色 | 全部私有 |

### 1.2 统一响应与错误码

```json
// 成功
{ "code": 0, "msg": "ok", "data": { } }

// 失败
{ "code": 40101, "msg": "登录已过期", "data": null }
```

| code 段 | 含义 |
| ------- | ---- |
| 0 | 成功 |
| 40xxx | 客户端错误：40100 未登录 / 40101 过期 / 40300 无权限 / 40400 资源不存在 |
| 41xxx | 业务错误：41001 商品已下架 / 41002 商品已被下单 / 41003 订单状态不允许此操作 / 41004 重复收藏 / 41101 验证码错误或已过期 / 41102 发送过于频繁 / 41103 超过当日发送上限 / 41104 验证码尝试次数过多 / 41203-41209 管理端业务（验证码、密码、引用约束等）/ 41301 AI 提问太频繁 / 41302 AI 当日提问达上限 / 41401 优惠券不可用 / 41402 未达用券门槛 / 41403 券已领完 / 41404 超过限领 / 41405 不在可领时段 / 41406 增值服务不可用 / 41501 会话不存在（含越权访问他人会话）/ 41601 该订单已评价过 / 41602 订单当前不可评价 / 41701 售后单不存在（含越权）/ 41702 已有进行中售后 / 41703 订单不可售后 / 41704 售后状态不允许此操作 / 41705 审核状态冲突请刷新 / 41801 请求过于频繁（限流）/ 41802 内容包含敏感词 / 41803 积分不足 / 41804 定金模式不支持优惠券 / 41805 邀请码无效 / 41806 订单当前不可开具健康证书（仅已完成单）/ 41901 供货商已被商品引用，禁删 / 41902 下单过于频繁（风控）/ 41903 评价包含联系方式（风控）/ 41904 黑名单账号限制操作 / 41905 请选择商品规格 / 41906 商品规格不可用 / 41907 拼团活动不可用 / 41908 购物车无可结算商品 / 41909 核销码错误或订单状态不允许核销 / 41910-41912 拼团下单限制（禁定金/服务/券） |
| 50xxx | 服务端错误：50000 系统异常 / 50001 微信接口异常 / 50002 支付下单失败 / 50003 短信发送失败 / 50005 AI 服务不可用 |

### 1.3 通用约定

- 鉴权：`Authorization: Bearer <access_token>`；
- 分页：请求 `page`（默认1）+ `pageSize`（默认10，最大50）；响应 `total + list`；
- 时间：ISO8601（`2026-09-10T14:00:00+08:00`）；
- 金额：接口层统一用**分（int64）**或字符串小数传输，禁止浮点。

---

## 2. 用户端接口（小程序 / H5）

### 2.1 认证

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/v1/auth/wechat/mini-login` | 小程序登录 `{code}` → token + 会员信息 |
| GET | `/api/v1/auth/wechat/h5-oauth-url?redirect=` | 获取微信网页授权跳转地址 |
| POST | `/api/v1/auth/wechat/h5-login` | H5 OAuth 回调 code 登录 |
| POST | `/api/v1/auth/sms/send` | 发送短信验证码 `{phone}`（60s 冷却、单号日上限） |
| POST | `/api/v1/auth/sms/login` | 手机号验证码登录 `{phone, code, inviteCode?}`，未注册自动注册；`inviteCode` 来自分享卡片带参（仅新注册生效，双方得积分）；支持测试公共验证码（仅非生产环境生效） |
| POST | `/api/v1/auth/refresh` | 刷新 token `{refreshToken}` |
| POST | `/api/v1/auth/logout` | 登出（吊销 refresh_token） |
| GET | `/api/v1/member/profile` | 当前会员信息 |
| PUT | `/api/v1/member/profile` | 更新昵称/头像/性别 |

**mini-login 请求/响应示例**

```json
// POST /api/v1/auth/wechat/mini-login
// Request
{ "code": "0a3XXXkkk23" }

// Response data
{
  "accessToken": "eyJhbGciOi...",
  "refreshToken": "r_9f8e7d...",
  "expiresIn": 7200,
  "member": {
    "id": "1943256789012345678",
    "nickname": "微信用户",
    "avatar": "https://thirdwx.qlogo.cn/...",
    "isNew": true
  }
}
```

**短信验证码登录示例（小程序 / H5 通用）**

```json
// POST /api/v1/auth/sms/send
// Request
{ "phone": "13800000000" }
// Response data
{ "interval": 60 }

// POST /api/v1/auth/sms/login
// Request
{ "phone": "13800000000", "code": "123456" }
// Response data：与 mini-login 完全相同（accessToken / refreshToken / member，isNew=true 表示本次自动注册）
```

### 2.2 首页 / 分类 / 品种

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/v1/home` | 首页聚合：运营位 banners、分类入口、热销推荐 |
| GET | `/api/v1/categories` | 宠物分类列表（含启用品种数） |
| GET | `/api/v1/breeds?categoryId=` | 品种列表（支持按分类过滤） |

### 2.3 宠物商品

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/v1/products` | 商品分页列表（筛选+排序） |
| GET | `/api/v1/products/:id` | 商品详情（含宠物档案、图集、品种卡） |
| POST | `/api/v1/products/:id/view` | 浏览计数（或由详情接口内部完成） |

**列表查询参数**

| 参数 | 类型 | 说明 |
| ---- | ---- | ---- |
| categoryId | int64 | 分类筛选 |
| breedId | int64 | 品种筛选 |
| gender | int | 1公 2母 |
| priceMin / priceMax | int | 价格区间（分） |
| keyword | string | 标题模糊搜索 |
| sort | string | `newest`(默认) / `sales` / `price_asc` / `price_desc` |
| page / pageSize | int | 分页 |

**详情响应示例（节选）**

```json
{
  "id": "1943256789000000001",
  "title": "纯种布偶猫 海双弟弟",
  "price": "3800.00",
  "originalPrice": "4200.00",
  "mainImage": "https://cdn.xxx.com/p/1.jpg",
  "images": ["...1.jpg", "...2.jpg"],
  "videoUrl": "https://cdn.xxx.com/v/1.mp4",
  "videoCover": "https://cdn.xxx.com/v/1_poster.jpg",
  "petProfile": {
    "gender": 1, "genderText": "公",
    "birthDate": "2026-04-12", "ageText": "5个月",
    "vaccineDesc": "三针疫苗已齐", "dewormDesc": "体内外驱虫已完成",
    "bodyType": "中型", "coatColor": "海豹双色",
    "personality": "黏人安静，亲人不怕生",
    "healthDesc": "健康承诺，支持到店验宠"
  },
  "breed": { "id": "1002", "name": "布偶猫", "intro": "..." },
  "category": { "id": "1001", "name": "猫" },
  "status": 1,
  "sales": 36, "favoriteCount": 128,
  "isFavorite": true
}
```

### 2.4 收藏

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/v1/favorites/:productId` | 收藏（幂等） |
| DELETE | `/api/v1/favorites/:productId` | 取消收藏（幂等） |
| GET | `/api/v1/favorites?page=&pageSize=` | 我的收藏列表（商品卡摘要） |

### 2.5 订单与支付

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/v1/orders` | 创建订单 → 返回订单 + 微信支付参数 |
| GET | `/api/v1/orders?status=&page=` | 我的订单列表（Tab：全部/待支付/已支付/已完成） |
| GET | `/api/v1/orders/:orderNo` | 订单详情 |
| POST | `/api/v1/orders/:orderNo/cancel` | 取消订单（仅待支付，释放商品） |
| POST | `/api/v1/orders/:orderNo/confirm` | 确认收货/完成（已支付→已完成） |

> 已实现订单视图新增 `reviewed`（是否已评价）与 `aftersaleStatus`（最新售后状态，0=无售后；1待审核 2已同意 3已拒绝 4已撤销），供订单列表/详情回显「去评价 / 申请售后 / 撤销售后」入口态。
>
> 迁移 012 起订单视图附带配送字段：`shipMethod`（方式名快照）、`shipFee`、`shipAddress`、`shipStatus`（0待配送 1配送中 2已送达）、`shipNo`、`shippedAt` / `deliveredAt` / `completedAt`（有值才返回）；下单请求新增 `shipMethodId`（必填）与 `shipAddress`（托运方式 kind=2 必填，≤255）。实付公式：`max(商品价+服务费-券抵扣, 0.01) + 运费`，运费不参与券抵扣；待支付关单时长 `Trade.PayTimeoutMinutes`（默认 15 分钟，`expireAt` 随之）。
>
> 迁移 014 起下单请求再增 `addressId`（地址簿 ID，优先于手填地址）与 `useDeposit`（定金锁宠：定金 = `（商品价+服务费）× DepositPercent%` 四舍五入 2 位，禁用优惠券；响应 `isDeposit=true` + `tailAmount`）；订单视图附带 `depositAmount / tailExpireAt / guaranteeDays`。命中秒杀的商品以 `salePrice` 计价并记 `flashSaleId`（响应自动体现为更低的 payAmount）。
| POST | `/api/v1/payments/wechat/prepay` | 待支付单重新拉起支付 `{orderNo}` → 支付参数 |
| GET | `/api/v1/payments/:paymentNo/status` | 轮询支付结果（前端支付后 2s 间隔轮询） |
| POST | `/api/v1/payment/notify` | **微信回调**（无鉴权，验签；不走统一响应格式） |

**创建订单请求/响应示例**

```json
// POST /api/v1/orders
// Request
{
  "productId": "1943256789000000001",
  "shipMethodId": "7002",
  "shipAddress": "上海市浦东新区 …（kind=2 托运必填）",
  "contactName": "张三",
  "contactPhone": "13800000000",
  "remark": "希望周末自提"
}

// Response data（订单 + 小程序拉起支付所需参数）
{
  "orderNo": "P20260910143015000123",
  "payAmount": "3800.00",
  "expireAt": "2026-09-10T15:00:15+08:00",
  "payParams": {
    "timeStamp": "1789102215",
    "nonceStr": "a1b2c3",
    "package": "prepay_id=wx201xx...",
    "signType": "RSA",
    "paySign": "..."
  },
  "h5PayUrl": ""   // H5 支付时返回 mweb_url，小程序端为空
}
```

### 2.5.1 配送方式（已实现，实际路径前缀为 /api，迁移 012）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/ship/methods` | 可用配送方式（仅启用，`sort ASC`）：`{id, name, kind(1自提/2托运配送), description, fee}` |

### 2.6 AI 智能客服（已实现，实际路径前缀为 /api）

基于平台知识库（`ai_knowledge` 表）+ 宠物档案的检索增强问答（RAG）。详情接口 `GET /api/products/:id` 返回 `aiEnabled` 标记是否展示 AI 入口。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/ai/ask` | 宠物问答（需登录；间隔 5s + 每日 20 次频控） |

```json
// POST /api/ai/ask
// Request
{
  "productId": "3001",
  "question": "它日常怎么喂养？",
  "history": [ { "role": "user", "content": "..." }, { "role": "assistant", "content": "..." } ]  // 可选，最近对话（服务端截最后6条）
}

// Response data
{
  "answer": "……",
  "refs": [ { "title": "布偶猫喂养与养护要点", "content": "……", "isBreedSpecific": true } ],  // 命中的知识条目
  "demo": false   // true=未接入大模型的演示模式（仅非生产）
}
```

实现要点：服务端组装提示词（品种 intro + 8 项档案字段 + 检索到的知识条目 top4），大模型走 OpenAI 兼容接口（`AI.BaseURL/ApiKey/Model` 三项配置启用）；未配置时非生产走演示模式（知识库+档案模板回答），生产返回 50005 且入口隐藏。

### 2.6.1 智能选宠推荐（已实现，实际路径前缀为 /api）

按用户条件（预算/家庭情况/生活习惯等自由描述）从在售宠物中推荐 3 只，按推荐分（0-100）降序。双通道：配置大模型时走 LLM 结构化推荐（目录 + 条件 → JSON，低温度采样，解析失败/调用失败自动降级），未配置时走本地规则打分（预算正则 + 关键词信号 × 档案字段匹配），**零配置可用**。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/ai/recommend` | 智能选宠（需登录；LLM 路径频控 `RecommendIntervalSeconds`=10s + `RecommendDailyLimit`=5 次/日，41303/41304；规则路径不频控） |

```json
// POST /api/ai/recommend
// Request
{ "requirement": "预算5000以内，住公寓，上班族，第一次养宠，想养温顺安静的猫" }  // ≤500 字

// Response data
{
  "items": [
    { "id": "9101", "title": "…", "mainImage": "", "price": "2800.00", "breedName": "英短蓝猫",
      "petGender": 1, "favoriteCount": 12, "sales": 28, "status": 1,
      "score": 94, "reason": "性格温顺安静，适合公寓和新手，疫苗齐全好养" }
  ],
  "source": "ai"   // ai=大模型推荐 rule=规则打分兜底
}
```

实现要点：在售目录（`status=1` 按销量取前 200，紧凑行含品种/价格/性别/月龄/体型/性格截断）与条件一同送入 LLM，要求只输出 `[{"productId","score","reason"}]`；目录外/重复编号过滤，不足 3 个用规则分候选补位（补位后按分重排）。规则打分：base 60 + 物种/预算/性格词/体型/疫苗/销量加权的 Top3，理由为命中信号模板拼接。

### 2.7 营销：增值服务与优惠券（已实现，实际路径前缀为 /api）

金额模型：`total = 商品价 + 服务费合计`，`pay = total - 券抵扣`（保底 0.01 元）。券生命周期：领取 `1可用` → 下单锁定 `2已锁定`（同事务 CAS，防并发重复用券）→ 支付成功核销 `3已使用`；取消/超时关单自动回滚（未过期回 `1`，已过期置 `4`）；退款不回券。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/services` | 增值服务列表（无需登录，status=1） |
| GET | `/api/coupons/center` | 领券中心（可领的模板，自动过滤已领完/已领过/不在时段） |
| POST | `/api/coupons/:id/claim` | 领取优惠券（CAS 扣库存，41403 领完 / 41404 限领） |
| GET | `/api/coupons?status=` | 我的优惠券（status: 1可用 2锁定 3已使用 4已过期；查询时惰性过期） |
| POST | `/api/coupons/usable` | 某笔消费可用券及预估抵扣 `{productId, serviceIds[]}`（门槛=商品价+服务费） |

```json
// POST /api/coupons/usable Response data
[
  { "id": "2098...", "name": "全场9折券", "type": 2, "threshold": "0.00",
    "discount": "88.00", "discountValid": false, "percent": 90, "validEnd": "..." },
  { "id": "2097...", "name": "新人立减券", "type": 3, "threshold": "0.00",
    "discount": "30.00", "discountValid": true, "percent": 0, "validEnd": "..." }
]
```

建单入参扩展（`POST /api/orders`）：`serviceIds []string`（增值服务 ID）、`couponId string`（用户券 ID，空 = 不用券）；响应 `payAmount` 为折后实付。订单视图（用户端/平台端）新增 `totalAmount / discountAmount / serviceFee / couponInfo / serviceItems`（服务快照 JSON 字符串）。

### 2.8 人工客服（已实现，实际路径前缀为 /api）

模型：每会员一个会话（首条消息自动创建，已结束会话再来新消息自动重开）；文本（≤500 字）与图片（复用 COS 直传）消息；发送走 REST 落库，WebSocket 仅做服务端推送。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/ws/cs?token=<accessToken>` | 建立 WS 长连接（token 走 query，WS 无法带 Authorization 头） |
| POST | `/api/cs/messages` | 发送消息 `{msgType?: 1文本/2图片, content}`（会员自动定位本人会话） |
| GET | `/api/cs/messages?before=&after=&limit=` | 历史消息游标分页：`before` 向上翻页取更早，`after` 断线补拉，均按时间正序返回 + `hasMore`（limit 默认 20 最大 100，无游标时取最新一页） |
| POST | `/api/cs/read` | 已读（清零会员侧未读并广播） |
| POST | `/api/upload-token` | 聊天图片 COS 直传凭证（目录固定 `chat`） |

**WS 协议**：客户端唯一上行帧是 `{"type":"pong"}`（收到 `ping` 即回，服务端据此判活），其余均为服务端推送：

```json
{"type":"ping","ts":1694000000}                  // 服务端每 20s 推送
{"type":"closing_warning","data":"连接即将断开,30s 内未恢复将自动断开"}  // 两阶段判死：>30s 无帧先预警，再等 30s 仍无帧则断开
{"type":"new_message","data":{"id":"1757...","sessionId":"1756...","senderRole":1,"msgType":1,"content":"...","createdAt":"..."}}
{"type":"session_update","data":{"sessionId":"1756...","status":1,"unreadAdmin":3,"unreadMember":0,"lastMessageText":"...","lastMessageAt":"..."}}
```

**弱网设计**（双端同语义）：指数退避（1s→30s 封顶）+ ±30% 抖动自动重连，网络恢复/页面可见立即重连；客户端 45s 无帧判死自愈；断线窗口消息不丢——客户端记录已见最大消息 ID，重连成功后 `?after=` 补拉缺口，全部消息按 ID 幂等去重；发送不受断连影响（REST），失败自动重试 3 次后标红供手动重发（用户端）。

### 2.9 订单评价（已实现，实际路径前缀为 /api）

规则：仅「已完成(status=30)」订单可评，一单一评（`order_no` 唯一）；星级 1-5 + 文字（≤500 字）+ 图片（≤9 张，COS 直传）。管理端隐藏/删除后公开侧不可见；官方回复（`reply`/`repliedAt`）由管理端维护并随评价视图下发，用户端评价区展示「商家回复」。订单/详情视图带 `reviewed` 标记回显入口态。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/orders/:orderNo/review` | 创建评价 `{orderNo, rating(1-5), content?, images?[]}`（41601 已评过 / 41602 不可评） |
| GET | `/api/products/:id/reviews?cursor=&limit=` | 公开评价列表（仅 status=1，游标 id 倒序，含会员昵称/头像） |
| GET | `/api/products/:id/review-summary` | 评分摘要 `{avgRating:"5.0", total, latest[3]}`（隐藏的不计入） |
| GET | `/api/member/reviews?cursor=&limit=` | 我的评价（本人全部，含被隐藏的，游标分页） |

### 2.10 售后申请（已实现，实际路径前缀为 /api）

规则：已支付(20)/已完成(30) 订单可发起，默认全额退款、原路退回，**管理端审核时可调金额**；同单同时仅一个进行中售后（41702）；审核同意即触发退款（确定性退款单号 `RF+售后单号` 幂等），拒绝须填备注。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/aftersale` | 发起 `{orderNo, reason(≤500字)}`（41703 订单不可售后 / 41702 已有进行中） |
| GET | `/api/aftersale?orderNo=` | 查订单当前售后（本人，最新一条；无则 `data:null`） |
| GET | `/api/aftersale/list?page=&pageSize=` | 我的售后单列表（分页 id 倒序，size≤50 默认 10） |
| POST | `/api/aftersale/:afterSaleNo/cancel` | 撤销（仅本人待审核单；41701 不存在/越权 / 41704 状态不允许） |

售后单状态：`1待审核 2已同意 3已拒绝 4已撤销`（`statusText` 同步返回）。

### 2.11 通知订阅与消息中心（已实现，实际路径前缀为 /api）

微信订阅消息（小程序）/模板消息（公众号 H5）离线触达：客服回复（仅用户不在线时）+ 订单事件（退款到账 / 超时关单 / 售后审核结果 / 自动确认收货）+ 优惠券到期提醒（3 天内到期，`biz_key=couponexp:<用户券ID>` 恰好一次）+ 疫苗/驱虫到期关怀提醒（scene=4，`biz_key=vaccine:v:{商品ID}:{日期}` / `deworm:w:{...}`）。服务端投递队列见文档 03 §3.20；模板未配置自动降级（不影响业务，站内消息中心不受影响）。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/notify/tmpl` | 返回双端模板 ID 配置 `{miniTmplCsReply, miniTmplOrder, h5TmplCsReply, h5TmplOrder}`（未配置为空串） |
| GET | `/api/notify/list?page=&pageSize=` | 站内消息中心：分页 `id DESC` + `unread` 未读数；`scene` 1客服回复/2订单/3优惠券，`readAt` 空即未读 |
| POST | `/api/notify/read` | 全部已读（未读行 `read_at=now`） |
| POST | `/api/notify/:id/read` | 单条已读（幂等；非本人/不存在返回 40400） |

> 小程序端必须在**用户点击回调内同步**调用 `wx.requestSubscribeMessage`（Taro 同名 API）订阅，模板 ID 从本接口预取（进入页面时缓存）；H5 端公众号模板消息无需订阅动作。已支付订单超过 `Trade.AutoConfirmDays`（默认 7 天，配置 ≤0 关闭）天未确认将自动完成并产生「订单已确认完成」通知。

### 2.12 用户增长（已实现，实际路径前缀为 /api，迁移 013/014）

**地址簿**

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/addresses` | 我的地址（默认在前，更新时间倒序） |
| POST | `/api/addresses` | 新建/编辑 `{id?, name, phone, address, isDefault?}`（设默认在同会员内先清后设） |
| PUT | `/api/addresses/:id/default` | 设为默认 |
| DELETE | `/api/addresses/:id` | 删除 |

**积分 / 邀请**

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/points/signin` | 每日签到 `{ok, balance}`（ok=false=今日已签） |
| GET | `/api/points?page=` | 积分余额 + 流水（append-only：sign/review/order/invite/exchange） |
| GET | `/api/invite` | 邀请概览 `{inviteCode, invited, rewardEach}`；新用户注册填邀请码，双方各得积分 |

**秒杀（迁移 014）**

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/flash-sales` | 生效中的秒杀活动（公开）：`{id, productId, productTitle, productImage, salePrice, stock, sold, startAt, endAt}`；下单自动按秒杀价计价，无需传参 |

**交易配置 / 演示支付**

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/trade-config` | 公开配置 `{depositPercent, depositHoldDays}`；携带用户 token 时附带 `myLevelName / myDiscount / myGrowthValue`（未登录或 V0 缺省） |
| POST | `/api/orders/:orderNo/mock-pay` | **演示支付**：未配商户号且非生产时直接落账（生产 404，前端自动回退微信支付提示） |

### 2.13 搜索增强（已实现，实际路径前缀为 /api，Redis 支撑）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/search/hot` | 热搜榜（公开）：`{list: [{keyword, score}]}`，Redis ZSET 7 天滑动窗口 |
| GET | `/api/search/complete?prefix=` | 搜索联想（公开）：品种名 5 + 分类名 3 + 标题命中 5，去重后 ≤10 条 `{list: string[]}` |
| POST | `/api/search/trace` | 记录搜索词（需登录）`{keyword}`：写热搜 ZSET + 个人历史（LIST 保留 10 条，7 天） |
| GET | `/api/search/history` | 我的搜索历史（时间倒序 ≤10 条） |
| DELETE | `/api/search/history` | 清空我的搜索历史 |

### 2.14 浏览历史与相关推荐（已实现，实际路径前缀为 /api）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/history/views?page=&pageSize=` | 浏览历史（需登录）：Redis LIST 记录最近 50 个商品（去重置顶），返回仅含在售商品卡 |
| GET | `/api/products/:id/related` | 看了又看（公开，登录后记入浏览历史）：同品种 → 同分类 → 热销兜底，共 4 个 |

> 商品详情接口内部异步记录浏览历史（登录态），浏览计数不变。

### 2.15 会员等级（已实现，实际路径前缀为 /api，迁移 017）

等级引擎（纯代码）：V1 ≥1000（98 折）、V2 ≥5000（95 折）、V3 ≥20000（92 折）；成长值 = 累计实付金额（元取整），支付落账只增不减。下单计价 `实付 = max(商品价+服务费-券抵扣, 0.01) × 等级折扣率 + 运费`，订单视图 `levelDiscount` 为等级优惠快照。首达 V1/V2/V3 恰好发放一次升级礼包券（5111/5112/5113）并推送通知。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/level` | 我的等级：`{growthValue, level, levelName, discount, nextThreshold, nextDiscount?}`（`nextThreshold=0` 表示已是顶级） |

### 2.16 电子健康证书（已实现，实际路径前缀为 /api）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/orders/:orderNo/certificate` | 本人已完成订单的健康证书：`{certNo("HC-"+orderNo), orderNo, memberMasked, productTitle, breedName, completedAt, guaranteeDays, guaranteeEndAt?, quarantineUrl?, supplierName?}`（41806 订单当前不可开具） |

> 证书页为 mobile 端精美卡片（截图保存），保障期 = 完成时间 + `guaranteeDays`；检疫证明与供货商名称有值才返回。

### 2.17 购物车（已实现，实际路径前缀为 /api，迁移 018）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/cart` | 我的购物车：`{list:[{id, productId, productTitle, productImage, price, origPrice, skuId?, skuSpecs?, checked, onSale, createdAt}], checkedCount}`（下架商品 `onSale=false` 置灰） |
| POST | `/api/cart` | 加购：`{productId, skuId?}`；同商品重复加购幂等回到勾选态（41906 规格不可用） |
| PUT | `/api/cart/:id` | 勾选切换：`{checked?: 0/1}` |
| DELETE | `/api/cart/:id` | 删除单项 |
| DELETE | `/api/cart` | 清空购物车 |

> 批量结算：`POST /api/orders` 传 `cartIds:[...]`（≤20）多商品一单；批量结算禁用增值服务与定金，券/等级折扣按整单生效。规格商品单购需传 `skuId`（41905 未选规格）。

### 2.18 拼团（已实现，实际路径前缀为 /api，迁移 018）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/group-buys` | 进行中的拼团活动：`{list:[{id, productId, productTitle, productImage, price, origPrice, size, hours, openTeams}]}` |
| POST | `/api/orders` | 拼团下单：`{productId, groupBuyId, skuId?, shipMethodId, ...}`（事务内自动凑团/开团，拼团价成交，禁券/定金/增值服务；超时未成团自动退款，`groupTeamId` 随订单返回） |

### 2.19 宠物档案（已实现，实际路径前缀为 /api，迁移 018）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/pets` | 我的宠物档案列表（品种/生日/体重/疫苗/驱虫到期日） |
| POST | `/api/pets` | 新建：`{name, breedName?, gender?, birthday?, weight?, vaccineAt?, nextVaccineDate?, nextDewormDate?, remark?}` |
| PUT | `/api/pets/:id` | 编辑（body 同上） |
| DELETE | `/api/pets/:id` | 删除 |

> 档案的 `next_vaccine_date` / `next_deworm_date` 进入每日护理提醒扫描（3 天内到期 → scene=4 关怀通知，biz_key 幂等）。

---

## 3. 平台端接口（/admin/api/v1，PC）

### 3.1 认证与账号

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/admin/api/v1/captcha` | 图形验证码 |
| POST | `/admin/api/v1/auth/login` | 账号密码登录 |
| POST | `/admin/api/v1/auth/refresh` | 刷新 token |
| GET | `/admin/api/v1/admin/profile` | 当前管理员信息 |
| PUT | `/admin/api/v1/admin/password` | 修改密码 |

### 3.2 商品管理

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/admin/api/v1/products` | 商品列表（多条件筛选：状态/分类/品种/关键字） |
| GET | `/admin/api/v1/products/:id` | 商品详情 |
| POST | `/admin/api/v1/products` | 创建商品（草稿；含 `supplierId?`/`quarantineCertUrl?`/`nextVaccineDate?`/`nextDewormDate?`（yyyy-MM-dd），迁移 016） |
| PUT | `/admin/api/v1/products/:id` | 编辑商品（同上，已售出/锁定单禁改） |
| PUT | `/admin/api/v1/products/:id/status` | 上架/下架 `{status}` |
| POST | `/admin/api/v1/media/upload-token` | 签发 COS 直传凭证 `{dir, contentType}` |
| GET | `/admin/api/v1/dashboard/overview` | 看板：在售数/今日订单/GMV/会员数 |

### 3.3 分类与品种管理

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET / POST | `/admin/api/v1/categories` | 列表 / 新建 |
| PUT / DELETE | `/admin/api/v1/categories/:id` | 编辑 / 删除（有关联商品则禁删） |
| GET / POST | `/admin/api/v1/breeds` | 列表（categoryId 过滤）/ 新建 |
| PUT / DELETE | `/admin/api/v1/breeds/:id` | 编辑 / 删除 |

### 3.4 订单管理

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/admin/api/v1/orders` | 订单列表（状态/时间/订单号/会员筛选，支持导出） |
| GET | `/admin/api/v1/orders/:orderNo` | 订单详情（含支付流水） |
| POST | `/admin/api/v1/orders/:orderNo/refund` | 发起退款 `{reason, amount}`（已支付/已完成） |
| POST | `/admin/api/v1/orders/:orderNo/ship` | 托运发货 `{shipNo}`（必填，≤64；仅已支付且待配送，自提单拒绝；重复发货幂等拒绝） |
| POST | `/admin/api/v1/orders/:orderNo/deliver` | 标记送达（仅配送中；落 `deliveredAt`） |
| GET | `/admin/api/v1/orders/export` | 订单导出 Excel |

> 发货 / 送达均按 `ship_status` CAS 推进并触发订阅消息（`ship:{orderNo}` / `deliver:{orderNo}` 幂等）；退款含运费全额原路退，无需单独处理。

### 3.5 会员管理

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/admin/api/v1/members` | 会员列表（注册时间/状态筛选） |
| GET | `/admin/api/v1/members/:id` | 会员详情（订单数/收藏数） |
| PUT | `/admin/api/v1/members/:id/status` | 启用/禁用 |

### 3.6 AI 知识库管理（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/ai/knowledge` | 知识条目分页（breedId/keyword 筛选） |
| POST | `/api/admin/ai/knowledge` | 新建条目 `{breedId?, title, keywords?, content, sort?, status?}` |
| PUT | `/api/admin/ai/knowledge/:id` | 编辑条目（breedId 空 = 平台通用） |
| DELETE | `/api/admin/ai/knowledge/:id` | 删除条目 |

### 3.7 营销管理（已实现，实际路径前缀为 /api/admin/marketing）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET / POST | `/api/admin/marketing/coupons` | 券模板分页（status/keyword） / 新建 |
| PUT / DELETE | `/api/admin/marketing/coupons/:id` | 编辑（已发放量不可改）/ 删除（有发放记录禁删 41206） |
| POST | `/api/admin/marketing/coupons/:id/issue` | 定向发放 `{memberIds[]}` → `{issued, failed}`（逐人按限领） |
| GET / POST | `/api/admin/marketing/services` | 增值服务列表 / 新建 `{name, description?, originalPrice?, price, guaranteeDays?(健康保障天数,0=无), sort?, status?}` |
| PUT / DELETE | `/api/admin/marketing/services/:id` | 编辑 / 删除 |

券模板入参：`{name, type(1满减/2折扣/3立减), thresholdAmount?, discountAmount?, discountPercent?, maxDiscountAmount?, totalCount?(0不限), perLimit?, newUserOnly?, pointsCost?(>0 需积分兑换，领取时事务内扣分), pickupStart?, pickupEnd?, validStart?, validEnd?, status?}`，时段格式 `YYYY-MM-DD HH:mm:ss` 或 RFC3339，空 = 不限。

### 3.8 人工客服工作台（已实现，实际路径前缀为 /api/admin）

公共会话池：所有登录管理员可回复任意会话；`GET /api/ws/cs` 同一连接（管理员 token 即客服身份）。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/cs/sessions` | 会话列表（进行中在前，按最后消息时间倒序，含会员昵称/手机号/头像摘要） |
| GET | `/api/admin/cs/sessions/:id/messages?before=&after=&limit=` | 历史消息（游标分页，同 2.8） |
| POST | `/api/admin/cs/sessions/:id/messages` | 回复 `{msgType?, content}` |
| POST | `/api/admin/cs/sessions/:id/read` | 已读（清零客服侧未读并广播） |
| POST | `/api/admin/cs/sessions/:id/close` | 结束会话（幂等；用户再发消息自动重开） |

### 3.9 评价管理（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/reviews?cursor=&limit=` | 全量评价（含隐藏，游标 id 倒序） |
| PUT | `/api/admin/reviews/:id/status` | 显示/隐藏 `{status: 1显示 0隐藏}` |
| POST | `/api/admin/reviews/:id/reply` | 官方回复 `{reply}`（≤200 字，可重复提交覆盖，`replied_at` 刷新） |
| DELETE | `/api/admin/reviews/:id` | 删除违规评价 |

### 3.10 售后管理（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/aftersales?status=&page=&pageSize=` | 售后列表（status 0=全部；含会员昵称/手机号） |
| POST | `/api/admin/aftersales/:afterSaleNo/audit` | 审核 `{agree, amount?, note?}`：同意可调退款金额（默认申请金额，≤实付）；**拒绝时 note 必填**；41705 审核冲突（已被并发处理/撤销） |

> 同意即触发原路退款（demo/微信见支付模块）；退款失败自动回滚审核状态为待审核，可修正后重试。

### 3.11 Banner 运营位管理（已实现，实际路径前缀为 /api/admin）

首页运营位（`pet_banner`，迁移 010）：管理端配置，`GET /api/home` 数据驱动下发（仅上架，`sort ASC`）。`jump_type` 白名单：`me / recommend / coupon / orders / notify / favorites / product / category / search / custom`；`target` 仅 product（商品 ID）/ category（分类 ID）/ search（关键词）/ custom（`/pages/` 前缀路径）时有效。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/banners?page=&pageSize=` | 列表（含下架，id 倒序） |
| POST | `/api/admin/banners` | 新建 `{title(≤32必填), subTitle?, icon?, jumpType(白名单), target?, sort?, status?}` |
| PUT | `/api/admin/banners/:id` | 编辑（可覆盖更新） |
| PUT | `/api/admin/banners/:id/status` | 上架/下架 `{status: 1/0}` |
| DELETE | `/api/admin/banners/:id` | 删除 |

### 3.12 配送方式管理（已实现，实际路径前缀为 /api/admin，迁移 012）

配送方式（`ship_method`）：用户下单选择，订单快照方式名与运费；停用后用户端不可见，历史订单不受影响。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/ship-methods?page=&pageSize=` | 列表（含停用，`sort ASC`） |
| POST | `/api/admin/ship-methods` | 新建 `{name(≤32必填), kind(1自提/2托运配送必选), fee(≥0), description?, sort?, status?}` |
| PUT | `/api/admin/ship-methods/:id` | 编辑（改价只影响新订单） |
| PUT | `/api/admin/ship-methods/:id/status` | 启用/停用 `{status: 1/0}` |
| DELETE | `/api/admin/ship-methods/:id` | 删除（历史订单靠快照回显） |

### 3.13 秒杀活动管理（已实现，实际路径前缀为 /api/admin，迁移 014）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/flash-sales?page=&pageSize=` | 全量活动（含未启用，id 倒序，带商品标题） |
| POST | `/api/admin/flash-sales` | 新建 `{productId, salePrice(≤商品原价), stock(1-9999), startAt, endAt, status?}` |
| PUT | `/api/admin/flash-sales/:id` | 编辑 |
| PUT | `/api/admin/flash-sales/:id/status` | 启用/停用（启用撞「每商品至多一个启用中」唯一索引 → 明确报错） |
| DELETE | `/api/admin/flash-sales/:id` | 删除（订单已快照秒杀价，名额回补按 ID 静默失效） |

### 3.14 经营报表（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/report?days=30` | 报表（days 1-90 默认 30）：`summary{todayGmv, todayOrders, monthGmv, monthOrders, pendingAfters}` + `daily[]`（日期/订单/成交额）+ `topProducts[]`（销量 Top10）+ `breeds[]`（品类分布）；GMV = `status IN (20,30,60)` 按 `paid_at` 汇总 |

### 3.15 审计日志与敏感词（已实现，实际路径前缀为 /api/admin，迁移 015）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/audit-logs?page=&pageSize=` | 管理端操作审计（非 GET 请求由中间件自动落库：管理员/方法/路径/IP，id 倒序） |
| GET / POST | `/api/admin/sensitive-words` | 敏感词分页 / 新增 `{word(≤64)}`（重复添加报错） |
| PUT | `/api/admin/sensitive-words/:id/status` | 启用/停用（进程内缓存 60s 刷新） |
| DELETE | `/api/admin/sensitive-words/:id` | 删除 |

> 敏感词生效点：客服聊天文本命中 → 拒绝发送；评价内容 / 售后理由命中 → `***` 打码后落库。

### 3.16 监控指标（运维）

`GET /metrics`（Prometheus 文本格式，容器网络内抓取）：`pet_api_http_requests_total{method,path,code}`、`pet_api_http_request_seconds_bucket`、`pet_api_orders_created_total`、`pet_api_orders_paid_total`、`pet_api_orders_closed_total{kind=pending|tail}`、`pet_api_pay_notify_total{result}`、`pet_api_notify_delivery_fail_total`、`pet_api_ratelimit_rejected_total{scope}`。部署编排见 `scripts/deploy/monitoring/`（Prometheus 抓取 + 告警规则 + Grafana）。

### 3.17 供货商管理（已实现，实际路径前缀为 /api/admin，迁移 016）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/suppliers?page=&pageSize=&keyword=&status=` | 供货商分页（含在售商品计数） |
| POST | `/api/admin/suppliers` | 新建 `{name(≤64必填), contact?, phone?, address?, licenseNo?, remark?, status?}` |
| PUT | `/api/admin/suppliers/:id` | 编辑 |
| PUT | `/api/admin/suppliers/:id/status` | 启用/停用 |
| DELETE | `/api/admin/suppliers/:id` | 删除（被商品引用 → 41901） |

### 3.18 财务对账（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/finance/payments?page=&pageSize=&orderNo=&status=&payType=` | 交易流水：`payment` 表逐笔（`status≥0`/`payType≥0` 过滤，默认 -1 全部；会员昵称/商品标题联查） |
| GET | `/api/admin/finance/daily?days=30` | 每日汇总（days 1-90 默认 30）：按支付时间聚合 `payAmount`（status=1 且 pay_type IN 1,3）与 `refundAmount`（status=3 且 pay_type=2），`netAmount = pay - refund` |

### 3.19 数据导出 CSV（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/export/orders` | 订单导出 CSV（≤20000 行，UTF-8 BOM 防 Excel 乱码，时间 `yyyy-MM-dd HH:mm:ss`） |
| GET | `/api/admin/export/members` | 会员导出（含成长值/等级） |
| GET | `/api/admin/export/points` | 积分流水导出 |

### 3.20 多角色权限（RBAC，已实现）

三角色：`super_admin`（全部权限）/ `operator`（运营：禁退款、禁会员与账号写操作）/ `support`（客服：白名单 = 客服会话全操作 + 看板/订单/评价只读 + 评价回复）。`adminAuth` 中间件按请求查库校验角色与启用状态后放行/403，停用账号即时失效；审计日志照常落库。

| 角色 | 可用（非只读部分） |
| ---- | ---- |
| super_admin | 全部 |
| operator | 商品/分类品种/营销/秒杀/Banner/配送方式/供货商/知识库 等运营写操作 |

### 3.21 拼团活动管理（已实现，实际路径前缀为 /api/admin，迁移 018）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/group-buys?page=&pageSize=` | 活动分页（含商品标题） |
| POST | `/api/admin/group-buys` | 新建 `{productId, price, size(2-10), hours(1-72), status?}` |
| PUT | `/api/admin/group-buys/:id` | 编辑 |
| PUT | `/api/admin/group-buys/:id/status` | 启用/停用 |
| DELETE | `/api/admin/group-buys/:id` | 删除（进行中的团由 closer 按到期收尾） |

### 3.22 自提核销 / 库存预警 / 风控 / 黑名单（已实现，实际路径前缀为 /api/admin，迁移 018）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/admin/orders/:orderNo/pickup-verify` | 自提核销：`{code}` 6 位核销码校验 + CAS 20→30（41909 核销码错误或状态不允许） |
| GET | `/api/admin/stock-alerts` | 库存预警：在售商品 `stock ≤ stock_warn_threshold` 清单 |
| GET | `/api/admin/risk-logs?page=&pageSize=&rule=` | 风控拦截记录（order_freq / review_contact / member_blacklist） |
| PUT | `/api/admin/members/blacklist` | 拉黑/解除：`{id, blacklist: 0/1}`（可登录但禁交易/评价/领券） |

> 商品上下行扩展字段（3.2）：`detailImages[]` 详情长图、`stockWarnThreshold` 预警阈值、`skus[]` SKU 规格（specs/price/sort/status，全量替换，`has_sku` 自动维护）。评价数据扩展：`healthScore/lookScore/serviceScore`（1-5，缺省 5），详情摘要接口返回三维均值。
| support | 客服会话（收发/已读/结束）、评价回复与显示状态；其余只读（看板/订单/评价列表） |

---

## 4. 安全清单

| 项 | 措施 |
| -- | ---- |
| 越权防护 | 双 JWT secret 物理隔离用户端/平台端；资源归属校验（订单只能查自己的） |
| 管理端 RBAC | 三角色（super_admin/operator/support）按请求查库鉴权，403 拦截越权写操作；停用即时失效 |
| 回调安全 | 微信支付 V3 验签 + 金额比对 + 幂等；回调地址不带敏感参数 |
| 注入 | 全部走 GORM 参数化；detail_html 存储前服务端 XSS 过滤（blueoxx/bluemonday） |
| 限流 | go-zero 内置限流：登录接口按 IP、支付轮询按用户 |
| 敏感配置 | appid/secret/商户密钥走环境变量或配置中心，不入库不入仓 |
| HTTPS | 全链路 HTTPS（小程序强制），H5 回调域名 HTTPS |
