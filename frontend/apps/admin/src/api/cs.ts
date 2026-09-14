import { ACCESS_KEY, client } from './client'

export interface CsMsgOut {
  id: string
  sessionId: string
  senderRole: number // 1会员 2客服
  msgType: number // 1文本 2图片
  content: string
  createdAt: string
}

export interface CsHistoryResp {
  list: CsMsgOut[]
  hasMore: boolean
}

export interface CsSessionItem {
  id: string
  memberId: string
  nickname: string
  phone: string
  avatar: string
  status: number // 1进行中 2已结束
  unreadAdmin: number
  lastMessageText: string
  lastMessageAt: string
  createdAt: string
}

export interface SessionUpdate {
  sessionId: string
  status: number
  unreadAdmin: number
  unreadMember: number
  lastMessageText: string
  lastMessageAt: string
}

export interface CsEvent {
  type: 'ping' | 'closing_warning' | 'new_message' | 'session_update'
  data?: unknown
}

export async function fetchCsSessions(): Promise<CsSessionItem[]> {
  return (await client.get('/admin/cs/sessions')) as any
}

export async function fetchCsHistory(id: string, params: { before?: string; after?: string; limit?: number }): Promise<CsHistoryResp> {
  return (await client.get(`/admin/cs/sessions/${id}/messages`, { params })) as any
}

export async function sendCsMessage(id: string, content: string): Promise<CsMsgOut> {
  return (await client.post(`/admin/cs/sessions/${id}/messages`, { msgType: 1, content })) as any
}

export async function sendCsImage(id: string, fileUrl: string): Promise<CsMsgOut> {
  return (await client.post(`/admin/cs/sessions/${id}/messages`, { msgType: 2, content: fileUrl })) as any
}

export async function markCsRead(id: string) {
  await client.post(`/admin/cs/sessions/${id}/read`)
}

export async function closeCsSession(id: string) {
  await client.post(`/admin/cs/sessions/${id}/close`)
}

export async function fetchUploadToken(contentType: string): Promise<{ uploadUrl: string; fileUrl: string; method: string }> {
  return (await client.post('/admin/upload-token', { dir: 'chat', contentType })) as any
}

export type CsSocketState = 'connected' | 'connecting'

interface CsSocketHandlers {
  onMessage: (e: CsEvent) => void
  /** 连接建立（含重连）后回调：调用方用它做 ?after=lastId 补拉 */
  onResync?: () => void
  onState?: (s: CsSocketState) => void
}

const RETRY_MAX_DELAY = 30000
const DEAD_AFTER = 45000 // 超过此时长无任何服务端帧 → 判定半开连接，主动重建

// 与 mobile/src/ws.ts 同一套语义：45s 无帧自愈、指数退避+抖动重连、online/visibility 即时重连
export function createCSWS(h: CsSocketHandlers) {
  let ws: WebSocket | null = null
  let destroyed = false
  let retry = 0
  let lastAlive = Date.now()
  let watchdog: number | undefined

  const connect = () => {
    if (destroyed) return
    const token = localStorage.getItem(ACCESS_KEY)
    if (!token) return
    h.onState?.('connecting')
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    let sock: WebSocket
    try {
      sock = new WebSocket(`${proto}://${location.host}/api/ws/cs?token=${encodeURIComponent(token)}`)
    } catch {
      scheduleRetry()
      return
    }
    ws = sock
    sock.onopen = () => {
      if (ws !== sock) return
      retry = 0
      lastAlive = Date.now()
      h.onState?.('connected')
      h.onResync?.()
    }
    sock.onmessage = (ev) => {
      if (ws !== sock) return
      lastAlive = Date.now()
      try {
        h.onMessage(JSON.parse(ev.data) as CsEvent)
      } catch {
        // 忽略非 JSON 帧
      }
    }
    sock.onclose = () => {
      if (ws !== sock) return
      ws = null
      scheduleRetry()
    }
    sock.onerror = () => {
      if (ws !== sock) return
      try {
        sock.close()
      } catch {
        // onclose 会跟进并触发重连
      }
    }
  }

  const scheduleRetry = () => {
    if (destroyed) return
    h.onState?.('connecting')
    const delay = Math.min(RETRY_MAX_DELAY, 1000 * 2 ** retry) * (0.7 + Math.random() * 0.6)
    retry++
    setTimeout(() => {
      if (!destroyed) connect()
    }, delay)
  }

  // 判死自愈：长时间无帧则主动断开并立即重建
  watchdog = window.setInterval(() => {
    if (destroyed) return
    if (Date.now() - lastAlive > DEAD_AFTER) {
      retry = 0
      if (ws) {
        const sock = ws
        ws = null
        try {
          sock.close()
        } catch {
          // 忽略
        }
      }
      connect()
    }
  }, 5000)

  /** 页面重新可见 / 网络恢复时调用：未连接则立即重建 */
  const reconnectNow = () => {
    if (destroyed || (ws && ws.readyState === WebSocket.OPEN)) return
    retry = 0
    if (ws) {
      const sock = ws
      ws = null
      try {
        sock.close()
      } catch {
        // 忽略
      }
    }
    connect()
  }

  const onWakeup = () => {
    if (document.visibilityState === 'visible') reconnectNow()
  }
  window.addEventListener('online', reconnectNow)
  document.addEventListener('visibilitychange', onWakeup)

  return {
    reconnectNow,
    close() {
      destroyed = true
      if (watchdog) clearInterval(watchdog)
      window.removeEventListener('online', reconnectNow)
      document.removeEventListener('visibilitychange', onWakeup)
      if (ws) {
        const sock = ws
        ws = null
        try {
          sock.close()
        } catch {
          // 忽略
        }
      }
    },
  }
}
