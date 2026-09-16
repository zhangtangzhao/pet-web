import { useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh, useRouter, useShareAppMessage } from '@tarojs/taro'
import { get, getToken } from '../../request'
import { BannerView, FlashSaleInfo, HomeResp, PageResp, ProductCard } from '../../types'
import './index.css'

// banner jump_type → 跳转（与后端 jumpTypes 白名单、管理端选项一一对应）
const BANNER_NAV: Record<string, (target: string) => void> = {
  me: () => Taro.navigateTo({ url: '/pages/me/index' }).catch(() => {}),
  recommend: () => Taro.navigateTo({ url: '/pages/recommend/index' }).catch(() => {}),
  coupon: () => Taro.navigateTo({ url: '/pages/coupon-center/index' }).catch(() => {}),
  orders: () => Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {}),
  notify: () => Taro.navigateTo({ url: '/pages/notify/index' }).catch(() => {}),
  favorites: () => Taro.navigateTo({ url: '/pages/favorites/index' }).catch(() => {}),
  product: (t) => Taro.navigateTo({ url: `/pages/detail/index?id=${t}` }).catch(() => {}),
  category: (t) => Taro.navigateTo({ url: `/pages/list/index?categoryId=${t}` }).catch(() => {}),
  search: (t) => Taro.navigateTo({ url: `/pages/list/index?keyword=${encodeURIComponent(t)}` }).catch(() => {}),
  custom: (t) => {
    if (t.startsWith('/pages/')) Taro.navigateTo({ url: t }).catch(() => {})
  },
}

export default function Home() {
  const { params } = useRouter()
  const [data, setData] = useState<HomeResp>()
  const [flash, setFlash] = useState<FlashSaleInfo[]>([])
  const [viewed, setViewed] = useState<ProductCard[]>([])

  const load = () =>
    get<HomeResp>('/home')
      .then(setData)
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))
      .finally(() => Taro.stopPullDownRefresh())

  useEffect(() => {
    // 分享带参：缓存邀请码，注册时自动归因
    if (params.inviteCode) Taro.setStorageSync('pet_invite_code', params.inviteCode)
    load()
    // 秒杀专区：公开接口已过滤（启用中 + 时间窗内 + 有余量）
    get<FlashSaleInfo[]>('/flash-sales')
      .then((l) => setFlash((l ?? []).slice(0, 6)))
      .catch(() => {})
    // 最近浏览（登录用户）
    if (getToken()) {
      get<PageResp<ProductCard>>('/history/views')
        .then((r) => setViewed((r?.list ?? []).slice(0, 10)))
        .catch(() => {})
      get<{ inviteCode: string }>('/invite')
        .then((inv) => inv?.inviteCode && Taro.setStorageSync('pet_invite_code', inv.inviteCode))
        .catch(() => {})
    }
  }, [])

  // 小程序分享：带上邀请码，好友注册自动归因
  useShareAppMessage(() => {
    const code = (Taro.getStorageSync('pet_invite_code') as string) || ''
    return {
      title: '宠物之家·遇见你的毛孩子',
      path: `/pages/index/index${code ? `?inviteCode=${code}` : ''}`,
    }
  })

  usePullDownRefresh(load)

  const goDetail = (p: ProductCard) => Taro.navigateTo({ url: `/pages/detail/index?id=${p.id}` })
  const goList = (categoryId?: string) =>
    Taro.navigateTo({ url: categoryId ? `/pages/list/index?categoryId=${categoryId}` : '/pages/list/index' })
  const openBanner = (b: BannerView) => BANNER_NAV[b.jumpType]?.(b.target)

  return (
    <View className='home'>
      <View className='banner card'>
        <Text className='banner-title'>遇见你的毛孩子</Text>
        <Text className='banner-sub'>健康活体 · 平台保障 · 售后无忧</Text>
        <View className='banner-search' onClick={() => goList()}>
          <Text className='banner-search-icon'>🔍</Text>
          <Text className='banner-search-text'>搜索心仪的宠物</Text>
        </View>
        {(data?.banners ?? []).map((b) => (
          <View className='banner-coupon' key={b.id} onClick={() => openBanner(b)}>
            {b.icon.startsWith('http') ? (
              <Image className='banner-coupon-img' src={b.icon} mode='aspectFill' />
            ) : (
              <Text className='banner-coupon-icon'>{b.icon || '🐾'}</Text>
            )}
            <Text>
              {b.title}
              {b.subTitle ? ` · ${b.subTitle}` : ''}
            </Text>
            <Text className='banner-coupon-arrow'>→</Text>
          </View>
        ))}
      </View>

      <View className='cats'>
        {(data?.categories ?? []).map((c) => (
          <View className='cat' key={c.id} onClick={() => goList(c.id)}>
            <View className='cat-icon'>{c.name.slice(0, 1)}</View>
            <Text className='cat-name'>{c.name}</Text>
          </View>
        ))}
      </View>

      {flash.length > 0 && (
        <>
          <View className='section-title section-flash'>
            ⚡ 限时秒杀
            <Text className='section-flash-tip'>手慢无</Text>
          </View>
          <View className='flash-row'>
            {flash.map((f) => (
              <View className='flash-card' key={f.id} onClick={() => Taro.navigateTo({ url: `/pages/detail/index?id=${f.productId}` }).catch(() => {})}>
                {f.productImage ? (
                  <Image className='flash-img' src={f.productImage} mode='aspectFill' lazyLoad />
                ) : (
                  <View className='flash-img flash-img-empty'>🐾</View>
                )}
                <View className='flash-main'>
                  <Text className='flash-title'>{f.productTitle}</Text>
                  <View className='flash-meta'>
                    <Text className='flash-price'>¥{Number(f.salePrice).toFixed(2)}</Text>
                    <Text className='flash-stock'>剩 {f.stock - f.sold}</Text>
                  </View>
                </View>
              </View>
            ))}
          </View>
        </>
      )}

      {viewed.length > 0 && (
        <>
          <View className='section-title'>最近看过</View>
          <View className='viewed-row'>
            {viewed.map((p) => (
              <View className='viewed-card' key={p.id} onClick={() => goDetail(p)}>
                {p.mainImage ? (
                  <Image className='viewed-img' src={p.mainImage} mode='aspectFill' lazyLoad />
                ) : (
                  <View className='viewed-img viewed-img-empty'>🐾</View>
                )}
                <Text className='viewed-title'>{p.title}</Text>
                <Text className='viewed-price'>¥{p.price}</Text>
              </View>
            ))}
          </View>
        </>
      )}

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
