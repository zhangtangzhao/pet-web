import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Select, Table, Tabs, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { PointsOrderRow, PointsProductRow, fetchPointsOrders, fetchPointsProducts, shipPointsOrder, upsertPointsProduct } from '../api/engagement'

function Products() {
  const [list, setList] = useState<PointsProductRow[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()
  const load = useCallback(() => { fetchPointsProducts().then(setList).catch((e: any) => message.error(e.message)) }, [])
  useEffect(() => { load() }, [load])
  const openEdit = (r?: PointsProductRow) => {
    form.resetFields()
    if (r) form.setFieldsValue({ ...r })
    else form.setFieldsValue({ type: 3, stock: 9999, status: 1 })
    setOpen(true)
  }
  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertPointsProduct(v)
      message.success('已保存'); setOpen(false); load()
    } catch (e: any) { if (e?.message) message.error(e.message) }
  }
  return (
    <>
      <Button size="small" type="primary" icon={<PlusOutlined />} style={{ marginBottom: 12 }} onClick={() => openEdit()}>新增商品</Button>
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '商品', dataIndex: 'name' },
        { title: '类型', dataIndex: 'typeText', width: 100, render: (v: string) => <Tag>{v}</Tag> },
        { title: '所需积分', dataIndex: 'pointsCost', width: 100 },
        { title: '库存', dataIndex: 'stock', width: 90 },
        { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => v === 1 ? <Tag color="green">启用</Tag> : <Tag>停用</Tag> },
        { title: '操作', width: 90, render: (_: unknown, r: PointsProductRow) => <Button size="small" type="link" onClick={() => openEdit(r)}>编辑</Button> },
      ] as never} />
      <Modal title="积分商品" open={open} onOk={submit} onCancel={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden><Input /></Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '必填' }]}><Input maxLength={64} /></Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Select options={[{ value: 1, label: '优惠券' }, { value: 2, label: '实物' }, { value: 3, label: '免运费卡' }]} />
          </Form.Item>
          <Form.Item noStyle shouldUpdate={(a, b) => a.type !== b.type}>
            {({ getFieldValue }) => getFieldValue('type') === 1 ? (
              <Form.Item name="couponTemplateId" label="券模板 ID"><Input placeholder="如 5111" /></Form.Item>
            ) : null}
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="pointsCost" label="所需积分" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="stock" label="库存" rules={[{ required: true }]}><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
          </div>
          <Form.Item name="description" label="说明"><Input maxLength={255} /></Form.Item>
          <Form.Item name="status" label="启用"><Select options={[{ value: 1, label: '启用' }, { value: 0, label: '停用' }]} /></Form.Item>
        </Form>
      </Modal>
    </>
  )
}

function Orders() {
  const [list, setList] = useState<PointsOrderRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [shipTarget, setShipTarget] = useState<PointsOrderRow>()
  const [shipNo, setShipNo] = useState('')
  const load = useCallback(async (p = 1) => {
    try { const r = await fetchPointsOrders({ page: p, pageSize: 15 }); setList(r.list); setTotal(r.total); setPage(p) }
    catch (e: any) { message.error(e.message) }
  }, [])
  useEffect(() => { load() }, [load])
  const doShip = async () => {
    if (!shipTarget) return
    try { await shipPointsOrder(shipTarget.id, shipNo.trim() || 'MANUAL'); message.success('已发货'); setShipTarget(undefined); setShipNo(''); load(page) }
    catch (e: any) { message.error(e.message) }
  }
  return (
    <>
    <Table rowKey="id" size="small" dataSource={list} pagination={{ current: page, total, pageSize: 15, onChange: load }} columns={[
      { title: '单号', dataIndex: 'orderNo', width: 190 },
      { title: '商品', dataIndex: 'productName' },
      { title: '积分', dataIndex: 'pointsCost', width: 90 },
      { title: '状态', dataIndex: 'statusText', width: 90 },
      { title: '收货', width: 220, render: (_: unknown, r: PointsOrderRow) => `${r.contact} ${r.phone} ${r.address}`.trim() || '-' },
      { title: '运单号', dataIndex: 'shipNo', width: 130 },
    ] as never} />
      <Modal title={`发货 · ${shipTarget?.orderNo ?? ''}`} open={!!shipTarget} onOk={doShip} onCancel={() => setShipTarget(undefined)}>
        <Input placeholder="运单号" value={shipNo} onChange={(e) => setShipNo(e.target.value)} />
      </Modal>
    </>
  )
}

export default function PointsShop() {
  return <Card title="积分商城"><Tabs items={[{ key: 'products', label: '商品管理', children: <Products /> }, { key: 'orders', label: '兑换订单', children: <Orders /> }]} /></Card>
}

