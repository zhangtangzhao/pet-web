import { useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'
export default function Vip() {
  const [info, setInfo] = useState<any>()
  const [loading, setLoading] = useState(false)
  useEffect(() => { if (getToken()) get<any>('/member/profile').then(setInfo).catch(() => {}) }, [])
  const buy = () => {
    if (loading) return
    setLoading(true)
    post('/vip/buy').then(() => { Taro.showToast({ title: '开通成功', icon: 'success' }); setTimeout(() => Taro.navigateBack().catch(() => {}), 800) }).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' })).finally(() => setLoading(false))
  }
  return (
    <View className='vip'>
      <View className='vip-card'>
        <Text className='vip-badge'>PLUS 会员</Text>
        <Text className='vip-price'>¥99 / 年</Text>
      </View>
      <View className='vip-benefits'>
        <Text className='vip-b-item'>🎁 每月赠免运费卡 1 张</Text>
        <Text className='vip-b-item'>🎫 每月赠 VIP 专属券</Text>
        <Text className='vip-b-item'>⚡ 优先发货</Text>
      </View>
      <View className={`vip-buy ${loading ? 'vip-off' : ''}`} onClick={buy}>{loading ? '开通中…' : '立即开通'}</View>
    </View>
  )
}
