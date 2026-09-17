import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Table, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { AuctionRow, fetchAuctions, upsertAuction } from '../api/play'
export default function Auction() {
  const [list, setList] = useState<AuctionRow[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm()
  const load = useCallback(() => { fetchAuctions().then(setList).catch((e: any) => message.error(e.message)) }, [])
  useEffect(() => { load() }, [load])
  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertAuction({ ...v, startAt: v.startAt?.format('YYYY-MM-DD HH:mm:ss'), endAt: v.endAt?.format('YYYY-MM-DD HH:mm:ss') })
      message.success('已保存'); setOpen(false); load()
    } catch (e: any) { if (e?.message) message.error(e.message) }
  }
  return (
    <Card title="竞拍管理" extra={<Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => { form.resetFields(); setOpen(true) }}>新增</Button>}>
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '商品', dataIndex: 'product_title', ellipsis: true },
        { title: '起拍价', dataIndex: 'start_price', width: 100 },
        { title: '加价', dataIndex: 'step_price', width: 80 },
        { title: '保证金', dataIndex: 'deposit_amount', width: 90 },
        { title: '最高价', dataIndex: 'highest_price', width: 100 },
        { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => { const m: Record<number, [string, string]> = { 0: ['default', '未开始'], 1: ['green', '进行中'], 2: ['blue', '已成交'], 3: ['red', '已流拍'] }; return <Tag color={m[v]?.[1]}>{m[v]?.[0]}</Tag> } },
        { title: '截止', dataIndex: 'end_at', width: 165 },
      ] as never} />
      <Modal title="竞拍活动" open={open} onOk={submit} onCancel={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="productId" label="商品 ID" rules={[{ required: true }]}><Input /></Form.Item>
          <Form.Item name="startPrice" label="起拍价" rules={[{ required: true }]}><InputNumber min={0.01} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="stepPrice" label="加价幅度" rules={[{ required: true }]}><InputNumber min={1} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="depositAmount" label="保证金"><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="startAt" label="开始时间" rules={[{ required: true }]}><Input placeholder="yyyy-MM-dd HH:mm:ss" /></Form.Item>
          <Form.Item name="endAt" label="截止时间" rules={[{ required: true }]}><Input placeholder="yyyy-MM-dd HH:mm:ss" /></Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
