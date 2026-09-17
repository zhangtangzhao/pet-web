import { useEffect, useState } from 'react'
import { Alert, Card, Table, Tag } from 'antd'
import { ActivityItem, fetchActivityCalendar } from '../api/engagement'

const TYPE_COLOR: Record<string, string> = { flash: 'red', group: 'orange', coupon: 'blue' }

export default function Calendar() {
  const [items, setItems] = useState<ActivityItem[]>([])
  const [conflicts, setConflicts] = useState<string[]>([])
  useEffect(() => { fetchActivityCalendar(30).then((r) => { setItems(r.items); setConflicts(r.conflicts) }).catch(() => {}) }, [])
  return (
    <Card title="活动日历（近 30 天排期）">
      {conflicts.length > 0 && <Alert type="warning" showIcon message="检测到同商品活动时间窗重叠" description={conflicts.map((c, i) => <div key={i}>{c}</div>)} style={{ marginBottom: 16 }} />}
      <Table rowKey={(r) => r.type + r.id} size="small" dataSource={items} pagination={false} columns={[
        { title: '类型', dataIndex: 'typeText', width: 100, render: (v: string, r: ActivityItem) => <Tag color={TYPE_COLOR[r.type]}>{v}</Tag> },
        { title: '活动', dataIndex: 'name' },
        { title: '商品', dataIndex: 'productId', width: 210 },
        { title: '开始', dataIndex: 'startAt', width: 160 },
        { title: '结束', dataIndex: 'endAt', width: 160 },
        { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => v === 1 ? <Tag color="green">启用中</Tag> : <Tag>停用</Tag> },
      ] as never} />
    </Card>
  )
}
