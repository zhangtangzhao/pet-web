import { useCallback, useEffect, useState } from 'react'
import { Avatar, Button, Input, message, Popconfirm, Switch, Table, Tag } from 'antd'
import { PageResp } from '../api/client'
import { fetchMembers, MemberRow, updateMemberStatus } from '../api/admin'

export default function Members() {
  const [data, setData] = useState<PageResp<MemberRow>>({ total: 0, list: [] })
  const [query, setQuery] = useState({ page: 1, pageSize: 10, keyword: '' })
  const [loading, setLoading] = useState(false)

  const load = useCallback(
    async (q = query) => {
      setLoading(true)
      try {
        setData(await fetchMembers(q))
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

  const toggle = async (r: MemberRow, enabled: boolean) => {
    try {
      await updateMemberStatus(r.id, enabled ? 1 : 2)
      message.success(enabled ? '已启用' : '已禁用')
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    {
      title: '会员',
      dataIndex: 'nickname',
      render: (v: string, r: MemberRow) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <Avatar size="small" src={r.avatar || undefined}>
            {v.slice(0, 1)}
          </Avatar>
          {v}
        </div>
      ),
    },
    { title: '手机号', dataIndex: 'phone', width: 140 },
    {
      title: '性别',
      dataIndex: 'gender',
      width: 80,
      render: (v: number) => (v === 1 ? '男' : v === 2 ? '女' : '未知'),
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v: number) => <Tag color={v === 1 ? 'green' : 'red'}>{v === 1 ? '正常' : '禁用'}</Tag>,
    },
    { title: '订单数', dataIndex: 'orderCount', width: 90 },
    { title: '收藏数', dataIndex: 'favoriteCount', width: 90 },
    {
      title: '注册时间',
      dataIndex: 'createdAt',
      width: 170,
      render: (v: string) => v?.replace('T', ' ').slice(0, 19),
    },
    {
      title: '操作',
      width: 110,
      render: (_: unknown, r: MemberRow) => (
        <Popconfirm title={`确认${r.status === 1 ? '禁用' : '启用'}该会员？`} onConfirm={() => toggle(r, r.status !== 1)}>
          <Switch checked={r.status === 1} size="small" />
        </Popconfirm>
      ),
    },
  ]

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Input.Search
          allowClear
          placeholder="昵称 / 手机号搜索"
          style={{ width: 260 }}
          onSearch={(v) => setQuery({ ...query, page: 1, keyword: v })}
        />
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
    </div>
  )
}
