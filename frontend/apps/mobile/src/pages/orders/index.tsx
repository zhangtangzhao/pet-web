import { useCallback, useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh, useRouter } from '@tarojs/taro'
import { askSubscribe, cancelAfterSale, cancelOrder, confirmOrder, goAfterSale, goReview, payOrder, prefetchTmplIds } from '../../orderActions'
import { get, getToken } from '../../request'
import { OrderView, PageResp } from '../../types'
import './index.css'

const CHIPS = [
  { status: 0, label: '全部' },
  { status: 10, label: '待支付' },
  { status: 15, label: '待补尾款' },
  { status: 20, label: '已支付' },
  { status: 30, label: '已完成' },
  { status: 60, label: '已退款' },
]

const AS_TEXT: Record<number, string> = { 1: '售后待审核', 2: '售后已同意', 3: '售后被拒绝', 4: '售后已撤销' }

const fmtLeft = (expireAt: string, now: number) => {
  const sec = Math.floor((new Date(expireAt).getTime() - now) / 1000)
  if (sec <= 0) return '已超时'
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return m > 0 ? `${m}分${String(s).padStart(2, '0')}秒内支付` : `${s}秒内支付`
}

const fmtShort = (iso: string) => {
  const d = new Date(iso)
  return `${d.getMonth() + 1}月${d.getDate()}日 ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

export default function Orders() {
  const { params } = useRouter()
  const [chip, setChip] = useState(Number(params.status) || 0)
  const [list, setList] = useState<OrderView[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)
  const [now, setNow] = useState(Date.now())

  // 待支付剩余时间小字：单一定时器驱动整页重渲染
  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 30000)
    return () => clearInterval(t)
  }, [])

  // 小程序订阅消息模板：进页预取，用户点击时同步唤起授权
  useEffect(() => {
    prefetchTmplIds()
  }, [])

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

  const reload = () => load(chip, 1, false)

  const stop = (fn: () => void) => (e: { stopPropagation: () => void }) => {
    e.stopPropagation()
    fn()
  }

  const goDetail = (o: OrderView) => Taro.navigateTo({ url: `/pages/order-detail/index?orderNo=${o.orderNo}` }).catch(() => {})

  const renderActions = (o: OrderView) => {
    const asPending = o.aftersaleStatus === 1
    return (
      <View className='od-actions'>
        {o.status === 10 && (
          <>
            <View className='od-btn' onClick={stop(() => cancelOrder(o.orderNo, reload))}>
              取消订单
            </View>
            <View className='od-btn od-btn-primary' onClick={stop(() => payOrder(o.orderNo, reload))}>
              去支付
            </View>
          </>
        )}
        {o.status === 15 && (
          <>
            <View className='od-btn' onClick={stop(() => cancelOrder(o.orderNo, reload))}>
              取消并退定金
            </View>
            <View className='od-btn od-btn-primary' onClick={stop(() => payOrder(o.orderNo, reload))}>
              补尾款
            </View>
          </>
        )}
        {o.status === 20 && (
          <>
            <View className='od-btn' onClick={stop(() => confirmOrder(o.orderNo, reload))}>
              确认收货
            </View>
            {(o.aftersaleStatus ?? 0) === 0 && (
              <View
                className='od-btn'
                onClick={stop(() => {
                  askSubscribe()
                  goAfterSale(o.orderNo)
                })}
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
              <View className='od-btn od-btn-primary' onClick={stop(() => goReview(o.orderNo))}>
                去评价
              </View>
            )}
            {(o.aftersaleStatus ?? 0) === 0 && (
              <View
                className='od-btn'
                onClick={stop(() => {
                  askSubscribe()
                  goAfterSale(o.orderNo)
                })}
              >
                申请售后
              </View>
            )}
          </>
        )}
        {asPending && (
          <View className='od-btn' onClick={stop(() => cancelAfterSale(o.orderNo, reload))}>
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
        <View className='od-item card' key={o.orderNo} onClick={() => goDetail(o)}>
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
          {o.status === 10 && (
            <View className='od-left'>
              <Text className={`od-left-text ${o.expireAt && new Date(o.expireAt).getTime() <= now ? 'od-left-expired' : ''}`}>
                {o.expireAt ? fmtLeft(o.expireAt, now) : ''}
              </Text>
            </View>
          )}
          {o.status === 15 && (
            <View className='od-left'>
              <Text className='od-left-text'>
                尾款 {Number(o.payAmount) - Number(o.depositAmount ?? '0')} 元，{o.tailExpireAt ? `${fmtShort(o.tailExpireAt)} 前补齐` : '请尽快补齐'}
              </Text>
            </View>
          )}
          <View className='od-foot'>
            <Text className='od-pay'>
              {o.status === 15 ? (
                <>
                  已付定金 <Text className='od-pay-num'>¥{o.depositAmount}</Text>
                </>
              ) : (
                <>
                  实付 <Text className='od-pay-num'>¥{o.payAmount}</Text>
                </>
              )}
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
