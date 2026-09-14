import { useCallback, useEffect, useRef, useState } from 'react'
import { Avatar, Badge, Button, Card, Empty, Image, Input, Popconfirm, Space, Tag, Upload, message } from 'antd'
import { CustomerServiceOutlined, PictureOutlined, SendOutlined } from '@ant-design/icons'
import {
  closeCsSession,
  createCSWS,
  type CsEvent,
  type CsMsgOut,
  type CsSessionItem,
  type SessionUpdate,
  fetchCsHistory,
  fetchCsSessions,
  fetchUploadToken,
  markCsRead,
  sendCsImage,
  sendCsMessage,
} from '../api/cs'

// 雪花 ID 字符串比较：先比长度再比字典序，等价数值比较
const cmpId = (a: string, b: string) => (a.length !== b.length ? a.length - b.length : a < b ? -1 : a > b ? 1 : 0)

const cmpTimeDesc = (a: string, b: string) => (a === b ? 0 : a > b ? -1 : 1)

const previewOf = (m: CsMsgOut) => (m.msgType === 2 ? '[图片]' : m.content)

const fmtTime = (iso: string) => {
  if (!iso) return ''
  const d = new Date(iso)
  const hm = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
  if (d.toDateString() === new Date().toDateString()) return hm
  return `${d.getMonth() + 1}-${d.getDate()} ${hm}`
}

export default function Support() {
  const [sessions, setSessions] = useState<CsSessionItem[]>([])
  const [current, setCurrent] = useState<CsSessionItem | null>(null)
  const [msgs, setMsgs] = useState<CsMsgOut[]>([])
  const [input, setInput] = useState('')
  const [connected, setConnected] = useState(false)
  const [hasMore, setHasMore] = useState(false)
  const [loadingMore, setLoadingMore] = useState(false)
  const currentRef = useRef<CsSessionItem | null>(null)
  const lastIdRef = useRef('') // 当前会话已见最大消息 ID，重连补拉游标
  const oldestRef = useRef('') // 当前会话已载最早消息 ID，向上翻页游标
  const listRef = useRef<HTMLDivElement>(null)
  const socketRef = useRef<ReturnType<typeof createCSWS>>()

  useEffect(() => {
    currentRef.current = current
  }, [current])

  const merge = useCallback((incoming: CsMsgOut[]) => {
    const sid = currentRef.current?.id
    const fresh = incoming.filter((m) => m.id && (!sid || m.sessionId === sid))
    if (!fresh.length) return
    setMsgs((prev) => {
      const seen = new Set(prev.map((m) => m.id))
      const add = fresh.filter((m) => !seen.has(m.id))
      if (!add.length) return prev
      for (const m of add) {
        if (!lastIdRef.current || cmpId(lastIdRef.current, m.id) < 0) lastIdRef.current = m.id
        if (!oldestRef.current || cmpId(oldestRef.current, m.id) > 0) oldestRef.current = m.id
      }
      return [...prev, ...add].sort((a, b) => cmpId(a.id, b.id))
    })
  }, [])

  const loadSessions = useCallback(async (selectFirst = false) => {
    try {
      const list = await fetchCsSessions()
      setSessions(list)
      setCurrent((cur) => {
        if (cur) return list.find((s) => s.id === cur.id) ?? cur
        return selectFirst ? (list[0] ?? null) : null
      })
    } catch (e: any) {
      message.error(e.message)
    }
  }, [])

  const openSession = useCallback(
    async (s: CsSessionItem) => {
      setCurrent(s)
      currentRef.current = s
      lastIdRef.current = ''
      oldestRef.current = ''
      setMsgs([])
      setHasMore(false)
      try {
        const h = await fetchCsHistory(s.id, { limit: 20 })
        merge(h.list)
        setHasMore(h.hasMore)
        if (s.unreadAdmin > 0) {
          markCsRead(s.id).catch(() => {})
          setSessions((prev) => prev.map((x) => (x.id === s.id ? { ...x, unreadAdmin: 0 } : x)))
        }
      } catch (e: any) {
        message.error(e.message)
      }
    },
    [merge],
  )

  // 重连成功：补拉断线窗口缺口 + 刷新列表
  const resync = useCallback(() => {
    const cur = currentRef.current
    if (cur && lastIdRef.current) {
      fetchCsHistory(cur.id, { after: lastIdRef.current, limit: 100 })
        .then((h) => merge(h.list))
        .catch(() => {})
    }
    loadSessions()
  }, [merge, loadSessions])

  const handleEvent = useCallback(
    (e: CsEvent) => {
      if (e.type === 'new_message') {
        const m = e.data as CsMsgOut
        const cur = currentRef.current
        if (cur && m.sessionId === cur.id) {
          merge([m])
          if (m.senderRole === 1) {
            markCsRead(cur.id).catch(() => {})
            setSessions((prev) =>
              prev.map((x) => (x.id === cur.id ? { ...x, unreadAdmin: 0, lastMessageText: previewOf(m), lastMessageAt: m.createdAt } : x)),
            )
          }
        }
      } else if (e.type === 'session_update') {
        const u = e.data as SessionUpdate
        setSessions((prev) => {
          if (!prev.some((x) => x.id === u.sessionId)) {
            fetchCsSessions().then(setSessions).catch(() => {}) // 新会话出现，重新拉列表
            return prev
          }
          const next = prev.map((x) =>
            x.id === u.sessionId
              ? { ...x, status: u.status, unreadAdmin: u.unreadAdmin, lastMessageText: u.lastMessageText, lastMessageAt: u.lastMessageAt }
              : x,
          )
          // 进行中在前，按最后消息时间倒序
          return next.sort(
            (a, b) => a.status - b.status || cmpTimeDesc(b.lastMessageAt || b.createdAt, a.lastMessageAt || a.createdAt),
          )
        })
        if (currentRef.current?.id === u.sessionId) {
          setCurrent((c) => (c && c.id === u.sessionId ? { ...c, status: u.status, unreadAdmin: u.unreadAdmin } : c))
        }
      }
    },
    [merge],
  )

  useEffect(() => {
    let disposed = false
    ;(async () => {
      try {
        const list = await fetchCsSessions()
        if (disposed) return
        setSessions(list)
        if (list.length) await openSession(list[0])
      } catch (e: any) {
        message.error(e.message)
      }
      if (!disposed) {
        socketRef.current = createCSWS({
          onMessage: handleEvent,
          onResync: resync,
          onState: (s) => setConnected(s === 'connected'),
        })
      }
    })()
    return () => {
      disposed = true
      socketRef.current?.close()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    const el = listRef.current
    if (el) el.scrollTop = el.scrollHeight
  }, [msgs])

  const loadMore = () => {
    const cur = currentRef.current
    if (loadingMore || !hasMore || !cur || !oldestRef.current) return
    setLoadingMore(true)
    fetchCsHistory(cur.id, { before: oldestRef.current, limit: 20 })
      .then((h) => {
        merge(h.list)
        setHasMore(h.hasMore)
      })
      .catch(() => {})
      .finally(() => setLoadingMore(false))
  }

  const send = async () => {
    const cur = currentRef.current
    const text = input.trim()
    if (!cur || !text) return
    setInput('')
    try {
      const out = await sendCsMessage(cur.id, text)
      merge([out])
    } catch (e: any) {
      message.error(e.message)
      setInput(text)
    }
  }

  const onPickImage = async (file: File) => {
    const cur = currentRef.current
    if (!cur) return false
    try {
      const t = await fetchUploadToken(file.type || 'image/jpeg')
      const resp = await fetch(t.uploadUrl, { method: 'PUT', body: file })
      if (!resp.ok) throw new Error('图片上传失败')
      const out = await sendCsImage(cur.id, t.fileUrl)
      merge([out])
    } catch (e: any) {
      message.error(e.message || '图片发送失败')
    }
    return false
  }

  const closeSession = async () => {
    const cur = currentRef.current
    if (!cur) return
    try {
      await closeCsSession(cur.id)
      setCurrent({ ...cur, status: 2 })
      setSessions((prev) => prev.map((x) => (x.id === cur.id ? { ...x, status: 2 } : x)))
    } catch (e: any) {
      message.error(e.message)
    }
  }

  return (
    <div style={{ display: 'flex', gap: 16, height: 'calc(100vh - 136px)' }}>
      <Card
        title={
          <Space>
            <CustomerServiceOutlined />
            会话列表
          </Space>
        }
        extra={<a onClick={() => loadSessions()}>刷新</a>}
        style={{ width: 340, flexShrink: 0, display: 'flex', flexDirection: 'column' }}
        styles={{ body: { flex: 1, padding: 0, overflowY: 'auto' } }}
      >
        {sessions.map((s) => (
          <div
            key={s.id}
            onClick={() => openSession(s)}
            style={{
              display: 'flex',
              gap: 12,
              alignItems: 'center',
              padding: '12px 16px',
              cursor: 'pointer',
              borderBottom: '1px solid #f5f5f5',
              background: current?.id === s.id ? '#fff7f2' : undefined,
            }}
          >
            <Badge count={s.unreadAdmin} size="small">
              <Avatar src={s.avatar} style={{ background: '#ffd8bf', color: '#ff7a45' }}>
                {s.nickname?.[0] || '客'}
              </Avatar>
            </Badge>
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
                <span style={{ fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {s.nickname || s.phone || `会员${s.memberId}`}
                </span>
                <span style={{ color: '#999', fontSize: 12, flexShrink: 0 }}>{fmtTime(s.lastMessageAt || s.createdAt)}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, marginTop: 2 }}>
                <span style={{ color: '#999', fontSize: 13, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {s.lastMessageText || '（无消息）'}
                </span>
                {s.status === 2 && <Tag style={{ flexShrink: 0, marginInlineEnd: 0 }}>已结束</Tag>}
              </div>
            </div>
          </div>
        ))}
        {!sessions.length && <Empty description="暂无会话" style={{ marginTop: 48 }} />}
      </Card>

      <Card
        title={
          current ? (
            <Space>
              <span>{current.nickname || current.phone || `会员${current.memberId}`}</span>
              {current.status === 2 && <Tag>已结束</Tag>}
            </Space>
          ) : (
            '人工客服'
          )
        }
        extra={
          current?.status === 1 && (
            <Popconfirm title="结束后用户再发消息将自动重新开启，确认结束？" onConfirm={closeSession}>
              <Button size="small" danger>
                结束会话
              </Button>
            </Popconfirm>
          )
        }
        style={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}
        styles={{ body: { flex: 1, display: 'flex', flexDirection: 'column', padding: 0, overflow: 'hidden' } }}
      >
        {!connected && (
          <div style={{ background: 'rgba(0,0,0,0.55)', color: '#fff', textAlign: 'center', fontSize: 12, padding: '4px 0' }}>
            连接中…（消息照常发送，恢复后自动补推）
          </div>
        )}
        <div ref={listRef} style={{ flex: 1, overflowY: 'auto', padding: 24, background: '#f7f7f7' }}>
          {hasMore && (
            <div style={{ textAlign: 'center', marginBottom: 12 }}>
              <a onClick={loadMore}>{loadingMore ? '加载中…' : '加载更早消息'}</a>
            </div>
          )}
          {msgs.map((m) => (
            <div
              key={m.id}
              style={{ display: 'flex', justifyContent: m.senderRole === 1 ? 'flex-start' : 'flex-end', marginBottom: 16 }}
            >
              <div
                style={{
                  maxWidth: '70%',
                  padding: '10px 14px',
                  borderRadius: 10,
                  lineHeight: 1.6,
                  wordBreak: 'break-all',
                  whiteSpace: 'pre-wrap',
                  fontSize: 14,
                  background: m.senderRole === 1 ? '#fff' : '#ff7a45',
                  color: m.senderRole === 1 ? '#333' : '#fff',
                  border: m.senderRole === 1 ? '1px solid #f0f0f0' : undefined,
                }}
              >
                {m.msgType === 2 ? <Image src={m.content} width={180} style={{ borderRadius: 6 }} /> : m.content}
                <div style={{ fontSize: 11, opacity: 0.65, marginTop: 4, textAlign: 'right' }}>{fmtTime(m.createdAt)}</div>
              </div>
            </div>
          ))}
          {!msgs.length && <Empty description={current ? '暂无消息' : '请选择左侧会话'} style={{ marginTop: 80 }} />}
        </div>
        <div style={{ borderTop: '1px solid #f0f0f0', padding: 12, display: 'flex', gap: 12, alignItems: 'center' }}>
          <Upload beforeUpload={onPickImage} showUploadList={false} accept="image/*" disabled={!current}>
            <Button icon={<PictureOutlined />} disabled={!current} />
          </Upload>
          <Input
            value={input}
            placeholder={current ? '回复会员…（Enter 发送）' : '请先选择会话'}
            disabled={!current}
            onChange={(e) => setInput(e.target.value)}
            onPressEnter={send}
          />
          <Button type="primary" icon={<SendOutlined />} onClick={send} disabled={!current || !input.trim()}>
            发送
          </Button>
        </div>
      </Card>
    </div>
  )
}
