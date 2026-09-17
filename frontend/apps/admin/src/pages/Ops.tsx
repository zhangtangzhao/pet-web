import { useCallback, useEffect, useState } from 'react'
import { Button, Card, Input, message, Popconfirm, Select, Space, Switch, Table, Tabs, Tag } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { deletePost, fetchPosts, PostRow, setPostStatus } from '../api/engagement'
import {
  AuditLogRow,
  deleteSensitiveWord,
  fetchAuditLogs,
  fetchRiskLogs,
  fetchSensitiveWords,
  RiskLogRow,
  saveSensitiveWord,
  SensitiveWordRow,
  updateSensitiveWordStatus,
} from '../api/admin'

function AuditLogs() {
  const [list, setList] = useState<AuditLogRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchAuditLogs({ page: nextPage, pageSize: 15 })
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

  const METHOD_COLOR: Record<string, string> = { POST: 'green', PUT: 'blue', DELETE: 'red' }

  const columns = [
    { title: '管理员', dataIndex: 'adminName', width: 120 },
    {
      title: '操作',
      dataIndex: 'method',
      width: 90,
      render: (v: string) => <Tag color={METHOD_COLOR[v] || 'default'}>{v}</Tag>,
    },
    { title: '路径', dataIndex: 'path', ellipsis: true, render: (v: string) => <code style={{ fontSize: 12 }}>{v}</code> },
    { title: 'IP', dataIndex: 'ip', width: 140 },
    { title: '时间', dataIndex: 'createdAt', width: 170, render: (v: string) => v?.replace('T', ' ').slice(0, 19) },
  ]

  return (
    <Table
      rowKey="id"
      loading={loading}
      size="small"
      columns={columns as never}
      dataSource={list}
      pagination={{
        current: page,
        total,
        pageSize: 15,
        showSizeChanger: false,
        onChange: (p) => load(p),
      }}
    />
  )
}

const RULE_TEXT: Record<string, string> = {
  order_freq: '下单频控',
  review_contact: '评价联系方式',
  member_blacklist: '黑名单',
}

function RiskLogs() {
  const [list, setList] = useState<RiskLogRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [rule, setRule] = useState('')
  const [loading, setLoading] = useState(false)

  const load = useCallback(
    async (nextPage: number, nextRule = rule) => {
      setLoading(true)
      try {
        const resp = await fetchRiskLogs({ page: nextPage, pageSize: 15, rule: nextRule || undefined })
        setList(resp.list)
        setTotal(resp.total)
        setPage(nextPage)
      } catch (e: any) {
        message.error(e.message)
      } finally {
        setLoading(false)
      }
    },
    [rule],
  )

  useEffect(() => {
    load(1)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return (
    <>
      <Space style={{ marginBottom: 12 }}>
        <Input
          allowClear
          placeholder="按规则筛选（order_freq / review_contact / member_blacklist）"
          style={{ width: 380 }}
          onChange={(e) => setRule(e.target.value.trim())}
          onPressEnter={() => load(1)}
        />
        <Button onClick={() => load(1)}>查询</Button>
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        size="small"
        columns={[
          {
            title: '规则',
            dataIndex: 'rule',
            width: 120,
            render: (v: string) => <Tag color="orange">{RULE_TEXT[v] ?? v}</Tag>,
          },
          { title: '会员ID', dataIndex: 'memberId', width: 200 },
          { title: '详情', dataIndex: 'detail', ellipsis: true },
          { title: '时间', dataIndex: 'createdAt', width: 170 },
        ]}
        dataSource={list}
        pagination={{
          current: page,
          total,
          pageSize: 15,
          showSizeChanger: false,
          onChange: (p) => load(p),
        }}
      />
    </>
  )
}

function SensitiveWords() {
  const [list, setList] = useState<SensitiveWordRow[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(false)
  const [word, setWord] = useState('')
  const [saving, setSaving] = useState(false)

  const load = useCallback(async (nextPage: number) => {
    setLoading(true)
    try {
      const resp = await fetchSensitiveWords({ page: nextPage, pageSize: 15 })
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

  const save = async () => {
    const w = word.trim()
    if (!w) return
    setSaving(true)
    try {
      await saveSensitiveWord(w)
      message.success('已添加')
      setWord('')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    } finally {
      setSaving(false)
    }
  }

  const toggleStatus = async (r: SensitiveWordRow) => {
    try {
      await updateSensitiveWordStatus(r.id, r.status === 1 ? 0 : 1)
      setList((prev) => prev.map((x) => (x.id === r.id ? { ...x, status: x.status === 1 ? 0 : 1 } : x)))
    } catch (e: any) {
      message.error(e.message)
      load(page)
    }
  }

  const remove = async (r: SensitiveWordRow) => {
    try {
      await deleteSensitiveWord(r.id)
      message.success('已删除')
      load(page)
    } catch (e: any) {
      message.error(e.message)
    }
  }

  return (
    <>
      <Space style={{ marginBottom: 16 }}>
        <Input
          placeholder="输入敏感词，回车添加"
          style={{ width: 320 }}
          value={word}
          maxLength={64}
          onChange={(e) => setWord(e.target.value)}
          onPressEnter={save}
        />
        <Button type="primary" icon={<PlusOutlined />} loading={saving} onClick={save}>
          添加
        </Button>
      </Space>
      <Table
        rowKey="id"
        loading={loading}
        size="small"
        dataSource={list}
        columns={[
          { title: '敏感词', dataIndex: 'word', ellipsis: true },
          {
            title: '启用',
            dataIndex: 'status',
            width: 90,
            render: (_: unknown, r: SensitiveWordRow) => (
              <Switch size="small" checked={r.status === 1} onChange={() => toggleStatus(r)} />
            ),
          },
          { title: '添加时间', dataIndex: 'createdAt', width: 170, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
          {
            title: '操作',
            width: 90,
            render: (_: unknown, r: SensitiveWordRow) => (
              <Popconfirm title="确定删除该敏感词？" onConfirm={() => remove(r)}>
                <Button size="small" type="link" danger>
                  删除
                </Button>
              </Popconfirm>
            ),
          },
        ]}
        pagination={{
          current: page,
          total,
          pageSize: 15,
          showSizeChanger: false,
          onChange: (p) => load(p),
        }}
      />
    </>
  )
}

function Posts() {
  const [list, setList] = useState<PostRow[]>([])
  const [status, setStatus] = useState(0)
  const load = useCallback(async (s = status) => {
    try { setList(await fetchPosts(s || undefined)) } catch (e: any) { message.error(e.message) }
  }, [status])
  useEffect(() => { load() }, [load])
  const moderate = async (r: PostRow, next: number) => {
    try { await setPostStatus(r.id, next); message.success('已更新'); load() } catch (e: any) { message.error(e.message) }
  }
  const remove = async (r: PostRow) => {
    try { await deletePost(r.id); message.success('已删除'); load() } catch (e: any) { message.error(e.message) }
  }
  return (
    <>
      <Space style={{ marginBottom: 12 }}>
        <Select style={{ width: 160 }} value={status} onChange={(v) => setStatus(v)}
          options={[{ value: 0, label: '全部' }, { value: 1, label: '显示中' }, { value: 2, label: '已隐藏' }]} />
      </Space>
      <Table rowKey="id" size="small" dataSource={list} pagination={false} columns={[
        { title: '会员', dataIndex: 'nickname', width: 120 },
        { title: '内容', dataIndex: 'content', ellipsis: true },
        { title: '图片', dataIndex: 'images', width: 80, render: (imgs: string[]) => (imgs?.length ?? 0) },
        { title: '点赞', dataIndex: 'likeCount', width: 80 },
        { title: '状态', dataIndex: 'status', width: 90, render: (v: number) => v === 1 ? <Tag color="green">显示中</Tag> : <Tag color="red">已隐藏</Tag> },
        { title: '操作', width: 200, render: (_: unknown, r: PostRow) => (<>
          <Button size="small" type="link" onClick={() => moderate(r, 1)}>通过</Button>
          <Button size="small" type="link" danger onClick={() => moderate(r, 2)}>隐藏</Button>
          <Button size="small" type="link" danger onClick={() => remove(r)}>删除</Button>
        </>) },
      ] as never} />
    </>
  )
}
export default function Ops() {
  return (
    <Card title="平台运营">
      <Tabs
        items={[
          { key: 'audit', label: '操作审计', children: <AuditLogs /> },
          { key: 'sensitive', label: '敏感词', children: <SensitiveWords /> },
          { key: 'risk', label: '风控记录', children: <RiskLogs /> },
          { key: 'posts', label: '晒单审核', children: <Posts /> },
        ]}
      />
    </Card>
  )
}


