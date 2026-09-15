import Taro from '@tarojs/taro'
import { get, post } from './request'

// 订单操作（orders 列表页 / order-detail 详情页共用），onDone 为操作成功后的刷新回调

let tmplIds: string[] | null = null

// 小程序订阅消息模板：进页预取，用户点击时同步唤起授权
export function prefetchTmplIds() {
  if (process.env.TARO_ENV !== 'weapp') return
  get<{ miniTmplOrder: string }>('/notify/tmpl')
    .then((t) => {
      tmplIds = [t.miniTmplOrder].filter(Boolean)
    })
    .catch(() => {})
}

export function askSubscribe() {
  if (process.env.TARO_ENV !== 'weapp' || !tmplIds || tmplIds.length === 0) return
  // entityIds 为 Taro 类型定义中 alipay 专有必填项，weapp 运行时忽略
  Taro.requestSubscribeMessage({ tmplIds, entityIds: [] }).catch(() => {})
}

export function payOrder(orderNo: string, onDone: () => void) {
  post(`/orders/${orderNo}/prepay`)
    .then((r: any) => {
      const p = r?.payParams
      if (!p) {
        Taro.showToast({ title: '微信支付待商户号联调后开放', icon: 'none' })
        return
      }
      Taro.requestPayment({
        timeStamp: p.timeStamp,
        nonceStr: p.nonceStr,
        package: p.package,
        signType: p.signType,
        paySign: p.paySign,
        success: () => onDone(),
      }).catch(() => {})
    })
    .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
}

export function cancelOrder(orderNo: string, onDone: () => void) {
  Taro.showModal({
    title: '取消订单',
    content: '确定取消该订单吗？',
    success: (r) => {
      if (!r.confirm) return
      post(`/orders/${orderNo}/cancel`)
        .then(() => {
          Taro.showToast({ title: '已取消', icon: 'success' })
          onDone()
        })
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    },
  })
}

export function confirmOrder(orderNo: string, onDone: () => void) {
  Taro.showModal({
    title: '确认收货',
    content: '确认已完成交易？确认后可对订单进行评价。',
    success: (r) => {
      if (!r.confirm) return
      post(`/orders/${orderNo}/confirm`)
        .then(() => {
          Taro.showToast({ title: '交易已完成', icon: 'success' })
          onDone()
        })
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    },
  })
}

export function cancelAfterSale(orderNo: string, onDone: () => void) {
  Taro.showModal({
    title: '撤销售后',
    content: '确定撤销该售后申请吗？',
    success: (r) => {
      if (!r.confirm) return
      get<{ afterSaleNo: string }>(`/aftersale?orderNo=${orderNo}`)
        .then((a) => post(`/aftersale/${a.afterSaleNo}/cancel`))
        .then(() => {
          Taro.showToast({ title: '已撤销', icon: 'success' })
          onDone()
        })
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    },
  })
}

export const goReview = (orderNo: string) => Taro.navigateTo({ url: `/pages/order-review/index?orderNo=${orderNo}` })

export const goAfterSale = (orderNo: string) => Taro.navigateTo({ url: `/pages/after-sale/index?orderNo=${orderNo}` })
