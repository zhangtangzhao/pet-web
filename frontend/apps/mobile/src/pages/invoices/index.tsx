import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken } from '../../request'
import './index.css'
export default function Invoices() {
  const [list, setList] = useState<any[]>([])
  const load = useCallback(() => { if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Finvoices%2Findex' }).catch(() => {}); return } get<any[]>('/invoices').then(setList).catch(() => {}) }, [])
  useEffect(() => { load() }, [load])
  return (
    <View className='inv'>
      <Text className='inv-title'>我的发票</Text>
      {list.length === 0 && <View className='inv-empty'>暂无发票记录，完成订单后可申请</View>}
      {list.map((it) => (
        <View className='inv-card' key={it.orderNo}>
          <View className='inv-row'><Text className='inv-order'>{it.orderNo}</Text><Text className={it.status === 1 ? 'inv-ok' : 'inv-pending'}>{it.status === 1 ? '已开票' : '待开票'}</Text></View>
          <Text className='inv-title-tag'>{it.title}</Text>
          <Text className='inv-amount'>¥{it.amount}</Text>
          {!!it.link && <Text className='inv-link'>下载发票</Text>}
        </View>
      ))}
    </View>
  )
}
