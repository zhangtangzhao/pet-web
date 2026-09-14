import Taro from '@tarojs/taro'
import { getToken } from './request'
import type { CsEvent } from './types'

// 人工客服 WebSocket 地址：H5 同源（nginx 反代 /api/ws），小程序用完整 wss 地址
function csWSUrl(token: string): string {
  if (process.env.TARO_ENV === 'h5') {
    return `${location.origin.replace(/^http/, 'ws')}/api/ws/cs?token=${encodeURIComponent(token)}`
  }
  const base = 'https://your-domain.com/api'.replace(/^http/, 'ws')
  return `${base}/ws/cs?token=${encodeURIComponent(token)}`
}

export type CsSocketState = 'connected' | 'connecting'

interface CsSocketHandlers {
  onMessage: (e: CsEvent) => void
  /** 连接建立（含重连）后回调：页面用它做 ?after=lastMsgID 补拉 */
  onResync?: () => void
  onState?: (s: CsSocketState) => void
}

const RETRY_MAX_DELAY = 30000
const DEAD_AFTER = 45000 // 超过此时长无任何服务端帧 → 判定半开连接，主动重建

/**
 * 人工客服连接管理：指数退避 + 抖动自动重连、45s 无帧判死自愈、
 * 网络恢复 / 页面重新可见时立即重连。对端恢复前用户完全无感知，消息由 REST 兜底。
 */
export function createCSSocket(h: CsSocketHandlers) {
  let task: Taro.SocketTask | null = null
  let destroyed = false
  let retry = 0
  let lastAlive = Date.now()
  let watchdog: ReturnType<typeof setInterval> | undefined

  const connect = () => {
    if (destroyed) return
    const token = getToken()
    if (!token) {
      scheduleRetry()
      return
    }
    h.onState?.('connecting')
    try {
      task = Taro.connectSocket({ url: csWSUrl(token) })
    } catch {
      task = null
      scheduleRetry()
      return
    }
    if (!task) {
      scheduleRetry()
      return
    }
    const sock = task // 关闭/重建竞态下忽略旧 socket 的迟到回调
    task.onOpen(() => {
      if (task !== sock) return
      retry = 0
      lastAlive = Date.now()
      h.onState?.('connected')
      h.onResync?.()
    })
    task.onMessage(({ data }) => {
      if (task !== sock) return
      lastAlive = Date.now()
      if (typeof data !== 'string') return
      try {
        h.onMessage(JSON.parse(data) as CsEvent)
      } catch {
        // 忽略非 JSON 帧
      }
    })
    task.onClose(() => {
      if (task !== sock) return
      task = null
      scheduleRetry()
    })
    task.onError(() => {
      if (task !== sock) return
      try {
        sock.close({})
      } catch {
        // onClose 会跟进并触发重连
      }
    })
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
  watchdog = setInterval(() => {
    if (destroyed) return
    if (Date.now() - lastAlive > DEAD_AFTER) {
      retry = 0
      try {
        task?.close({})
      } catch {
        // 忽略
      }
      task = null
      connect()
    }
  }, 5000)

  // 弱网恢复：网络重新可用时立即重连，不等退避计时
  Taro.onNetworkStatusChange?.((res) => {
    if (destroyed || !res.isConnected) return
    retry = 0
    try {
      task?.close({})
    } catch {
      // 忽略
    }
    task = null
    connect()
  })

  return {
    /** 页面重新可见时调用：若未连接则立即重建 */
    reconnectNow() {
      if (destroyed || task) return
      retry = 0
      connect()
    },
    close() {
      destroyed = true
      if (watchdog) clearInterval(watchdog)
      try {
        task?.close({})
      } catch {
        // 忽略
      }
    },
  }
}
