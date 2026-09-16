import { useCallback, useEffect, useState } from 'react'
import { Card, Col, Progress, Row, Select, Statistic, Table } from 'antd'
import { fetchReport, RankRow, ReportRow, ReportResp } from '../api/admin'

const DAYS_OPTS = [7, 14, 30, 60, 90].map((d) => ({ value: d, label: `近 ${d} 天` }))

const rankColumns = (amountTitle: string) => [
  { title: '名称', dataIndex: 'name', ellipsis: true },
  { title: '成交', dataIndex: 'count', width: 90 },
  { title: amountTitle, dataIndex: 'amount', width: 120, render: (v: string) => `¥${v}` },
]

export default function Report() {
  const [days, setDays] = useState(30)
  const [data, setData] = useState<ReportResp>()
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (d: number) => {
    setLoading(true)
    try {
      setData(await fetchReport(d))
    } catch {
      // 顶部条已有全局错误提示，这里静默保持空态
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    load(days)
  }, [days, load])

  const dailyMax = Math.max(1, ...(data?.daily ?? []).map((d: ReportRow) => d.orders))
  const topMax = Math.max(1, ...(data?.topProducts ?? []).map((r: RankRow) => r.count))

  return (
    <Card
      title="经营报表"
      loading={loading && !data}
      extra={<Select value={days} options={DAYS_OPTS} onChange={setDays} style={{ width: 120 }} />}
    >
      <Row gutter={16}>
        <Col span={6}>
          <Card size="small">
            <Statistic title="今日成交额" prefix="¥" value={data?.summary.todayGmv ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic title="今日订单" value={data?.summary.todayOrders ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic title="本月成交额" prefix="¥" value={data?.summary.monthGmv ?? 0} />
          </Card>
        </Col>
        <Col span={6}>
          <Card size="small">
            <Statistic title="本月订单" value={data?.summary.monthOrders ?? 0} suffix={`· 待处理售后 ${data?.summary.pendingAfters ?? 0}`} />
          </Card>
        </Col>
      </Row>

      <Card size="small" title={`近 ${days} 天成交`} style={{ marginTop: 16 }}>
        <Table
          rowKey="date"
          size="small"
          pagination={false}
          dataSource={data?.daily ?? []}
          columns={[
            { title: '日期', dataIndex: 'date', width: 100 },
            { title: '订单', dataIndex: 'orders', width: 100 },
            {
              title: '占比',
              dataIndex: 'orders',
              render: (v: number) => <Progress percent={Math.round((v / dailyMax) * 100)} showInfo={false} strokeColor="#ff7a45" />,
            },
            { title: '成交额', dataIndex: 'gmv', width: 130, render: (v: string) => `¥${v}` },
          ]}
        />
      </Card>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={12}>
          <Card size="small" title="商品销量 Top10">
            <Table
              rowKey="name"
              size="small"
              pagination={false}
              dataSource={data?.topProducts ?? []}
              columns={[
                { title: '商品', dataIndex: 'name', ellipsis: true },
                {
                  title: '销量',
                  dataIndex: 'count',
                  width: 160,
                  render: (v: number) => <Progress percent={Math.round((v / topMax) * 100)} strokeColor="#1677ff" format={() => v} />,
                },
                { title: '金额', dataIndex: 'amount', width: 110, render: (v: string) => `¥${v}` },
              ]}
            />
          </Card>
        </Col>
        <Col span={12}>
          <Card size="small" title="品类成交分布">
            <Table
              rowKey="name"
              size="small"
              pagination={false}
              dataSource={data?.breeds ?? []}
              columns={rankColumns('成交额') as never}
            />
          </Card>
        </Col>
      </Row>
    </Card>
  )
}
