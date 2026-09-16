import { useCallback, useEffect, useState } from 'react'
import { Button, Empty, Form, Image, Input, message, Modal, Popconfirm, Rate, Space, Switch, Table, Tag } from 'antd'
import { DeleteOutlined, MessageOutlined, ReloadOutlined } from '@ant-design/icons'
import { deleteReview, fetchReviews, replyReview, ReviewRow, updateReviewStatus } from '../api/admin'

export default function Reviews() {
  const [list, setList] = useState<ReviewRow[]>([])
  const [hasMore, setHasMore] = useState(false)
  const [cursor, setCursor] = useState('')
  const [loading, setLoading] = useState(false)
  const [replyOpen, setReplyOpen] = useState(false)
  const [replyRow, setReplyRow] = useState<ReviewRow>()
  const [replyForm] = Form.useForm()

  const load = useCallback(
    async (nextCursor: string, append: boolean) => {
      setLoading(true)
      try {
        const resp = await fetchReviews({ cursor: nextCursor || undefined, limit: 20 })
        setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
        setHasMore(resp.hasMore)
        setCursor(resp.list.length ? resp.list[resp.list.length - 1].id : nextCursor)
      } catch (e: any) {
        message.error(e.message)
      } finally {
        setLoading(false)
      }
    },
    [],
  )

  useEffect(() => {
    load('', false)
  }, [load])

  const toggleStatus = async (r: ReviewRow) => {
    try {
      await updateReviewStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const remove = async (r: ReviewRow) => {
    try {
      await deleteReview(r.id)
      message.success('已删除')
      load('', false)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const openReply = (r: ReviewRow) => {
    setReplyRow(r)
    replyForm.setFieldsValue({ reply: r.reply || '' })
    setReplyOpen(true)
  }

  const submitReply = async () => {
    if (!replyRow) return
    try {
      const v = await replyForm.validateFields()
      await replyReview(replyRow.id, v.reply)
      message.success('已回复')
      setList((prev) => prev.map((x) => (x.id === replyRow.id ? { ...x, reply: v.reply, repliedAt: new Date().toISOString() } : x)))
      setReplyOpen(false)
    } catch (e: any) {
      if (e?.message) message.error(e.message)
    }
  }

  const columns = [
    { title: '会员', dataIndex: 'nickname', width: 110 },
    { title: '商品', dataIndex: 'productTitle', ellipsis: true },
    { title: '评分', dataIndex: 'rating', width: 140, render: (v: number) => <Rate disabled value={v} style={{ fontSize: 14 }} /> },
    {
      title: '三维分',
      width: 130,
      render: (_: unknown, r: ReviewRow) => (
        <span style={{ fontSize: 12, color: '#666' }}>
          健康 {r.healthScore ?? 5} / 品相 {r.lookScore ?? 5} / 服务 {r.serviceScore ?? 5}
        </span>
      ),
    },
    { title: '内容', dataIndex: 'content', ellipsis: true, render: (v: string) => v || '-' },
    {
      title: '图片',
      dataIndex: 'images',
      width: 150,
      render: (imgs: string[]) =>
        imgs?.length ? (
          <Image.PreviewGroup>
            <Space>
              {imgs.slice(0, 3).map((u, i) => (
                <Image key={i} src={u} width={40} height={40} style={{ objectFit: 'cover' }} />
              ))}
            </Space>
          </Image.PreviewGroup>
        ) : (
          '-'
        ),
    },
    { title: '订单号', dataIndex: 'orderNo', width: 190 },
    { title: '官方回复', dataIndex: 'reply', ellipsis: true, render: (v: string) => v || '-' },
    {
      title: '状态',
      dataIndex: 'status',
      width: 80,
      render: (v: number) => <Tag color={v === 1 ? 'green' : 'default'}>{v === 1 ? '显示' : '隐藏'}</Tag>,
    },
    { title: '时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 200,
      fixed: 'right' as const,
      render: (_: unknown, r: ReviewRow) => (
        <Space size={4}>
          <Switch
            size="small"
            checked={r.status === 1}
            checkedChildren="显"
            unCheckedChildren="隐"
            onChange={() => toggleStatus(r)}
          />
          <Button size="small" type="link" icon={<MessageOutlined />} onClick={() => openReply(r)}>
            回复
          </Button>
          <Popconfirm title="确定删除该评价？" onConfirm={() => remove(r)}>
            <Button size="small" type="link" danger icon={<DeleteOutlined />} />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <span style={{ fontSize: 16, fontWeight: 600 }}>订单评价</span>
        <Button icon={<ReloadOutlined />} onClick={() => load('', false)}>
          刷新
        </Button>
      </div>
      <Table
        rowKey="id"
        loading={loading}
        columns={columns as never}
        dataSource={list}
        pagination={false}
        scroll={{ x: 1400 }}
        locale={{ emptyText: <Empty description="暂无评价" /> }}
        footer={() =>
          hasMore ? (
            <div style={{ textAlign: 'center' }}>
              <Button onClick={() => load(cursor, true)} loading={loading}>
                加载更多
              </Button>
            </div>
          ) : null
        }
      />
      <Modal
        title={`回复评价${replyRow ? ` · ${replyRow.nickname || '匿名用户'}` : ''}`}
        open={replyOpen}
        onOk={submitReply}
        onCancel={() => setReplyOpen(false)}
        destroyOnClose
      >
        {replyRow?.content && (
          <div
            style={{
              background: '#f5f5f5',
              borderRadius: 8,
              padding: '8px 12px',
              marginBottom: 16,
              color: '#666',
              fontSize: 13,
            }}
          >
            {replyRow.content}
          </div>
        )}
        <Form form={replyForm} layout="vertical">
          <Form.Item
            name="reply"
            label="官方回复内容（将展示在用户端评价区，可重复提交覆盖）"
            rules={[{ required: true, message: '必填' }]}
          >
            <Input.TextArea rows={4} maxLength={200} showCount placeholder="感谢您的评价，祝您和毛孩子生活愉快…" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
