import { useCallback, useEffect, useState } from 'react'
import {
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Select,
  Switch,
  Table,
  Tag,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import {
  deleteShipMethod,
  fetchShipMethods,
  ShipMethodRow,
  updateShipMethodStatus,
  upsertShipMethod,
} from '../api/admin'

const KIND_OPTIONS = [
  { value: 1, label: '到店自提' },
  { value: 2, label: '托运配送' },
]

const KIND_LABEL: Record<number, string> = { 1: '到店自提', 2: '托运配送' }

export default function ShipMethods() {
  const [list, setList] = useState<ShipMethodRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()
  const kind = Form.useWatch('kind', form)

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchShipMethods({ page: nextPage, pageSize: 10 })
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
    load(1)
  }, [load])

  const openEdit = (r?: ShipMethodRow) => {
    form.resetFields()
    if (r) {
      form.setFieldsValue({ ...r, status: r.status === 1 })
    } else {
      form.setFieldsValue({ status: true, sort: 0, kind: 2, fee: 0 })
    }
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertShipMethod({
        id: v.id,
        name: v.name,
        kind: v.kind,
        description: v.description || '',
        fee: v.fee != null ? String(v.fee) : '0',
        sort: v.sort ?? 0,
        status: v.status ? 1 : 0,
      })
      message.success('已保存')
      setModalOpen(false)
      load(page)
    } catch (e: any) {
      if (e?.message) message.error(e.message)
    }
  }

  const toggleStatus = async (r: ShipMethodRow) => {
    try {
      await updateShipMethodStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page)
    }
  }

  const remove = async (r: ShipMethodRow) => {
    try {
      await deleteShipMethod(r.id)
      message.success('已删除')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name', width: 130 },
    {
      title: '类型',
      dataIndex: 'kind',
      width: 100,
      render: (v: number) => <Tag color={v === 1 ? 'blue' : 'orange'}>{KIND_LABEL[v] || v}</Tag>,
    },
    {
      title: '运费',
      dataIndex: 'fee',
      width: 100,
      render: (v: string) => (Number(v) > 0 ? `¥${v}` : '免运费'),
    },
    { title: '说明', dataIndex: 'description', ellipsis: true, render: (v: string) => v || '-' },
    { title: '排序', dataIndex: 'sort', width: 70 },
    {
      title: '启用',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: ShipMethodRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 130,
      fixed: 'right' as const,
      render: (_: unknown, r: ShipMethodRow) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该配送方式？" onConfirm={() => remove(r)}>
            <Button size="small" type="link" danger>
              删除
            </Button>
          </Popconfirm>
        </>
      ),
    },
  ]

  return (
    <Card
      title="配送方式"
      extra={
        <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>
          新增方式
        </Button>
      }
    >
      <Table
        rowKey="id"
        loading={loading}
        size="small"
        columns={columns as never}
        dataSource={list}
        pagination={{
          current: page,
          total,
          pageSize: 10,
          showSizeChanger: false,
          onChange: (p) => load(p),
        }}
      />
      <Modal title="配送方式" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '必填' }]}>
            <Input placeholder="如：高铁托运" maxLength={32} showCount />
          </Form.Item>
          <Form.Item name="kind" label="类型" rules={[{ required: true, message: '必选' }]}>
            <Select options={KIND_OPTIONS} />
          </Form.Item>
          <Form.Item name="fee" label="运费（元）" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={0} precision={2} style={{ width: '100%' }} addonAfter="元" />
          </Form.Item>
          <Form.Item name="description" label="说明">
            <Input placeholder={kind === 1 ? '如：到店自提免运费，凭订单号取宠' : '如：专业宠物专车，全程视频可查'} maxLength={128} showCount />
          </Form.Item>
          <Form.Item name="sort" label="排序（越小越靠前）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
