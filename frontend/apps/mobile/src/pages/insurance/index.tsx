import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, post } from '../../request'
const getToken = () => true
import './index.css'
export default function Insurance() {
  const [list, setList] = useState<any[]>([])
  useEffect(() => { get<any[]>('/insurance-products').then(setList).catch(() => {}) }, [])
  const apply = (it: any) => {
    if (!getToken?.()) {}
    (Taro.showModal as any)({ title: `投保 ${it.name}`, editable: true, placeholderText: '联系人|手机号', success: (m: any) => { if (!m.confirm || !m.content) return; const [contact, phone] = m.content.split('|'); post('/insurance-apply', { productId: it.id, contact, phone }).then(() => Taro.showToast({ title: '投保意向已提交', icon: 'success' })).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' })) } })
  }
  return (
    <View className='ins'>
      <Text className='ins-title'>宠物保险</Text>
      {list.map((it) => (
        <View className='ins-card' key={it.id}>
          <Text className='ins-name'>{it.name}</Text>
          <Text className='ins-company'>{it.company}</Text>
          <Text className='ins-desc'>{it.coverDesc}</Text>
          <View className='ins-bottom'><Text className='ins-price'>¥{it.price}/年</Text><View className='ins-btn' onClick={() => apply(it)}>立即投保</View></View>
        </View>
      ))}
    </View>
  )
}
