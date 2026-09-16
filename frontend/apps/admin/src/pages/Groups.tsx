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
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import {
  AdminProduct,
  deleteGroupBuy,
  fetchGroupBuys,
  fetchProducts,
  GroupBuyRow,
  updateGroupBuyStatus,
  upsertGroupBuy,
} from '../api/admin'

interface FormValues {
  id?: string
  productId: string
  price: string
  size: number
  hours: number
  status: boolean
}

export default function Groups() {
  const [list, setList] = useState<GroupBuyRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [productOpts, setProductOpts] = useState<AdminProduct[]>([])
  const [form] = Form.useForm()

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchGroupBuys({ page: nextPage, pageSize: 10 })
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

  const searchProducts = async (keyword?: string) => {
    try {
      const resp = await fetchProducts({ page: 1, pageSize: 20, keyword })
      setProductOpts(resp.list)
    } catch {
      setProductOpts([])
    }
  }

  const openEdit = async (r?: GroupBuyRow) => {
    form.resetFields()
    if (r) {
      form.setFieldsValue({
        id: r.id,
        productId: r.productId,
        price: r.price,
        size: r.size,
        hours: r.hours,
        status: r.status === 1,
      })
      await searchProducts()
      if (!productOpts.some((p) => p.id === r.productId)) {
        setProductOpts((prev) => [{ id: r.productId, title: r.productTitle } as AdminProduct, ...prev])
      }
    } else {
      form.setFieldsValue({ status: true, size: 2, hours: 24 })
      await searchProducts()
    }
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const v = (await form.validateFields()) as FormValues
      await upsertGroupBuy({
        id: v.id,
        productId: v.productId,
        price: String(v.price),
        size: v.size,
        hours: v.hours,
        status: v.status ? 1 : 0,
      })
      message.success('已保存')
      setModalOpen(false)
      load(page)
    } catch (e: any) {
      if (e?.message) message.error(e.message)
    }
  }

  const toggleStatus = async (r: GroupBuyRow) => {
    try {
      await updateGroupBuyStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page)
    }
  }

  const remove = async (r: GroupBuyRow) => {
    try {
      await deleteGroupBuy(r.id)
      message.success('已删除')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '商品', dataIndex: 'productTitle', ellipsis: true },
    {
      title: '拼团价',
      dataIndex: 'price',
      width: 100,
      render: (v: string) => <span style={{ color: '#f5222d', fontWeight: 600 }}>¥{v}</span>,
    },
    { title: '成团人数', dataIndex: 'size', width: 90, render: (v: number) => `${v} 人团` },
    { title: '成团时限', dataIndex: 'hours', width: 100, render: (v: number) => `${v} 小时` },
    {
      title: '启用',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: GroupBuyRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 165 },
    {
      title: '操作',
      width: 130,
      fixed: 'right' as const,
      render: (_: unknown, r: GroupBuyRow) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该活动？" onConfirm={() => remove(r)}>
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
      title="拼团活动"
      extra={
        <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>
          新增活动
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
      <Modal title="拼团活动" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="productId" label="商品" rules={[{ required: true, message: '必选' }]}>
            <Select
              showSearch
              filterOption={false}
              placeholder="搜索商品标题"
              onSearch={(kw) => searchProducts(kw)}
              options={productOpts.map((p) => ({ value: p.id, label: `${p.title}（¥${p.price}）` }))}
            />
          </Form.Item>
          <Form.Item name="price" label="拼团价" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={0.01} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="size" label="成团人数（2-10）" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={2} max={10} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="hours" label="成团时限（小时，1-72，超时未成团自动退款）" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={1} max={72} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
