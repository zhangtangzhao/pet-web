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
  Radio,
  Select,
  Space,
  Switch,
  Table,
  Tabs,
  Tag,
} from 'antd'
import { PlusOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import {
  CouponTemplate,
  MemberRow,
  ServiceItemRow,
  deleteCoupon,
  deleteService,
  fetchCoupons,
  fetchMembers,
  fetchServices,
  issueCoupon,
  upsertCoupon,
  upsertService,
} from '../api/admin'

const TYPE_OPTIONS = [
  { value: 1, label: '满减券' },
  { value: 2, label: '折扣券' },
  { value: 3, label: '立减券' },
]

const fmt = (s?: string) => (s ? s.replace('T', ' ').slice(0, 19) : '-')

interface CouponFormValues {
  id?: string
  name: string
  type: number
  thresholdAmount?: number
  discountAmount?: number
  discountPercent?: number
  maxDiscountAmount?: number
  totalCount?: number
  perLimit?: number
  newUserOnly?: boolean
  pickupRange?: [dayjs.Dayjs, dayjs.Dayjs]
  validRange?: [dayjs.Dayjs, dayjs.Dayjs]
  status?: boolean
}

interface ServiceFormValues {
  id?: string
  name: string
  description?: string
  originalPrice?: number
  price: number
  guaranteeDays?: number
  sort?: number
  status?: boolean
}

function CouponFaceText(t: CouponTemplate) {
  if (t.type === 2) return `${(t.discountPercent / 10).toFixed(1)} 折${Number(t.maxDiscountAmount) > 0 ? `（上限¥${t.maxDiscountAmount}）` : ''}`
  const cond = Number(t.thresholdAmount) > 0 ? `满${Number(t.thresholdAmount)}减` : '立减'
  return `${cond}¥${Number(t.discountAmount).toFixed(2).replace(/\.00$/, '')}`
}

export default function Marketing() {
  return (
    <Card>
      <Tabs
        items={[
          { key: 'coupons', label: '优惠券', children: <CouponTab /> },
          { key: 'services', label: '增值服务', children: <ServiceTab /> },
        ]}
      />
    </Card>
  )
}

// ─────────────────────────── 优惠券模板 ───────────────────────────

function CouponTab() {
  const [rows, setRows] = useState<CouponTemplate[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [modal, setModal] = useState(false)
  const [issueTarget, setIssueTarget] = useState<CouponTemplate>()
  const [form] = Form.useForm()

  const load = useCallback(
    async (p = page, kw = keyword) => {
      const resp = await fetchCoupons({ page: p, pageSize: 10, keyword: kw || undefined })
      setRows(resp.list)
      setTotal(resp.total)
      setPage(p)
    },
    [page, keyword],
  )

  useEffect(() => {
    load(1).catch((e) => message.error(e.message))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const openCreate = () => {
    form.setFieldsValue({
      id: undefined,
      name: '',
      type: 3,
      thresholdAmount: undefined,
      discountAmount: undefined,
      discountPercent: undefined,
      maxDiscountAmount: undefined,
      totalCount: 0,
      perLimit: 1,
      newUserOnly: false,
      pickupRange: undefined,
      validRange: undefined,
      status: true,
    })
    setModal(true)
  }

  const openEdit = (r: CouponTemplate) => {
    form.setFieldsValue({
      id: r.id,
      name: r.name,
      type: r.type,
      thresholdAmount: Number(r.thresholdAmount) || undefined,
      discountAmount: Number(r.discountAmount) || undefined,
      discountPercent: r.discountPercent || undefined,
      maxDiscountAmount: Number(r.maxDiscountAmount) || undefined,
      totalCount: r.totalCount,
      perLimit: r.perLimit,
      newUserOnly: r.newUserOnly === 1,
      pickupRange: r.pickupStart ? [dayjs(r.pickupStart), dayjs(r.pickupEnd)] : undefined,
      validRange: r.validStart ? [dayjs(r.validStart), dayjs(r.validEnd)] : undefined,
      status: r.status === 1,
    })
    setModal(true)
  }

  const submit = async () => {
    const v: CouponFormValues = await form.validateFields()
    const range = (key: 'pickup' | 'valid', r?: [dayjs.Dayjs, dayjs.Dayjs]) =>
      r ? { [`${key}Start`]: r[0].format('YYYY-MM-DD HH:mm:ss'), [`${key}End`]: r[1].format('YYYY-MM-DD HH:mm:ss') } : {}
    try {
      await upsertCoupon({
        id: v.id,
        name: v.name,
        type: v.type,
        thresholdAmount: v.thresholdAmount != null ? v.thresholdAmount.toFixed(2) : '0',
        discountAmount: v.discountAmount != null ? v.discountAmount.toFixed(2) : '0',
        discountPercent: v.discountPercent,
        maxDiscountAmount: v.maxDiscountAmount != null ? v.maxDiscountAmount.toFixed(2) : '0',
        totalCount: v.totalCount ?? 0,
        perLimit: v.perLimit ?? 1,
        newUserOnly: v.newUserOnly ? 1 : 0,
        ...range('pickup', v.pickupRange),
        ...range('valid', v.validRange),
        status: v.status === false ? 0 : 1,
      })
      message.success('已保存')
      setModal(false)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const remove = async (r: CouponTemplate) => {
    try {
      await deleteCoupon(r.id)
      message.success('已删除')
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const type = Form.useWatch('type', form)

  const columns: ColumnsType<CouponTemplate> = [
    { title: '名称', dataIndex: 'name', ellipsis: true },
    { title: '类型', dataIndex: 'typeText', width: 90 },
    { title: '面额', width: 180, render: (_, r) => <Tag color="orange">{CouponFaceText(r)}</Tag> },
    {
      title: '发放量',
      width: 110,
      render: (_, r) => (r.totalCount === 0 ? `${r.issuedCount} / 不限` : `${r.issuedCount} / ${r.totalCount}`),
    },
    { title: '限领', dataIndex: 'perLimit', width: 70 },
    {
      title: '有效期',
      width: 200,
      render: (_, r) => (
        <span style={{ fontSize: 12 }}>
          {r.validStart ? fmt(r.validStart).slice(0, 10) : '不限'} ~ {r.validEnd ? fmt(r.validEnd).slice(0, 10) : '不限'}
        </span>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s) => (s === 1 ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>),
    },
    { title: '更新时间', dataIndex: 'updatedAt', width: 150, render: fmt },
    {
      title: '操作',
      width: 190,
      render: (_, r) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEdit(r)}>编辑</Button>
          <Button type="link" size="small" onClick={() => setIssueTarget(r)}>发放</Button>
          <Popconfirm title="确定删除该优惠券模板？" onConfirm={() => remove(r)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Space>
          <Input
            allowClear
            prefix={<SearchOutlined />}
            placeholder="券名称"
            style={{ width: 220 }}
            onPressEnter={(e) => load(1, (e.target as HTMLInputElement).value)}
            onBlur={(e) => setKeyword(e.target.value)}
          />
          <Button icon={<ReloadOutlined />} onClick={() => load()} />
        </Space>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建优惠券</Button>
      </div>
      <Table
        rowKey="id"
        loading={!rows.length && total === 0}
        columns={columns as never}
        dataSource={rows}
        pagination={{
          total,
          current: page,
          pageSize: 10,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p) => load(p),
        }}
      />

      <Modal
        title={form.getFieldValue('id') ? '编辑优惠券' : '新建优惠券'}
        open={modal}
        onOk={submit}
        onCancel={() => setModal(false)}
        width={560}
        destroyOnClose
      >
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="id" hidden><Input /></Form.Item>
          <Form.Item name="name" label="券名称" rules={[{ required: true, message: '请输入券名称' }]}>
            <Input placeholder="如：满1000减100券" maxLength={64} />
          </Form.Item>
          <Form.Item name="type" label="类型" rules={[{ required: true }]}>
            <Radio.Group options={TYPE_OPTIONS} optionType="button" />
          </Form.Item>
          {type === 1 && (
            <>
              <Form.Item name="thresholdAmount" label="使用门槛（元）" rules={[{ required: true, message: '请输入门槛金额' }]}>
                <InputNumber style={{ width: 200 }} min={0.01} precision={2} placeholder="如 1000" />
              </Form.Item>
              <Form.Item name="discountAmount" label="抵扣金额（元）" rules={[{ required: true, message: '请输入抵扣金额' }]}>
                <InputNumber style={{ width: 200 }} min={0.01} precision={2} placeholder="如 100" />
              </Form.Item>
            </>
          )}
          {type === 2 && (
            <>
              <Form.Item name="discountPercent" label="折扣（90 = 9 折）" rules={[{ required: true, message: '请输入折扣' }]}>
                <InputNumber style={{ width: 200 }} min={1} max={99} precision={0} placeholder="如 90" />
              </Form.Item>
              <Form.Item name="maxDiscountAmount" label="最高抵扣（元，0 = 不封顶）">
                <InputNumber style={{ width: 200 }} min={0} precision={2} placeholder="如 200" />
              </Form.Item>
            </>
          )}
          {type === 3 && (
            <Form.Item name="discountAmount" label="立减金额（元）" rules={[{ required: true, message: '请输入立减金额' }]}>
              <InputNumber style={{ width: 200 }} min={0.01} precision={2} placeholder="如 30" />
            </Form.Item>
          )}
          <Space size="large">
            <Form.Item name="totalCount" label="发放总量（0 = 不限）">
              <InputNumber style={{ width: 140 }} min={0} precision={0} />
            </Form.Item>
            <Form.Item name="perLimit" label="每人限领">
              <InputNumber style={{ width: 120 }} min={1} precision={0} />
            </Form.Item>
          </Space>
          <Form.Item name="newUserOnly" label="新用户注册自动赠送" valuePropName="checked">
            <Switch />
          </Form.Item>
          <Form.Item name="pickupRange" label="可领取时段（不选 = 不限）">
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="validRange" label="可用时段（不选 = 不限）">
            <DatePicker.RangePicker showTime style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="status" label="启用" valuePropName="checked">
            <Switch />
          </Form.Item>
        </Form>
      </Modal>

      <IssueModal target={issueTarget} onClose={() => setIssueTarget(undefined)} />
    </>
  )
}

// ─────────────────────────── 定向发放 ───────────────────────────

function IssueModal({ target, onClose }: { target?: CouponTemplate; onClose: () => void }) {
  const [rows, setRows] = useState<MemberRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [selected, setSelected] = useState<string[]>([])

  const load = useCallback(
    async (p = page, kw = keyword) => {
      const resp = await fetchMembers({ page: p, pageSize: 8, keyword: kw || undefined })
      setRows(resp.list)
      setTotal(resp.total)
      setPage(p)
    },
    [page, keyword],
  )

  useEffect(() => {
    if (target) {
      setSelected([])
      load(1).catch((e) => message.error(e.message))
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [target])

  const submit = async () => {
    if (!target || selected.length === 0) return
    try {
      const r = await issueCoupon(target.id, selected)
      message.success(`已发放 ${r.issued} 张${r.failed > 0 ? `，${r.failed} 人因限领/库存跳过` : ''}`)
      onClose()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  return (
    <Modal
      title={`定向发放：${target?.name ?? ''}`}
      open={!!target}
      onOk={submit}
      okButtonProps={{ disabled: selected.length === 0 }}
      onCancel={onClose}
      width={640}
    >
      <Input.Search
        allowClear
        placeholder="昵称 / 手机号搜索"
        style={{ marginBottom: 12 }}
        onSearch={(v) => {
          setKeyword(v)
          load(1, v)
        }}
      />
      <Table
        rowKey="id"
        size="small"
        dataSource={rows}
        columns={[
          { title: '昵称', dataIndex: 'nickname', ellipsis: true },
          { title: '手机号', dataIndex: 'phone', width: 130 },
          { title: '注册时间', dataIndex: 'createdAt', width: 110, render: (v: string) => fmt(v).slice(0, 10) },
        ] as never}
        rowSelection={{ selectedRowKeys: selected, onChange: (keys) => setSelected(keys as string[]) }}
        pagination={{
          total,
          current: page,
          pageSize: 8,
          size: 'small',
          onChange: (p) => load(p),
        }}
      />
    </Modal>
  )
}

// ─────────────────────────── 增值服务 ───────────────────────────

function ServiceTab() {
  const [rows, setRows] = useState<ServiceItemRow[]>([])
  const [modal, setModal] = useState(false)
  const [form] = Form.useForm()

  const load = useCallback(
    () => fetchServices().then(setRows).catch((e) => message.error(e.message)),
    [],
  )

  useEffect(() => {
    load()
  }, [load])

  const openCreate = () => {
    form.setFieldsValue({ id: undefined, name: '', description: '', originalPrice: undefined, price: undefined, guaranteeDays: 0, sort: 0, status: true })
    setModal(true)
  }

  const openEdit = (r: ServiceItemRow) => {
    form.setFieldsValue({
      id: r.id,
      name: r.name,
      description: r.description,
      originalPrice: Number(r.originalPrice) || undefined,
      price: Number(r.price),
      guaranteeDays: r.guaranteeDays ?? 0,
      sort: r.sort ?? 0,
      status: r.status === 1,
    })
    setModal(true)
  }

  const submit = async () => {
    const v: ServiceFormValues = await form.validateFields()
    try {
      await upsertService({
        id: v.id,
        name: v.name,
        description: v.description,
        originalPrice: v.originalPrice?.toFixed(2),
        price: v.price.toFixed(2),
        guaranteeDays: v.guaranteeDays ?? 0,
        sort: v.sort,
        status: v.status === false ? 0 : 1,
      })
      message.success('已保存')
      setModal(false)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const remove = async (r: ServiceItemRow) => {
    try {
      await deleteService(r.id)
      message.success('已删除')
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns: ColumnsType<ServiceItemRow> = [
    { title: '服务名称', dataIndex: 'name', width: 180 },
    { title: '说明', dataIndex: 'description', ellipsis: true },
    {
      title: '价格',
      width: 200,
      render: (_, r) => (
        <span>
          <span style={{ color: '#fa541c', fontWeight: 600 }}>¥{r.price}</span>
          {Number(r.originalPrice) > Number(r.price) && (
            <span style={{ color: '#bbb', textDecoration: 'line-through', marginLeft: 8 }}>¥{r.originalPrice}</span>
          )}
        </span>
      ),
    },
    {
      title: '健康保障',
      dataIndex: 'guaranteeDays',
      width: 90,
      render: (_: unknown, r: ServiceItemRow) => (r.guaranteeDays && r.guaranteeDays > 0 ? `${r.guaranteeDays} 天` : '-'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (s) => (s === 1 ? <Tag color="green">启用</Tag> : <Tag>停用</Tag>),
    },
    {
      title: '操作',
      width: 140,
      render: (_, r) => (
        <Space>
          <Button type="link" size="small" onClick={() => openEdit(r)}>编辑</Button>
          <Popconfirm title="确定删除该服务项？" onConfirm={() => remove(r)}>
            <Button type="link" size="small" danger>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <>
      <div style={{ marginBottom: 16, textAlign: 'right' }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新建服务</Button>
      </div>
      <Table rowKey="id" columns={columns as never} dataSource={rows} pagination={false} />
      <Modal title={form.getFieldValue('id') ? '编辑服务' : '新建服务'} open={modal} onOk={submit} onCancel={() => setModal(false)} destroyOnClose>
        <Form form={form} layout="vertical" preserve={false}>
          <Form.Item name="id" hidden><Input /></Form.Item>
          <Form.Item name="name" label="服务名称" rules={[{ required: true, message: '请输入服务名称' }]}>
            <Input placeholder="如：专业托运" maxLength={64} />
          </Form.Item>
          <Form.Item name="description" label="服务说明">
            <Input.TextArea rows={2} maxLength={255} placeholder="展示在订单加购列表" />
          </Form.Item>
          <Space size="large">
            <Form.Item name="originalPrice" label="原价（划线展示）">
              <InputNumber style={{ width: 160 }} min={0} precision={2} placeholder="如 300" />
            </Form.Item>
            <Form.Item name="price" label="服务价（下单计费）" rules={[{ required: true, message: '请输入服务价' }]}>
              <InputNumber style={{ width: 160 }} min={0.01} precision={2} placeholder="如 200" />
            </Form.Item>
            <Form.Item name="guaranteeDays" label="健康保障天数（0=无）">
              <InputNumber style={{ width: 160 }} min={0} max={365} precision={0} placeholder="如 30" />
            </Form.Item>
          </Space>
          <Space size="large">
            <Form.Item name="sort" label="排序（小在前）">
              <InputNumber style={{ width: 120 }} min={0} precision={0} />
            </Form.Item>
            <Form.Item name="status" label="启用" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </>
  )
}
