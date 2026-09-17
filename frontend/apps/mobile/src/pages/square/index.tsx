import { useCallback, useEffect, useState } from 'react'
import { Image, Text, Textarea, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { del, get, getToken, post } from '../../request'
import './index.css'

interface Post { id: string; nickname: string; avatar: string; content: string; images: string[]; likeCount: number; liked: boolean; status: number; createdAt: string }
interface ListResp { list: Post[]; hasMore: boolean }

export default function Square() {
  const [list, setList] = useState<Post[]>([])
  const [cursor, setCursor] = useState('')
  const [hasMore, setHasMore] = useState(false)
  const [draft, setDraft] = useState('')
  const [showForm, setShowForm] = useState(false)

  const load = useCallback((cur = '') => {
    get<ListResp>('/posts?limit=10' + (cur ? '&cursor=' + cur : ''))
      .then((r) => {
        setList((prev) => (cur ? prev.concat(r.list) : r.list))
        setCursor(r.list.length ? r.list[r.list.length - 1].id : cur)
        setHasMore(r.hasMore)
      })
      .catch(() => {})
  }, [])

  useEffect(() => { load() }, [load])

  const publish = () => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fsquare%2Findex' }).catch(() => {}); return }
    if (!draft.trim()) { Taro.showToast({ title: '写点什么吧', icon: 'none' }); return }
    post('/posts', { content: draft.trim() })
      .then(() => { Taro.showToast({ title: '已提交，审核通过后展示', icon: 'none' }); setDraft(''); setShowForm(false) })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const like = (p: Post) => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fsquare%2Findex' }).catch(() => {}); return }
    post('/posts/' + p.id + '/like').then((r: any) => {
      setList((prev) => prev.map((x) => x.id === p.id ? { ...x, liked: r.liked, likeCount: x.likeCount + (r.liked ? 1 : -1) } : x))
    }).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const remove = (p: Post) => {
    Taro.showModal({ title: '删除动态', content: '确定删除这条动态吗？', success: (m) => { if (m.confirm) del('/posts/' + p.id).then(() => load()).catch(() => {}) } })
  }

  return (
    <View className='sq'>
      <View className='sq-head'>
        <Text className='sq-title'>晒单广场</Text>
        <View className='sq-add' onClick={() => setShowForm(!showForm)}>{showForm ? '收起' : '+ 晒一晒'}</View>
      </View>
      {showForm && (
        <View className='sq-form'>
          <Textarea className='sq-textarea' maxlength={500} placeholder='分享你的养宠日常…（审核通过后展示）' value={draft} onInput={(e) => setDraft(e.detail.value)} />
          <View className='sq-publish' onClick={publish}>发布</View>
        </View>
      )}
      {list.length === 0 && <View className='sq-empty'>还没有动态，来发第一条吧</View>}
      {list.map((p) => (
        <View className='sq-card' key={p.id}>
          <View className='sq-card-head'>
            <Image className='sq-avatar' src={p.avatar || 'https://cdn.example.com/def.png'} />
            <Text className='sq-nick'>{p.nickname || '铲屎官'}</Text>
            <Text className='sq-time'>{p.createdAt.slice(5, 16).replace('T', ' ')}</Text>
          </View>
          <Text className='sq-content'>{p.content}</Text>
          {p.images.length > 0 && (
            <View className='sq-imgs'>{p.images.map((u) => <Image key={u} className='sq-img' src={u} mode='aspectFill' onClick={() => Taro.previewImage({ urls: p.images, current: u })} />)}</View>
          )}
          <View className='sq-actions'>
            <Text className={p.liked ? 'sq-liked' : ''} onClick={() => like(p)}>{p.liked ? '♥' : '♡'} {p.likeCount}</Text>
            <Text className='sq-del' onClick={() => remove(p)}>删除</Text>
          </View>
        </View>
      ))}
      {hasMore && <View className='sq-more' onClick={() => load(cursor)}>加载更多</View>}
    </View>
  )
}
