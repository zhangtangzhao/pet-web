import { useCallback, useEffect, useState } from 'react'
import {
  Button,
  Card,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Switch,
  Table,
} from 'antd'
import { PlusOutlined, SearchOutlined } from '@ant-design/icons'
import {
  deleteSupplier,
  fetchSuppliers,
  SupplierRow,
  updateSupplierStatus,
  upsertSupplier,
} from '../api/admin'

export default function Suppliers() {
  const [list, setList] = useState<SupplierRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()

  const load = useCallback(
    async (nextPage: number, kw: string) => {
      setLoading(true)
      try {
        const resp = await fetchSuppliers({ page: nextPage, pageSize: 10, keyword: kw || undefined })
        setList(resp.list)
        setTotal(resp.total)
        setPage(nextPage)
      } catch (e: any) {
        message.error(e.message)
      } finally {
        setLoading(false)
      }
    },
    [],
  )

  useEffect(() => {
    load(1, '')
  }, [load])

  const openEdit = (r?: SupplierRow) => {
    form.resetFields()
    if (r) {
      form.setFieldsValue({ ...r, status: r.status === 1 })
    } else {
      form.setFieldsValue({ status: true })
    }
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertSupplier({
        id: v.id,
        name: v.name,
        contact: v.contact || '',
        phone: v.phone || '',
        address: v.address || '',
      })
      message.success('已保存')
      setModalOpen(false)
      load(page, keyword)
    } catch (e: any) {
      if (e?.message) message.error(e.message)
    }
  }

  const toggleStatus = async (r: SupplierRow) => {
    try {
      await updateSupplierStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page, keyword)
    }
  }

  const remove = async (r: SupplierRow) => {
    try {
      await deleteSupplier(r.id)
      message.success('已删除')
      load(page, keyword)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '名称', dataIndex: 'name', width: 160 },
    { title: '联系人', dataIndex: 'contact', width: 100, render: (v: string) => v || '-' },
    { title: '电话', dataIndex: 'phone', width: 130, render: (v: string) => v || '-' },
    { title: '地址', dataIndex: 'address', ellipsis: true, render: (v: string) => v || '-' },
    { title: '在售商品', dataIndex: 'productCount', width: 90 },
    {
      title: '启用',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: SupplierRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 130,
      fixed: 'right' as const,
      render: (_: unknown, r: SupplierRow) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r)}>
            编辑
          </Button>
          <Popconfirm title="关联商品存在时将无法删除，确定删除？" onConfirm={() => remove(r)}>
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
      title="供货商管理"
      extra={
        <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>
          新增供货商
        </Button>
      }
    >
      <Input
        allowClear
        placeholder="搜索名称/联系人/电话"
        prefix={<SearchOutlined />}
        style={{ width: 260, marginBottom: 12 }}
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        onPressEnter={() => load(1, keyword)}
        onClear={() => {
          setKeyword('')
          load(1, '')
        }}
      />
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
          onChange: (p) => load(p, keyword),
        }}
      />
      <Modal title="供货商" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '必填' }]}>
            <Input placeholder="如：XX犬舍" maxLength={64} showCount />
          </Form.Item>
          <Form.Item name="contact" label="联系人">
            <Input maxLength={32} />
          </Form.Item>
          <Form.Item name="phone" label="电话">
            <Input maxLength={20} />
          </Form.Item>
          <Form.Item name="address" label="地址">
            <Input maxLength={255} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
