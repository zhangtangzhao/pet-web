import { useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { clearTokens, get, getToken, post } from '../../request'
import { LevelView } from '../../types'
import './index.css'
import TabBar from '../../components/TabBar'

interface MemberInfo {
  id: string
  nickname: string
  avatar: string
  phone: string
  gender: number
  hasWxBind: boolean
  createdAt: string
}

const ENTRIES: { icon: string; label: string; url: string }[] = [
  { icon: '📦', label: '我的订单', url: '/pages/orders/index' },
  { icon: '⭐', label: '我的评价', url: '/pages/my-reviews/index' },
  { icon: '🛠', label: '我的售后', url: '/pages/my-aftersales/index' },
  { icon: '🎫', label: '我的优惠券', url: '/pages/coupon-center/index' },
  { icon: '🎁', label: '积分签到', url: '/pages/points/index' },
  { icon: '🛒', label: '购物车', url: '/pages/cart/index' },
  { icon: '📸', label: '晒单广场', url: '/pages/square/index' },
  { icon: '📋', label: '任务中心', url: '/pages/tasks/index' },
  { icon: '👑', label: 'VIP 会员', url: '/pages/vip/index' },
  { icon: '🔨', label: '我的砍价', url: '/pages/bargain/index' },
  { icon: '🔨', label: '竞拍大厅', url: '/pages/auction/index' },
  { icon: '📅', label: '服务预约', url: '/pages/booking/index' },
  { icon: '🛡', label: '宠物保险', url: '/pages/insurance/index' },
  { icon: '❤️', label: '配种服务', url: '/pages/stud/index' },
  { icon: '🧾', label: '我的发票', url: '/pages/invoices/index' },
  { icon: '🎁', label: '积分商城', url: '/pages/shop/index' },
  { icon: '📖', label: '养宠百科', url: '/pages/encyclopedia/index' },
  { icon: '🐶', label: '宠物档案', url: '/pages/pet/index' },
  { icon: '🤝', label: '邀请有礼', url: '/pages/invite/index' },
  { icon: '📍', label: '地址簿', url: '/pages/address/index' },
  { icon: '❤️', label: '我的收藏', url: '/pages/favorites/index' },
  { icon: '🔔', label: '消息中心', url: '/pages/notify/index' },
  { icon: '🐾', label: '智能选宠', url: '/pages/recommend/index' },
  { icon: '💬', label: '联系客服', url: '/pages/service-chat/index' },
]

const fmtDate = (s: string) => (s ? s.slice(0, 10) : '')

export default function Me() {
  const [member, setMember] = useState<MemberInfo>()
  const [level, setLevel] = useState<LevelView>()

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/me/index')}`,
      }).catch(() => {})
      return
    }
    get<MemberInfo>('/member/profile')
      .then(setMember)
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
    get<LevelView>('/level').then(setLevel).catch(() => {})
  }, [])

  const goEntry = (url: string) => Taro.navigateTo({ url }).catch(() => {})

  const logout = () => {
    Taro.showModal({
      title: '退出登录',
      content: '确定退出当前账号吗？',
      success: (r) => {
        if (!r.confirm) return
        // 无论服务端吊销是否成功，本地一律登出
        post('/auth/logout', { refreshToken: Taro.getStorageSync('pet_refresh') }).catch(() => {})
        clearTokens()
        Taro.showToast({ title: '已退出登录', icon: 'success' })
        setTimeout(() => Taro.reLaunch({ url: '/pages/index/index' }), 600)
      },
    })
  }

  return (
    <View className='me'>
      <View className='me-head'>
        <View className='me-avatar'>
          {member?.avatar ? (
            <Image className='me-avatar-img' src={member.avatar} mode='aspectFill' />
          ) : (
            <Text>🐾</Text>
          )}
        </View>
        <View>
          <View className='me-nickname'>{member?.nickname || '宠物爱好者'}</View>
          <View className='me-sub'>
            <Text>{member?.phone || '未绑定手机号'}</Text>
            {member?.createdAt && <Text>{fmtDate(member.createdAt)} 加入</Text>}
          </View>
        </View>
      </View>

      {level && (
        <View className='me-level'>
          <View className='me-level-row'>
            <Text className='me-level-badge'>{level.levelName}</Text>
            <Text className='me-level-desc'>
              成长值 {level.growthValue}
              {level.discount !== '1.00' && ` · 商品享 ${Math.round(Number(level.discount) * 100)} 折`}
            </Text>
          </View>
          {level.nextThreshold > 0 && (
            <View className='me-level-next'>
              <View className='me-level-bar'>
                <View
                  className='me-level-bar-in'
                  style={{ width: `${Math.min(100, Math.round((level.growthValue / level.nextThreshold) * 100))}%` }}
                />
              </View>
              <Text className='me-level-tip'>
                还差 {level.nextThreshold - level.growthValue} 成长值升级，享 {Math.round(Number(level.nextDiscount ?? '1') * 100)} 折
              </Text>
            </View>
          )}
        </View>
      )}

      <View className='me-grid'>
        {ENTRIES.map((e) => (
          <View className='me-grid-item' key={e.url} onClick={() => goEntry(e.url)}>
            <Text className='me-grid-icon'>{e.icon}</Text>
            <Text className='me-grid-label'>{e.label}</Text>
          </View>
        ))}
      </View>

      <View className='me-logout' onClick={logout}>
        退出登录
      </View>
    <TabBar active='me' />
    </View>
  )
}



