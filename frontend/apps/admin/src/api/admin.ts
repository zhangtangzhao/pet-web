import { client, ACCESS_KEY, REFRESH_KEY, PageResp } from './client'

export interface AdminInfo {
  id: string
  username: string
  nickname: string
  role: string
}

export interface AdminProduct {
  id: string
  spuNo: string
  title: string
  mainImage: string
  price: string
  originalPrice: string
  breedName: string
  petGender: number
  status: number
  sales: number
  viewCount: number
  favoriteCount: number
  createdAt: string
}

export interface AdminProductDetail {
  id: string
  title: string
  categoryId: string
  breedId: string
  price: string
  originalPrice: string
  status: number
  petGender: number
  birthDate: string
  vaccineDesc: string
  dewormDesc: string
  bodyType: string
  coatColor: string
  personality: string
  healthDesc: string
  mainImage: string
  images: string[]
  videoUrl: string
  videoCover: string
  detailHtml: string
  supplierId?: string
  quarantineCertUrl?: string
  nextVaccineDate?: string
  nextDewormDate?: string
  detailImages?: string[]
  stockWarnThreshold?: string
  skus?: ProductSkuRow[]
}

export interface ProductSkuRow {
  id: string
  specs: string
  price: string
  sort: number
  status: number
}

export interface SkuUpsertItem {
  specs: string
  price: string
  sort?: number
  status?: number
}

export interface OrderView {
  orderNo: string
  status: number
  statusText: string
  totalAmount: string
  discountAmount: string
  serviceFee: string
  payAmount: string
  couponInfo: string
  serviceItems: string
  contactName: string
  contactPhone: string
  remark: string
  shipMethod: string
  shipFee: string
  shipAddress: string
  shipStatus: number
  shipNo: string
  shippedAt: string
  deliveredAt: string
  completedAt: string
  expireAt: string
  paidAt: string
  createdAt: string
  levelDiscount?: string
  depositAmount?: string
  tailExpireAt?: string
  guaranteeDays?: number
  isPickup?: boolean
  pickupCode?: string
  groupTeamId?: string
  items: {
    productId: string
    productTitle: string
    productImage: string
    breedName: string
    price: string
    quantity: number
  }[]
}

export interface MemberRow {
  id: string
  nickname: string
  avatar: string
  phone: string
  gender: number
  status: number
  blacklist: number
  orderCount: number
  favoriteCount: number
  growthValue: number
  levelName: string
  createdAt: string
}

export interface CategoryItem {
  id: string
  name: string
  icon: string
  breedCount: number
}

export interface BreedItem {
  id: string
  categoryId: string
  name: string
  cover: string
}

export async function getCaptcha(): Promise<{ captchaId: string; imageB64: string }> {
  return (await client.get('/admin/captcha')) as any
}

export async function adminLogin(
  username: string,
  password: string,
  captchaId?: string,
  captchaCode?: string,
): Promise<{ accessToken: string; refreshToken: string; expiresIn: number; admin: AdminInfo }> {
  const data = (await client.post('/admin/login', {
    username,
    password,
    captchaId,
    captchaCode,
  })) as any
  localStorage.setItem(ACCESS_KEY, data.accessToken)
  localStorage.setItem(REFRESH_KEY, data.refreshToken)
  return data
}

export async function adminLogout() {
  try {
    await client.post('/admin/logout', { refreshToken: localStorage.getItem(REFRESH_KEY) })
  } finally {
    localStorage.removeItem(ACCESS_KEY)
    localStorage.removeItem(REFRESH_KEY)
  }
}

export async function adminChangePassword(oldPassword: string, newPassword: string) {
  await client.put('/admin/password', { oldPassword, newPassword })
}

export async function fetchOverview(): Promise<{
  onSaleCount: number
  todayOrders: number
  todayGmv: string
  memberCount: number
}> {
  return (await client.get('/admin/overview')) as any
}

export async function fetchProducts(params: {
  page?: number
  pageSize?: number
  status?: number
  keyword?: string
}): Promise<PageResp<AdminProduct>> {
  return (await client.get('/admin/products', { params })) as any
}

export async function fetchProductDetail(id: string): Promise<AdminProductDetail> {
  return (await client.get(`/admin/products/${id}`)) as any
}

export async function createProduct(payload: Record<string, unknown>) {
  await client.post('/admin/products', payload)
}

export async function updateProduct(id: string, payload: Record<string, unknown>) {
  await client.put(`/admin/products/${id}`, payload)
}

export async function updateProductStatus(id: string, status: number) {
  await client.put(`/admin/products/${id}/status`, { status })
}

export async function fetchCategories(): Promise<CategoryItem[]> {
  return (await client.get('/categories')) as any
}

export async function fetchBreeds(categoryId?: string): Promise<BreedItem[]> {
  return (await client.get('/breeds', { params: { categoryId } })) as any
}

export async function upsertCategory(payload: { id?: string; name: string; icon?: string; sort?: number }) {
  if (payload.id) await client.put(`/admin/categories/${payload.id}`, payload)
  else await client.post('/admin/categories', payload)
}

export async function deleteCategory(id: string) {
  await client.delete(`/admin/categories/${id}`)
}

export async function upsertBreed(payload: {
  id?: string
  categoryId: string
  name: string
  intro?: string
  cover?: string
  sort?: number
}) {
  if (payload.id) await client.put(`/admin/breeds/${payload.id}`, payload)
  else await client.post('/admin/breeds', payload)
}

export async function deleteBreed(id: string) {
  await client.delete(`/admin/breeds/${id}`)
}

export async function fetchOrders(params: {
  page?: number
  pageSize?: number
  status?: number
  orderNo?: string
  keyword?: string
}): Promise<PageResp<OrderView>> {
  return (await client.get('/admin/orders', { params })) as any
}

export async function refundOrder(orderNo: string, reason: string, amount?: string) {
  await client.post(`/admin/orders/${orderNo}/refund`, { reason, amount })
}

export async function shipOrder(orderNo: string, shipNo: string) {
  await client.post(`/admin/orders/${orderNo}/ship`, { shipNo })
}

export async function deliverOrder(orderNo: string) {
  await client.post(`/admin/orders/${orderNo}/deliver`)
}

export async function fetchMembers(params: {
  page?: number
  pageSize?: number
  keyword?: string
  status?: number
}): Promise<PageResp<MemberRow>> {
  return (await client.get('/admin/members', { params })) as any
}

export async function updateMemberStatus(id: string, status: number) {
  await client.put(`/admin/members/${id}/status`, { status })
}

export interface KnowledgeRow {
  id: string
  breedId: string
  breedName: string
  title: string
  keywords: string
  content: string
  sort: number
  status: number
  updatedAt: string
}

export async function fetchKnowledge(params: {
  page?: number
  pageSize?: number
  breedId?: string
  keyword?: string
}): Promise<PageResp<KnowledgeRow>> {
  return (await client.get('/admin/ai/knowledge', { params })) as any
}

export async function upsertKnowledge(payload: {
  id?: string
  breedId?: string
  title: string
  keywords?: string
  content: string
  sort?: number
  status?: number
}) {
  if (payload.id) await client.put(`/admin/ai/knowledge/${payload.id}`, payload)
  else await client.post('/admin/ai/knowledge', payload)
}

export async function deleteKnowledge(id: string) {
  await client.delete(`/admin/ai/knowledge/${id}`)
}

// ───────── 营销：优惠券 / 增值服务 ─────────

export interface CouponTemplate {
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
  pickupStart: string
  pickupEnd: string
  validStart: string
  validEnd: string
  status: number
  updatedAt: string
}

export interface ServiceItemRow {
  id: string
  name: string
  description: string
  originalPrice: string
  price: string
  guaranteeDays?: number
  sort?: number
  status: number
}

export async function fetchCoupons(params: {
  page?: number
  pageSize?: number
  status?: number
  keyword?: string
}): Promise<PageResp<CouponTemplate>> {
  return (await client.get('/admin/marketing/coupons', { params })) as any
}

export async function upsertCoupon(payload: {
  id?: string
  name: string
  type: number
  thresholdAmount?: string
  discountAmount?: string
  discountPercent?: number
  maxDiscountAmount?: string
  totalCount?: number
  perLimit?: number
  newUserOnly?: number
  pickupStart?: string
  pickupEnd?: string
  validStart?: string
  validEnd?: string
  status?: number
}) {
  if (payload.id) await client.put(`/admin/marketing/coupons/${payload.id}`, payload)
  else await client.post('/admin/marketing/coupons', payload)
}

export async function deleteCoupon(id: string) {
  await client.delete(`/admin/marketing/coupons/${id}`)
}

export async function issueCoupon(id: string, memberIds: string[]) {
  return (await client.post(`/admin/marketing/coupons/${id}/issue`, { memberIds })) as any
}

export async function fetchServices(): Promise<ServiceItemRow[]> {
  return (await client.get('/admin/marketing/services')) as any
}

export async function upsertService(payload: {
  id?: string
  name: string
  description?: string
  originalPrice?: string
  price: string
  guaranteeDays?: number
  sort?: number
  status?: number
}) {
  if (payload.id) await client.put(`/admin/marketing/services/${payload.id}`, payload)
  else await client.post('/admin/marketing/services', payload)
}

export async function deleteService(id: string) {
  await client.delete(`/admin/marketing/services/${id}`)
}

// ───────── 评价 / 售后 ─────────

export interface ReviewRow {
  id: string
  orderNo: string
  memberId: string
  nickname: string
  avatar: string
  productId: string
  productTitle: string
  rating: number
  healthScore?: number
  lookScore?: number
  serviceScore?: number
  content: string
  images: string[]
  status: number
  createdAt: string
  reply?: string
  repliedAt?: string
}

export async function fetchReviews(params: {
  cursor?: string
  limit?: number
}): Promise<{ list: ReviewRow[]; hasMore: boolean }> {
  return (await client.get('/admin/reviews', { params })) as any
}

export async function updateReviewStatus(id: string, status: number) {
  await client.put(`/admin/reviews/${id}/status`, { status })
}

export async function replyReview(id: string, reply: string) {
  await client.post(`/admin/reviews/${id}/reply`, { reply })
}

export async function deleteReview(id: string) {
  await client.delete(`/admin/reviews/${id}`)
}

export interface AfterSaleRow {
  id: string
  afterSaleNo: string
  orderNo: string
  memberId: string
  nickname: string
  phone: string
  reason: string
  refundAmount: string
  status: number
  statusText: string
  adminNote: string
  refundPaymentNo: string
  auditAt: string
  createdAt: string
}

export async function fetchAfterSales(params: {
  page?: number
  pageSize?: number
  status?: number
}): Promise<PageResp<AfterSaleRow>> {
  return (await client.get('/admin/aftersales', { params })) as any
}

export async function auditAfterSale(afterSaleNo: string, agree: boolean, amount?: string, note?: string) {
  await client.post(`/admin/aftersales/${afterSaleNo}/audit`, { agree, amount, note })
}

// ───────── Banner 运营位 ─────────

export interface BannerRow {
  id: string
  title: string
  subTitle: string
  icon: string
  jumpType: string
  target: string
  sort: number
  status: number
  createdAt: string
}

export async function fetchBanners(params: { page?: number; pageSize?: number }): Promise<PageResp<BannerRow>> {
  return (await client.get('/admin/banners', { params })) as any
}

export async function upsertBanner(payload: {
  id?: string
  title: string
  subTitle?: string
  icon?: string
  jumpType: string
  target?: string
  sort?: number
  status?: number
}) {
  if (payload.id) {
    await client.put(`/admin/banners/${payload.id}`, payload)
  } else {
    await client.post('/admin/banners', payload)
  }
}

export async function updateBannerStatus(id: string, status: number) {
  await client.put(`/admin/banners/${id}/status`, { status })
}

export async function deleteBanner(id: string) {
  await client.delete(`/admin/banners/${id}`)
}

// ───────── 配送方式 / 托运 ─────────

export interface ShipMethodRow {
  id: string
  name: string
  kind: number
  description: string
  fee: string
  sort: number
  status: number
  createdAt: string
}

export async function fetchShipMethods(params: { page?: number; pageSize?: number }): Promise<PageResp<ShipMethodRow>> {
  return (await client.get('/admin/ship-methods', { params })) as any
}

export async function upsertShipMethod(payload: {
  id?: string
  name: string
  kind: number
  description?: string
  fee?: string
  sort?: number
  status?: number
}) {
  if (payload.id) {
    await client.put(`/admin/ship-methods/${payload.id}`, payload)
  } else {
    await client.post('/admin/ship-methods', payload)
  }
}

export async function updateShipMethodStatus(id: string, status: number) {
  await client.put(`/admin/ship-methods/${id}/status`, { status })
}

export async function deleteShipMethod(id: string) {
  await client.delete(`/admin/ship-methods/${id}`)
}

// ───────── 报表 / 秒杀 / 审计 / 敏感词 ─────────

export interface ReportRow {
  date: string
  orders: number
  gmv: string
}

export interface RankRow {
  name: string
  count: number
  amount: string
}

export interface ReportResp {
  summary: {
    todayGmv: string
    todayOrders: number
    monthGmv: string
    monthOrders: number
    pendingAfters: number
  }
  daily: ReportRow[]
  topProducts: RankRow[]
  breeds: RankRow[]
}

export async function fetchReport(days?: number): Promise<ReportResp> {
  return (await client.get('/admin/report', { params: { days } })) as any
}

export interface FlashSaleRow {
  id: string
  productId: string
  productTitle: string
  salePrice: string
  stock: number
  sold: number
  startAt: string
  endAt: string
  status: number
  createdAt: string
}

export async function fetchFlashSales(params: { page?: number; pageSize?: number }): Promise<PageResp<FlashSaleRow>> {
  return (await client.get('/admin/flash-sales', { params })) as any
}

export async function upsertFlashSale(payload: {
  id?: string
  productId: string
  salePrice: string
  stock: number
  startAt: string
  endAt: string
  status?: number
}) {
  if (payload.id) await client.put(`/admin/flash-sales/${payload.id}`, payload)
  else await client.post('/admin/flash-sales', payload)
}

export async function updateFlashSaleStatus(id: string, status: number) {
  await client.put(`/admin/flash-sales/${id}/status`, { status })
}

export async function deleteFlashSale(id: string) {
  await client.delete(`/admin/flash-sales/${id}`)
}

export interface AuditLogRow {
  id: string
  adminName: string
  method: string
  path: string
  ip: string
  createdAt: string
}

export async function fetchAuditLogs(params: { page?: number; pageSize?: number }): Promise<PageResp<AuditLogRow>> {
  return (await client.get('/admin/audit-logs', { params })) as any
}

export interface SensitiveWordRow {
  id: string
  word: string
  status: number
  createdAt: string
}

export async function fetchSensitiveWords(params: { page?: number; pageSize?: number }): Promise<PageResp<SensitiveWordRow>> {
  return (await client.get('/admin/sensitive-words', { params })) as any
}

export async function saveSensitiveWord(word: string) {
  await client.post('/admin/sensitive-words', { word })
}

export async function updateSensitiveWordStatus(id: string, status: number) {
  await client.put(`/admin/sensitive-words/${id}/status`, { status })
}

export async function deleteSensitiveWord(id: string) {
  await client.delete(`/admin/sensitive-words/${id}`)
}

// ───────── 供货商 / 财务对账 / 数据导出 ─────────

export interface SupplierRow {
  id: string
  name: string
  contact: string
  phone: string
  address: string
  productCount: number
  status: number
  createdAt: string
}

export async function fetchSuppliers(params: {
  page?: number
  pageSize?: number
  keyword?: string
  status?: number
}): Promise<PageResp<SupplierRow>> {
  return (await client.get('/admin/suppliers', { params })) as any
}

export async function upsertSupplier(payload: {
  id?: string
  name: string
  contact?: string
  phone?: string
  address?: string
  status?: number
}) {
  if (payload.id) await client.put(`/admin/suppliers/${payload.id}`, payload)
  else await client.post('/admin/suppliers', payload)
}

export async function updateSupplierStatus(id: string, status: number) {
  await client.put(`/admin/suppliers/${id}/status`, { status })
}

export async function deleteSupplier(id: string) {
  await client.delete(`/admin/suppliers/${id}`)
}

// ───────── 拼团活动 ─────────

export interface GroupBuyRow {
  id: string
  productId: string
  productTitle: string
  price: string
  size: number
  hours: number
  openTeams?: number
  status: number
  createdAt: string
}

export async function fetchGroupBuys(params: { page?: number; pageSize?: number }): Promise<PageResp<GroupBuyRow>> {
  return (await client.get('/admin/group-buys', { params })) as any
}

export async function upsertGroupBuy(payload: {
  id?: string
  productId: string
  price: string
  size: number
  hours: number
  status?: number
}) {
  if (payload.id) await client.put(`/admin/group-buys/${payload.id}`, payload)
  else await client.post('/admin/group-buys', payload)
}

export async function updateGroupBuyStatus(id: string, status: number) {
  await client.put(`/admin/group-buys/${id}/status`, { status })
}

export async function deleteGroupBuy(id: string) {
  await client.delete(`/admin/group-buys/${id}`)
}

// ───────── 库存预警 / 风控 / 自提核销 / 会员黑名单 ─────────

export interface StockAlertRow {
  productId: string
  title: string
  stock: number
  threshold: number
  sales: number
}

export async function fetchStockAlerts(): Promise<StockAlertRow[]> {
  const r = (await client.get('/admin/stock-alerts')) as any
  return r?.list ?? []
}

export interface RiskLogRow {
  id: string
  memberId: string
  rule: string
  detail: string
  createdAt: string
}

export async function fetchRiskLogs(params: {
  page?: number
  pageSize?: number
  rule?: string
}): Promise<{ total: number; list: RiskLogRow[] }> {
  return (await client.get('/admin/risk-logs', { params })) as any
}

export async function pickupVerify(orderNo: string, code: string) {
  await client.post(`/admin/orders/${orderNo}/pickup-verify`, { code })
}

export async function setMemberBlacklist(id: string, blacklist: number) {
  await client.put('/admin/members/blacklist', { id, blacklist })
}

export interface FinancePaymentRow {
  orderNo: string
  memberName: string
  productTitle: string
  amount: string
  refundAmount: string
  payType: number
  payTypeText: string
  status: number
  statusText: string
  paymentNo: string
  paidAt: string
  createdAt: string
}

export async function fetchFinancePayments(params: {
  page?: number
  pageSize?: number
  orderNo?: string
  status?: number
  payType?: number
}): Promise<PageResp<FinancePaymentRow>> {
  return (await client.get('/admin/finance/payments', { params })) as any
}

export interface FinanceDailyRow {
  date: string
  orderCount: number
  payAmount: string
  refundAmount: string
  netAmount: string
}

export async function fetchFinanceDaily(days?: number): Promise<FinanceDailyRow[]> {
  return (await client.get('/admin/finance/daily', { params: { days } })) as any
}

export async function downloadCSV(kind: 'orders' | 'members' | 'points', filename: string) {
  // client 的响应拦截器已解包 resp.data，此处返回值即 Blob 本体
  const blob = (await client.get(`/admin/export/${kind}`, { responseType: 'blob' })) as unknown as Blob
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

