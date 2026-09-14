import { useEffect, useRef, useState } from 'react'
import { Image, Input, ScrollView, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import { createCSSocket } from '../../ws'
import type { CsEvent, CsHistoryResp, CsMsgOut, CsSessionUpdate, UploadTokenResp } from '../../types'
import './index.css'

interface ChatMsg extends CsMsgOut {
  key: string // 服务端 id 或乐观发送的本地 key
  state?: 'pending' | 'failed' // 乐观发送状态，服务端确认后移除
  retries?: number
}

const MAX_RETRIES = 3

let seq = 0
const nextKey = () => `m${Date.now()}-${seq++}`

// 雪花 ID 字符串比较：先比长度再比字典序，等价数值比较
const cmpId = (a: string, b: string) => (a.length !== b.length ? a.length - b.length : a < b ? -1 : a > b ? 1 : 0)

export default function ServiceChat() {
  const [items, setItems] = useState<ChatMsg[]>([])
  const [input, setInput] = useState('')
  const [connected, setConnected] = useState(false)
  const [ended, setEnded] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const endRef = useRef('')
  const itemsRef = useRef<ChatMsg[]>([])
  const lastIdRef = useRef('') // 已见最大消息 ID，重连补拉游标
  const oldestRef = useRef('') // 已载最早消息 ID，向上翻页游标
  const loadingMoreRef = useRef(false)
  const inflightRef = useRef(new Set<string>())
  const socketRef = useRef<ReturnType<typeof createCSSocket>>()

  useEffect(() => {
    itemsRef.current = items
    endRef.current = items.length ? items[items.length - 1].key : ''
  }, [items])

  // 幂等合并：所有来源（历史/推送/补拉/发送回执）按 id 去重后归位
  const merge = (incoming: CsMsgOut[]) => {
    const fresh = incoming.filter((m) => m.id)
    if (fresh.length === 0) return
    setItems((prev) => {
      const seen = new Set(prev.filter((m) => m.id).map((m) => m.id))
      const add: ChatMsg[] = []
      for (const m of fresh) {
        if (seen.has(m.id)) continue
        seen.add(m.id)
        add.push({ ...m, key: m.id })
      }
      if (add.length === 0) return prev
      for (const m of add) {
        if (!lastIdRef.current || cmpId(lastIdRef.current, m.id) < 0) lastIdRef.current = m.id
        if (!oldestRef.current || cmpId(oldestRef.current, m.id) > 0) oldestRef.current = m.id
      }
      const real = [...prev.filter((m) => m.id), ...add].sort((a, b) => cmpId(a.id, b.id))
      return [...real, ...prev.filter((m) => !m.id)]
    })
  }

  const doSend = (m: ChatMsg) => {
    if (inflightRef.current.has(m.key)) return
    inflightRef.current.add(m.key)
    post<CsMsgOut>('/cs/messages', { msgType: m.msgType, content: m.content })
      .then((out) => {
        inflightRef.current.delete(m.key)
        setItems((prev) => prev.filter((p) => p.key !== m.key))
        merge([out])
      })
      .catch(() => {
        inflightRef.current.delete(m.key)
        setItems((prev) =>
          prev.map((p) => {
            if (p.key !== m.key) return p
            const retries = (p.retries ?? 0) + 1
            return { ...p, retries, state: retries >= MAX_RETRIES ? ('failed' as const) : ('pending' as const) }
          }),
        )
      })
  }

  const send = (msgType: number, content: string) => {
    const m: ChatMsg = {
      key: nextKey(),
      id: '',
      sessionId: '',
      senderRole: 1,
      msgType,
      content,
      createdAt: '',
      state: 'pending',
      retries: 0,
    }
    setItems((prev) => [...prev, m])
    doSend(m)
  }

  const resend = (m: ChatMsg) => {
    const target = { ...m, state: 'pending' as const, retries: 0 }
    setItems((prev) => prev.map((p) => (p.key === m.key ? target : p)))
    doSend(target)
  }

  // 重连成功 / 网络恢复：补拉断线窗口缺口 + 自动重发未达消息（超次数的保留手动重试）
  const resync = () => {
    const q = lastIdRef.current ? `?after=${lastIdRef.current}&limit=100` : '?limit=50'
    get<CsHistoryResp>(`/cs/messages${q}`)
      .then((h) => merge(h.list))
      .catch(() => {})
    for (const m of itemsRef.current) {
      if ((m.state === 'pending' || m.state === 'failed') && (m.retries ?? 0) < MAX_RETRIES) doSend(m)
    }
  }

  const handleEvent = (e: CsEvent) => {
    if (e.type === 'new_message') {
      merge([e.data as CsMsgOut])
      post('/cs/read').catch(() => {})
    } else if (e.type === 'session_update') {
      setEnded((e.data as CsSessionUpdate).status === 2)
    }
  }

  const loadMore = () => {
    if (loadingMoreRef.current || !hasMore || !oldestRef.current) return
    loadingMoreRef.current = true
    get<CsHistoryResp>(`/cs/messages?before=${oldestRef.current}&limit=20`)
      .then((h) => {
        merge(h.list)
        setHasMore(h.hasMore)
      })
      .catch(() => {})
      .finally(() => {
        loadingMoreRef.current = false
      })
  }

  const uploadImage = async (filePath: string): Promise<string> => {
    const t = await post<UploadTokenResp>('/upload-token', { dir: 'chat', contentType: 'image/jpeg' })
    if (process.env.TARO_ENV === 'h5') {
      const blob = await fetch(filePath).then((r) => r.blob())
      const resp = await fetch(t.uploadUrl, { method: 'PUT', body: blob })
      if (!resp.ok) throw new Error('图片上传失败')
    } else {
      const buf = Taro.getFileSystemManager().readFileSync(filePath) as ArrayBuffer
      const resp = await Taro.request({
        url: t.uploadUrl,
        method: 'PUT',
        data: buf,
        header: { 'Content-Type': 'image/jpeg' },
      })
      if (resp.statusCode >= 300) throw new Error('图片上传失败')
    }
    return t.fileUrl
  }

  const chooseImage = () => {
    Taro.chooseImage({ count: 1 })
      .then((res) => {
        const filePath = res.tempFilePaths?.[0]
        if (!filePath) return
        uploadImage(filePath)
          .then((url) => send(2, url))
          .catch((e: Error) => Taro.showToast({ title: e.message || '图片发送失败', icon: 'none' }))
      })
      .catch(() => {})
  }

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/service-chat/index')}`,
      }).catch(() => {})
      return
    }
    let disposed = false
    get<CsHistoryResp>('/cs/messages?limit=20')
      .then((h) => {
        if (disposed) return
        merge(h.list)
        setHasMore(h.hasMore)
        post('/cs/read').catch(() => {})
        socketRef.current = createCSSocket({
          onMessage: handleEvent,
          onResync: resync,
          onState: (s) => setConnected(s === 'connected'),
        })
      })
      .catch((e: Error) => Taro.showToast({ title: e.message, icon: 'none' }))
    Taro.onNetworkStatusChange?.((n) => {
      if (n.isConnected) resync()
    })
    return () => {
      disposed = true
      socketRef.current?.close()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // 页面重新可见时若已断线立即重建，不等退避计时
  Taro.useDidShow(() => socketRef.current?.reconnectNow())

  const sendInput = () => {
    const text = input.trim()
    if (!text) return
    setInput('')
    send(1, text)
  }

  const renderBubble = (m: ChatMsg) =>
    m.msgType === 2 ? (
      <Image
        className='cs-img'
        src={m.content}
        mode='aspectFill'
        onClick={() => Taro.previewImage({ urls: [m.content], current: m.content })}
      />
    ) : (
      <Text>{m.content}</Text>
    )

  return (
    <View className='cs'>
      {!connected && <View className='cs-status'>连接中…</View>}
      {ended && <View className='cs-ended'>会话已结束，发送新消息将重新开启</View>}

      <ScrollView
        scrollY
        className='cs-list'
        scrollIntoView={endRef.current}
        scrollWithAnimation
        onScrollToUpper={loadMore}
        upperThreshold={80}
      >
        {items.length === 0 && <View className='cs-tip'>您好，我是平台客服，请描述您遇到的问题～</View>}
        {items.map((m) =>
          m.senderRole === 1 ? (
            <View key={m.key} id={m.key} className='cs-row-user'>
              {(m.state === 'pending' || m.state === 'failed') && (
                <View className='cs-meta'>
                  {m.state === 'pending' ? (
                    <Text className='cs-pending'>发送中…</Text>
                  ) : (
                    <Text className='cs-retry' onClick={() => resend(m)}>
                      发送失败，点击重试
                    </Text>
                  )}
                </View>
              )}
              <View className='cs-bubble cs-bubble-user'>{renderBubble(m)}</View>
            </View>
          ) : (
            <View key={m.key} id={m.key} className='cs-row-admin'>
              <View className='cs-bubble cs-bubble-admin'>{renderBubble(m)}</View>
            </View>
          ),
        )}
        <View className='cs-space' />
      </ScrollView>

      <View className='cs-input-bar'>
        <View className='cs-attach' onClick={chooseImage}>
          ＋
        </View>
        <Input
          className='cs-input'
          value={input}
          placeholder='请描述您的问题…'
          onInput={(e) => setInput(e.detail.value)}
          onConfirm={sendInput}
          confirmType='send'
        />
        <View className={`cs-send ${input.trim() ? 'cs-send-on' : ''}`} onClick={sendInput}>
          发送
        </View>
      </View>
    </View>
  )
}
