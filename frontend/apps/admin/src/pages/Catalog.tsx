import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Col, Form, Input, InputNumber, message, Modal, Popconfirm, Row, Select, Table } from 'antd'
import {
  BreedItem,
  CategoryItem,
  deleteBreed,
  deleteCategory,
  fetchBreeds,
  fetchCategories,
  upsertBreed,
  upsertCategory,
} from '../api/admin'

export default function Catalog() {
  const [cats, setCats] = useState<CategoryItem[]>([])
  const [active, setActive] = useState<CategoryItem>()
  const [breeds, setBreeds] = useState<BreedItem[]>([])

  const [catModal, setCatModal] = useState(false)
  const [breedModal, setBreedModal] = useState(false)
  const [catForm] = Form.useForm()
  const [breedForm] = Form.useForm()

  const loadCats = useCallback(async () => {
    const list = await fetchCategories()
    setCats(list)
    setActive((prev) => prev && list.find((c) => c.id === prev.id) ? prev : list[0])
  }, [])

  const loadBreeds = useCallback(async (cid?: string) => {
    if (!cid) return
    setBreeds(await fetchBreeds(cid))
  }, [])

  useEffect(() => {
    loadCats().catch((e) => message.error(e.message))
  }, [loadCats])

  useEffect(() => {
    if (active) loadBreeds(active.id).catch(() => {})
  }, [active, loadBreeds])

  const submitCat = async () => {
    const v = await catForm.validateFields()
    try {
      await upsertCategory(v)
      message.success('已保存')
      setCatModal(false)
      await loadCats()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const submitBreed = async () => {
    const v = await breedForm.validateFields()
    try {
      await upsertBreed({ ...v, categoryId: v.categoryId ?? active?.id })
      message.success('已保存')
      setBreedModal(false)
      loadBreeds(active?.id)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const catColumns = [
    { title: 'ID', dataIndex: 'id', width: 90 },
    { title: '名称', dataIndex: 'name', width: 120 },
    { title: '品种数', dataIndex: 'breedCount', width: 90 },
    {
      title: '操作',
      render: (_: unknown, r: CategoryItem) => (
        <>
          <Button
            size="small"
            type="link"
            onClick={() => {
              catForm.setFieldsValue(r)
              setCatModal(true)
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="删除分类？需先清空其下品种/商品"
            onConfirm={async () => {
              try {
                await deleteCategory(r.id)
                message.success('已删除')
                loadCats()
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

  const breedColumns = [
    { title: '名称', dataIndex: 'name', width: 140 },
    { title: '介绍', dataIndex: 'intro', ellipsis: true },
    {
      title: '操作',
      width: 150,
      render: (_: unknown, r: BreedItem) => (
        <>
          <Button
            size="small"
            type="link"
            onClick={() => {
              breedForm.setFieldsValue(r)
              setBreedModal(true)
            }}
          >
            编辑
          </Button>
          <Popconfirm
            title="删除品种？需先清空其下商品"
            onConfirm={async () => {
              try {
                await deleteBreed(r.id)
                message.success('已删除')
                loadBreeds(active?.id)
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
    <Row gutter={16}>
      <Col span={10}>
        <Card
          title="宠物分类"
          extra={
            <Button
              size="small"
              type="primary"
              onClick={() => {
                catForm.resetFields()
                setCatModal(true)
              }}
            >
              新增分类
            </Button>
          }
        >
          <Table rowKey="id" size="small" pagination={false} columns={catColumns as never} dataSource={cats} />
        </Card>
      </Col>
      <Col span={14}>
        <Card
          title={active ? `「${active.name}」品种` : '品种'}
          extra={
            <Button
              size="small"
              type="primary"
              onClick={() => {
                breedForm.resetFields()
                breedForm.setFieldsValue({ categoryId: active?.id })
                setBreedModal(true)
              }}
            >
              新增品种
            </Button>
          }
        >
          <Table rowKey="id" size="small" pagination={false} columns={breedColumns as never} dataSource={breeds} />
        </Card>
      </Col>
      <Modal title="分类" open={catModal} onOk={submitCat} onCancel={() => setCatModal(false)} destroyOnClose>
        <Form form={catForm} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '必填' }]}>
            <Input placeholder="如：猫" />
          </Form.Item>
          <Form.Item name="sort" label="排序（越小越靠前）">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
      <Modal title="品种" open={breedModal} onOk={submitBreed} onCancel={() => setBreedModal(false)} destroyOnClose>
        <Form form={breedForm} layout="vertical">
          <Form.Item name="id" hidden>
            <Input />
          </Form.Item>
          <Form.Item name="categoryId" label="所属分类" rules={[{ required: true, message: '必选' }]}>
            <Select
              placeholder="选择分类"
              options={cats.map((c) => ({ value: c.id, label: c.name }))}
            />
          </Form.Item>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '必填' }]}>
            <Input placeholder="如：布偶猫" />
          </Form.Item>
          <Form.Item name="intro" label="介绍">
            <Input.TextArea rows={3} />
          </Form.Item>
          <Form.Item name="sort" label="排序">
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Row>
  )
}
