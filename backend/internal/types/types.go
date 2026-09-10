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
	Phone string `json:"phone"`
	Code  string `json:"code"`
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
}

type FavoriteListReq struct {
	PageReq
}

type HomeResp struct {
	Categories []CategoryItem `json:"categories"`
	Hot        []ProductCard  `json:"hot"`
}

// ─────────────────────────── 用户端 · 订单支付 ───────────────────────────

type CreateOrderReq struct {
	ProductID    string `json:"productId"`
	ContactName  string `json:"contactName"`
	ContactPhone string `json:"contactPhone"`
	Remark       string `json:"remark,optional"`
}

type WxPayParams struct {
	TimeStamp string `json:"timeStamp"`
	NonceStr  string `json:"nonceStr"`
	Package   string `json:"package"`
	SignType  string `json:"signType"`
	PaySign   string `json:"paySign"`
}

type CreateOrderResp struct {
	OrderNo   string       `json:"orderNo"`
	PayAmount string       `json:"payAmount"`
	ExpireAt  string       `json:"expireAt"`
	PayParams *WxPayParams `json:"payParams"` // 小程序 JSAPI 支付参数
	H5PayURL  string       `json:"h5PayUrl"`  // H5 支付跳转地址
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
	OrderNo      string          `json:"orderNo"`
	Status       int             `json:"status"`
	StatusText   string          `json:"statusText"`
	TotalAmount  string          `json:"totalAmount"`
	PayAmount    string          `json:"payAmount"`
	ContactName  string          `json:"contactName"`
	ContactPhone string          `json:"contactPhone"`
	Remark       string          `json:"remark"`
	ExpireAt     string          `json:"expireAt"`
	PaidAt       string          `json:"paidAt"`
	CreatedAt    string          `json:"createdAt"`
	Items        []OrderItemView `json:"items"`
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
