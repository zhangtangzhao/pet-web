import { useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import { get } from '../../request'
import './index.css'
export default function Stud() {
  const [list, setList] = useState<any[]>([])
  useEffect(() => { get<any[]>('/stud-services').then(setList).catch(() => {}) }, [])
  return (
    <View className='stud'>
      <Text className='stud-title'>配种服务</Text>
      {list.map((s) => (
        <View className='stud-card' key={s.id}>
          <Text className='stud-name'>{s.petName}（{s.breedName}）</Text>
          <Text className='stud-desc'>{s.description}</Text>
          {!!s.healthCerts && <Text className='stud-certs'>🏥 {s.healthCerts}</Text>}
          <View className='stud-bottom'><Text className='stud-price'>¥{s.price}/次</Text></View>
        </View>
      ))}
      {list.length === 0 && <View className='stud-empty'>暂无配种服务</View>}
    </View>
  )
}
