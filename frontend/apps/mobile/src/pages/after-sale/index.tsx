import { useEffect, useState } from 'react'
import { Textarea, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { AfterSaleView } from '../../types'
import './index.css'

const AS_CLASS: Record<number, string> = { 1: 'as-tag-pending', 2: 'as-tag-ok', 3: 'as-tag-refused', 4: 'as-tag-cancel' }

export default function AfterSalePage() {
  const { params } = useRouter()
  const [orderNo] = useState(params.orderNo ?? '')
  const [as, setAs] = useState<AfterSaleView>()
  const [reason, setReason] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = () =>
    get<AfterSaleView>(`/aftersale?orderNo=${orderNo}`)
      .then(setAs)
      .catch(() => setAs(undefined))

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent(`/pages/after-sale/index?orderNo=${orderNo}`)}`,
      }).catch(() => {})
      return
    }
    load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orderNo])

  const submit = async () => {
    if (submitting) return
    if (reason.trim() === '') {
      Taro.showToast({ title: '请填写售后原因', icon: 'none' })
      return
    }
    setSubmitting(true)
    try {
      await post('/aftersale', { orderNo, reason: reason.trim() })
      Taro.showToast({ title: '已提交，等待平台审核', icon: 'success' })
      setReason('')
      load()
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  const cancel = () => {
    if (!as) return
    Taro.showModal({
      title: '撤销售后',
      content: '确定撤销该售后申请吗？',
      success: (r) => {
        if (!r.confirm) return
        post(`/aftersale/${as.afterSaleNo}/cancel`)
          .then(() => {
            Taro.showToast({ title: '已撤销', icon: 'success' })
            load()
          })
          .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      },
    })
  }

  return (
    <View className='as'>
      {as ? (
        <View className='as-card card'>
          <View className='as-head'>
            <View className={`as-tag ${AS_CLASS[as.status]}`}>{as.statusText}</View>
            <Text className='as-no'>{as.afterSaleNo}</Text>
          </View>
          <View className='as-row'>
            <Text className='as-label'>订单号</Text>
            <Text className='as-value'>{as.orderNo}</Text>
          </View>
          <View className='as-row'>
            <Text className='as-label'>售后原因</Text>
            <Text className='as-value'>{as.reason}</Text>
          </View>
          <View className='as-row'>
            <Text className='as-label'>退款金额</Text>
            <Text className='as-value'>¥{as.refundAmount}</Text>
          </View>
          {as.adminNote && (
            <View className='as-row'>
              <Text className='as-label'>平台备注</Text>
              <Text className='as-value'>{as.adminNote}</Text>
            </View>
          )}
          {as.status === 1 && (
            <View className='as-cancel' onClick={cancel}>
              撤销申请
            </View>
          )}
        </View>
      ) : (
        <View className='as-card card'>
          <View className='as-title'>申请退款售后</View>
          <View className='as-desc'>订单号 {orderNo}</View>
          <View className='as-desc'>提交后平台将尽快审核，退款默认全额、原路退回。</View>
          <Textarea
            className='as-textarea'
            maxlength={500}
            placeholder='请填写售后原因（如：宠物健康问题、与描述不符等）'
            value={reason}
            onInput={(e) => setReason(e.detail.value)}
          />
          <View className={`as-submit ${submitting ? 'as-submit-off' : ''}`} onClick={submit}>
            提交申请
          </View>
        </View>
      )}
    </View>
  )
}
