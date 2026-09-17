import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, InputNumber, message, Modal, Popconfirm, Select, Table, Tag } from 'antd'
import { PageResp } from '../api/client'
import {
  AdminProduct,
  AdminProductDetail,
  BreedItem,
  CategoryItem,
  createProduct,
  fetchBreeds,
  fetchCategories,
  fetchProductDetail,
  fetchProducts,
  fetchSuppliers,
  SupplierRow,
  updateProduct,
  updateProductStatus,
} from '../api/admin'

const STATUS_TAG: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'green', text: '在售' },
  2: { color: 'orange', text: '下架' },
  3: { color: 'blue', text: '锁定中' },
  4: { color: 'red', text: '已售出' },
}

interface EditForm {
  title: string
  categoryId: string
  breedId: string
  price: string
  originalPrice?: string
  petGender: number
  birthDate?: string
  vaccineDesc?: string
  dewormDesc?: string
  bodyType?: string
  coatColor?: string
  personality?: string
  healthDesc?: string
  mainImage: string
  images?: string[]
  videoUrl?: string
  videoCover?: string
  detailHtml?: string
  supplierId?: string
  quarantineCertUrl?: string
  nextVaccineDate?: string
  nextDewormDate?: string
  detailImages?: string
  stockWarnThreshold?: number
  skus?: { specs: string; price: string; sort?: number; status?: number }[]
}

export default function Products() {
  const [data, setData] = useState<PageResp<AdminProduct>>({ total: 0, list: [] })
  const [query, setQuery] = useState({ page: 1, pageSize: 10, status: -1, keyword: '' })
  const [loading, setLoading] = useState(false)
  const [open, setOpen] = useState(false)
  const [editId, setEditId] = useState<string>()
  const [form] = Form.useForm<EditForm>()
  const [cats, setCats] = useState<CategoryItem[]>([])
  const [breeds, setBreeds] = useState<BreedItem[]>([])
  const [suppliers, setSuppliers] = useState<SupplierRow[]>([])

  const load = useCallback(
    async (q = query) => {
      setLoading(true)
      try {
        setData(await fetchProducts(q))
      } catch (e: any) {
        message.error(e.message)
      } finally {
        setLoading(false)
      }
    },
    [query],
  )

  useEffect(() => {
    load()
  }, [query])

  useEffect(() => {
    fetchCategories().then(setCats).catch(() => {})
    fetchSuppliers({ page: 1, pageSize: 100 }).then((r) => setSuppliers(r.list)).catch(() => {})
  }, [])

  const openEdit = async (id?: string) => {
    setEditId(id)
    form.resetFields()
    if (id) {
      const d: AdminProductDetail = await fetchProductDetail(id)
      const rows = d as unknown as EditForm & { detailImages?: string[]; skus?: { specs: string; price: string; sort?: number; status?: number }[] }
      form.setFieldsValue({
        ...(rows as unknown as EditForm),
        detailImages: (rows.detailImages ?? []).join('\n'),
        skus: rows.skus ?? [],
      })
    }
    setOpen(true)
  }

  const onCategoryChange = (cid: string) => {
    form.setFieldValue('breedId', undefined)
    fetchBreeds(cid).then(setBreeds).catch(() => {})
  }

  const onSubmit = async () => {
    const raw = await form.validateFields()
    const v = {
      ...raw,
      price: String(raw.price ?? ''),
      originalPrice: raw.originalPrice != null ? String(raw.originalPrice) : '',
      images:
        typeof raw.images === 'string' && (raw.images as string).trim()
          ? (raw.images as unknown as string)
              .split(/[,，\n]/)
              .map((s) => s.trim())
              .filter(Boolean)
          : [],
      detailImages:
        typeof raw.detailImages === 'string' && (raw.detailImages as string).trim()
          ? (raw.detailImages as unknown as string).split('\n').map((s) => s.trim()).filter(Boolean)
          : [],
      stockWarnThreshold: raw.stockWarnThreshold != null ? String(raw.stockWarnThreshold) : '',
      skus: (raw.skus ?? []).map((s) => ({ specs: s.specs, price: String(s.price), sort: s.sort, status: s.status })),
    }
    try {
      if (editId) await updateProduct(editId, v)
      else await createProduct(v)
      message.success(editId ? '已保存' : '已创建（草稿，上架后可见）')
      setOpen(false)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const toggleStatus = async (row: AdminProduct) => {
    try {
      await updateProductStatus(row.id, row.status === 1 ? 2 : 1)
      message.success('状态已更新')
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    {
      title: 'SPU',
      dataIndex: 'spuNo',
      width: 150,
    },
    {
      title: '标题',
      dataIndex: 'title',
      ellipsis: true,
    },
    { title: '品种', dataIndex: 'breedName', width: 110 },
    {
      title: '价格',
      dataIndex: 'price',
      width: 110,
      render: (v: string, r: AdminProduct) => (
        <>
          ¥{v}
          {Number(r.originalPrice) > Number(r.price) && (
            <span style={{ color: '#999', textDecoration: 'line-through', marginLeft: 4, fontSize: 12 }}>
              ¥{r.originalPrice}
            </span>
          )}
        </>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v: number) => <Tag color={STATUS_TAG[v]?.color}>{STATUS_TAG[v]?.text ?? v}</Tag>,
    },
    { title: '销量', dataIndex: 'sales', width: 70 },
    { title: '浏览', dataIndex: 'viewCount', width: 70 },
    { title: '收藏', dataIndex: 'favoriteCount', width: 70 },
    {
      title: '操作',
      width: 170,
      render: (_: unknown, r: AdminProduct) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r.id)}>
            编辑
          </Button>
          {[0, 1, 2].includes(r.status) && (
            <Popconfirm title={`确认${r.status === 1 ? '下架' : '上架'}？`} onConfirm={() => toggleStatus(r)}>
              <Button size="small" type="link">
                {r.status === 1 ? '下架' : '上架'}
              </Button>
            </Popconfirm>
          )}
        </>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
        <Select
          allowClear
          placeholder="全部状态"
          style={{ width: 120 }}
          value={query.status}
          onChange={(v) => setQuery({ ...query, page: 1, status: v ?? -1 })}
          options={[
            { value: -1, label: '全部状态' },
            { value: 0, label: '草稿' },
            { value: 1, label: '在售' },
            { value: 2, label: '下架' },
            { value: 3, label: '锁定中' },
            { value: 4, label: '已售出' },
          ]}
        />
        <Input.Search
          placeholder="标题 / SPU 搜索"
          style={{ width: 260 }}
          onSearch={(kw) => setQuery({ ...query, page: 1, keyword: kw })}
        />
        <Button type="primary" onClick={() => openEdit()}>
          新增宠物
        </Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns as never}
        dataSource={data.list}
        pagination={{
          total: data.total,
          current: query.page,
          pageSize: query.pageSize,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery({ ...query, page, pageSize }),
        }}
      />
      <Modal
        title={editId ? '编辑宠物' : '新增宠物'}
        open={open}
        onOk={onSubmit}
        onCancel={() => setOpen(false)}
        width={720}
        destroyOnClose
      >
        <Form form={form} layout="vertical" initialValues={{ petGender: 0 }}>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input placeholder="如：赛级布偶猫妹妹 海豹双色" />
          </Form.Item>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 12 }}>
            <Form.Item name="categoryId" label="分类" rules={[{ required: true, message: '必选' }]}>
              <Select
                placeholder="分类"
                options={cats.map((c) => ({ value: c.id, label: c.name }))}
                onChange={onCategoryChange}
              />
            </Form.Item>
            <Form.Item name="breedId" label="品种" rules={[{ required: true, message: '必选' }]}>
              <Select placeholder="先选分类" options={breeds.map((b) => ({ value: b.id, label: b.name }))} />
            </Form.Item>
            <Form.Item name="petGender" label="性别">
              <Select
                options={[
                  { value: 0, label: '未知' },
                  { value: 1, label: '公' },
                  { value: 2, label: '母' },
                ]}
              />
            </Form.Item>
            <Form.Item name="price" label="售价（元）" rules={[{ required: true, message: '必填' }]}>
              <InputNumber min={0.01} precision={2} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="originalPrice" label="划线价（元）">
              <InputNumber min={0} precision={2} style={{ width: '100%' }} />
            </Form.Item>
            <Form.Item name="birthDate" label="出生日期">
              <Input placeholder="yyyy-MM-dd" />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="mainImage" label="主图 URL" rules={[{ required: true, message: '必填（接 COS 后改为直传）' }]}>
              <Input />
            </Form.Item>
            <Form.Item name="videoUrl" label="视频 URL">
              <Input />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="vaccineDesc" label="疫苗情况">
              <Input placeholder="如：三针疫苗已齐" />
            </Form.Item>
            <Form.Item name="dewormDesc" label="驱虫情况">
              <Input placeholder="如：体内外驱虫已完成" />
            </Form.Item>
            <Form.Item name="bodyType" label="体型">
              <Input placeholder="小型/中型/大型" />
            </Form.Item>
            <Form.Item name="coatColor" label="毛色">
              <Input />
            </Form.Item>
            <Form.Item name="personality" label="性格">
              <Input placeholder="如：黏人安静" />
            </Form.Item>
            <Form.Item name="healthDesc" label="健康说明">
              <Input />
            </Form.Item>
            <Form.Item name="nextVaccineDate" label="疫苗到期日">
              <Input placeholder="yyyy-MM-dd" />
            </Form.Item>
            <Form.Item name="nextDewormDate" label="驱虫到期日">
              <Input placeholder="yyyy-MM-dd" />
            </Form.Item>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <Form.Item name="certType" label="血统证书类型"><Input placeholder="CFA / CKU / TICA" /></Form.Item>
            <Form.Item name="certNo" label="血统证书编号"><Input /></Form.Item>
            <Form.Item name="chipNo" label="芯片号"><Input /></Form.Item>
            <Form.Item name="supplierId" label="供货商">
              <Select
                allowClear
                placeholder="选择供货商（可选）"
                options={suppliers.map((s) => ({ value: s.id, label: s.name }))}
              />
            </Form.Item>
            <Form.Item name="quarantineCertUrl" label="检疫证明 URL">
              <Input placeholder="https://...（可选）" maxLength={512} />
            </Form.Item>
          </div>
          <Form.Item name="images" label="相册 URL（逗号分隔，按顺序展示）">
            <Input.TextArea rows={2} placeholder="https://... , https://..." />
          </Form.Item>
          <Form.Item name="detailHtml" label="图文详情 HTML">
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item name="detailImages" label="详情长图 URL（每行一个）">
            <Input.TextArea rows={3} placeholder={'https://...\nhttps://...'} />
          </Form.Item>
          <Form.Item name="preSale" label="预售模式"><Select options={[{value:0,label:'否'},{value:1,label:'是'}]} /></Form.Item>
          <Form.Item name="preSalePrice" label="预售价格"><InputNumber min={0} precision={2} style={{ width: '100%' }} /></Form.Item>
          <Form.Item name="preSaleEta" label="预计到窝日期"><Input placeholder="yyyy-MM-dd" /></Form.Item>
          <Form.Item name="scheduledOffSaleAt" label="定时下架时间"><Input placeholder="yyyy-MM-dd HH:mm" /></Form.Item>
          <Form.Item name="stockWarnThreshold" label="库存预警阈值（在售库存 ≤ 该值时进入预警）">
            <InputNumber min={0} max={100} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="SKU 规格（规格商品下单必选其一，价格以规格为准）">
            <Form.List name="skus">
              {(fields, { add, remove }) => (
                <>
                  {fields.map((field) => (
                    <div key={field.key} style={{ display: 'grid', gridTemplateColumns: '2fr 1fr auto', gap: 8 }}>
                      <Form.Item name={[field.name, 'specs']} noStyle rules={[{ required: true, message: '规格描述' }]}>
                        <Input placeholder="如：3个月|含三针疫苗" />
                      </Form.Item>
                      <Form.Item name={[field.name, 'price']} noStyle rules={[{ required: true, message: '价格' }]}>
                        <InputNumber min={0.01} precision={2} placeholder="价格" style={{ width: '100%' }} />
                      </Form.Item>
                      <Button danger type="link" onClick={() => remove(field.name)}>
                        删除
                      </Button>
                    </div>
                  ))}
                  <Button type="dashed" block onClick={() => add({ status: 1 })}>
                    + 添加规格
                  </Button>
                </>
              )}
            </Form.List>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

