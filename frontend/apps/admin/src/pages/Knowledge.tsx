import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Form, Input, InputNumber, message, Modal, Popconfirm, Select, Space, Switch, Table } from 'antd'
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons'
import { BreedItem, KnowledgeRow, deleteKnowledge, fetchBreeds, fetchKnowledge, upsertKnowledge } from '../api/admin'

interface FormValues {
  breedId?: string
  title: string
  keywords?: string
  content: string
  sort?: number
  status: boolean
}

export default function Knowledge() {
  const [rows, setRows] = useState<KnowledgeRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [keyword, setKeyword] = useState('')
  const [breeds, setBreeds] = useState<BreedItem[]>([])
  const [modal, setModal] = useState(false)
  const [form] = Form.useForm()

  const load = useCallback(
    async (p = page, kw = keyword) => {
      const resp = await fetchKnowledge({ page: p, pageSize: 10, keyword: kw || undefined })
      setRows(resp.list)
      setTotal(resp.total)
      setPage(p)
    },
    [page, keyword],
  )

  useEffect(() => {
    load(1).catch((e) => message.error(e.message))
    fetchBreeds().then(setBreeds).catch(() => {})
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const openCreate = () => {
    form.setFieldsValue({ id: undefined, breedId: undefined, title: '', keywords: '', content: '', sort: 0, status: true })
    setModal(true)
  }

  const openEdit = (r: KnowledgeRow) => {
    form.setFieldsValue({
      id: r.id,
      breedId: r.breedId || undefined,
      title: r.title,
      keywords: r.keywords,
      content: r.content,
      sort: r.sort,
      status: r.status === 1,
    })
    setModal(true)
  }

  const submit = async () => {
    const v: FormValues = await form.validateFields()
    try {
      await upsertKnowledge({
        id: form.getFieldValue('id') || undefined,
        breedId: v.breedId,
        title: v.title,
        keywords: v.keywords,
        content: v.content,
        sort: v.sort,
        status: v.status ? 1 : 0,
      })
      message.success('已保存')
      setModal(false)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const toggleStatus = async (r: KnowledgeRow) => {
    try {
      await upsertKnowledge({
        id: r.id,
        breedId: r.breedId || undefined,
        title: r.title,
        keywords: r.keywords,
        content: r.content,
        sort: r.sort,
        status: r.status === 1 ? 0 : 1,
      })
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '标题', dataIndex: 'title', width: 200, ellipsis: true },
    { title: '归属', dataIndex: 'breedName', width: 110 },
    { title: '关键词', dataIndex: 'keywords', width: 160, ellipsis: true },
    {
      title: '内容摘要',
      dataIndex: 'content',
      ellipsis: true,
      render: (v: string) => <span style={{ color: '#888' }}>{v}</span>,
    },
    { title: '排序', dataIndex: 'sort', width: 70 },
    {
      title: '启用',
      dataIndex: 'status',
      width: 70,
      render: (_: unknown, r: KnowledgeRow) => (
        <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
      ),
    },
    { title: '更新时间', dataIndex: 'updatedAt', width: 160 },
    {
      title: '操作',
      width: 130,
      render: (_: unknown, r: KnowledgeRow) => (
        <>
          <Button size="small" type="link" onClick={() => openEdit(r)}>
            编辑
          </Button>
          <Popconfirm
            title="删除该知识条目？"
            onConfirm={async () => {
              try {
                await deleteKnowledge(r.id)
                message.success('已删除')
                load()
              } catch (e: any) {
                message.error(e.message)
              }
            }}
          >
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
      title="AI 知识库"
      extra={
        <Space>
          <Input
            allowClear
            placeholder="标题/关键词搜索"
            prefix={<SearchOutlined />}
            style={{ width: 220 }}
            onChange={(e) => setKeyword(e.target.value)}
            onPressEnter={() => load(1)}
          />
          <Button onClick={() => load(1)}>查询</Button>
          <Button icon={<ReloadOutlined />} onClick={() => load()} />
          <Button type="primary" onClick={openCreate}>
            新建条目
          </Button>
        </Space>
      }
    >
      <Table
        rowKey="id"
        columns={columns}
        dataSource={rows}
        pagination={{ current: page, pageSize: 10, total, onChange: (p) => load(p) }}
      />
      <Modal title="知识条目" open={modal} onOk={submit} onCancel={() => setModal(false)} width={640}>
        <Form form={form} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="breedId" label="归属品种（不选 = 平台通用，全品种适用）">
            <Select allowClear placeholder="平台通用" options={breeds.map((b) => ({ value: b.id, label: b.name }))} />
          </Form.Item>
          <Form.Item name="title" label="标题" rules={[{ required: true, message: '请输入标题' }]}>
            <Input maxLength={128} placeholder="如：布偶猫喂养与养护要点" />
          </Form.Item>
          <Form.Item name="keywords" label="检索关键词（英文逗号分隔，用户提问命中后该条目优先被引用）">
            <Input maxLength={255} placeholder="如：喂养,猫粮,幼猫" />
          </Form.Item>
          <Form.Item name="content" label="内容" rules={[{ required: true, message: '请输入内容' }]}>
            <Input.TextArea rows={6} maxLength={4000} showCount />
          </Form.Item>
          <Space>
            <Form.Item name="sort" label="排序（越小越靠前）">
              <InputNumber min={0} />
            </Form.Item>
            <Form.Item name="status" label="启用" valuePropName="checked">
              <Switch />
            </Form.Item>
          </Space>
        </Form>
      </Modal>
    </Card>
  )
}
