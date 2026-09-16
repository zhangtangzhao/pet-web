import { useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken } from '../../request'
import { InviteResp } from '../../types'
import './index.css'

export default function Invite() {
  const [data, setData] = useState<InviteResp>()

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Finvite%2Findex' }).catch(() => {})
      return
    }
    get<InviteResp>('/invite')
      .then(setData)
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }, [])

  const copy = () => {
    if (!data?.inviteCode) return
    Taro.setClipboardData({ data: data.inviteCode }).catch(() => {})
  }

  return (
    <View className='iv'>
      <View className='iv-hero'>
        <View className='iv-title'>邀请有礼</View>
        <View className='iv-desc'>好友注册时填写你的邀请码，双方各得 {data?.rewardEach ?? 0} 积分</View>
        <View className='iv-code-box'>
          <Text className='iv-code'>{data?.inviteCode || '····'}</Text>
        </View>
        <View className='iv-copy' onClick={copy}>
          复制邀请码
        </View>
      </View>

      <View className='iv-card'>
        <View className='iv-stat'>
          <Text className='iv-stat-num'>{data?.invited ?? 0}</Text>
          <Text className='iv-stat-label'>已邀请好友</Text>
        </View>
        <View className='iv-stat'>
          <Text className='iv-stat-num'>{(data?.invited ?? 0) * (data?.rewardEach ?? 0)}</Text>
          <Text className='iv-stat-label'>累计获得积分</Text>
        </View>
      </View>

      <View className='iv-tips'>
        <View className='iv-tips-title'>玩法说明</View>
        <Text className='iv-tip'>1. 将邀请码分享给好友；</Text>
        <Text className='iv-tip'>2. 好友首次注册（手机号登录）时填写邀请码；</Text>
        <Text className='iv-tip'>3. 注册成功后双方立即到账积分，可在积分页兑换好礼。</Text>
      </View>
    </View>
  )
}
