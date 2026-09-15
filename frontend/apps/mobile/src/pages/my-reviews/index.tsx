import { useCallback, useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get, getToken } from '../../request'
import { ReviewListResp, ReviewView } from '../../types'
import './index.css'

const PAGE_SIZE = 10

export default function MyReviews() {
  const [list, setList] = useState<ReviewView[]>([])
  const [cursor, setCursor] = useState('')
  const [hasMore, setHasMore] = useState(false)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (nextCursor: string, append: boolean) => {
    setLoading(true)
    try {
      const qs = nextCursor ? `&cursor=${nextCursor}` : ''
      const resp = await get<ReviewListResp>(`/member/reviews?limit=${PAGE_SIZE}${qs}`)
      setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
      setHasMore(resp.hasMore)
      if (resp.list.length > 0) setCursor(resp.list[resp.list.length - 1].id)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/my-reviews/index')}`,
      }).catch(() => {})
      return
    }
    load('', false)
  }, [load])

  usePullDownRefresh(() => {
    load('', false).finally(() => Taro.stopPullDownRefresh())
  })

  const loadMore = () => {
    if (loading || !hasMore) return
    load(cursor, true)
  }

  const goProduct = (r: ReviewView) => Taro.navigateTo({ url: `/pages/detail/index?id=${r.productId}` }).catch(() => {})

  return (
    <View className='mrv'>
      {list.length === 0 && !loading && <View className='mrv-empty'>还没有评价，完成订单后去评价吧～</View>}
      {list.map((r) => (
        <View className='mrv-item card' key={r.id} onClick={() => goProduct(r)}>
          <View className='mrv-head'>
            <Text className='mrv-title'>{r.productTitle}</Text>
            <Text className='mrv-stars'>{'★'.repeat(r.rating)}{'☆'.repeat(5 - r.rating)}</Text>
          </View>
          {r.content && <Text className='mrv-content'>{r.content}</Text>}
          {r.images.length > 0 && (
            <View className='mrv-imgs'>
              {r.images.map((img, i) => (
                <Image key={i} className='mrv-img' src={img} mode='aspectFill' lazyLoad />
              ))}
            </View>
          )}
          {r.reply && (
            <View className='mrv-reply'>
              <Text className='mrv-reply-tag'>商家回复</Text>
              <Text className='mrv-reply-text'>{r.reply}</Text>
            </View>
          )}
          {r.status === 0 && <Text className='mrv-hidden'>该评价已被平台隐藏</Text>}
          <Text className='mrv-time'>{r.createdAt.slice(0, 16).replace('T', ' ')}</Text>
        </View>
      ))}
      {hasMore && (
        <View className='mrv-more' onClick={loadMore}>
          {loading ? '加载中…' : '加载更多'}
        </View>
      )}
    </View>
  )
}
