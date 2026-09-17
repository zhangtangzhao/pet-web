import { useEffect, useState } from 'react'
import { RichText, View } from '@tarojs/components'
import { get } from '../../request'

export default function Agreement() {
  const [data, setData] = useState<{ version: string; title: string; content: string }>()
  useEffect(() => {
    get<{ version: string; title: string; content: string }>('/agreement').then(setData).catch(() => {})
  }, [])
  return (
    <View className='agmt'>
      <View className='agmt-title'>{data?.title ?? '宠物活体购买协议'}</View>
      <View className='agmt-ver'>版本 {data?.version ?? 'v1'}</View>
      <View className='agmt-body'>{data?.content ?? ''}</View>
    </View>
  )
}
