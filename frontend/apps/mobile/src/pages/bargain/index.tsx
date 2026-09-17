import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'
export default function Bargain() {
  const [list, setList] = useState<any[]>([])
  const load = useCallback(() => { if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fbargain%2Findex' }).catch(() => {}); return } get<any[]>('/bargain-launches').then(setList).catch(() => {}) }, [])
  useEffect(() => { load() }, [load])
  return (
    <View className='bg-page'>
      <Text className='bg-page-title'>我的砍价</Text>
      {list.length === 0 && <View className='bg-empty'>暂无砍价记录，去商品详情发起吧</View>}
      {list.map((l) => (
        <View className='bg-card' key={l.id}>
          <Text className='bg-name'>{l.productTitle}</Text>
          <View className='bg-prices'><Text className='bg-orig'>¥{l.originPrice}</Text><Text className='bg-cur'>¥{l.currentPrice}</Text></View>
          <Text className='bg-meta'>{l.statusText} · 已助力 {l.helperCount}/{l.maxHelpers} 人 · {l.expireAt} 截止</Text>
        </View>
      ))}
    </View>
  )
}
