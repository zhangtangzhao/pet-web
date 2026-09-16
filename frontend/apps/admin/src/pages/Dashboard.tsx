import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic, Table, Tag } from 'antd'
import { fetchOverview, fetchStockAlerts, StockAlertRow } from '../api/admin'

export default function Dashboard() {
  const [ov, setOv] = useState<Awaited<ReturnType<typeof fetchOverview>>>()
  const [alerts, setAlerts] = useState<StockAlertRow[]>([])

  useEffect(() => {
    fetchOverview().then(setOv).catch(() => {})
    fetchStockAlerts().then(setAlerts).catch(() => {})
  }, [])

  return (
    <>
      <Row gutter={16}>
      <Col span={6}>
        <Card>
          <Statistic title="在售宠物" value={ov?.onSaleCount ?? '-'} suffix="只" />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic title="今日成交订单" value={ov?.todayOrders ?? '-'} suffix="单" />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic title="今日成交额" prefix="¥" value={ov?.todayGmv ?? '-'} precision={2} />
        </Card>
      </Col>
      <Col span={6}>
        <Card>
          <Statistic title="注册会员" value={ov?.memberCount ?? '-'} suffix="人" />
        </Card>
      </Col>
      </Row>
      {alerts.length > 0 && (
        <Card title={<span>库存预警 <Tag color="red">{alerts.length}</Tag></span>} style={{ marginTop: 16 }}>
          <Table
            rowKey="productId"
            size="small"
            pagination={false}
            dataSource={alerts}
            columns={[
              { title: '商品', dataIndex: 'title', ellipsis: true },
              {
                title: '库存/阈值',
                width: 120,
                render: (_: unknown, r: StockAlertRow) => (
                  <span>
                    <Tag color={r.stock === 0 ? 'red' : 'orange'}>{r.stock}</Tag>/{r.threshold}
                  </span>
                ),
              },
              { title: '累计销量', dataIndex: 'sales', width: 100 },
            ]}
          />
        </Card>
      )}
    </>
  )
}
