import { useEffect, useState } from 'react'
import { Card, Table, Tag } from 'antd'
import { fetchInsuranceApplies, InsuranceApplyRow } from '../api/play'
export default function InsuranceApplies() {
  const [list, setList] = useState<InsuranceApplyRow[]>([])
  useEffect(() => { fetchInsuranceApplies().then(setList).catch(() => {}) }, [])
  return (
    <Card title="保险投保意向">
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '会员', dataIndex: 'nickname' },
        { title: '保险产品', dataIndex: 'product_name' },
        { title: '联系人', dataIndex: 'contact', width: 120 },
        { title: '电话', dataIndex: 'phone', width: 140 },
        { title: '状态', dataIndex: 'status', width: 100, render: (v: number) => v === 0 ? <Tag color="orange">待处理</Tag> : <Tag color="green">已联系</Tag> },
      ] as never} />
    </Card>
  )
}
