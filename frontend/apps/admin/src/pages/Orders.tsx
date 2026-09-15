import { useCallback, useEffect, useState } from 'react'
import { Button, Descriptions, Drawer, Form, Input, InputNumber, message, Modal, Popconfirm, Table, Tag } from 'antd'
import { PageResp } from '../api/client'
import { deliverOrder, fetchOrders, OrderView, refundOrder, shipOrder } from '../api/admin'

const STATUS_TAG: Record<number, string> = {
  10: 'gold',
  20: 'green',
  30: 'blue',
  40: 'default',
  45: 'default',
  50: 'orange',
  60: 'red',
}

const SHIP_TEXT: Record<number, string> = { 0: '待配送', 1: '配送中', 2: '已送达' }

export default function Orders() {
  const [data, setData] = useState<PageResp<OrderView>>({ total: 0, list: [] })
  const [query, setQuery] = useState({ page: 1, pageSize: 10, status: 0, orderNo: '', keyword: '' })
  const [loading, setLoading] = useState(false)
  const [detail, setDetail] = useState<OrderView>()
  const [refundTarget, setRefundTarget] = useState<OrderView>()
  const [refundForm] = Form.useForm()
  const [shipTarget, setShipTarget] = useState<OrderView>()
  const [shipForm] = Form.useForm()

  const load = useCallback(
    async (q = query) => {
      setLoading(true)
      try {
        setData(await fetchOrders(q))
      } catch (e: any) {
        message.error(e.message)
      } finally {
        setLoading(false)
      }
    },
    [query],
  )

  useEffect(() => {
    load()
  }, [query])

  const submitRefund = async () => {
    const v = await refundForm.validateFields()
    if (!refundTarget) return
    try {
      await refundOrder(refundTarget.orderNo, v.reason, v.amount != null ? String(v.amount) : undefined)
      message.success('退款已受理')
      setRefundTarget(undefined)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const submitShip = async () => {
    const v = await shipForm.validateFields()
    if (!shipTarget) return
    try {
      await shipOrder(shipTarget.orderNo, v.shipNo.trim())
      message.success('已发货')
      setShipTarget(undefined)
      setDetail(undefined)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const submitDeliver = async (orderNo: string) => {
    try {
      await deliverOrder(orderNo)
      message.success('已标记送达')
      setDetail(undefined)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '订单号', dataIndex: 'orderNo', width: 200 },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (v: number, r: OrderView) => <Tag color={STATUS_TAG[v]}>{r.statusText}</Tag>,
    },
    { title: '实付', dataIndex: 'payAmount', width: 110, render: (v: string) => `¥${v}` },
    { title: '联系人', dataIndex: 'contactName', width: 100 },
    { title: '手机号', dataIndex: 'contactPhone', width: 130 },
    { title: '商品', dataIndex: ['items', 0, 'productTitle'], ellipsis: true },
    { title: '下单时间', dataIndex: 'createdAt', width: 170, render: (v: string) => v?.replace('T', ' ').slice(0, 19) },
    {
      title: '操作',
      width: 170,
      fixed: 'right' as const,
      render: (_: unknown, r: OrderView) => (
        <>
          <Button size="small" type="link" onClick={() => setDetail(r)}>
            详情
          </Button>
          {[20, 30].includes(r.status) && (
            <Button
              size="small"
              type="link"
              danger
              onClick={() => {
                refundForm.resetFields()
                setRefundTarget(r)
              }}
            >
              退款
            </Button>
          )}
        </>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
        <Input.Search
          allowClear
          placeholder="订单号"
          style={{ width: 240 }}
          onSearch={(v) => setQuery({ ...query, page: 1, orderNo: v })}
        />
        <Input.Search
          allowClear
          placeholder="联系人 / 手机号"
          style={{ width: 220 }}
          onSearch={(v) => setQuery({ ...query, page: 1, keyword: v })}
        />
      </div>
      <Table
        rowKey="orderNo"
        loading={loading}
        scroll={{ x: 1100 }}
        columns={columns as never}
        dataSource={data.list}
        pagination={{
          total: data.total,
          current: query.page,
          pageSize: query.pageSize,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery({ ...query, page, pageSize }),
        }}
      />
      <Drawer title={`订单 ${detail?.orderNo ?? ''}`} width={560} open={!!detail} onClose={() => setDetail(undefined)}>
        {detail && (
          <>
            <Descriptions column={1} size="small" bordered>
              <Descriptions.Item label="状态">
                <Tag color={STATUS_TAG[detail.status]}>{detail.statusText}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="实付金额">¥{detail.payAmount}</Descriptions.Item>
              {(Number(detail.serviceFee) > 0 || Number(detail.discountAmount) > 0 || detail.couponInfo) && (
                <>
                  <Descriptions.Item label="商品金额">¥{detail.totalAmount}</Descriptions.Item>
                  {Number(detail.serviceFee) > 0 && (
                    <Descriptions.Item label="增值服务费">¥{detail.serviceFee}</Descriptions.Item>
                  )}
                  {Number(detail.discountAmount) > 0 && (
                    <Descriptions.Item label="券抵扣">
                      <span style={{ color: '#fa541c' }}>
                        -¥{detail.discountAmount}
                        {detail.couponInfo ? `（${detail.couponInfo}）` : ''}
                      </span>
                    </Descriptions.Item>
                  )}
                </>
              )}
              <Descriptions.Item label="联系人">
                {detail.contactName} {detail.contactPhone}
              </Descriptions.Item>
              <Descriptions.Item label="备注">{detail.remark || '-'}</Descriptions.Item>
              <Descriptions.Item label="配送方式">
                {detail.shipMethod || '-'}
                {Number(detail.shipFee) > 0 ? `（运费 ¥${detail.shipFee}）` : ''}
              </Descriptions.Item>
              {detail.shipAddress && <Descriptions.Item label="收货地址">{detail.shipAddress}</Descriptions.Item>}
              <Descriptions.Item label="配送状态">{SHIP_TEXT[detail.shipStatus] ?? '-'}</Descriptions.Item>
              {detail.shipNo && <Descriptions.Item label="运单号">{detail.shipNo}</Descriptions.Item>}
              <Descriptions.Item label="下单时间">{detail.createdAt?.replace('T', ' ').slice(0, 19)}</Descriptions.Item>
              {detail.paidAt && (
                <Descriptions.Item label="支付时间">{detail.paidAt.replace('T', ' ').slice(0, 19)}</Descriptions.Item>
              )}
              {detail.shippedAt && (
                <Descriptions.Item label="发货时间">{detail.shippedAt.replace('T', ' ').slice(0, 19)}</Descriptions.Item>
              )}
              {detail.deliveredAt && (
                <Descriptions.Item label="送达时间">{detail.deliveredAt.replace('T', ' ').slice(0, 19)}</Descriptions.Item>
              )}
            </Descriptions>
            {detail.status === 20 && (
              <div style={{ marginTop: 16, display: 'flex', justifyContent: 'flex-end', gap: 8 }}>
                {detail.shipStatus === 0 && (
                  <Button
                    type="primary"
                    size="small"
                    onClick={() => {
                      shipForm.resetFields()
                      setShipTarget(detail)
                    }}
                  >
                    发货
                  </Button>
                )}
                {detail.shipStatus === 1 && (
                  <Popconfirm title="确认已送达？" onConfirm={() => submitDeliver(detail.orderNo)}>
                    <Button type="primary" size="small">
                      标记送达
                    </Button>
                  </Popconfirm>
                )}
              </div>
            )}
            <Table
              style={{ marginTop: 16 }}
              rowKey="productId"
              size="small"
              pagination={false}
              columns={[
                { title: '商品', dataIndex: 'productTitle' },
                { title: '品种', dataIndex: 'breedName', width: 100 },
                { title: '单价', dataIndex: 'price', width: 100, render: (v: string) => `¥${v}` },
                { title: '数量', dataIndex: 'quantity', width: 70 },
              ]}
              dataSource={detail.items}
            />
            {detail.serviceItems && detail.serviceItems !== '[]' && (
              <div style={{ marginTop: 12, fontSize: 13, color: '#666' }}>
                增值服务：
                {(() => {
                  try {
                    return (JSON.parse(detail.serviceItems) as { name: string; price: string }[])
                      .map((s) => `${s.name} ¥${s.price}`)
                      .join('、')
                  } catch {
                    return '-'
                  }
                })()}
              </div>
            )}
          </>
        )}
      </Drawer>
      <Modal
        title={`退款 · ${refundTarget?.orderNo ?? ''}`}
        open={!!refundTarget}
        onOk={submitRefund}
        onCancel={() => setRefundTarget(undefined)}
        destroyOnClose
      >
        <Form form={refundForm} layout="vertical">
          <Form.Item name="reason" label="退款原因" rules={[{ required: true, message: '请填写退款原因' }]}>
            <Input placeholder="如：买家协商取消" />
          </Form.Item>
          <Form.Item name="amount" label="退款金额（元，留空为全额）">
            <InputNumber
              min={0.01}
              precision={2}
              style={{ width: '100%' }}
              addonAfter="元"
              placeholder={`默认全额 ¥${refundTarget?.payAmount ?? ''}`}
            />
          </Form.Item>
        </Form>
      </Modal>
      <Modal
        title={`发货 · ${shipTarget?.orderNo ?? ''}`}
        open={!!shipTarget}
        onOk={submitShip}
        onCancel={() => setShipTarget(undefined)}
        destroyOnClose
      >
        <Form form={shipForm} layout="vertical">
          <Form.Item name="shipNo" label="运单号" rules={[{ required: true, message: '请填写运单号' }]}>
            <Input placeholder="托运单号 / 专车运单号" maxLength={64} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
