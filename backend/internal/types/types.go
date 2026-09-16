package types

// ─────────────────────────── 通用 ───────────────────────────

type PageReq struct {
	Page     int `json:",default=1" form:"page,default=1"`
	PageSize int `json:",default=10" form:"pageSize,default=10"`
}

type PageResp struct {
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

type IDPathReq struct {
	ID string `path:"id"`
}

// ─────────────────────────── 用户端 · 认证 ───────────────────────────

type WechatMiniLoginReq struct {
	Code string `json:"code"`
}

type WechatH5OAuthUrlReq struct {
	Redirect string `form:"redirect,optional"`
}

type WechatH5LoginReq struct {
	Code string `json:"code"`
}

type SmsSendReq struct {
	Phone string `json:"phone"`
}

type SmsLoginReq struct {
	Phone      string `json:"phone"`
	Code       string `json:"code"`
	InviteCode string `json:"inviteCode,optional"` // 邀请码（新注册归因）
}

type SmsSendResp struct {
	Interval int `json:"interval"` // 可再次发送的间隔秒数
}

type CancelOrderReq struct {
	Reason string `json:"reason,optional"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

type MemberInfo struct {
	ID        string `json:"id"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Phone     string `json:"phone"`
	Gender    int    `json:"gender"`
	HasWxBind bool   `json:"hasWxBind"`
	IsNew     bool   `json:"isNew"`
	CreatedAt string `json:"createdAt"`
}

type LoginResp struct {
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
	ExpiresIn    int        `json:"expiresIn"`
	Member       MemberInfo `json:"member"`
}

type OAuthUrlResp struct {
	URL string `json:"url"`
}

type UpdateProfileReq struct {
	Nickname string `json:"nickname,optional"`
	Avatar   string `json:"avatar,optional"`
	Gender   int    `json:"gender,optional"`
}

// ─────────────────────────── 用户端 · 宠物商品 ───────────────────────────

type CategoryItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	BreedCnt int64  `json:"breedCount"`
}

type BreedItem struct {
	ID         string `json:"id"`
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	Cover      string `json:"cover"`
}

type ProductListReq struct {
	PageReq
	CategoryID string `form:"categoryId,optional"`
	BreedID    string `form:"breedId,optional"`
	Gender     int    `form:"gender,optional"`
	PriceMin   int64  `form:"priceMin,optional"` // 分
	PriceMax   int64  `form:"priceMax,optional"`
	Keyword    string `form:"keyword,optional"`
	Sort       string `form:"sort,optional"`
}

type PetProfile struct {
	Gender      int    `json:"gender"`
	GenderText  string `json:"genderText"`
	BirthDate   string `json:"birthDate"`
	AgeText     string `json:"ageText"`
	VaccineDesc string `json:"vaccineDesc"`
	DewormDesc  string `json:"dewormDesc"`
	BodyType    string `json:"bodyType"`
	CoatColor   string `json:"coatColor"`
	Personality string `json:"personality"`
	HealthDesc  string `json:"healthDesc"`
}

type ProductCard struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	MainImage     string `json:"mainImage"`
	Price         string `json:"price"`
	OriginalPrice string `json:"originalPrice"`
	BreedName     string `json:"breedName"`
	PetGender     int    `json:"petGender"`
	FavoriteCount int    `json:"favoriteCount"`
	Sales         int    `json:"sales"`
	Status        int    `json:"status"`
}

type CategoryRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProductDetail struct {
	ProductCard
	Images     []string    `json:"images"`
	VideoURL   string      `json:"videoUrl"`
	VideoCover string      `json:"videoCover"`
	DetailHTML string      `json:"detailHtml"`
	PetProfile PetProfile  `json:"petProfile"`
	Breed      BreedItem   `json:"breed"`
	Category   CategoryRef `json:"category"`
	IsFavorite bool        `json:"isFavorite"`
	AIEnabled  bool        `json:"aiEnabled"`
}

type FavoriteListReq struct {
	PageReq
}

type HomeResp struct {
	Categories []CategoryItem `json:"categories"`
	Hot        []ProductCard  `json:"hot"`
	Banners    []BannerView   `json:"banners"`
}

// ─────────────────────────── 运营位 banner ───────────────────────────

type BannerView struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	SubTitle  string `json:"subTitle"`
	Icon      string `json:"icon"`
	JumpType  string `json:"jumpType"`
	Target    string `json:"target"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type BannerUpsertReq struct {
	ID       string `json:"id,optional"`
	Title    string `json:"title"`
	SubTitle string `json:"subTitle,optional"`
	Icon     string `json:"icon,optional"`
	JumpType string `json:"jumpType"`
	Target   string `json:"target,optional"`
	Sort     int    `json:"sort,optional"`
	Status   int    `json:"status,optional"`
}

type BannerListReq struct {
	PageReq
}

type BannerStatusReq struct {
	Status int `json:"status"`
}

// ─────────────────────────── 用户端 · 订单支付 ───────────────────────────

type CreateOrderReq struct {
	ProductID    string   `json:"productId"`
	ServiceIDs   []string `json:"serviceIds,optional"`  // 增值服务
	CouponID     string   `json:"couponId,optional"`    // 用户券 ID，空 = 不用券
	ShipMethodID string   `json:"shipMethodId"`         // 配送方式
	ShipAddress  string   `json:"shipAddress,optional"` // 收货地址（托运配送类必填）
	AddressID    string   `json:"addressId,optional"`   // 地址簿地址，优先于手填
	UseDeposit   bool     `json:"useDeposit,optional"`  // 定金锁宠模式
	ContactName  string   `json:"contactName"`
	ContactPhone string   `json:"contactPhone"`
	Remark       string   `json:"remark,optional"`
}

type WxPayParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type CreateOrderResp struct {
	OrderNo    string       `json:"orderNo"`
	PaymentNo  string       `json:"paymentNo,optional"`
	PayAmount  string       `json:"payAmount"` // 定金单为定金金额
	ExpireAt   string       `json:"expireAt"`
	PayParams  *WxPayParams `json:"payParams"`           // 小程序 JSAPI 支付参数
	H5PayURL   string       `json:"h5PayUrl"`            // H5 支付跳转地址
	IsDeposit  bool         `json:"isDeposit"`           // 定金锁宠单
	TailAmount string       `json:"tailAmount,optional"` // 尾款金额（定金单）
}

type OrderListReq struct {
	PageReq
	Status int `form:"status,optional"` // 0=全部
}

type OrderItemView struct {
	ProductID    string `json:"productId"`
	ProductTitle string `json:"productTitle"`
	ProductImage string `json:"productImage"`
	BreedName    string `json:"breedName"`
	Price        string `json:"price"`
	Quantity     int    `json:"quantity"`
}

type OrderView struct {
	OrderNo         string          `json:"orderNo"`
	Status          int             `json:"status"`
	StatusText      string          `json:"statusText"`
	TotalAmount     string          `json:"totalAmount"`
	DiscountAmount  string          `json:"discountAmount"`
	ServiceFee      string          `json:"serviceFee"`
	ShipFee         string          `json:"shipFee"`
	PayAmount       string          `json:"payAmount"`
	CouponInfo      string          `json:"couponInfo"`
	ServiceItems    string          `json:"serviceItems"` // 服务快照 JSON [{id,name,price}]
	ContactName     string          `json:"contactName"`
	ContactPhone    string          `json:"contactPhone"`
	Remark          string          `json:"remark"`
	ShipMethod      string          `json:"shipMethod"` // 配送方式名快照
	ShipAddress     string          `json:"shipAddress"`
	ShipStatus      int             `json:"shipStatus"` // 0待配送 1配送中 2已送达
	ShipNo          string          `json:"shipNo"`
	ExpireAt        string          `json:"expireAt"`
	PaidAt          string          `json:"paidAt"`
	ShippedAt       string          `json:"shippedAt,optional"`
	DeliveredAt     string          `json:"deliveredAt,optional"`
	CompletedAt     string          `json:"completedAt,optional"`
	CreatedAt       string          `json:"createdAt"`
	Items           []OrderItemView `json:"items"`
	Reviewed        bool            `json:"reviewed"`        // 已评价（status=30 入口态）
	AftersaleStatus int             `json:"aftersaleStatus"` // 最新售后单状态，0=无售后
	DepositAmount   string          `json:"depositAmount"`   // 定金，0=非定金单
	TailExpireAt    string          `json:"tailExpireAt,optional"`
	GuaranteeDays   int             `json:"guaranteeDays"` // 健康保障天数快照，0=无
}

type ShipMethodView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Kind        int    `json:"kind"` // 1自提 2托运配送
	Description string `json:"description"`
	Fee         string `json:"fee"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"`
	CreatedAt   string `json:"createdAt"`
}

type TradeConfigResp struct {
	DepositPercent  int `json:"depositPercent"`  // 定金比例（%），0=未开启定金模式
	DepositHoldDays int `json:"depositHoldDays"` // 尾款补款期限（天）
}

type ShipMethodListReq struct {
	PageReq
}

type ShipMethodUpsertReq struct {
	ID          string `json:"id,optional"`
	Name        string `json:"name"`
	Kind        int    `json:"kind"`
	Description string `json:"description,optional"`
	Fee         string `json:"fee"`
	Sort        int    `json:"sort,optional"`
	Status      int    `json:"status,optional"`
}

type ShipReq struct {
	ShipNo string `json:"shipNo"`
}

type OrderNoPathReq struct {
	OrderNo string `path:"orderNo"`
}

type PrepayReq struct {
	OrderNo string `json:"orderNo"`
}

type PaymentStatusReq struct {
	PaymentNo string `path:"paymentNo"`
}

type PaymentStatusResp struct {
	Status    int    `json:"status"`    // 0待支付 1成功 2失败 3已退款
	PayStatus string `json:"payStatus"` // 文本
}

// ─────────────────────────── 平台端 ───────────────────────────

type AdminLoginReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captchaId,optional"`
	CaptchaCode string `json:"captchaCode,optional"`
}

type AdminInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type ChangePasswordReq struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

type CaptchaResp struct {
	CaptchaID string `json:"captchaId"`
	ImageB64  string `json:"imageB64"` // data:image/png;base64,...
}

type AdminLoginResp struct {
	AccessToken  string    `json:"accessToken"`
	RefreshToken string    `json:"refreshToken"`
	ExpiresIn    int       `json:"expiresIn"`
	Admin        AdminInfo `json:"admin"`
}

type AdminProductListReq struct {
	PageReq
	Status     int    `form:"status,default=-1"`
	CategoryID string `form:"categoryId,optional"`
	BreedID    string `form:"breedId,optional"`
	Keyword    string `form:"keyword,optional"`
}

type PetProfileUpsert struct {
	PetGender   int    `json:"petGender,optional"`
	BirthDate   string `json:"birthDate,optional"` // yyyy-MM-dd
	VaccineDesc string `json:"vaccineDesc,optional"`
	DewormDesc  string `json:"dewormDesc,optional"`
	BodyType    string `json:"bodyType,optional"`
	CoatColor   string `json:"coatColor,optional"`
	Personality string `json:"personality,optional"`
	HealthDesc  string `json:"healthDesc,optional"`
}

type ProductUpsertReq struct {
	ID            string `json:"id,optional"` // 有值为编辑
	Title         string `json:"title"`
	CategoryID    string `json:"categoryId"`
	BreedID       string `json:"breedId"`
	Price         string `json:"price"` // 元，字符串小数
	OriginalPrice string `json:"originalPrice,optional"`
	PetProfileUpsert
	MainImage  string   `json:"mainImage"`
	Images     []string `json:"images"`
	VideoURL   string   `json:"videoUrl,optional"`
	VideoCover string   `json:"videoCover,optional"`
	DetailHTML string   `json:"detailHtml,optional"`
}

type ProductStatusReq struct {
	ID     string `path:"id"`
	Status int    `json:"status"` // 1上架 2下架
}

type AdminProductView struct {
	ID            string `json:"id"`
	SpuNo         string `json:"spuNo"`
	Title         string `json:"title"`
	MainImage     string `json:"mainImage"`
	Price         string `json:"price"`
	OriginalPrice string `json:"originalPrice"`
	BreedName     string `json:"breedName"`
	PetGender     int    `json:"petGender"`
	Status        int    `json:"status"`
	Sales         int    `json:"sales"`
	ViewCount     int    `json:"viewCount"`
	FavoriteCount int    `json:"favoriteCount"`
	CreatedAt     string `json:"createdAt"`
}

type AdminProductDetailResp struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	CategoryID    string `json:"categoryId"`
	BreedID       string `json:"breedId"`
	Price         string `json:"price"`
	OriginalPrice string `json:"originalPrice"`
	Status        int    `json:"status"`
	PetProfileUpsert
	MainImage  string   `json:"mainImage"`
	Images     []string `json:"images"`
	VideoURL   string   `json:"videoUrl"`
	VideoCover string   `json:"videoCover"`
	DetailHTML string   `json:"detailHtml"`
}

type UploadTokenReq struct {
	Dir         string `json:"dir,optional"`
	ContentType string `json:"contentType,optional"`
}

type UploadTokenResp struct {
	UploadURL string `json:"uploadUrl"` // PUT 直传地址
	FileURL   string `json:"fileUrl"`   // 上传成功后的访问地址
	Method    string `json:"method"`
}

type AdminOrderListReq struct {
	PageReq
	Status  int    `form:"status,default=-1"`
	OrderNo string `form:"orderNo,optional"`
	Keyword string `form:"keyword,optional"` // 联系人/手机号
}

type RefundReq struct {
	OrderNo string `path:"orderNo"`
	Reason  string `json:"reason"`
	Amount  string `json:"amount,optional"` // 默认全额
}

type MemberListReq struct {
	PageReq
	Keyword string `form:"keyword,optional"`
	Status  int    `form:"status,default=-1"`
}

type MemberAdminView struct {
	ID         string `json:"id"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Phone      string `json:"phone"`
	Gender     int    `json:"gender"`
	Status     int    `json:"status"`
	OrderCount int64  `json:"orderCount"`
	FavCount   int64  `json:"favoriteCount"`
	CreatedAt  string `json:"createdAt"`
}

type MemberStatusReq struct {
	ID     string `path:"id"`
	Status int    `json:"status"` // 1正常 2禁用
}

type CategoryUpsertReq struct {
	ID   string `json:"id,optional"`
	Name string `json:"name"`
	Icon string `json:"icon,optional"`
	Sort int    `json:"sort,optional"`
}

type BreedUpsertReq struct {
	ID         string `json:"id,optional"`
	CategoryID string `json:"categoryId"`
	Name       string `json:"name"`
	Intro      string `json:"intro,optional"`
	Cover      string `json:"cover,optional"`
	Sort       int    `json:"sort,optional"`
}

type OverviewResp struct {
	OnSaleCount int64  `json:"onSaleCount"`
	TodayOrders int64  `json:"todayOrders"`
	TodayGMV    string `json:"todayGmv"`
	MemberCount int64  `json:"memberCount"`
}

// ─────────────────────────── AI 智能客服 ───────────────────────────

type AIChatMessage struct {
	Role    string `json:"role"` // user / assistant
	Content string `json:"content"`
}

type AIAskReq struct {
	ProductID string          `json:"productId"`
	Question  string          `json:"question"`
	History   []AIChatMessage `json:"history,optional"` // 客户端携带的最近对话（服务端截断）
}

type AIKnowledgeRef struct {
	Title           string `json:"title"`
	Content         string `json:"content"`
	IsBreedSpecific bool   `json:"isBreedSpecific"`
}

type AIAskResp struct {
	Answer string           `json:"answer"`
	Refs   []AIKnowledgeRef `json:"refs"` // 引用的知识库条目
	Demo   bool             `json:"demo"` // 是否演示模式（未接入真实大模型）
}

type AIRecommendReq struct {
	Requirement string `json:"requirement"` // 预算/要求/家庭情况/生活习惯等自由描述，≤500 字
}

type RecommendItem struct {
	ProductCard
	Score  int    `json:"score"`  // 推荐分 0-100
	Reason string `json:"reason"` // 推荐理由（≤60 字）
}

type AIRecommendResp struct {
	Items  []RecommendItem `json:"items"`  // 恰好 3 个（候选不足时按实际数量），按分数降序
	Source string          `json:"source"` // ai=大模型推荐 rule=规则打分兜底
}

type KnowledgeListReq struct {
	PageReq
	BreedID string `form:"breedId,optional"`
	Keyword string `form:"keyword,optional"`
}

type KnowledgeItem struct {
	ID        string `json:"id"`
	BreedID   string `json:"breedId"` // 空 = 平台通用
	BreedName string `json:"breedName"`
	Title     string `json:"title"`
	Keywords  string `json:"keywords"`
	Content   string `json:"content"`
	Sort      int    `json:"sort"`
	Status    int    `json:"status"`
	UpdatedAt string `json:"updatedAt"`
}

type KnowledgeUpsertReq struct {
	ID       string `json:"id,optional"`
	BreedID  string `json:"breedId,optional"`
	Title    string `json:"title"`
	Keywords string `json:"keywords,optional"`
	Content  string `json:"content"`
	Sort     int    `json:"sort,optional"`
	Status   int    `json:"status,optional"`
}

// ─────────────────────────── 营销 · 增值服务/优惠券 ───────────────────────────

type ServiceItemView struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	OriginalPrice string `json:"originalPrice"`
	Price         string `json:"price"`
	GuaranteeDays int    `json:"guaranteeDays"` // >0 为健康保障服务（完成 N 天内可申请售后）
}

type ServiceUpsertReq struct {
	ID            string `json:"id,optional"`
	Name          string `json:"name"`
	Description   string `json:"description,optional"`
	OriginalPrice string `json:"originalPrice,optional"`
	Price         string `json:"price"`
	GuaranteeDays int    `json:"guaranteeDays,optional"` // >0 为健康保障服务
	Sort          int    `json:"sort,optional"`
	Status        int    `json:"status,optional"`
}

type CouponTemplateView struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Type              int    `json:"type"`
	TypeText          string `json:"typeText"`
	ThresholdAmount   string `json:"thresholdAmount"`
	DiscountAmount    string `json:"discountAmount"`
	DiscountPercent   int    `json:"discountPercent"`
	MaxDiscountAmount string `json:"maxDiscountAmount"`
	TotalCount        int    `json:"totalCount"`
	IssuedCount       int    `json:"issuedCount"`
	PerLimit          int    `json:"perLimit"`
	NewUserOnly       int    `json:"newUserOnly"`
	PointsCost        int    `json:"pointsCost"` // >0 需用积分兑换领取
	PickupStart       string `json:"pickupStart"`
	PickupEnd         string `json:"pickupEnd"`
	ValidStart        string `json:"validStart"`
	ValidEnd          string `json:"validEnd"`
	Status            int    `json:"status"`
	UpdatedAt         string `json:"updatedAt"`
}

type CouponUpsertReq struct {
	ID                string `json:"id,optional"`
	Name              string `json:"name"`
	Type              int    `json:"type"`
	ThresholdAmount   string `json:"thresholdAmount,optional"`
	DiscountAmount    string `json:"discountAmount,optional"`
	DiscountPercent   int    `json:"discountPercent,optional"`
	MaxDiscountAmount string `json:"maxDiscountAmount,optional"`
	TotalCount        int    `json:"totalCount,optional"`
	PerLimit          int    `json:"perLimit,optional"`
	NewUserOnly       int    `json:"newUserOnly,optional"`
	PointsCost        int    `json:"pointsCost,optional"` // >0 积分兑换券
	PickupStart       string `json:"pickupStart,optional"`
	PickupEnd         string `json:"pickupEnd,optional"`
	ValidStart        string `json:"validStart,optional"`
	ValidEnd          string `json:"validEnd,optional"`
	Status            int    `json:"status,optional"`
}

type CouponIssueReq struct {
	MemberIDs []string `json:"memberIds"`
}

type MyCouponView struct {
	ID            string `json:"id"`
	TemplateID    string `json:"templateId"`
	Name          string `json:"name"`
	Type          int    `json:"type"`
	Threshold     string `json:"threshold"`
	Discount      string `json:"discount"`
	DiscountValid bool   `json:"discountValid"` // 折扣券展示：true=按 percent 展示
	Percent       int    `json:"percent"`
	ValidEnd      string `json:"validEnd"`
	Status        int    `json:"status"`
	StatusText    string `json:"statusText"`
	ReceivedAt    string `json:"receivedAt"`
}

type UsableCouponView struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Type          int    `json:"type"`
	Threshold     string `json:"threshold"`
	Discount      string `json:"discount"` // 预估抵扣金额
	DiscountValid bool   `json:"discountValid"`
	Percent       int    `json:"percent"`
	ValidEnd      string `json:"validEnd"`
}

type UsableCouponsReq struct {
	ProductID  string   `json:"productId"`
	ServiceIDs []string `json:"serviceIds,optional"`
}

type CouponAdminListReq struct {
	PageReq
	Status  int    `form:"status,optional"` // 模板状态筛选
	Keyword string `form:"keyword,optional"`
}

// ─────────────────────────── 人工客服 ───────────────────────────

type CsSendReq struct {
	MsgType int    `json:"msgType,optional"` // 1文本 2图片，默认 1
	Content string `json:"content"`
}

type CsMsgOut struct {
	ID         string `json:"id"` // 雪花 ID 以字符串下发，避免 JS 精度丢失
	SessionID  string `json:"sessionId"`
	SenderRole int    `json:"senderRole"`
	SenderID   string `json:"senderId"`
	MsgType    int    `json:"msgType"`
	Content    string `json:"content"`
	CreatedAt  string `json:"createdAt"`
}

type CsHistoryReq struct {
	Before int64 `form:"before,optional"` // 取 id < before 的消息（向上翻页）
	After  int64 `form:"after,optional"`  // 取 id > after 的消息（断线补拉）
	Limit  int   `form:"limit,optional"`  // 默认 20，最大 100
}

type CsHistoryResp struct {
	List    []CsMsgOut `json:"list"`    // 按时间正序
	HasMore bool       `json:"hasMore"` // before 模式下是否还有更早消息
}

type CsSessionItem struct {
	ID              string `json:"id"`
	MemberID        string `json:"memberId"`
	Nickname        string `json:"nickname"`
	Phone           string `json:"phone"`
	Avatar          string `json:"avatar"`
	Status          int    `json:"status"`
	UnreadAdmin     int    `json:"unreadAdmin"`
	LastMessageText string `json:"lastMessageText"`
	LastMessageAt   string `json:"lastMessageAt"`
	CreatedAt       string `json:"createdAt"`
}

// ─────────────────────────── 订单评价 ───────────────────────────

type ReviewCreateReq struct {
	OrderNo string   `json:"orderNo,optional"` // 实际取 path 参数，optional 仅为满足 go-zero mapping 必填校验
	Rating  int      `json:"rating"`
	Content string   `json:"content,optional"`
	Images  []string `json:"images,optional"`
}

type ReviewView struct {
	ID           string   `json:"id"`
	OrderNo      string   `json:"orderNo"`
	MemberID     string   `json:"memberId"`
	Nickname     string   `json:"nickname"`
	Avatar       string   `json:"avatar"`
	ProductID    string   `json:"productId"`
	ProductTitle string   `json:"productTitle"`
	Rating       int      `json:"rating"`
	Content      string   `json:"content"`
	Images       []string `json:"images"`
	Status       int      `json:"status"`
	CreatedAt    string   `json:"createdAt"`
	Reply        string   `json:"reply"`
	RepliedAt    string   `json:"repliedAt,optional"`
}

type ReviewListReq struct {
	Cursor string `form:"cursor,optional"`
	Limit  int    `form:"limit,optional"`
}

type ReviewListResp struct {
	List    []ReviewView `json:"list"`
	HasMore bool         `json:"hasMore"`
}

type ReviewSummaryResp struct {
	AvgRating string       `json:"avgRating"`
	Total     int64        `json:"total"`
	Latest    []ReviewView `json:"latest"`
}

type ReviewStatusReq struct {
	Status int `json:"status"`
}

type ReviewReplyReq struct {
	Reply string `json:"reply"`
}

// ─────────────────────────── 售后 ───────────────────────────

type AfterSaleApplyReq struct {
	OrderNo string `json:"orderNo"`
	Reason  string `json:"reason"`
}

type AfterSaleView struct {
	ID              string `json:"id"`
	AfterSaleNo     string `json:"afterSaleNo"`
	OrderNo         string `json:"orderNo"`
	MemberID        string `json:"memberId"`
	Nickname        string `json:"nickname"`
	Phone           string `json:"phone"`
	Reason          string `json:"reason"`
	RefundAmount    string `json:"refundAmount"`
	Status          int    `json:"status"`
	StatusText      string `json:"statusText"`
	AdminNote       string `json:"adminNote"`
	RefundPaymentNo string `json:"refundPaymentNo"`
	AuditAt         string `json:"auditAt"`
	CreatedAt       string `json:"createdAt"`
}

type AfterSaleOrderReq struct {
	OrderNo string `form:"orderNo"`
}

type AfterSaleNoPathReq struct {
	AfterSaleNo string `path:"afterSaleNo"`
}

type AfterSaleAdminListReq struct {
	PageReq
	Status int `form:"status,optional"` // 0=全部
}

type AfterSaleAuditReq struct {
	Agree  bool   `json:"agree"`
	Amount string `json:"amount,optional"` // 同意时可调，默认全额
	Note   string `json:"note,optional"`   // 拒绝时必填
}

// ─────────────────────────── 微信通知 ───────────────────────────

type NotifyTmplResp struct {
	MiniTmplCsReply string `json:"miniTmplCsReply"`
	MiniTmplOrder   string `json:"miniTmplOrder"`
	H5TmplCsReply   string `json:"h5TmplCsReply"`
	H5TmplOrder     string `json:"h5TmplOrder"`
}

type NotifyView struct {
	ID        string `json:"id"`    // 字符串 ID，避免 JS 精度丢失
	Scene     int    `json:"scene"` // 1客服回复 2订单 3优惠券
	Title     string `json:"title"`
	Content   string `json:"content"`
	OrderNo   string `json:"orderNo"`
	ReadAt    string `json:"readAt,optional"` // 空=未读
	CreatedAt string `json:"createdAt"`
}

type NotifyListResp struct {
	PageResp
	Unread int `json:"unread"`
}

// ─────────────────────────── 用户端 · 地址簿 ───────────────────────────

type AddressView struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	IsDefault int    `json:"isDefault"`
}

type AddressSaveReq struct {
	ID        string `json:"id,optional"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Address   string `json:"address"`
	IsDefault bool   `json:"isDefault,optional"`
}

type AddressIDReq struct {
	ID string `path:"id"`
}

// ─────────────────────────── 用户端 · 积分 / 邀请 ───────────────────────────

type PointsLogView struct {
	ID        string `json:"id"`
	Change    int64  `json:"change"` // 正负
	Balance   int64  `json:"balance"`
	Reason    string `json:"reason"`
	Ref       string `json:"ref"`
	CreatedAt string `json:"createdAt"`
}

type PointsResp struct {
	Points int64           `json:"points"`
	Total  int64           `json:"total"`
	List   []PointsLogView `json:"list"`
}

type PointsPageReq struct {
	PageReq
}

type SignInResp struct {
	OK      bool  `json:"ok"`      // false=今日已签到
	Balance int64 `json:"balance"` // 签到后余额（未开启活动时 0）
}

type InviteResp struct {
	InviteCode string `json:"inviteCode"`
	Invited    int64  `json:"invited"`
	RewardEach int64  `json:"rewardEach"` // 每邀请一人双方各得积分
}

// ─────────────────────────── 秒杀 ───────────────────────────

type FlashSaleView struct {
	ID           string `json:"id"`
	ProductID    string `json:"productId"`
	ProductTitle string `json:"productTitle"`
	ProductImage string `json:"productImage"` // 主图（首页秒杀专区展示）
	SalePrice    string `json:"salePrice"`
	Stock        int    `json:"stock"`
	Sold         int    `json:"sold"`
	StartAt      string `json:"startAt"`
	EndAt        string `json:"endAt"`
	Status       int    `json:"status"`
	CreatedAt    string `json:"createdAt"`
}

type FlashSaleUpsertReq struct {
	ID        string `json:"id,optional"`
	ProductID string `json:"productId"`
	SalePrice string `json:"salePrice"`
	Stock     int    `json:"stock"`
	StartAt   string `json:"startAt"`
	EndAt     string `json:"endAt"`
	Status    int    `json:"status,optional"`
}

// ─────────────────────────── 平台端 · 报表 / 审计 / 敏感词 ───────────────────────────

type AdminDailyRow struct {
	Date   string `json:"date"`
	Orders int64  `json:"orders"`
	GMV    string `json:"gmv"`
}

type AdminRankRow struct {
	Name   string `json:"name"`
	Count  int64  `json:"count"`
	Amount string `json:"amount"`
}

type AdminReportResp struct {
	Summary     AdminReportSummary `json:"summary"`
	Daily       []AdminDailyRow    `json:"daily"`       // 近 N 天成交
	TopProducts []AdminRankRow     `json:"topProducts"` // 商品销量 Top10
	Breeds      []AdminRankRow     `json:"breeds"`      // 品类成交分布
}

type AdminReportSummary struct {
	TodayGMV      string `json:"todayGmv"`
	TodayOrders   int64  `json:"todayOrders"`
	MonthGMV      string `json:"monthGmv"`
	MonthOrders   int64  `json:"monthOrders"`
	PendingAfters int64  `json:"pendingAfters"`
}

type AuditLogRow struct {
	ID        string `json:"id"`
	AdminName string `json:"adminName"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	IP        string `json:"ip"`
	CreatedAt string `json:"createdAt"`
}

type SensitiveWordRow struct {
	ID        string `json:"id"`
	Word      string `json:"word"`
	Status    int    `json:"status"`
	CreatedAt string `json:"createdAt"`
}

type SensitiveWordSaveReq struct {
	Word string `json:"word"`
}
