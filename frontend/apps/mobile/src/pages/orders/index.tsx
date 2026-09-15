import { useCallback, useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh, useRouter } from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { OrderView, PageResp } from '../../types'
import './index.css'

const CHIPS = [
  { status: 0, label: '全部' },
  { status: 10, label: '待支付' },
  { status: 20, label: '已支付' },
  { status: 30, label: '已完成' },
  { status: 60, label: '已退款' },
]

const AS_TEXT: Record<number, string> = { 1: '售后待审核', 2: '售后已同意', 3: '售后被拒绝', 4: '售后已撤销' }

export default function Orders() {
  const { params } = useRouter()
  const [chip, setChip] = useState(Number(params.status) || 0)
  const [list, setList] = useState<OrderView[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [tmplIds, setTmplIds] = useState<string[]>([])

  // 小程序订阅消息模板：进页预取，用户点击时同步唤起授权
  useEffect(() => {
    if (process.env.TARO_ENV !== 'weapp') return
    get<{ miniTmplOrder: string }>('/notify/tmpl')
      .then((t) => setTmplIds([t.miniTmplOrder].filter(Boolean)))
      .catch(() => {})
  }, [])

  const askSubscribe = () => {
    if (process.env.TARO_ENV !== 'weapp' || tmplIds.length === 0) return
    // entityIds 为 Taro 类型定义中 alipay 专有必填项，weapp 运行时忽略
    Taro.requestSubscribeMessage({ tmplIds, entityIds: [] }).catch(() => {})
  }

  const load = useCallback(
    async (nextChip: number, nextPage: number, append: boolean) => {
      setLoading(true)
      try {
        const resp = await get<PageResp<OrderView>>(
          `/orders?page=${nextPage}&pageSize=10${nextChip ? `&status=${nextChip}` : ''}`,
        )
        setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
        setTotal(resp.total)
      } catch (e: any) {
        Taro.showToast({ title: e.message, icon: 'none' })
      } finally {
        setLoading(false)
      }
    },
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/orders/index')}`,
      }).catch(() => {})
      return
    }
    load(chip, 1, false)
  }, [chip, load])

  usePullDownRefresh(() => {
    load(chip, 1, false).finally(() => Taro.stopPullDownRefresh())
  })

  const pickChip = (s: number) => {
    setChip(s)
    setPage(1)
  }

  const loadMore = () => {
    if (loading || list.length >= total) return
    const next = page + 1
    setPage(next)
    load(chip, next, true)
  }

  const confirmOrder = (o: OrderView) => {
    Taro.showModal({
      title: '确认收货',
      content: '确认已完成交易？确认后可对订单进行评价。',
      success: (r) => {
        if (!r.confirm) return
        post(`/orders/${o.orderNo}/confirm`)
          .then(() => {
            Taro.showToast({ title: '交易已完成', icon: 'success' })
            load(chip, 1, false)
          })
          .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      },
    })
  }

  const payOrder = (o: OrderView) => {
    post(`/orders/${o.orderNo}/prepay`)
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
          success: () => load(chip, 1, false),
        }).catch(() => {})
      })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const cancelOrder = (o: OrderView) => {
    Taro.showModal({
      title: '取消订单',
      content: '确定取消该订单吗？',
      success: (r) => {
        if (!r.confirm) return
        post(`/orders/${o.orderNo}/cancel`)
          .then(() => {
            Taro.showToast({ title: '已取消', icon: 'success' })
            load(chip, 1, false)
          })
          .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      },
    })
  }

  const cancelAfterSale = (o: OrderView) => {
    Taro.showModal({
      title: '撤销售后',
      content: '确定撤销该售后申请吗？',
      success: (r) => {
        if (!r.confirm) return
        get<{ afterSaleNo: string }>(`/aftersale?orderNo=${o.orderNo}`)
          .then((a) => post(`/aftersale/${a.afterSaleNo}/cancel`))
          .then(() => {
            Taro.showToast({ title: '已撤销', icon: 'success' })
            load(chip, 1, false)
          })
          .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      },
    })
  }

  const goReview = (o: OrderView) => Taro.navigateTo({ url: `/pages/order-review/index?orderNo=${o.orderNo}` })
  const goAfterSale = (o: OrderView) => Taro.navigateTo({ url: `/pages/after-sale/index?orderNo=${o.orderNo}` })

  const renderActions = (o: OrderView) => {
    const asPending = o.aftersaleStatus === 1
    return (
      <View className='od-actions'>
        {o.status === 10 && (
          <>
            <View className='od-btn' onClick={() => cancelOrder(o)}>
              取消订单
            </View>
            <View className='od-btn od-btn-primary' onClick={() => payOrder(o)}>
              去支付
            </View>
          </>
        )}
        {o.status === 20 && (
          <>
            <View className='od-btn' onClick={() => confirmOrder(o)}>
              确认收货
            </View>
            {(o.aftersaleStatus ?? 0) === 0 && (
              <View
                className='od-btn'
                onClick={() => {
                  askSubscribe()
                  goAfterSale(o)
                }}
              >
                申请售后
              </View>
            )}
          </>
        )}
        {o.status === 30 && (
          <>
            {o.reviewed ? (
              <View className='od-btn od-btn-disabled'>已评价</View>
            ) : (
              <View className='od-btn od-btn-primary' onClick={() => goReview(o)}>
                去评价
              </View>
            )}
            {(o.aftersaleStatus ?? 0) === 0 && (
              <View
                className='od-btn'
                onClick={() => {
                  askSubscribe()
                  goAfterSale(o)
                }}
              >
                申请售后
              </View>
            )}
          </>
        )}
        {asPending && (
          <View className='od-btn' onClick={() => cancelAfterSale(o)}>
            撤销售后
          </View>
        )}
      </View>
    )
  }

  return (
    <View className='od'>
      <View className='od-chips'>
        {CHIPS.map((c) => (
          <View key={c.status} className={`od-chip ${chip === c.status ? 'od-chip-on' : ''}`} onClick={() => pickChip(c.status)}>
            {c.label}
          </View>
        ))}
      </View>

      {list.length === 0 && !loading && <View className='od-empty'>暂无订单，去逛逛吧～</View>}
      {list.map((o) => (
        <View className='od-item card' key={o.orderNo}>
          <View className='od-head'>
            <Text className='od-no'>{o.orderNo}</Text>
            <Text className={`od-status od-status-${o.status}`}>{o.statusText}</Text>
          </View>
          {o.items.map((it) => (
            <View className='od-prod' key={it.productId}>
              {it.productImage ? (
                <Image className='od-img' src={it.productImage} mode='aspectFill' />
              ) : (
                <View className='od-img'>🐾</View>
              )}
              <View className='od-prod-main'>
                <Text className='od-title'>{it.productTitle}</Text>
                <Text className='od-breed'>{it.breedName}</Text>
              </View>
              <Text className='price'>¥{it.price}</Text>
            </View>
          ))}
          <View className='od-foot'>
            <Text className='od-pay'>
              实付 <Text className='od-pay-num'>¥{o.payAmount}</Text>
            </Text>
            {(o.aftersaleStatus ?? 0) > 0 && <Text className='od-as'>{AS_TEXT[o.aftersaleStatus!]}</Text>}
          </View>
          {renderActions(o)}
        </View>
      ))}
      {list.length < total && (
        <View className='od-more' onClick={loadMore}>
          {loading ? '加载中…' : '加载更多'}
        </View>
      )}
    </View>
  )
}
