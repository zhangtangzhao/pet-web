import { useCallback, useEffect, useState } from 'react'
import {
  Button,
  Card,
  DatePicker,
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
import dayjs, { Dayjs } from 'dayjs'
import {
  AdminProduct,
  deleteFlashSale,
  fetchFlashSales,
  fetchProducts,
  FlashSaleRow,
  updateFlashSaleStatus,
  upsertFlashSale,
} from '../api/admin'

interface FormValues {
  id?: string
  productId: string
  salePrice: string
  stock: number
  range: [Dayjs, Dayjs]
  status: boolean
}

export default function FlashSales() {
  const [list, setList] = useState<FlashSaleRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [productOpts, setProductOpts] = useState<AdminProduct[]>([])
  const [form] = Form.useForm()

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchFlashSales({ page: nextPage, pageSize: 10 })
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

  const openEdit = async (r?: FlashSaleRow) => {
    form.resetFields()
    if (r) {
      form.setFieldsValue({
        id: r.id,
        productId: r.productId,
        salePrice: r.salePrice,
        stock: r.stock,
        range: [dayjs(r.startAt), dayjs(r.endAt)],
        status: r.status === 1,
      })
      await searchProducts()
      if (!productOpts.some((p) => p.id === r.productId)) {
        setProductOpts((prev) => [
          { id: r.productId, title: r.productTitle } as AdminProduct,
          ...prev,
        ])
      }
    } else {
      form.setFieldsValue({ status: true, stock: 10 })
      await searchProducts()
    }
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const v = (await form.validateFields()) as FormValues
      await upsertFlashSale({
        id: v.id,
        productId: v.productId,
        salePrice: String(v.salePrice),
        stock: v.stock,
        startAt: v.range[0].format('YYYY-MM-DD HH:mm:ss'),
        endAt: v.range[1].format('YYYY-MM-DD HH:mm:ss'),
        status: v.status ? 1 : 0,
      })
      message.success('已保存')
      setModalOpen(false)
      load(page)
    } catch (e: any) {
      if (e?.message) message.error(e.message)
    }
  }

  const toggleStatus = async (r: FlashSaleRow) => {
    try {
      await updateFlashSaleStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page)
    }
  }

  const remove = async (r: FlashSaleRow) => {
    try {
      await deleteFlashSale(r.id)
      message.success('已删除')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '商品', dataIndex: 'productTitle', ellipsis: true },
    {
      title: '秒杀价',
      dataIndex: 'salePrice',
      width: 100,
      render: (v: string) => <span style={{ color: '#f5222d', fontWeight: 600 }}>¥{v}</span>,
    },
    {
      title: '销量/名额',
      width: 110,
      render: (_: unknown, r: FlashSaleRow) => `${r.sold}/${r.stock}`,
    },
    {
      title: '活动时间',
      width: 300,
      render: (_: unknown, r: FlashSaleRow) => (
        <span style={{ fontSize: 12 }}>
          {r.startAt?.replace('T', ' ').slice(0, 16)} ~ {r.endAt?.replace('T', ' ').slice(0, 16)}
        </span>
      ),
    },
    {
      title: '启用',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: FlashSaleRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 130,
      fixed: 'right' as const,
      render: (_: unknown, r: FlashSaleRow) => (
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
      title="秒杀活动"
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
      <Modal title="秒杀活动" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
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
          <Form.Item name="salePrice" label="秒杀价（不高于商品原价）" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={0.01} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="stock" label="活动名额" rules={[{ required: true, message: '必填' }]}>
            <InputNumber min={1} max={9999} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="range" label="活动时间" rules={[{ required: true, message: '必选' }]}>
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
