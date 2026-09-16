import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Input, message, Table, Tabs } from 'antd'
import { DownloadOutlined, SearchOutlined } from '@ant-design/icons'
import { downloadCSV, fetchFinanceDaily, fetchFinancePayments, FinanceDailyRow, FinancePaymentRow } from '../api/admin'

function ExportButtons() {
  const [busy, setBusy] = useState('')
  const run = async (kind: 'orders' | 'members' | 'points') => {
    setBusy(kind)
    try {
      await downloadCSV(kind, `pet-${kind}-${new Date().toISOString().slice(0, 10)}.csv`)
      message.success('导出成功')
    } catch (e: any) {
      message.error(e.message)
    } finally {
      setBusy('')
    }
  }
  return (
    <div style={{ display: 'flex', gap: 8 }}>
      <Button size="small" icon={<DownloadOutlined />} loading={busy === 'orders'} onClick={() => run('orders')}>
        订单导出
      </Button>
      <Button size="small" icon={<DownloadOutlined />} loading={busy === 'members'} onClick={() => run('members')}>
        会员导出
      </Button>
      <Button size="small" icon={<DownloadOutlined />} loading={busy === 'points'} onClick={() => run('points')}>
        积分导出
      </Button>
    </div>
  )
}

function Payments() {
  const [list, setList] = useState<FinancePaymentRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [orderNo, setOrderNo] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (nextPage: number, no: string) => {
    setLoading(true)
    try {
      const resp = await fetchFinancePayments({ page: nextPage, pageSize: 10, orderNo: no || undefined })
      setList(resp.list)
      setTotal(resp.total)
      setPage(nextPage)
    } catch (e: any) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load(1, '')
  }, [load])

  const columns = [
    { title: '订单号', dataIndex: 'orderNo', width: 170 },
    { title: '会员', dataIndex: 'memberName', width: 110, ellipsis: true },
    { title: '商品', dataIndex: 'productTitle', ellipsis: true },
    { title: '实付', dataIndex: 'amount', width: 100, render: (v: string) => `¥${v}` },
    { title: '退款', dataIndex: 'refundAmount', width: 100, render: (v: string) => (v && v !== '0.00' ? <span style={{ color: '#cf1322' }}>-¥{v}</span> : '-') },
    { title: '方式', dataIndex: 'payTypeText', width: 90 },
    { title: '支付单号', dataIndex: 'paymentNo', width: 170, ellipsis: true, render: (v: string) => v || '-' },
    { title: '状态', dataIndex: 'statusText', width: 80 },
    { title: '支付时间', dataIndex: 'paidAt', width: 165, render: (v: string) => (v ? v.replace('T', ' ').slice(0, 16) : '-') },
    { title: '下单时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
  ]

  return (
    <>
      <div style={{ marginBottom: 12, display: 'flex', gap: 8 }}>
        <Input
          allowClear
          placeholder="按订单号搜索"
          prefix={<SearchOutlined />}
          style={{ width: 260 }}
          value={orderNo}
          onChange={(e) => setOrderNo(e.target.value)}
          onPressEnter={() => load(1, orderNo)}
          onClear={() => {
            setOrderNo('')
            load(1, '')
          }}
        />
      </div>
      <Table
        rowKey="orderNo"
        loading={loading}
        size="small"
        columns={columns as never}
        dataSource={list}
        scroll={{ x: 1200 }}
        pagination={{
          current: page,
          total,
          pageSize: 10,
          showSizeChanger: false,
          onChange: (p) => load(p, orderNo),
        }}
      />
    </>
  )
}

function Daily() {
  const [list, setList] = useState<FinanceDailyRow[]>([])
  const [days, setDays] = useState(30)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (d: number) => {
    setLoading(true)
    try {
      setList(await fetchFinanceDaily(d))
    } catch (e: any) {
      message.error(e.message)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load(30)
  }, [load])

  const columns = [
    { title: '日期', dataIndex: 'date', width: 120 },
    { title: '成交单数', dataIndex: 'orderCount', width: 100 },
    { title: '支付金额', dataIndex: 'payAmount', width: 120, render: (v: string) => `¥${v}` },
    { title: '退款金额', dataIndex: 'refundAmount', width: 120, render: (v: string) => (v && v !== '0.00' ? <span style={{ color: '#cf1322' }}>-¥{v}</span> : '-') },
    { title: '净收入', dataIndex: 'netAmount', width: 120, render: (v: string) => <b>¥{v}</b> },
  ]

  return (
    <>
      <div style={{ marginBottom: 12, display: 'flex', gap: 8 }}>
        {[7, 14, 30, 60, 90].map((d) => (
          <Button key={d} size="small" type={days === d ? 'primary' : 'default'} onClick={() => { setDays(d); load(d) }}>
            近 {d} 天
          </Button>
        ))}
      </div>
      <Table
        rowKey="date"
        loading={loading}
        size="small"
        columns={columns as never}
        dataSource={list}
        pagination={false}
      />
    </>
  )
}

export default function Finance() {
  return (
    <Card title="财务对账" extra={<ExportButtons />}>
      <Tabs
        defaultActiveKey="payments"
        items={[
          { key: 'payments', label: '交易流水', children: <Payments /> },
          { key: 'daily', label: '每日汇总', children: <Daily /> },
        ]}
      />
    </Card>
  )
}
