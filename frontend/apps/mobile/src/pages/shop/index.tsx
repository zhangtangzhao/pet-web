import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'

interface ShopItem { id: string; name: string; pointsCost: number; stock: number; type: number; typeText: string; couponName: string; description: string }
interface PointsOrder { orderNo: string; productName: string; pointsCost: number; status: number; statusText: string; shipNo: string }

export default function PointsShop() {
  const [items, setItems] = useState<ShopItem[]>([])
  const [orders, setOrders] = useState<PointsOrder[]>([])

  const load = useCallback(() => {
    get<{ list: ShopItem[] }>('/points-shop').then((r) => setItems(r.list ?? [])).catch(() => {})
    if (getToken()) get<{ list: PointsOrder[] }>('/points-orders').then((r) => setOrders(r.list ?? [])).catch(() => {})
  }, [])

  useEffect(() => { load() }, [load])

  const exchange = (it: ShopItem) => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fshop%2Findex' }).catch(() => {}); return }
    const doEx = (extra?: Record<string, string>) =>
      post('/points-shop/' + it.id + '/exchange', extra ?? {})
        .then(() => { Taro.showToast({ title: '兑换成功', icon: 'success' }); load() })
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    if (it.type === 2) {
      (Taro.showModal as any)({
        title: '实物兑换', editable: true, placeholderText: '收货人|手机号|地址',
        success: (m) => {
          if (!m.confirm || !m.content) return
          const parts = m.content.split('|')
          if (parts.length < 3) { Taro.showToast({ title: '格式：收货人|手机号|地址', icon: 'none' }); return }
          doEx({ contact: parts[0], phone: parts[1], address: parts.slice(2).join('|') })
        },
      })
      return
    }
    doEx()
  }

  return (
    <View className='pshop'>
      <View className='pshop-title'>积分商城</View>
      {items.map((it) => (
        <View className='pshop-card' key={it.id}>
          <View className='pshop-main'>
            <Text className='pshop-name'>{it.name}</Text>
            {!!it.description && <Text className='pshop-desc'>{it.description}</Text>}
            <Text className='pshop-meta'>{it.typeText} · 剩 {it.stock}</Text>
          </View>
          <View className='pshop-right'>
            <Text className='pshop-cost'>{it.pointsCost} 积分</Text>
            <View className='pshop-btn' onClick={() => exchange(it)}>兑换</View>
          </View>
        </View>
      ))}
      {orders.length > 0 && (
        <View className='pshop-orders'>
          <View className='pshop-sub'>我的兑换</View>
          {orders.map((o) => (
            <View className='pshop-order' key={o.orderNo}>
              <Text>{o.productName}（{o.pointsCost} 积分）</Text>
              <Text className='pshop-status'>{o.statusText}{o.shipNo ? ' · ' + o.shipNo : ''}</Text>
            </View>
          ))}
        </View>
      )}
    </View>
  )
}

