package types

// ───────── 砍价 ─────────

type BargainActivityView struct {
	ID            string `json:"id"`
	ProductID     string `json:"productId"`
	ProductTitle  string `json:"productTitle"`
	ProductImage  string `json:"productImage"`
	Price         string `json:"price"`
	BottomPrice   string `json:"bottomPrice"`
	DurationHours int    `json:"durationHours"`
	MaxHelpers    int    `json:"maxHelpers"`
	Status        int    `json:"status"`
}

type BargainLaunchView struct {
	ID           string `json:"id"`
	ActivityID   string `json:"activityId"`
	ProductTitle string `json:"productTitle"`
	ProductImage string `json:"productImage"`
	OriginPrice  string `json:"originPrice"`
	CurrentPrice string `json:"currentPrice"`
	BottomPrice  string `json:"bottomPrice"`
	HelperCount  int    `json:"helperCount"`
	MaxHelpers   int    `json:"maxHelpers"`
	Status       int    `json:"status"`
	StatusText   string `json:"statusText"`
	ExpireAt     string `json:"expireAt"`
}

type BargainLaunchReq struct {
	ActivityID string `json:"activityId"`
}

type BargainHelpReq struct {
	LaunchID string `json:"launchId"`
}

// ───────── 竞拍 ─────────

type AuctionView struct {
	ID            string `json:"id"`
	ProductID     string `json:"productId"`
	ProductTitle  string `json:"productTitle"`
	ProductImage  string `json:"productImage"`
	StartPrice    string `json:"startPrice"`
	StepPrice     string `json:"stepPrice"`
	DepositAmount string `json:"depositAmount"`
	StartAt       string `json:"startAt"`
	EndAt         string `json:"endAt"`
	Status        int    `json:"status"`
	StatusText    string `json:"statusText"`
	HighestPrice  string `json:"highestPrice"`
	HighestMember string `json:"highestMember,optional"`
	DepositPaid   bool   `json:"depositPaid"`
}

type AuctionUpsertReq struct {
	ID            string `json:"id,optional"`
	ProductID     string `json:"productId"`
	StartPrice    string `json:"startPrice"`
	StepPrice     string `json:"stepPrice"`
	DepositAmount string `json:"depositAmount"`
	StartAt       string `json:"startAt"` // yyyy-MM-dd HH:mm:ss
	EndAt         string `json:"endAt"`
}

type AuctionBidReq struct {
	ID    string `path:"id"`
	Price string `json:"price"`
}

// ───────── 任务中心 ─────────

type TaskView struct {
	Key     string `json:"key"`
	Title   string `json:"title"`
	Desc    string `json:"desc"`
	Reward  int    `json:"reward"`
	Done    bool   `json:"done"`
	Claimed bool   `json:"claimed"`
	Type    string `json:"type"` // daily / growth
}

type TaskListResp struct {
	List   []TaskView `json:"list"`
	Streak int        `json:"streak"`
}

type TaskClaimReq struct {
	Key string `json:"key"`
}

// ───────── VIP / 预约 / 发票 ─────────

type VipBuyResp struct {
	PaymentNo string `json:"paymentNo"`
	Price     string `json:"price"`
}

type BookingReq struct {
	ServiceID string `json:"serviceId"`
	StoreID   string `json:"storeId"`
	Date      string `json:"date"` // yyyy-MM-dd
	Slot      string `json:"slot"` // 09:00-11:00 等
	Contact   string `json:"contact"`
	Phone     string `json:"phone"`
	PetName   string `json:"petName,optional"`
}

type BookingView struct {
	BookingNo   string `json:"bookingNo"`
	ServiceName string `json:"serviceName"`
	StoreName   string `json:"storeName,optional"`
	Date        string `json:"date"`
	Slot        string `json:"slot"`
	Contact     string `json:"contact"`
	Phone       string `json:"phone"`
	PetName     string `json:"petName,optional"`
	Price       string `json:"price"`
	Status      int    `json:"status"`
	StatusText  string `json:"statusText"`
	VerifyCode  string `json:"verifyCode,optional"`
}

type BookingVerifyReq struct {
	BookingNo string `path:"bookingNo"`
	Code      string `json:"code"`
}

type InvoiceApplyReq struct {
	OrderNo   string `json:"orderNo"`
	TitleType int    `json:"titleType"` // 1个人 2企业
	Title     string `json:"title"`
	TaxNo     string `json:"taxNo,optional"`
}

type InvoiceView struct {
	OrderNo   string `json:"orderNo"`
	TitleType int    `json:"titleType"`
	Title     string `json:"title"`
	TaxNo     string `json:"taxNo,optional"`
	Amount    string `json:"amount"`
	Status    int    `json:"status"`
	Link      string `json:"link,optional"`
	CreatedAt string `json:"createdAt"`
}

type InvoiceIssueReq struct {
	ID   string `path:"id"`
	Link string `json:"link,optional"`
}

// ───────── 分群 / 定价 ─────────

type SegmentRow struct {
	Segment string `json:"segment"`
	Name    string `json:"name"`
	Count   int64  `json:"count"`
}

type SegmentMemberRow struct {
	MemberID   string `json:"memberId"`
	Nickname   string `json:"nickname"`
	Phone      string `json:"phone,optional"`
	Total      string `json:"total"`
	OrderCount int64  `json:"orderCount"`
	LastPaid   string `json:"lastPaid"`
}

type SegmentIssueReq struct {
	Segment          string `json:"segment"`
	CouponTemplateID string `json:"couponTemplateId"`
}

type PriceSuggestResp struct {
	BreedID     string `json:"breedId"`
	SoldCount   int64  `json:"soldCount"`
	OnSaleCount int64  `json:"onSaleCount"`
	AvgPrice    string `json:"avgPrice"`
	MinPrice    string `json:"minPrice"`
	MaxPrice    string `json:"maxPrice"`
	SuggestLow  string `json:"suggestLow,optional"`
	SuggestHigh string `json:"suggestHigh,optional"`
}

type BargainActivityUpsert struct {
	ID            string `json:"id,optional"`
	ProductID     string `json:"productId"`
	BottomPrice   string `json:"bottomPrice"`
	DurationHours int    `json:"durationHours,optional"`
	MaxHelpers    int    `json:"maxHelpers,optional"`
}
