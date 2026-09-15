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
  expireAt: string
  paidAt: string
  createdAt: string
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
  orderCount: number
  favoriteCount: number
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
