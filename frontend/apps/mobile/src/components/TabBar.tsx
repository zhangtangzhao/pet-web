import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import './TabBar.css'

const TABS = [
  { key: 'home', icon: '🏠', label: '首页', url: '/pages/index/index' },
  { key: 'category', icon: '📋', label: '分类', url: '/pages/list/index' },
  { key: 'message', icon: '💬', label: '消息', url: '/pages/notify/index' },
  { key: 'me', icon: '👤', label: '我的', url: '/pages/me/index' },
]

export default function TabBar({ active }: { active: string }) {
  return (
    <View className='tabbar'>
      {TABS.map((t) => (
        <View key={t.key} className={`tabbar-item ${t.key === active ? 'tabbar-on' : ''}`}
          onClick={() => { if (t.key !== active) Taro.redirectTo({ url: t.url }).catch(() => {}) }}>
          <Text className='tabbar-icon'>{t.icon}</Text>
          <Text className='tabbar-label'>{t.label}</Text>
        </View>
      ))}
    </View>
  )
}
