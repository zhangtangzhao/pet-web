// Package hub 客服 WebSocket 连接注册表与推送（仅标准库 + coder/websocket，避免与 svc 循环依赖）。
// 单实例内存实现；多实例部署时可将其替换为 Redis pub/sub 广播，对外接口不变。
package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

// 连接角色
const (
	RoleMember = "member"
	RoleAdmin  = "admin"
)

// 心跳与判死参数
const (
	pingInterval  = 20 * time.Second // 应用层 ping 周期
	warnThreshold = 30 * time.Second // 超过此时长未收到任何帧 → 推送 closing_warning
	closeGrace    = 30 * time.Second // 预警后再等此时长仍无帧 → 断开
	writeTimeout  = 10 * time.Second
	sendBuffer    = 64
)

// Event 服务端推送事件信封
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data,omitempty"`
}

// Conn 单条 WebSocket 连接
type Conn struct {
	ws     *websocket.Conn
	send   chan []byte
	alive  atomic.Int64 // 最后收到帧的时间（unix nano）
	warned atomic.Bool  // 是否已推送 closing_warning
	closed chan struct{}
	once   sync.Once
}

// Hub 连接注册表：member/admin 各自 id → 连接集合（同账号多标签页多条连接）
type Hub struct {
	mu      sync.RWMutex
	members map[int64]map[*Conn]struct{}
	admins  map[int64]map[*Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{
		members: make(map[int64]map[*Conn]struct{}),
		admins:  make(map[int64]map[*Conn]struct{}),
	}
}

// Serve 接管已 Accept 的连接：注册 → 读写泵 → 判死看护，直到连接结束并注销
func (h *Hub) Serve(role string, id int64, ws *websocket.Conn) {
	c := &Conn{
		ws:     ws,
		send:   make(chan []byte, sendBuffer),
		closed: make(chan struct{}),
	}
	c.alive.Store(time.Now().UnixNano())
	h.Register(role, id, c)
	defer h.Unregister(role, id, c)

	go c.writePump()
	go c.watchdog()
	c.readPump()
}

// Register 登记连接
func (h *Hub) Register(role string, id int64, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m := h.table(role)
	if m[id] == nil {
		m[id] = make(map[*Conn]struct{})
	}
	m[id][c] = struct{}{}
}

// Unregister 注销连接
func (h *Hub) Unregister(role string, id int64, c *Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	m := h.table(role)
	if set, ok := m[id]; ok {
		delete(set, c)
		if len(set) == 0 {
			delete(m, id)
		}
	}
}

// MemberOnline 判断会员是否有任一在线连接（用于离线时才投递微信通知）
func (h *Hub) MemberOnline(id int64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.members[id]) > 0
}

// PushMember 推送给指定会员的全部连接
func (h *Hub) PushMember(id int64, v any) {
	h.mu.RLock()
	set := make([]*Conn, 0, 1)
	for c := range h.members[id] {
		set = append(set, c)
	}
	h.mu.RUnlock()
	pushAll(set, v)
}

// PushAdmins 推送给全部在线客服连接
func (h *Hub) PushAdmins(v any) {
	h.mu.RLock()
	set := make([]*Conn, 0, 8)
	for _, conns := range h.admins {
		for c := range conns {
			set = append(set, c)
		}
	}
	h.mu.RUnlock()
	pushAll(set, v)
}

func (h *Hub) table(role string) map[int64]map[*Conn]struct{} {
	if role == RoleAdmin {
		return h.admins
	}
	return h.members
}

func pushAll(conns []*Conn, v any) {
	if len(conns) == 0 {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	for _, c := range conns {
		select {
		case c.send <- data:
		default: // 缓冲满（对端假死/网络拥塞），丢帧防阻塞；判死机制会回收该连接
		}
	}
}

// readPump 阻塞读；收到任何帧即刷新存活时间并清除预警
func (c *Conn) readPump() {
	for {
		if _, _, err := c.ws.Read(context.Background()); err != nil {
			c.shutdown()
			return
		}
		c.alive.Store(time.Now().UnixNano())
		c.warned.Store(false)
	}
}

// writePump 唯一写协程：串行化业务帧与应用层 ping
func (c *Conn) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	for {
		select {
		case data := <-c.send:
			if !c.write(data) {
				return
			}
		case <-ticker.C:
			if !c.write([]byte(fmt.Sprintf(`{"type":"ping","ts":%d}`, time.Now().Unix()))) {
				return
			}
		case <-c.closed:
			return
		}
	}
}

func (c *Conn) write(data []byte) bool {
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	if err := c.ws.Write(ctx, websocket.MessageText, data); err != nil {
		c.shutdown()
		return false
	}
	return true
}

// watchdog 两阶段判死：超 warnThreshold 无帧推送 closing_warning，再宽限 closeGrace 仍无帧则断开
func (c *Conn) watchdog() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.closed:
			return
		case <-ticker.C:
			elapsed := time.Since(time.Unix(0, c.alive.Load()))
			if elapsed > warnThreshold+closeGrace {
				log.Printf("hub: close dead conn, silent for %s", elapsed.Truncate(time.Second))
				c.shutdown()
				return
			}
			if elapsed > warnThreshold && !c.warned.Swap(true) {
				select {
				case c.send <- []byte(`{"type":"closing_warning","data":"连接即将断开,30s 内未恢复将自动断开"}`):
				default:
				}
			}
		}
	}
}

func (c *Conn) shutdown() {
	c.once.Do(func() {
		close(c.closed)
		_ = c.ws.Close(websocket.StatusNormalClosure, "")
	})
}
