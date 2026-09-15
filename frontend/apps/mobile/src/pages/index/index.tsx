import { useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get } from '../../request'
import { HomeResp, ProductCard } from '../../types'
import './index.css'

export default function Home() {
  const [data, setData] = useState<HomeResp>()

  const load = () =>
    get<HomeResp>('/home')
      .then(setData)
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))
      .finally(() => Taro.stopPullDownRefresh())

  useEffect(() => {
    load()
  }, [])

  usePullDownRefresh(load)

  const goDetail = (p: ProductCard) => Taro.navigateTo({ url: `/pages/detail/index?id=${p.id}` })
  const goCoupons = () => Taro.navigateTo({ url: '/pages/coupon-center/index' })
  const goOrders = () => Taro.navigateTo({ url: '/pages/orders/index' })
  const goNotify = () => Taro.navigateTo({ url: '/pages/notify/index' })

  return (
    <View className='home'>
      <View className='banner card'>
        <Text className='banner-title'>遇见你的毛孩子</Text>
        <Text className='banner-sub'>健康活体 · 平台保障 · 售后无忧</Text>
        <View className='banner-coupon' onClick={goCoupons}>
          <Text className='banner-coupon-icon'>🎫</Text>
          <Text>领券中心 · 新人立省 30 元</Text>
          <Text className='banner-coupon-arrow'>→</Text>
        </View>
        <View className='banner-coupon' onClick={goOrders}>
          <Text className='banner-coupon-icon'>📦</Text>
          <Text>我的订单 · 支付/收货/评价/售后</Text>
          <Text className='banner-coupon-arrow'>→</Text>
        </View>
        <View className='banner-coupon' onClick={goNotify}>
          <Text className='banner-coupon-icon'>🔔</Text>
          <Text>消息中心 · 订单/客服/优惠券提醒</Text>
          <Text className='banner-coupon-arrow'>→</Text>
        </View>
      </View>

      <View className='cats'>
        {(data?.categories ?? []).map((c) => (
          <View className='cat' key={c.id}>
            <View className='cat-icon'>{c.name.slice(0, 1)}</View>
            <Text className='cat-name'>{c.name}</Text>
          </View>
        ))}
      </View>

      <View className='section-title'>热销推荐</View>
      <View className='grid'>
        {(data?.hot ?? []).map((p) => (
          <View className='pet-card card' key={p.id} onClick={() => goDetail(p)}>
            {p.mainImage ? (
              <Image className='pet-img' src={p.mainImage} mode='aspectFill' lazyLoad />
            ) : (
              <View className='pet-img pet-img-empty'>🐾</View>
            )}
            <View className='pet-body'>
              <Text className='pet-title'>{p.title}</Text>
              <View className='pet-meta'>
                <Text className='price'>¥{p.price}</Text>
                <Text className='pet-sales'>已售{p.sales}</Text>
              </View>
            </View>
          </View>
        ))}
      </View>
    </View>
  )
}
