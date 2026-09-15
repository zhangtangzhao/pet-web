import { useCallback, useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { askSubscribe, cancelAfterSale, cancelOrder, confirmOrder, goAfterSale, goReview, payOrder, prefetchTmplIds } from '../../orderActions'
import { get, getToken } from '../../request'
import { OrderView } from '../../types'
import './index.css'

const SHIP_TEXT: Record<number, string> = { 0: '待配送', 1: '配送中', 2: '已送达' }

// 时间线节点（自提单无发货/送达环节，自动省略）
function buildTimeline(o: OrderView) {
  const nodes = [{ label: '创建订单', time: o.createdAt }]
  if (o.shipAddress) {
    nodes.push({ label: '支付成功', time: o.paidAt || '' })
    nodes.push({ label: '商家发货', time: o.shippedAt || '' })
    nodes.push({ label: '送达', time: o.deliveredAt || '' })
  } else {
    nodes.push({ label: '支付成功', time: o.paidAt || '' })
  }
  nodes.push({ label: '交易完成', time: o.completedAt || '' })
  return nodes
}

const fmtTime = (s: string) => (s ? s.slice(0, 16).replace('T', ' ') : '')

export default function OrderDetail() {
  const { params } = useRouter()
  const orderNo = params.orderNo || ''
  const [o, setO] = useState<OrderView | null>(null)
  const [leftSec, setLeftSec] = useState(0)

  const load = useCallback(() => {
    return get<OrderView>(`/orders/${orderNo}`)
      .then(setO)
      .catch((e: any) => {
        Taro.showToast({ title: e.message, icon: 'none' })
      })
  }, [orderNo])

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent(`/pages/order-detail/index?orderNo=${orderNo}`)}`,
      }).catch(() => {})
      return
    }
    prefetchTmplIds()
    load()
  }, [load])

  // 待支付倒计时（与后端关单同源 expireAt）
  useEffect(() => {
    if (!o || o.status !== 10) return
    const tick = () => setLeftSec(Math.max(0, Math.floor((new Date(o.expireAt).getTime() - Date.now()) / 1000)))
    tick()
    const t = setInterval(tick, 1000)
    return () => clearInterval(t)
  }, [o])

  if (!o) return <View className='odt' />

  const reload = () => load()
  const asStatus = o.aftersaleStatus ?? 0
  const goodsAmount = (Number(o.totalAmount) - Number(o.serviceFee)).toFixed(2)
  const nodes = buildTimeline(o)
  const doneCount = nodes.filter((n) => n.time).length
  const terminal = o.status === 40 || o.status === 45 || o.status === 60
  const mm = Math.floor(leftSec / 60)
  const ss = String(leftSec % 60).padStart(2, '0')

  const renderActions = () => {
    return (
      <View className='odt-actions'>
        {o.status === 10 && (
          <>
            <View className='odt-btn' onClick={() => cancelOrder(o.orderNo, reload)}>
              取消订单
            </View>
            <View className='odt-btn odt-btn-primary' onClick={() => payOrder(o.orderNo, reload)}>
              去支付
            </View>
          </>
        )}
        {o.status === 20 && (
          <>
            <View className='odt-btn' onClick={() => confirmOrder(o.orderNo, reload)}>
              确认收货
            </View>
            {asStatus === 0 && (
              <View
                className='odt-btn'
                onClick={() => {
                  askSubscribe()
                  goAfterSale(o.orderNo)
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
              <View className='odt-btn odt-btn-disabled'>已评价</View>
            ) : (
              <View className='odt-btn odt-btn-primary' onClick={() => goReview(o.orderNo)}>
                去评价
              </View>
            )}
            {asStatus === 0 && (
              <View
                className='odt-btn'
                onClick={() => {
                  askSubscribe()
                  goAfterSale(o.orderNo)
                }}
              >
                申请售后
              </View>
            )}
          </>
        )}
        {asStatus === 1 && (
          <View className='odt-btn' onClick={() => cancelAfterSale(o.orderNo, reload)}>
            撤销售后
          </View>
        )}
      </View>
    )
  }

  return (
    <View className='odt'>
      {/* 状态卡 */}
      <View className={`odt-hero odt-hero-${o.status}`}>
        <Text className='odt-hero-status'>{o.statusText}</Text>
        {o.status === 10 && (
          <Text className='odt-hero-tip'>{leftSec > 0 ? `请在 ${mm}分${ss}秒 内完成支付` : '订单已超时，即将自动关闭'}</Text>
        )}
        {o.status === 20 && o.shipAddress && <Text className='odt-hero-tip'>{SHIP_TEXT[o.shipStatus]}，请保持电话畅通</Text>}
        {o.status === 30 && <Text className='odt-hero-tip'>感谢您的信任，欢迎评价本次交易</Text>}
        {terminal && <Text className='odt-hero-tip'>{o.status === 60 ? '退款已原路退回' : '订单已关闭'}</Text>}
      </View>

      {/* 状态时间线 */}
      {!terminal && (
        <View className='odt-card'>
          <Text className='odt-card-title'>订单进度</Text>
          <View className='odt-timeline'>
            {nodes.map((n, i) => (
              <View className='odt-node' key={n.label}>
                <View className='odt-node-left'>
                  <View className={`odt-dot ${n.time ? 'odt-dot-on' : ''} ${i === doneCount - 1 ? 'odt-dot-cur' : ''}`} />
                  {i < nodes.length - 1 && <View className='odt-line' />}
                </View>
                <View className='odt-node-main'>
                  <Text className={`odt-node-label ${n.time ? 'odt-node-label-on' : ''}`}>{n.label}</Text>
                  <Text className='odt-node-time'>{n.time ? fmtTime(n.time) : '待完成'}</Text>
                </View>
              </View>
            ))}
          </View>
        </View>
      )}
      {terminal && (o.status === 40 || o.status === 45) && (
        <View className='odt-card'>
          <Text className='odt-card-title'>关闭原因</Text>
          <Text className='odt-reason'>{o.status === 45 ? '超时未支付，系统自动关闭' : '买家主动取消订单'}</Text>
        </View>
      )}

      {/* 商品 */}
      <View className='odt-card'>
        {o.items.map((it) => (
          <View className='odt-prod' key={it.productId}>
            {it.productImage ? (
              <Image className='odt-img' src={it.productImage} mode='aspectFill' />
            ) : (
              <View className='odt-img'>🐾</View>
            )}
            <View className='odt-prod-main'>
              <Text className='odt-title'>{it.productTitle}</Text>
              <Text className='odt-breed'>{it.breedName}</Text>
              <Text className='price'>¥{it.price}</Text>
            </View>
          </View>
        ))}
      </View>

      {/* 金额明细 */}
      <View className='odt-card'>
        <View className='odt-amt-row'>
          <Text>商品金额</Text>
          <Text>¥{goodsAmount}</Text>
        </View>
        {Number(o.serviceFee) > 0 && (
          <View className='odt-amt-row'>
            <Text>增值服务</Text>
            <Text>¥{o.serviceFee}</Text>
          </View>
        )}
        {Number(o.shipFee) > 0 && (
          <View className='odt-amt-row'>
            <Text>运费</Text>
            <Text>¥{o.shipFee}</Text>
          </View>
        )}
        {Number(o.discountAmount) > 0 && (
          <View className='odt-amt-row odt-amt-discount'>
            <Text>优惠券{o.couponInfo ? `（${o.couponInfo}）` : ''}</Text>
            <Text>-¥{o.discountAmount}</Text>
          </View>
        )}
        <View className='odt-amt-row odt-amt-pay'>
          <Text>实付款</Text>
          <Text className='odt-amt-pay-num'>¥{o.payAmount}</Text>
        </View>
      </View>

      {/* 配送信息 */}
      <View className='odt-card'>
        <Text className='odt-card-title'>配送信息</Text>
        <View className='odt-kv'>
          <Text className='odt-kv-label'>配送方式</Text>
          <Text className='odt-kv-value'>
            {o.shipMethod}（¥{o.shipFee}）
          </Text>
        </View>
        {o.shipAddress && (
          <View className='odt-kv'>
            <Text className='odt-kv-label'>收货地址</Text>
            <Text className='odt-kv-value'>{o.shipAddress}</Text>
          </View>
        )}
        {o.shipAddress && (
          <View className='odt-kv'>
            <Text className='odt-kv-label'>配送状态</Text>
            <Text className='odt-kv-value'>
              {SHIP_TEXT[o.shipStatus]}
              {o.shipNo ? ` · 单号 ${o.shipNo}` : ''}
            </Text>
          </View>
        )}
        <View className='odt-kv'>
          <Text className='odt-kv-label'>联系人</Text>
          <Text className='odt-kv-value'>
            {o.contactName} {o.contactPhone}
          </Text>
        </View>
        {o.remark && (
          <View className='odt-kv'>
            <Text className='odt-kv-label'>订单备注</Text>
            <Text className='odt-kv-value'>{o.remark}</Text>
          </View>
        )}
      </View>

      {/* 售后进度入口 */}
      {asStatus > 0 && (
        <View className='odt-card odt-as' onClick={() => goAfterSale(o.orderNo)}>
          <Text className='odt-card-title'>售后进度</Text>
          <View className='odt-kv'>
            <Text className='odt-kv-value'>查看售后进度 ›</Text>
          </View>
        </View>
      )}

      {renderActions()}

      <View className='odt-back' onClick={() => Taro.redirectTo({ url: '/pages/orders/index' })}>
        返回订单列表
      </View>
    </View>
  )
}
