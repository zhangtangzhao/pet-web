import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'

interface Task { key: string; title: string; desc: string; reward: number; claimed: boolean; type: string }

export default function TasksPage() {
  const [tasks, setTasks] = useState<Task[]>([])
  const [loading, setLoading] = useState(false)

  const load = useCallback(() => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Ftasks%2Findex' }).catch(() => {}); return }
    get<{ list: Task[] }>('/tasks').then((r) => setTasks(r.list ?? [])).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])

  const claim = (t: Task) => {
    if (loading) return
    setLoading(true)
    post('/tasks/claim', { key: t.key })
      .then(() => { Taro.showToast({ title: `+${t.reward} 积分`, icon: 'success' }); load() })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      .finally(() => setLoading(false))
  }

  return (
    <View className='tk'>
      <Text className='tk-title'>任务中心</Text>
      {tasks.map((t) => (
        <View className='tk-card' key={t.key}>
          <View className='tk-main'>
            <Text className='tk-name'>{t.title}</Text>
            <Text className='tk-desc'>{t.desc}</Text>
          </View>
          <View className={`tk-btn ${t.claimed ? 'tk-done' : ''}`} onClick={() => !t.claimed && claim(t)}>
            {t.claimed ? '已完成' : `+${t.reward}`}
          </View>
        </View>
      ))}
      {tasks.length === 0 && <View className='tk-empty'>加载中…</View>}
    </View>
  )
}
