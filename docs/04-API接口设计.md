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
| 41xxx | 业务错误：41001 商品已下架 / 41002 商品已被下单 / 41003 订单状态不允许此操作 / 41004 重复收藏 / 41101 验证码错误或已过期 / 41102 发送过于频繁 / 41103 超过当日发送上限 / 41104 验证码尝试次数过多 / 41203-41209 管理端业务（验证码、密码、引用约束等）/ 41301 AI 提问太频繁 / 41302 AI 当日提问达上限 |
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
