import { useEffect, useMemo, useState } from 'react'
import { Image, Input, Text, Textarea, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import {
  AddressView,
  CreateOrderResult,
  FlashSaleInfo,
  ProductDetail,
  ServiceItemView,
  ShipMethod,
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
  const [shipMethods, setShipMethods] = useState<ShipMethod[]>([])
  const [shipMethodId, setShipMethodId] = useState('')
  const [address, setAddress] = useState('')
  const [flash, setFlash] = useState<FlashSaleInfo | null>(null)
  const [addrs, setAddrs] = useState<AddressView[]>([])
  const [addrId, setAddrId] = useState('')
  const [depositPercent, setDepositPercent] = useState(0)
  const [levelRate, setLevelRate] = useState(1)
  const [levelName, setLevelName] = useState('')
  const [useDeposit, setUseDeposit] = useState(false)
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
    get<ShipMethod[]>('/ship/methods')
      .then((list) => {
        setShipMethods(list ?? [])
        setShipMethodId((cur) => (list?.some((m) => m.id === cur) ? cur : (list?.[0]?.id ?? '')))
      })
      .catch(() => {})
    get<FlashSaleInfo[]>('/flash-sales')
      .then((list) => setFlash((list ?? []).find((f) => f.productId === params.id) ?? null))
      .catch(() => {})
    get<AddressView[]>('/addresses')
      .then((list) => {
        setAddrs(list ?? [])
        const def = (list ?? []).find((a) => a.isDefault === 1)
        if (def) setAddrId(def.id)
      })
      .catch(() => {})
    get<{ depositPercent: number; depositHoldDays: number; myLevelName?: string; myDiscount?: string }>('/trade-config')
      .then((t) => {
        setDepositPercent(t.depositPercent || 0)
        if (t.myDiscount) {
          const r = Number(t.myDiscount)
          if (r > 0 && r < 1) {
            setLevelRate(r)
            setLevelName(t.myLevelName || '')
          }
        }
      })
      .catch(() => {})
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
  const goodsPrice = flash ? Number(flash.salePrice) : d ? Number(d.price) : 0
  const total = goodsPrice + serviceFee
  const coupon = coupons.find((c) => c.id === couponId)
  const discount = useDeposit ? 0 : coupon ? Number(coupon.discount) : 0
  const shipMethod = shipMethods.find((m) => m.id === shipMethodId)
  const shipFee = shipMethod ? Number(shipMethod.fee) : 0
  const round2 = (n: number) => Math.round(n * 100) / 100
  // 与后端一致：券抵扣（保底 0.01）→ 会员等级折扣作用于券后商品额 → 加运费（券不抵运费）
  const afterCoupon = Math.max(total - discount, 0.01)
  const goodsPay = levelRate < 1 ? round2(afterCoupon * levelRate) : afterCoupon
  const levelOff = round2(afterCoupon - goodsPay)
  const pay = goodsPay + shipFee
  // 定金锁宠：定金 = 商品+服务费的 N%（不支持用券），尾款 = 总额-定金
  const depositAmt = Math.round(Math.max(total, 0.01) * depositPercent) / 100
  const tailAmt = Math.max(pay - depositAmt, 0)

  const submit = async () => {
    if (submitting) return
    if (!name || !phone) {
      Taro.showToast({ title: '请填写联系人和手机号', icon: 'none' })
      return
    }
    if (!shipMethodId) {
      Taro.showToast({ title: '请选择配送方式', icon: 'none' })
      return
    }
    if (shipMethod?.kind === 2 && !addrId && !address.trim()) {
      Taro.showToast({ title: '请填写收货地址', icon: 'none' })
      return
    }
    setSubmitting(true)
    try {
      const r = await post<CreateOrderResult>('/orders', {
        productId: params.id,
        serviceIds: pickedSvc,
        couponId: useDeposit ? '' : couponId,
        shipMethodId,
        shipAddress: addrId ? '' : address.trim(),
        addressId: addrId,
        useDeposit,
        contactName: name,
        contactPhone: phone,
      })
      Taro.showModal({
        title: '订单已创建',
        content: r.isDeposit
          ? `订单号 ${r.orderNo}，已付定金 ¥${r.payAmount}，尾款 ¥${r.tailAmount} 请尽快补齐。`
          : `订单号 ${r.orderNo}，实付 ¥${r.payAmount}。微信支付待商户号联调后开放。`,
        cancelText: '返回首页',
        confirmText: '查看订单',
        success: (m) => {
          Taro.redirectTo({
            url: m.confirm ? `/pages/order-detail/index?orderNo=${r.orderNo}` : '/pages/index/index',
          })
        },
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
            <View className='co-prod-price'>
              <Text className='price'>¥{goodsPrice.toFixed(2)}</Text>
              {flash && <Text className='co-flash-tag'>秒杀</Text>}
              {flash && Number(d.price) > goodsPrice && <Text className='co-prod-orig'>¥{Number(d.price).toFixed(2)}</Text>}
            </View>
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

      {shipMethods.length > 0 && (
        <View className='co-card card'>
          <View className='co-sec-title'>配送方式</View>
          <View className='co-ship'>
            {shipMethods.map((m) => {
              const on = m.id === shipMethodId
              return (
                <View className={`co-ship-item ${on ? 'co-ship-item-on' : ''}`} key={m.id} onClick={() => setShipMethodId(m.id)}>
                  <View className='co-ship-main'>
                    <Text className='co-ship-name'>{m.name}</Text>
                    <Text className='co-ship-desc'>{m.description}</Text>
                  </View>
                  <Text className='co-ship-fee'>{Number(m.fee) > 0 ? `¥${Number(m.fee).toFixed(2)}` : '免运费'}</Text>
                </View>
              )
            })}
          </View>
          {shipMethod?.kind === 2 && (
            <View className='co-addr'>
              {addrs.length > 0 && (
                <View className='co-addr-chips'>
                  {addrs.map((a) => (
                    <View
                      key={a.id}
                      className={`co-addr-chip ${addrId === a.id ? 'co-addr-chip-on' : ''}`}
                      onClick={() => {
                        setAddrId(addrId === a.id ? '' : a.id)
                        setAddress('')
                      }}
                    >
                      {a.name}·{a.address.length > 14 ? `${a.address.slice(0, 14)}…` : a.address}
                      {a.isDefault === 1 ? ' (默认)' : ''}
                    </View>
                  ))}
                </View>
              )}
              <Textarea
                className='co-addr-input'
                placeholder={addrId ? '已选用地址簿地址' : '请填写收货地址（含城市 / 到达机场或门牌号）'}
                maxlength={255}
                value={addrId ? addrs.find((a) => a.id === addrId)?.address ?? '' : address}
                disabled={!!addrId}
                onInput={(e) => setAddress(e.detail.value)}
              />
            </View>
          )}
        </View>
      )}

      <View className='co-card card'>
        {depositPercent > 0 && (
          <View className='co-deposit' onClick={() => setUseDeposit((v) => !v)}>
            <View className='co-deposit-main'>
              <Text className='co-deposit-name'>定金锁宠</Text>
              <Text className='co-deposit-desc'>先付 {depositPercent}% 定金锁定宝贝，超时未补尾款定金不退</Text>
            </View>
            <View className={`co-svc-check ${useDeposit ? 'co-svc-check-on' : ''}`}>{useDeposit ? '✓' : ''}</View>
          </View>
        )}
        {useDeposit ? (
          <View className='co-coupon co-coupon-off'>
            <Text className='co-coupon-left'>优惠券</Text>
            <Text className='co-coupon-arrow'>定金锁宠暂不支持优惠券</Text>
          </View>
        ) : (
          <View className='co-coupon' onClick={() => setSheetOpen(true)}>
            <Text className='co-coupon-left'>优惠券</Text>
            <Text className='co-coupon-arrow'>
              {coupon ? `-¥${coupon.discount}` : `${coupons.length > 0 ? `${coupons.length} 张可用` : '暂无可用'} >`}
            </Text>
          </View>
        )}
      </View>

      <View className='co-card card'>
        <View className='co-sum-row'>
          <Text>商品价{flash ? '（秒杀价）' : ''}</Text>
          <Text>¥{goodsPrice.toFixed(2)}</Text>
        </View>
        {serviceFee > 0 && (
          <View className='co-sum-row'>
            <Text>增值服务</Text>
            <Text>¥{serviceFee.toFixed(2)}</Text>
          </View>
        )}
        {shipFee > 0 && (
          <View className='co-sum-row'>
            <Text>运费（{shipMethod?.name}）</Text>
            <Text>¥{shipFee.toFixed(2)}</Text>
          </View>
        )}
        {discount > 0 && (
          <View className='co-sum-row'>
            <Text>券抵扣（{coupon?.name}）</Text>
            <Text className='co-sum-discount'>-¥{discount.toFixed(2)}</Text>
          </View>
        )}
        {levelOff > 0 && (
          <View className='co-sum-row'>
            <Text>会员优惠{levelName ? `（${levelName} ${Math.round(levelRate * 100)}折）` : ''}</Text>
            <Text className='co-sum-discount'>-¥{levelOff.toFixed(2)}</Text>
          </View>
        )}
        {useDeposit && (
          <View className='co-sum-row'>
            <Text>定金（{depositPercent}%）</Text>
            <Text>¥{depositAmt.toFixed(2)}</Text>
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
          {useDeposit ? (
            <>
              定金 <Text className='pay-num'>¥{depositAmt.toFixed(2)}</Text>
              <Text className='pay-tail'>尾款 ¥{tailAmt.toFixed(2)}</Text>
            </>
          ) : (
            <>
              实付 <Text className='pay-num'>¥{pay.toFixed(2)}</Text>
            </>
          )}
        </View>
        <View className={`btn-submit ${submitting ? 'btn-submit-off' : ''}`} onClick={submit}>
          {useDeposit ? '付定金锁宠' : '提交订单'}
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
