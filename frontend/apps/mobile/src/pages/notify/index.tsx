import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { NotifyListResp, NotifyView } from '../../types'
import './index.css'

const SCENE_ICON: Record<number, string> = { 1: '💬', 2: '📦', 3: '🎫' }

export default function NotifyPage() {
  const [list, setList] = useState<NotifyView[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [unread, setUnread] = useState(0)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (nextPage: number, append: boolean) => {
    setLoading(true)
    try {
      const resp = await get<NotifyListResp>(`/notify/list?page=${nextPage}&pageSize=10`)
      setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
      setTotal(resp.total)
      setUnread(resp.unread)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/notify/index')}`,
      }).catch(() => {})
      return
    }
    load(1, false)
  }, [load])

  const loadMore = () => {
    if (loading || list.length >= total) return
    const next = page + 1
    setPage(next)
    load(next, true)
  }

  usePullDownRefresh(() => {
    load(1, false).finally(() => Taro.stopPullDownRefresh())
  })

  const readAll = () => {
    if (unread === 0) return
    post('/notify/read')
      .then(() => {
        setUnread(0)
        setList((prev) => prev.map((n) => ({ ...n, readAt: n.readAt || 'now' })))
        Taro.showToast({ title: '已全部已读', icon: 'success' })
      })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const open = (n: NotifyView) => {
    if (!n.readAt) {
      // 乐观更新：先消红点，失败回滚重查
      setList((prev) => prev.map((x) => (x.id === n.id ? { ...x, readAt: x.readAt || 'now' } : x)))
      setUnread((u) => Math.max(0, u - 1))
      post(`/notify/${n.id}/read`).catch(() => load(1, false))
    }
    if (n.orderNo) Taro.navigateTo({ url: '/pages/orders/index' }).catch(() => {})
  }

  const fmt = (s: string) => (s ? s.slice(0, 16).replace('T', ' ') : '')

  return (
    <View className='ntf'>
      <View className='ntf-head'>
        <Text className='ntf-unread'>{unread > 0 ? `${unread} 条未读` : '全部已读'}</Text>
        <View className={`ntf-readall ${unread === 0 ? 'ntf-readall-off' : ''}`} onClick={readAll}>
          全部已读
        </View>
      </View>

      {list.length === 0 && !loading && <View className='ntf-empty'>暂无消息～</View>}
      {list.map((n) => (
        <View className='ntf-item card' key={n.id} onClick={() => open(n)}>
          <View className='ntf-icon'>{SCENE_ICON[n.scene] ?? '🔔'}</View>
          <View className='ntf-main'>
            <View className='ntf-title-row'>
              <Text className={`ntf-title ${n.readAt ? 'ntf-read' : ''}`}>{n.title}</Text>
              {!n.readAt && <View className='ntf-dot' />}
            </View>
            <Text className={`ntf-content ${n.readAt ? 'ntf-read' : ''}`}>{n.content}</Text>
            <Text className='ntf-time'>{fmt(n.createdAt)}</Text>
          </View>
        </View>
      ))}
      {list.length < total && (
        <View className='ntf-more' onClick={loadMore}>
          {loading ? '加载中…' : '加载更多'}
        </View>
      )}
    </View>
  )
}
