export interface CategoryItem {
  id: string
  name: string
  icon: string
  breedCount: number
}

export interface ProductCard {
  id: string
  title: string
  mainImage: string
  price: string
  originalPrice: string
  breedName: string
  petGender: number
  favoriteCount: number
  sales: number
  status: number
}

export interface BannerView {
  id: string
  title: string
  subTitle: string
  icon: string // emoji 或图片 URL
  jumpType: string
  target: string
  sort: number
  status: number
  createdAt: string
}

export interface HomeResp {
  categories: CategoryItem[]
  hot: ProductCard[]
  banners?: BannerView[]
}

export interface PetProfile {
  gender: number
  genderText: string
  birthDate: string
  ageText: string
  vaccineDesc: string
  dewormDesc: string
  bodyType: string
  coatColor: string
  personality: string
  healthDesc: string
}

export interface ProductDetail extends ProductCard {
  images: string[]
  videoUrl: string
  videoCover: string
  detailHtml: string
  petProfile: PetProfile
  breed: { id: string; categoryId: string; name: string; cover: string }
  category: { id: string; name: string }
  isFavorite: boolean
  aiEnabled?: boolean
}

export interface AIChatMessage {
  role: 'user' | 'assistant'
  content: string
}

export interface AIKnowledgeRef {
  title: string
  content: string
  isBreedSpecific: boolean
}

export interface AIAskResp {
  answer: string
  refs: AIKnowledgeRef[]
  demo: boolean
}

export interface PageResp<T> {
  total: number
  list: T[]
}

export interface LoginResp {
  accessToken: string
  refreshToken: string
  expiresIn: number
  member: { id: string; nickname: string; avatar: string; phone: string; isNew: boolean }
}

export interface OrderView {
  orderNo: string
  status: number
  statusText: string
  totalAmount: string
  discountAmount: string
  serviceFee: string
  shipFee: string
  payAmount: string
  couponInfo: string
  serviceItems: string
  contactName: string
  contactPhone: string
  remark: string
  shipMethod: string
  shipAddress: string
  shipStatus: number // 0待配送 1配送中 2已送达
  shipNo: string
  expireAt: string
  createdAt: string
  paidAt?: string
  shippedAt?: string
  deliveredAt?: string
  completedAt?: string
  reviewed?: boolean
  aftersaleStatus?: number
  depositAmount?: string // 定金，0=非定金单
  tailExpireAt?: string // 尾款截止
  guaranteeDays?: number // 健康保障天数，0=无
  items: { productId: string; productTitle: string; productImage: string; breedName: string; price: string; quantity: number }[]
}

export interface ShipMethod {
  id: string
  name: string
  kind: number // 1自提 2托运配送
  description: string
  fee: string
}

// ───────── 营销 ─────────

export interface ServiceItemView {
  id: string
  name: string
  description: string
  originalPrice: string
  price: string
}

export interface CouponTemplateView {
  id: string
  name: string
  type: number
  typeText: string
  thresholdAmount: string
  discountAmount: string
  discountPercent: number
  maxDiscountAmount: string
  totalCount: number
  issuedCount: number
  perLimit: number
  newUserOnly: number
  pointsCost?: number // >0 积分兑换
  validStart: string
  validEnd: string
}

export interface MyCouponView {
  id: string
  templateId: string
  name: string
  type: number
  threshold: string
  discount: string
  discountValid: boolean
  percent: number
  validEnd: string
  status: number
  statusText: string
  receivedAt: string
}

export interface UsableCouponView {
  id: string
  name: string
  type: number
  threshold: string
  discount: string
  discountValid: boolean
  percent: number
  validEnd: string
}

export interface CreateOrderResult {
  orderNo: string
  paymentNo?: string
  payAmount: string
  expireAt: string
  payParams: unknown | null
  h5PayUrl: string
  isDeposit?: boolean
  tailAmount?: string
}

// ───────── 增长（积分/邀请/秒杀/地址簿） ─────────

export interface AddressView {
  id: string
  name: string
  phone: string
  address: string
  isDefault: number
}

export interface PointsLogView {
  id: string
  change: number
  balance: number
  reason: string
  ref: string
  createdAt: string
}

export interface PointsResp {
  points: number
  total: number
  list: PointsLogView[]
}

export interface SignInResp {
  ok: boolean
  balance: number
}

export interface InviteResp {
  inviteCode: string
  invited: number
  rewardEach: number
}

export interface FlashSaleInfo {
  id: string
  productId: string
  productTitle: string
  productImage: string
  salePrice: string
  stock: number
  sold: number
  startAt: string
  endAt: string
  status: number
}


// ─────────────────────────── 人工客服 ───────────────────────────

export interface CsMsgOut {
  id: string // 雪花 ID 字符串,避免 JS 精度丢失
  sessionId: string
  senderRole: number // 1会员 2客服
  msgType: number // 1文本 2图片
  content: string
  createdAt: string
}

export interface CsHistoryResp {
  list: CsMsgOut[]
  hasMore: boolean
}

export interface CsSessionUpdate {
  sessionId: string
  status: number // 1进行中 2已结束
  unreadAdmin: number
  unreadMember: number
  lastMessageText: string
  lastMessageAt: string
}

export interface CsEvent {
  type: 'ping' | 'closing_warning' | 'new_message' | 'session_update'
  data?: unknown
}

export interface UploadTokenResp {
  uploadUrl: string
  fileUrl: string
  method: string
}

// ─────────────────────────── 订单评价 / 售后 / 通知 ───────────────────────────

export interface ReviewView {
  id: string // 雪花 ID 字符串，避免 JS 精度丢失
  orderNo: string
  memberId: string
  nickname: string
  avatar: string
  productId: string
  productTitle: string
  rating: number
  content: string
  images: string[]
  status: number
  createdAt: string
  reply?: string // 商家回复
  repliedAt?: string
}

export interface ReviewListResp {
  list: ReviewView[]
  hasMore: boolean
}

export interface ReviewSummaryResp {
  avgRating: string
  total: number
  latest: ReviewView[]
}

export interface AfterSaleView {
  id: string
  afterSaleNo: string
  orderNo: string
  memberId: string
  reason: string
  refundAmount: string
  status: number // 1待审核 2已同意 3已拒绝 4已撤销
  statusText: string
  adminNote: string
  refundPaymentNo: string
  auditAt: string
  createdAt: string
}

export interface NotifyTmplResp {
  miniTmplCsReply: string
  miniTmplOrder: string
  h5TmplCsReply: string
  h5TmplOrder: string
}

export interface NotifyView {
  id: string
  scene: number // 1客服回复 2订单 3优惠券
  title: string
  content: string
  orderNo: string
  readAt: string // 空=未读
  createdAt: string
}

export interface NotifyListResp extends PageResp<NotifyView> {
  unread: number
}

export interface RecommendItem {
  id: string
  title: string
  mainImage: string
  price: string
  originalPrice: string
  breedName: string
  petGender: number
  favoriteCount: number
  sales: number
  status: number
  score: number // 推荐分 0-100
  reason: string // 推荐理由
}

export interface AIRecommendResp {
  items: RecommendItem[]
  source: 'ai' | 'rule' // ai=大模型推荐 rule=规则打分兜底
}
