import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { PointsResp, SignInResp } from '../../types'
import './index.css'

export default function Points() {
  const [data, setData] = useState<PointsResp>()
  const [signing, setSigning] = useState(false)

  const load = useCallback(
    () =>
      get<PointsResp>('/points')
        .then(setData)
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' })),
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fpoints%2Findex' }).catch(() => {})
      return
    }
    load()
  }, [load])

  const signin = () => {
    if (signing) return
    setSigning(true)
    post<SignInResp>('/points/signin')
      .then((r) => {
        Taro.showToast({ title: r.ok ? '签到成功' : '今日已签到', icon: 'success' })
        load()
      })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      .finally(() => setSigning(false))
  }

  const goExchange = () => Taro.navigateTo({ url: '/pages/coupon-center/index' }).catch(() => {})

  return (
    <View className='pt'>
      <View className='pt-hero'>
        <View className='pt-balance'>
          <Text className='pt-num'>{data?.points ?? 0}</Text>
          <Text className='pt-unit'>当前积分</Text>
        </View>
        <View className={`pt-signin ${signing ? 'pt-off' : ''}`} onClick={signin}>
          每日签到
        </View>
        <View className='pt-exchange' onClick={goExchange}>
          积分兑好礼 ›
        </View>
      </View>

      <View className='pt-card'>
        <View className='pt-title'>积分明细</View>
        {(data?.list ?? []).length === 0 && <View className='pt-empty'>暂无积分记录</View>}
        {(data?.list ?? []).map((it) => (
          <View className='pt-row' key={it.id}>
            <View className='pt-row-main'>
              <Text className='pt-reason'>{it.reason}</Text>
              <Text className='pt-time'>{it.createdAt.slice(0, 16).replace('T', ' ')}</Text>
            </View>
            <View className='pt-row-right'>
              <Text className={`pt-change ${it.change >= 0 ? 'pt-change-plus' : ''}`}>
                {it.change >= 0 ? `+${it.change}` : it.change}
              </Text>
              <Text className='pt-balance-log'>余额 {it.balance}</Text>
            </View>
          </View>
        ))}
      </View>
    </View>
  )
}
