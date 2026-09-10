import { useEffect, useState } from 'react'
import { Card, Col, Row, Statistic } from 'antd'
import { fetchOverview } from '../api/admin'

export default function Dashboard() {
  const [ov, setOv] = useState<Awaited<ReturnType<typeof fetchOverview>>>()

  useEffect(() => {
    fetchOverview().then(setOv).catch(() => {})
  }, [])

  return (
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
  )
}
