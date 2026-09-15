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
| 41xxx | 业务错误：41001 商品已下架 / 41002 商品已被下单 / 41003 订单状态不允许此操作 / 41004 重复收藏 / 41101 验证码错误或已过期 / 41102 发送过于频繁 / 41103 超过当日发送上限 / 41104 验证码尝试次数过多 / 41203-41209 管理端业务（验证码、密码、引用约束等）/ 41301 AI 提问太频繁 / 41302 AI 当日提问达上限 / 41401 优惠券不可用 / 41402 未达用券门槛 / 41403 券已领完 / 41404 超过限领 / 41405 不在可领时段 / 41406 增值服务不可用 / 41501 会话不存在（含越权访问他人会话）/ 41601 该订单已评价过 / 41602 订单当前不可评价 / 41701 售后单不存在（含越权）/ 41702 已有进行中售后 / 41703 订单不可售后 / 41704 售后状态不允许此操作 / 41705 审核状态冲突请刷新 |
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
| POST | `/api/v1/auth/sms/login` | 手机号验证码登录 `{phone, code}`，未注册自动注册；支持测试公共验证码（仅非生产环境生效） |
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
| GET | `/api/v1/home` | 首页聚合：轮播位、分类入口、热销推荐 |
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
| POST | `/api/v1/payments/wechat/prepay` | 待支付单重新拉起支付 `{orderNo}` → 支付参数 |
| GET | `/api/v1/payments/:paymentNo/status` | 轮询支付结果（前端支付后 2s 间隔轮询） |
| POST | `/api/v1/payment/notify` | **微信回调**（无鉴权，验签；不走统一响应格式） |

**创建订单请求/响应示例**

```json
// POST /api/v1/orders
// Request
{
  "productId": "1943256789000000001",
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

规则：仅「已完成(status=30)」订单可评，一单一评（`order_no` 唯一）；星级 1-5 + 文字（≤500 字）+ 图片（≤9 张，COS 直传）。管理端隐藏/删除后公开侧不可见。订单/详情视图带 `reviewed` 标记回显入口态。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/orders/:orderNo/review` | 创建评价 `{orderNo, rating(1-5), content?, images?[]}`（41601 已评过 / 41602 不可评） |
| GET | `/api/products/:id/reviews?cursor=&limit=` | 公开评价列表（仅 status=1，游标 id 倒序，含会员昵称/头像） |
| GET | `/api/products/:id/review-summary` | 评分摘要 `{avgRating:"5.0", total, latest[3]}`（隐藏的不计入） |

### 2.10 售后申请（已实现，实际路径前缀为 /api）

规则：已支付(20)/已完成(30) 订单可发起，默认全额退款、原路退回，**管理端审核时可调金额**；同单同时仅一个进行中售后（41702）；审核同意即触发退款（确定性退款单号 `RF+售后单号` 幂等），拒绝须填备注。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| POST | `/api/aftersale` | 发起 `{orderNo, reason(≤500字)}`（41703 订单不可售后 / 41702 已有进行中） |
| GET | `/api/aftersale?orderNo=` | 查订单当前售后（本人，最新一条；无则 `data:null`） |
| POST | `/api/aftersale/:afterSaleNo/cancel` | 撤销（仅本人待审核单；41701 不存在/越权 / 41704 状态不允许） |

售后单状态：`1待审核 2已同意 3已拒绝 4已撤销`（`statusText` 同步返回）。

### 2.11 通知订阅（已实现，实际路径前缀为 /api）

微信订阅消息（小程序）/模板消息（公众号 H5）离线触达：客服回复（仅用户不在线时）+ 订单事件（退款到账 / 超时关单 / 售后审核结果）。服务端投递队列见文档 03 §3.20；模板未配置自动降级（不影响业务）。

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/notify/tmpl` | 返回双端模板 ID 配置 `{miniTmplCsReply, miniTmplOrder, h5TmplCsReply, h5TmplOrder}`（未配置为空串） |

> 小程序端必须在**用户点击回调内同步**调用 `wx.requestSubscribeMessage`（Taro 同名 API）订阅，模板 ID 从本接口预取（进入页面时缓存）；H5 端公众号模板消息无需订阅动作。

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
| POST | `/admin/api/v1/products` | 创建商品（草稿） |
| PUT | `/admin/api/v1/products/:id` | 编辑商品 |
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
| GET | `/admin/api/v1/orders/export` | 订单导出 Excel |

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
| GET / POST | `/api/admin/marketing/services` | 增值服务列表 / 新建 `{name, description?, originalPrice?, price, sort?, status?}` |
| PUT / DELETE | `/api/admin/marketing/services/:id` | 编辑 / 删除 |

券模板入参：`{name, type(1满减/2折扣/3立减), thresholdAmount?, discountAmount?, discountPercent?, maxDiscountAmount?, totalCount?(0不限), perLimit?, newUserOnly?, pickupStart?, pickupEnd?, validStart?, validEnd?, status?}`，时段格式 `YYYY-MM-DD HH:mm:ss` 或 RFC3339，空 = 不限。

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
| DELETE | `/api/admin/reviews/:id` | 删除违规评价 |

### 3.10 售后管理（已实现，实际路径前缀为 /api/admin）

| 方法 | 路径 | 说明 |
| ---- | ---- | ---- |
| GET | `/api/admin/aftersales?status=&page=&pageSize=` | 售后列表（status 0=全部；含会员昵称/手机号） |
| POST | `/api/admin/aftersales/:afterSaleNo/audit` | 审核 `{agree, amount?, note?}`：同意可调退款金额（默认申请金额，≤实付）；**拒绝时 note 必填**；41705 审核冲突（已被并发处理/撤销） |

> 同意即触发原路退款（demo/微信见支付模块）；退款失败自动回滚审核状态为待审核，可修正后重试。

---

## 4. 安全清单

| 项 | 措施 |
| -- | ---- |
| 越权防护 | 双 JWT secret 物理隔离用户端/平台端；资源归属校验（订单只能查自己的） |
| 回调安全 | 微信支付 V3 验签 + 金额比对 + 幂等；回调地址不带敏感参数 |
| 注入 | 全部走 GORM 参数化；detail_html 存储前服务端 XSS 过滤（blueoxx/bluemonday） |
| 限流 | go-zero 内置限流：登录接口按 IP、支付轮询按用户 |
| 敏感配置 | appid/secret/商户密钥走环境变量或配置中心，不入库不入仓 |
| HTTPS | 全链路 HTTPS（小程序强制），H5 回调域名 HTTPS |
