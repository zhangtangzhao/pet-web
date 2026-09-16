import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { CouponTemplateView, MyCouponView } from '../../types'
import './index.css'

const fmtDate = (iso: string) => {
  if (!iso) return '长期有效'
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// 券卡片左侧展示：立减/满减显示金额，折扣显示折扣率
function CouponFace({ type, amount, percent, threshold, off }: {
  type: number; amount: string; percent: number; threshold: string; off?: boolean
}) {
  const cond = Number(threshold) > 0 ? `满${Number(threshold)}可用` : '无门槛'
  return (
    <View className={`cp-left ${off ? 'cp-left-off' : ''}`}>
      {type === 2 ? (
        <View>
          <Text className='cp-val'>{percent / 10}</Text>
          <Text className='cp-val-unit'>折</Text>
          <View className='cp-cond'>{cond}</View>
        </View>
      ) : (
        <View>
          <Text className='cp-val-unit'>¥</Text>
          <Text className='cp-val'>{Number(amount).toFixed(2).replace(/\.00$/, '')}</Text>
          <View className='cp-cond'>{cond}</View>
        </View>
      )}
    </View>
  )
}

export default function CouponCenter() {
  const [tab, setTab] = useState<0 | 1>(0)
  const [center, setCenter] = useState<CouponTemplateView[]>([])
  const [mine, setMine] = useState<MyCouponView[]>([])

  const loadCenter = useCallback(
    () => get<CouponTemplateView[]>('/coupons/center').then(setCenter).catch((e) => Taro.showToast({ title: e.message, icon: 'none' })),
    [],
  )
  const loadMine = useCallback(
    () => get<MyCouponView[]>('/coupons').then(setMine).catch((e) => Taro.showToast({ title: e.message, icon: 'none' })),
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fcoupon-center%2Findex' }).catch(() => {})
      return
    }
    loadCenter()
    loadMine()
  }, [loadCenter, loadMine])

  const claim = async (t: CouponTemplateView) => {
    const doClaim = () =>
      post(`/coupons/${t.id}/claim`)
        .then(() => {
          Taro.showToast({ title: (t.pointsCost ?? 0) > 0 ? '兑换成功' : '领取成功', icon: 'success' })
          loadCenter()
          loadMine()
        })
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    if ((t.pointsCost ?? 0) > 0) {
      Taro.showModal({
        title: '积分兑换',
        content: `确定消耗 ${t.pointsCost} 积分兑换「${t.name}」吗？`,
        success: (m) => {
          if (m.confirm) doClaim()
        },
      })
      return
    }
    doClaim()
  }

  return (
    <View className='cc'>
      <View className='cc-tabs'>
        <Text className={`cc-tab ${tab === 0 ? 'cc-tab-on' : ''}`} onClick={() => setTab(0)}>领券中心</Text>
        <Text className={`cc-tab ${tab === 1 ? 'cc-tab-on' : ''}`} onClick={() => setTab(1)}>我的优惠券</Text>
      </View>

      {tab === 0 &&
        (center.length === 0 ? (
          <View className='cc-empty'>暂无可领取的优惠券</View>
        ) : (
          center.map((t) => (
            <View className='cp-item' key={t.id}>
              <CouponFace type={t.type} amount={t.discountAmount} percent={t.discountPercent} threshold={t.thresholdAmount} />
              <View className='cp-main'>
                <Text className='cp-name'>{t.name}</Text>
                <Text className='cp-date'>
                  {fmtDate(t.validEnd)} 前有效 · 每人限领 {t.perLimit}
                  {t.newUserOnly ? ' · 新人专享' : ''}
                </Text>
              </View>
              <View className='btn-claim' onClick={() => claim(t)}>
                {(t.pointsCost ?? 0) > 0 ? `${t.pointsCost}积分` : '领取'}
              </View>
            </View>
          ))
        ))}

      {tab === 1 &&
        (mine.length === 0 ? (
          <View className='cc-empty'>
            还没有优惠券
            <View className='cc-tip'>去领券中心逛逛吧</View>
          </View>
        ) : (
          mine.map((c) => (
            <View className='cp-item' key={c.id}>
              <CouponFace
                type={c.type}
                amount={c.discount}
                percent={c.percent}
                threshold={c.threshold}
                off={c.status !== 1}
              />
              <View className='cp-main'>
                <Text className='cp-name'>{c.name}</Text>
                <Text className='cp-date'>{c.validEnd ? `${fmtDate(c.validEnd)} 前有效` : '长期有效'}</Text>
              </View>
              <Text className='cp-status'>{c.statusText}</Text>
            </View>
          ))
        ))}
    </View>
  )
}
