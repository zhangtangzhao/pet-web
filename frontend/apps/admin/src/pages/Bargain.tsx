import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Select, Table, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { BargainRow, fetchBargains, upsertBargain } from '../api/play'
import { fetchProducts } from '../api/admin'
export default function Bargain() {
  const [list, setList] = useState<BargainRow[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()
  const load = useCallback(() => { fetchBargains().then(setList).catch((e: any) => message.error(e.message)) }, [])
  useEffect(() => { load() }, [load])
  const submit = async () => {
    try { const v = await form.validateFields(); await upsertBargain(v); message.success('已保存'); setOpen(false); load() } catch (e: any) { if (e?.message) message.error(e.message) }
  }
  return (
    <Card title="砍价活动" extra={<Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setOpen(true) }}>新增</Button>}>
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '商品', dataIndex: 'product_title', ellipsis: true },
        { title: '底价', dataIndex: 'bottom_price', width: 100, render: (v: string) => <span style={{ color: '#f5222d', fontWeight: 600 }}>¥{v}</span> },
        { title: '时限(h)', dataIndex: 'duration_hours', width: 80 },
        { title: '助力人数', dataIndex: 'max_helpers', width: 90 },
        { title: '状态', dataIndex: 'status', width: 80, render: (v: number) => v === 1 ? <Tag color="green">启用</Tag> : <Tag>停用</Tag> },
      ] as never} />
      <Modal title="砍价活动" open={open} onOk={submit} onCancel={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="productId" label="商品 ID" rules={[{ required: true }]}><Input placeholder="商品 ID" /></Form.Item>
          <Form.Item name="bottomPrice" label="底价" rules={[{ required: true }]}><InputNumber min={0.01} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="durationHours" label="砍价时限（小时）"><InputNumber min={1} max={72} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="maxHelpers" label="助力人数"><InputNumber min={2} max={10} style={{ width: '100%' }} /></Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
