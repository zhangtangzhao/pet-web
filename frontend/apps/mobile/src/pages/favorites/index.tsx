import { useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { useDidShow } from '@tarojs/taro'
import { del, get, getToken } from '../../request'
import { PageResp, ProductCard } from '../../types'
import './index.css'

export default function Favorites() {
  const [list, setList] = useState<ProductCard[]>([])
  const [total, setTotal] = useState(0)

  const load = () => {
    if (!getToken()) {
      Taro.navigateTo({ url: '/pages/login/index?redirect=/pages/favorites/index' })
      return
    }
    get<PageResp<ProductCard>>('/favorites')
      .then((r) => {
        setList(r.list)
        setTotal(r.total)
      })
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  useDidShow(load)

  const unfav = async (id: string) => {
    try {
      await del(`/favorites/${id}`)
      setList(list.filter((p) => p.id !== id))
      setTotal(total - 1)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    }
  }

  return (
    <View className='favs'>
      {list.length === 0 && <View className='empty'>暂无收藏，去逛逛吧～</View>}
      {list.map((p) => (
        <View className='fav-item card' key={p.id}>
          <View onClick={() => Taro.navigateTo({ url: `/pages/detail/index?id=${p.id}` })}>
            {p.mainImage ? (
              <Image className='fav-img' src={p.mainImage} mode='aspectFill' />
            ) : (
              <View className='fav-img fav-img-empty'>🐾</View>
            )}
          </View>
          <View className='fav-body' onClick={() => Taro.navigateTo({ url: `/pages/detail/index?id=${p.id}` })}>
            <Text className='fav-title'>{p.title}</Text>
            <Text className='price'>¥{p.price}</Text>
          </View>
          <View className='fav-op' onClick={() => unfav(p.id)}>
            <Text>取消收藏</Text>
          </View>
        </View>
      ))}
      {total > 0 && <View className='count'>共 {total} 件收藏</View>}
    </View>
  )
}
