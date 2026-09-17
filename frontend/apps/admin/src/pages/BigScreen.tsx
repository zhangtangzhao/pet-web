import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic, Tag } from 'antd'
import { fetchRealtime, RealtimeData } from '../api/engagement'

export default function BigScreen() {
  const [d, setD] = useState<RealtimeData>()
  useEffect(() => {
    const load = () => fetchRealtime().then(setD).catch(() => {})
    load()
    const t = setInterval(load, 30000)
    return () => clearInterval(t)
  }, [])
  const maxHour = Math.max(1, ...(d?.hourly ?? []).map((h) => Number(h.gmv)))
  return (
    <div>
      <Row gutter={16}>
        <Col span={8}><Card><Statistic title="今日 GMV" prefix="¥" value={d?.todayGmv ?? '-'} precision={2} /></Card></Col>
        <Col span={8}><Card><Statistic title="今日订单" value={d?.todayOrders ?? '-'} suffix="单" /></Card></Col>
        <Col span={8}><Card><Statistic title="更新时间" value={d?.generatedAt ?? '-'} /></Card></Col>
      </Row>
      <Card title="24 小时成交分布" style={{ marginTop: 16 }}>
        {(d?.hourly ?? []).map((h) => (
          <div key={h.hour} style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
            <span style={{ width: 34, fontSize: 12, color: '#888' }}>{h.hour}时</span>
            <div style={{ flex: 1, background: '#f5f5f5', borderRadius: 4 }}>
              <div style={{ width: `${(Number(h.gmv) / maxHour) * 100}%`, minWidth: 2, height: 14, borderRadius: 4, background: 'linear-gradient(90deg,#ff6b35,#ff8f5c)' }} />
            </div>
            <span style={{ width: 130, fontSize: 12, color: '#666' }}>¥{h.gmv} / {h.orders}单</span>
          </div>
        ))}
      </Card>
      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card title="下单漏斗（近 30 天）">
            {Object.entries(d?.funnel ?? {}).map(([k, v]) => (
              <div key={k} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0' }}>
                <span style={{ fontSize: 13 }}>{k}</span><Tag>{v}</Tag>
              </div>
            ))}
          </Card>
        </Col>
        <Col span={12}>
          <Card title="地域分布 Top10（收货省份）">
            {(d?.regions ?? []).map((r) => (
              <div key={r.province} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0' }}>
                <span style={{ fontSize: 13 }}>{r.province}</span><Tag color="orange">{r.orders} 单</Tag>
              </div>
            ))}
          </Card>
        </Col>
      </Row>
    </div>
  )
}
