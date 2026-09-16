import { useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, getToken } from '../../request'
import { CertView } from '../../types'
import './index.css'

const fmtDate = (s: string) => (s ? s.slice(0, 10).replace('T', ' ') : '')

export default function Certificate() {
  const { params } = useRouter()
  const [c, setC] = useState<CertView>()

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent(`/pages/certificate/index?orderNo=${params.orderNo}`)}`,
      }).catch(() => {})
      return
    }
    get<CertView>(`/orders/${params.orderNo}/certificate`)
      .then(setC)
      .catch((e: any) => {
        Taro.showToast({ title: e.message, icon: 'none' })
        setTimeout(() => Taro.navigateBack().catch(() => {}), 1200)
      })
  }, [params.orderNo])

  if (!c) return <View className='cert' />

  return (
    <View className='cert'>
      <View className='cert-card'>
        <View className='cert-crest'>🐾</View>
        <View className='cert-brand'>宠物之家 · 电子健康证书</View>
        <View className='cert-no'>{c.certNo}</View>
        <View className='cert-line' />
        <View className='cert-row'>
          <Text className='cert-label'>持有人</Text>
          <Text className='cert-value'>{c.memberMasked}</Text>
        </View>
        <View className='cert-row'>
          <Text className='cert-label'>宠物</Text>
          <Text className='cert-value'>{c.productTitle}</Text>
        </View>
        <View className='cert-row'>
          <Text className='cert-label'>品种</Text>
          <Text className='cert-value'>{c.breedName}</Text>
        </View>
        <View className='cert-row'>
          <Text className='cert-label'>交易完成</Text>
          <Text className='cert-value'>{fmtDate(c.completedAt)}</Text>
        </View>
        {c.guaranteeDays > 0 && (
          <View className='cert-row'>
            <Text className='cert-label'>健康保障</Text>
            <Text className='cert-value'>
              {c.guaranteeDays} 天{c.guaranteeEndAt ? `（至 ${c.guaranteeEndAt}）` : ''}
            </Text>
          </View>
        )}
        {c.supplierName && (
          <View className='cert-row'>
            <Text className='cert-label'>供货商</Text>
            <Text className='cert-value'>{c.supplierName}</Text>
          </View>
        )}
        {c.quarantineUrl && (
          <View className='cert-row'>
            <Text className='cert-label'>检疫证明</Text>
            <Text className='cert-value cert-link' onClick={() => Taro.previewImage({ urls: [c.quarantineUrl as string] })}>
              点击查看 ›
            </Text>
          </View>
        )}
        <View className='cert-seal'>健康保障</View>
      </View>
      <View className='cert-tip'>本证书与订单绑定生成，截图即可保存分享</View>
    </View>
  )
}
