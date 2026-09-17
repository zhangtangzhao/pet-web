import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'
export default function AuctionPage() {
  const [list, setList] = useState<any[]>([])
  const load = useCallback(() => { get<any[]>('/auctions').then(setList).catch(() => {}) }, [])
  useEffect(() => { load() }, [load])
  const deposit = (a: any) => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fauction%2Findex' }).catch(() => {}); return }
    post(`/auctions/${a.id}/deposit`).then(() => { Taro.showToast({ title: '保证金已缴', icon: 'success' }); load() }).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }
  const bid = (a: any) => {
    (Taro.showModal as any)({ title: '出价', editable: true, placeholderText: `不低于 ¥${(Number(a.highestPrice) || Number(a.startPrice)) + Number(a.stepPrice)}`, success: (m: any) => { if (!m.confirm || !m.content) return; post(`/auctions/${a.id}/bid`, { price: m.content }).then(() => { Taro.showToast({ title: '出价成功', icon: 'success' }); load() }).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' })) } })
  }
  return (
    <View className='auc'>
      <Text className='auc-title'>竞拍大厅</Text>
      {list.map((a) => (
        <View className='auc-card' key={a.id}>
          <Text className='auc-name'>{a.productTitle}</Text>
          <Text className='auc-price'>当前 {a.highestPrice !== '0.00' ? `¥${a.highestPrice}` : `起拍 ¥${a.startPrice}`}</Text>
          <Text className='auc-meta'>{a.statusText} · 加价 ¥{a.stepPrice} · 保证金 ¥{a.depositAmount} · {a.endAt} 截止</Text>
          <View className='auc-actions'>
            {a.status === 1 && !a.depositPaid && <View className='auc-btn auc-btn-dep' onClick={() => deposit(a)}>缴保证金</View>}
            {a.status === 1 && a.depositPaid && <View className='auc-btn auc-btn-bid' onClick={() => bid(a)}>出价</View>}
          </View>
        </View>
      ))}
      {list.length === 0 && <View className='auc-empty'>暂无竞拍活动</View>}
    </View>
  )
}
