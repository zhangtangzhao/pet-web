import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get, getToken } from '../../request'
import { AfterSaleView, PageResp } from '../../types'
import './index.css'

const TAG_CLASS: Record<number, string> = {
  1: 'mas-tag-pending',
  2: 'mas-tag-ok',
  3: 'mas-tag-refused',
  4: 'mas-tag-cancel',
}

const PAGE_SIZE = 10

export default function MyAfterSales() {
  const [list, setList] = useState<AfterSaleView[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (nextPage: number, append: boolean) => {
    setLoading(true)
    try {
      const resp = await get<PageResp<AfterSaleView>>(`/aftersale/list?page=${nextPage}&pageSize=${PAGE_SIZE}`)
      setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
      setTotal(resp.total)
      setPage(nextPage)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/my-aftersales/index')}`,
      }).catch(() => {})
      return
    }
    load(1, false)
  }, [load])

  usePullDownRefresh(() => {
    load(1, false).finally(() => Taro.stopPullDownRefresh())
  })

  const loadMore = () => {
    if (loading || list.length >= total) return
    load(page + 1, true)
  }

  const goDetail = (a: AfterSaleView) => Taro.navigateTo({ url: `/pages/after-sale/index?orderNo=${a.orderNo}` })

  return (
    <View className='mas'>
      {list.length === 0 && !loading && <View className='mas-empty'>暂无售后记录～</View>}
      {list.map((a) => (
        <View className='mas-item card' key={a.id} onClick={() => goDetail(a)}>
          <View className='mas-head'>
            <Text className={`mas-tag ${TAG_CLASS[a.status] || 'mas-tag-cancel'}`}>{a.statusText}</Text>
            <Text className='mas-order'>{a.orderNo}</Text>
          </View>
          <View className='mas-row'>
            <Text className='mas-label'>退款金额</Text>
            <Text className='mas-value'>¥{a.refundAmount}</Text>
          </View>
          <View className='mas-row'>
            <Text className='mas-label'>申请原因</Text>
            <Text className='mas-value'>{a.reason || '-'}</Text>
          </View>
          {a.adminNote && (
            <View className='mas-row'>
              <Text className='mas-label'>平台备注</Text>
              <Text className='mas-value'>{a.adminNote}</Text>
            </View>
          )}
          <View className='mas-row'>
            <Text className='mas-label'>申请时间</Text>
            <Text className='mas-value'>{a.createdAt.slice(0, 16).replace('T', ' ')}</Text>
          </View>
          <View className='mas-link'>查看进度 ›</View>
        </View>
      ))}
      {list.length < total && (
        <View className='mas-more' onClick={loadMore}>
          {loading ? '加载中…' : '加载更多'}
        </View>
      )}
    </View>
  )
}
