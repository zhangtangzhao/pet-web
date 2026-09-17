import { client } from './client'
export interface BargainRow { id: string; product_id: string; product_title: string; price: string; bottom_price: string; duration_hours: number; max_helpers: number; status: number }
export const fetchBargains = async () => (await client.get('/admin/bargain-activities')) as any as BargainRow[]
export const upsertBargain = async (p: any) => client.post('/admin/bargain-activities', p)
export interface AuctionRow { id: string; product_id: string; product_title: string; start_price: string; step_price: string; deposit_amount: string; start_at: string; end_at: string; status: number; highest_price: string }
export const fetchAuctions = async () => (await client.get('/admin/auctions')) as any as AuctionRow[]
export const upsertAuction = async (p: any) => client.post('/admin/auctions', p)
export interface BookingRow { bookingNo: string; serviceName: string; storeName: string; date: string; slot: string; contact: string; phone: string; price: string; status: number; statusText: string; verifyCode: string }
export const fetchBookings = async (params: { page?: number; pageSize?: number }) => (await client.get('/admin/bookings', { params })) as any as { list: BookingRow[] }
export const verifyBooking = async (bookingNo: string, code: string) => client.post('/admin/bookings/verify', { bookingNo, code })
export interface InsuranceApplyRow { id: string; nickname: string; product_name: string; contact: string; phone: string; status: number }
export const fetchInsuranceApplies = async () => (await client.get('/admin/insurance-applies')) as any as InsuranceApplyRow[]
export interface SegmentRow { segment: string; name: string; count: number }
export const fetchSegments = async () => (await client.get('/admin/member-segments')) as any as SegmentRow[]
export const issueSegmentCoupon = async (segment: string, couponTemplateId: string) => {
  const r = (await client.post('/admin/member-segments/issue', { segment, couponTemplateId })) as any
  return r?.granted ?? 0
}
