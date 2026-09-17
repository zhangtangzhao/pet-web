import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Popconfirm, Switch, Table } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { StoreRow, deleteStore, fetchStores, updateStoreStatus, upsertStore } from '../api/engagement'

interface FormValues { id?: string; name: string; address: string; phone?: string; businessHours?: string; sort?: number; status: boolean }

export default function Stores() {
  const [list, setList] = useState<StoreRow[]>([])
  const [open, setOpen] = useState(false)
  const [form] = Form.useForm<FormValues>()

  const load = useCallback(() => { fetchStores().then(setList).catch((e: any) => message.error(e.message)) }, [])
  useEffect(() => { load() }, [load])

  const openEdit = (r?: StoreRow) => {
    form.resetFields()
    if (r) form.setFieldsValue({ ...r, status: r.status === 1 })
    else form.setFieldsValue({ status: true, sort: 0 })
    setOpen(true)
  }

  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertStore({ ...v, status: v.status ? 1 : 0, sort: v.sort ?? 0 })
      message.success('已保存'); setOpen(false); load()
    } catch (e: any) { if (e?.message) message.error(e.message) }
  }

  const toggle = async (r: StoreRow) => {
    try { await updateStoreStatus(r.id, r.status === 1 ? 0 : 1); setList((p) => p.map((x) => x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)) }
    catch (e: any) { message.error(e.message); load() }
  }

  const remove = async (r: StoreRow) => {
    try { await deleteStore(r.id); message.success('已删除'); load() } catch (e: any) { message.error(e.message) }
  }

  return (
    <Card title="自提门店" extra={<Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>新增门店</Button>}>
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '门店', dataIndex: 'name' },
        { title: '地址', dataIndex: 'address', ellipsis: true },
        { title: '电话', dataIndex: 'phone', width: 140 },
        { title: '营业时间', dataIndex: 'businessHours', width: 150 },
        { title: '启用', dataIndex: 'status', width: 80, render: (_: unknown, r: StoreRow) => <Switch size="small" checked={r.status === 1} onChange={() => toggle(r)} /> },
        { title: '操作', width: 130, render: (_: unknown, r: StoreRow) => (<>
          <Button size="small" type="link" onClick={() => openEdit(r)}>编辑</Button>
          <Popconfirm title="确定删除该门店？" onConfirm={() => remove(r)}><Button size="small" type="link" danger>删除</Button></Popconfirm>
        </>) },
      ] as never} />
      <Modal title="自提门店" open={open} onOk={submit} onCancel={() => setOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden><Input /></Form.Item>
          <Form.Item name="name" label="门店名称" rules={[{ required: true, message: '必填' }]}><Input maxLength={64} /></Form.Item>
          <Form.Item name="address" label="地址" rules={[{ required: true, message: '必填' }]}><Input.TextArea rows={2} /></Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="phone" label="电话"><Input /></Form.Item>
            <Form.Item name="businessHours" label="营业时间"><Input placeholder="如 09:00-21:00" /></Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="sort" label="排序"><InputNumber min={0} style={{ width: '100%' }} /></Form.Item>
            <Form.Item name="status" label="启用" valuePropName="checked"><Switch /></Form.Item>
          </div>
        </Form>
      </Modal>
    </Card>
  )
}
