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
import { BannerRow, deleteBanner, fetchBanners, updateBannerStatus, upsertBanner } from '../api/admin'

const JUMP_TYPES: { value: string; label: string }[] = [
  { value: 'me', label: '我的' },
  { value: 'recommend', label: '智能选宠' },
  { value: 'coupon', label: '领券中心' },
  { value: 'orders', label: '我的订单' },
  { value: 'notify', label: '消息中心' },
  { value: 'favorites', label: '收藏' },
  { value: 'product', label: '商品详情' },
  { value: 'category', label: '商品分类' },
  { value: 'search', label: '搜索关键词' },
  { value: 'custom', label: '自定义页面' },
]

const JUMP_LABEL: Record<string, string> = Object.fromEntries(JUMP_TYPES.map((t) => [t.value, t.label]))

// 需要 target 参数的跳转类型 → 输入提示
const TARGET_PLACEHOLDER: Record<string, string> = {
  product: '商品 ID（如 3001）',
  category: '分类 ID',
  search: '搜索关键词',
  custom: '小程序页面路径（/pages/xxx/index）',
}

export default function Banners() {
  const [list, setList] = useState<BannerRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [modalOpen, setModalOpen] = useState(false)
  const [form] = Form.useForm()
  const jumpType = Form.useWatch('jumpType', form)
  const needTarget = jumpType === 'product' || jumpType === 'category' || jumpType === 'search' || jumpType === 'custom'

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchBanners({ page: nextPage, pageSize: 10 })
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

  const openEdit = (r?: BannerRow) => {
    form.resetFields()
    if (r) {
      form.setFieldsValue({ ...r, status: r.status === 1 })
    } else {
      form.setFieldsValue({ status: true, sort: 0, jumpType: 'me' })
    }
    setModalOpen(true)
  }

  const submit = async () => {
    try {
      const v = await form.validateFields()
      await upsertBanner({
        id: v.id,
        title: v.title,
        subTitle: v.subTitle || '',
        icon: v.icon || '',
        jumpType: v.jumpType,
        target: v.target || '',
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

  const toggleStatus = async (r: BannerRow) => {
    try {
      await updateBannerStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page)
    }
  }

  const remove = async (r: BannerRow) => {
    try {
      await deleteBanner(r.id)
      message.success('已删除')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '标题', dataIndex: 'title', width: 140 },
    { title: '副标题', dataIndex: 'subTitle', ellipsis: true, render: (v: string) => v || '-' },
    { title: '图标', dataIndex: 'icon', width: 70, render: (v: string) => (v?.startsWith('http') ? '图片' : v || '-') },
    {
      title: '跳转类型',
      dataIndex: 'jumpType',
      width: 110,
      render: (v: string) => <Tag color="orange">{JUMP_LABEL[v] || v}</Tag>,
    },
    { title: '跳转参数', dataIndex: 'target', ellipsis: true, render: (v: string) => v || '-' },
    { title: '排序', dataIndex: 'sort', width: 70 },
    {
      title: '上架',
      dataIndex: 'status',
      width: 80,
      render: (_: unknown, r: BannerRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '创建时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 130,
      fixed: 'right' as const,
      render: (_: unknown, r: BannerRow) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r)}>
            编辑
          </Button>
          <Popconfirm title="确定删除该 banner？" onConfirm={() => remove(r)}>
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
      title="首页 Banner 运营位"
      extra={
        <Button size="small" type="primary" icon={<PlusOutlined />} onClick={() => openEdit()}>
          新增 Banner
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
      <Modal title="Banner" open={modalOpen} onOk={submit} onCancel={() => setModalOpen(false)} destroyOnClose>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '必填' }]}>
            <Input placeholder="如：领券中心" maxLength={32} showCount />
          </Form.Item>
          <Form.Item name="subTitle" label="副标题">
            <Input placeholder="如：新人立省 30 元" maxLength={64} />
          </Form.Item>
          <Form.Item name="icon" label="图标（emoji 或图片 URL）">
            <Input placeholder="如 🎫 或 https://..." maxLength={255} />
          </Form.Item>
          <Form.Item name="jumpType" label="跳转类型" rules={[{ required: true, message: '必选' }]}>
            <Select options={JUMP_TYPES} />
          </Form.Item>
          {needTarget && (
            <Form.Item name="target" label="跳转参数" rules={[{ required: true, message: '必填' }]}>
              <Input placeholder={TARGET_PLACEHOLDER[jumpType] || ''} maxLength={255} />
            </Form.Item>
          )}
          <Form.Item name="sort" label="排序（越小越靠前）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="上架" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>
    </Card>
  )
}
