import { client, PageResp } from './client'

export interface StoreRow { id: string; name: string; address: string; phone: string; businessHours: string; status: number; sort: number; createdAt: string }
export const fetchStores = async () => (await client.get('/admin/stores')) as any as StoreRow[]
export const upsertStore = async (p: { id?: string; name: string; address: string; phone?: string; businessHours?: string; sort?: number; status?: number }) => {
  if (p.id) await client.put(`/admin/stores/${p.id}`, p)
  else await client.post('/admin/stores', p)
}
export const updateStoreStatus = async (id: string, status: number) => client.put(`/admin/stores/${id}/status`, { status })
export const deleteStore = async (id: string) => client.delete(`/admin/stores/${id}`)

export interface PointsProductRow { id: string; name: string; pointsCost: number; stock: number; type: number; typeText: string; status: number }
export const fetchPointsProducts = async () => {
  const r = (await client.get('/admin/points-products')) as any
  return r?.list ?? []
}
export const upsertPointsProduct = async (p: any) => {
  if (p.id) await client.put(`/admin/points-products/${p.id}`, p)
  else await client.post('/admin/points-products', p)
}
export interface PointsOrderRow { id: string; orderNo: string; productName: string; pointsCost: number; status: number; statusText: string; shipNo: string; contact: string; phone: string; address: string }
export const fetchPointsOrders = async (params: { page?: number; pageSize?: number }) =>
  (await client.get('/admin/points-orders', { params })) as any as PageResp<PointsOrderRow>
export const shipPointsOrder = async (id: string, shipNo: string) =>
  client.put(`/admin/points-orders/${id}/ship`, { shipNo })

export interface PostRow { id: string; nickname: string; content: string; images: string[]; likeCount: number; status: number; createdAt: string }
export const fetchPosts = async (status?: number) => (await client.get('/admin/posts', { params: { status } })) as any as PostRow[]
export const setPostStatus = async (id: string, status: number) => client.put(`/admin/posts/${id}/status`, { status })
export const deletePost = async (id: string) => client.delete(`/admin/posts/${id}`)

export interface ActivityItem { type: string; typeText: string; id: string; name: string; productId?: string; startAt: string; endAt: string; status: number }
export const fetchActivityCalendar = async (days = 30) =>
  (await client.get('/admin/activity-calendar', { params: { days } })) as any as { items: ActivityItem[]; conflicts: string[] }

export interface RealtimeData {
  todayGmv: string
  todayOrders: number
  hourly: { hour: string; gmv: string; orders: number }[]
  funnel: Record<string, number>
  regions: { province: string; orders: number }[]
  generatedAt: string
}
export const fetchRealtime = async () => (await client.get('/admin/dashboard/realtime')) as any as RealtimeData

export const confirmExchange = async (afterSaleNo: string, exchangeShipNo: string) =>
  client.post(`/admin/aftersales/${afterSaleNo}/return-confirm`, { exchangeShipNo })
