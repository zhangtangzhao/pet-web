import { useEffect, useMemo, useState } from 'react'
import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import {
  CreateOrderResult,
  ProductDetail,
  ServiceItemView,
  UsableCouponView,
} from '../../types'
import './index.css'

const fmtDate = (iso: string) => {
  if (!iso) return '长期有效'
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export default function Checkout() {
  const { params } = useRouter()
  const [d, setD] = useState<ProductDetail>()
  const [services, setServices] = useState<ServiceItemView[]>([])
  const [pickedSvc, setPickedSvc] = useState<string[]>([])
  const [coupons, setCoupons] = useState<UsableCouponView[]>([])
  const [couponId, setCouponId] = useState('')
  const [sheetOpen, setSheetOpen] = useState(false)
  const [name, setName] = useState('')
  const [phone, setPhone] = useState('')
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent(`/pages/checkout/index?id=${params.id}`)}`,
      }).catch(() => {})
      return
    }
    get<ProductDetail>(`/products/${params.id}`)
      .then(setD)
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))
    get<ServiceItemView[]>('/services').then(setServices).catch(() => {})
  }, [params.id])

  // 勾选服务变化后重新拉取可用券（门槛含服务费）
  useEffect(() => {
    if (!d) return
    post<UsableCouponView[]>('/coupons/usable', {
      productId: params.id,
      serviceIds: pickedSvc,
    })
      .then((list) => {
        setCoupons(list ?? [])
        setCouponId((cur) => (list?.some((c) => c.id === cur) ? cur : ''))
      })
      .catch(() => setCoupons([]))
  }, [d, pickedSvc.join(',')])

  const toggleSvc = (id: string) =>
    setPickedSvc((cur) => (cur.includes(id) ? cur.filter((x) => x !== id) : [...cur, id]))

  const serviceFee = useMemo(
    () => services.filter((s) => pickedSvc.includes(s.id)).reduce((n, s) => n + Number(s.price), 0),
    [services, pickedSvc],
  )
  const goodsPrice = d ? Number(d.price) : 0
  const total = goodsPrice + serviceFee
  const coupon = coupons.find((c) => c.id === couponId)
  const discount = coupon ? Number(coupon.discount) : 0
  const pay = Math.max(total - discount, 0.01)

  const submit = async () => {
    if (submitting) return
    if (!name || !phone) {
      Taro.showToast({ title: '请填写联系人和手机号', icon: 'none' })
      return
    }
    setSubmitting(true)
    try {
      const r = await post<CreateOrderResult>('/orders', {
        productId: params.id,
        serviceIds: pickedSvc,
        couponId,
        contactName: name,
        contactPhone: phone,
      })
      Taro.showModal({
        title: '订单已创建',
        content: `订单号 ${r.orderNo}，实付 ¥${r.payAmount}。微信支付待商户号联调后开放。`,
        showCancel: false,
        success: () => Taro.redirectTo({ url: '/pages/index/index' }),
      })
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  if (!d) return <View className='checkout'>加载中…</View>

  return (
    <View className='checkout'>
      <View className='co-card card'>
        <View className='co-prod'>
          {d.mainImage ? (
            <Image className='co-prod-img' src={d.mainImage} mode='aspectFill' />
          ) : (
            <View className='co-prod-img'>🐾</View>
          )}
          <View className='co-prod-main'>
            <Text className='co-prod-title'>{d.title}</Text>
            <Text className='price'>¥{d.price}</Text>
          </View>
        </View>
      </View>

      {services.length > 0 && (
        <View className='co-card card'>
          <View className='co-sec-title'>增值服务</View>
          {services.map((s) => {
            const on = pickedSvc.includes(s.id)
            return (
              <View className='co-svc' key={s.id} onClick={() => toggleSvc(s.id)}>
                <View className={`co-svc-check ${on ? 'co-svc-check-on' : ''}`}>{on ? '✓' : ''}</View>
                <View className='co-svc-main'>
                  <Text className='co-svc-name'>{s.name}</Text>
                  <Text className='co-svc-desc'>{s.description}</Text>
                </View>
                <View className='co-svc-price'>
                  <Text className='co-svc-now'>¥{s.price}</Text>
                  {Number(s.originalPrice) > Number(s.price) && (
                    <Text className='co-svc-orig'>¥{s.originalPrice}</Text>
                  )}
                </View>
              </View>
            )
          })}
        </View>
      )}

      <View className='co-card card'>
        <View className='co-coupon' onClick={() => setSheetOpen(true)}>
          <Text className='co-coupon-left'>优惠券</Text>
          <Text className='co-coupon-arrow'>
            {coupon ? `-¥${coupon.discount}` : `${coupons.length > 0 ? `${coupons.length} 张可用` : '暂无可用'} >`}
          </Text>
        </View>
      </View>

      <View className='co-card card'>
        <View className='co-sum-row'>
          <Text>商品价</Text>
          <Text>¥{goodsPrice.toFixed(2)}</Text>
        </View>
        {serviceFee > 0 && (
          <View className='co-sum-row'>
            <Text>增值服务</Text>
            <Text>¥{serviceFee.toFixed(2)}</Text>
          </View>
        )}
        {discount > 0 && (
          <View className='co-sum-row'>
            <Text>券抵扣（{coupon?.name}）</Text>
            <Text className='co-sum-discount'>-¥{discount.toFixed(2)}</Text>
          </View>
        )}
      </View>

      <View className='co-card card'>
        <View className='co-sec-title'>联系人信息</View>
        <Input className='co-input' placeholder='联系人姓名' value={name} onInput={(e) => setName(e.detail.value)} />
        <Input className='co-input' placeholder='联系手机号' type='number' maxlength={11} value={phone} onInput={(e) => setPhone(e.detail.value)} />
      </View>

      <View className='footer'>
        <View className='pay'>
          实付 <Text className='pay-num'>¥{pay.toFixed(2)}</Text>
        </View>
        <View className={`btn-submit ${submitting ? 'btn-submit-off' : ''}`} onClick={submit}>
          提交订单
        </View>
      </View>

      {sheetOpen && (
        <View>
          <View className='mask' onClick={() => setSheetOpen(false)} />
          <View className='sheet'>
            <View className='sheet-title'>选择优惠券</View>
            {coupons.length === 0 && <View className='cp-empty'>暂无可用优惠券</View>}
            {coupons.map((c) => (
              <View
                className={`cp-item ${c.id === couponId ? 'cp-item-on' : ''}`}
                key={c.id}
                onClick={() => {
                  setCouponId(c.id === couponId ? '' : c.id)
                  setSheetOpen(false)
                }}
              >
                <View className='cp-left'>
                  {c.discountValid ? (
                    <View>
                      <Text className='cp-val-unit'>¥</Text>
                      <Text className='cp-val'>{Number(c.discount).toFixed(2).replace(/\.00$/, '')}</Text>
                      <View className='cp-cond'>{Number(c.threshold) > 0 ? `满${Number(c.threshold)}可用` : '无门槛'}</View>
                    </View>
                  ) : (
                    <View>
                      <Text className='cp-val'>{c.percent / 10}</Text>
                      <Text className='cp-val-unit'>折</Text>
                      <View className='cp-cond'>省¥{Number(c.discount).toFixed(2)}</View>
                    </View>
                  )}
                </View>
                <View className='cp-main'>
                  <Text className='cp-name'>{c.name}</Text>
                  <Text className='cp-date'>{fmtDate(c.validEnd)} 前使用</Text>
                </View>
                <Text className={`cp-pick ${c.id === couponId ? 'cp-pick-on' : ''}`}>
                  {c.id === couponId ? '已选' : '选择'}
                </Text>
              </View>
            ))}
          </View>
        </View>
      )}
    </View>
  )
}
