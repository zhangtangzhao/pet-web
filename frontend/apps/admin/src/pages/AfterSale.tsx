import { useCallback, useEffect, useState } from 'react'
import { Button, Form, Input, InputNumber, message, Modal, Radio, Table, Tabs, Tag } from 'antd'
import { PageResp } from '../api/client'
import { confirmExchange } from '../api/engagement'
import { AfterSaleRow, auditAfterSale, fetchAfterSales } from '../api/admin'

const STATUS_TAG: Record<number, { color: string; text: string }> = {
  1: { color: 'gold', text: '待审核' },
  2: { color: 'green', text: '已同意' },
  3: { color: 'red', text: '已拒绝' },
  4: { color: 'default', text: '已撤销' },
}

export default function AfterSale() {
  const [data, setData] = useState<PageResp<AfterSaleRow>>({ total: 0, list: [] })
  const [query, setQuery] = useState({ page: 1, pageSize: 10, status: 0 })
  const [loading, setLoading] = useState(false)
  const [target, setTarget] = useState<AfterSaleRow>()
  const [form] = Form.useForm()
  const agree: boolean = Form.useWatch('agree', form) ?? true

  const load = useCallback(
    async (q = query) => {
      setLoading(true)
      try {
        setData(await fetchAfterSales(q))
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

  const openAudit = (r: AfterSaleRow) => {
    form.setFieldsValue({ agree: true, amount: Number(r.refundAmount), note: '' })
    setTarget(r)
  }

  const submitAudit = async () => {
    const v = await form.validateFields()
    if (!target) return
    try {
      await auditAfterSale(
        target.afterSaleNo,
        v.agree,
        v.agree && v.amount != null ? String(v.amount) : undefined,
        v.note,
      )
      message.success(v.agree ? '已同意并受理退款' : '已拒绝')
      setTarget(undefined)
      load()
    } catch (e: any) {
      message.error(e.message)
    }
  }

  const columns = [
    { title: '售后单号', dataIndex: 'afterSaleNo', width: 200 },
    { title: '订单号', dataIndex: 'orderNo', width: 200 },
    { title: '会员', dataIndex: 'nickname', width: 110 },
    { title: '手机号', dataIndex: 'phone', width: 130 },
    { title: '原因', dataIndex: 'reason', ellipsis: true },
    { title: '退款金额', dataIndex: 'refundAmount', width: 100, render: (v: string) => `¥${v}` },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: (v: number, r: AfterSaleRow) => (
        <Tag color={STATUS_TAG[v]?.color}>{r.statusText || STATUS_TAG[v]?.text}</Tag>
      ),
    },
    { title: '申请时间', dataIndex: 'createdAt', width: 165, render: (v: string) => v?.replace('T', ' ').slice(0, 16) },
    {
      title: '操作',
      width: 200,
      fixed: 'right' as const,
      render: (_: unknown, r: AfterSaleRow) =>
        r.status === 1 ? (
          <Button size="small" type="link" onClick={() => openAudit(r)}>
            审核
          </Button>
        ) : r.status === 5 ? (
          <Button
            size="small"
            type="primary"
            onClick={() => {
              Modal.confirm({
                title: `确认换出 · ${r.afterSaleNo}`,
                content: '输入换出发货的运单号',
                okText: '确认换出',
                onOk: async () => {
                  const el = document.querySelector<HTMLInputElement>('.ant-modal-confirm .ant-input')
                  try { await confirmExchange(r.afterSaleNo, el?.value?.trim() || 'MANUAL'); message.success('换货完成'); load() } catch (e: any) { message.error(e.message) }
                },
              })
            }}
          >
            确认换出
          </Button>
        ) : (
          <span style={{ color: '#999' }}>{r.adminNote || '-'}</span>
        ),
    },
  ]

  return (
    <div>
      <Tabs
        activeKey={String(query.status)}
        onChange={(k) => setQuery({ ...query, page: 1, status: Number(k) })}
        items={[
          { key: '0', label: '全部' },
          { key: '1', label: '待审核' },
          { key: '2', label: '已同意' },
          { key: '3', label: '已拒绝' },
          { key: '4', label: '已撤销' },
        ]}
      />
      <Table
        rowKey="id"
        loading={loading}
        columns={columns as never}
        dataSource={data.list}
        scroll={{ x: 1350 }}
        pagination={{
          total: data.total,
          current: query.page,
          pageSize: query.pageSize,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (page, pageSize) => setQuery({ ...query, page, pageSize }),
        }}
      />
      <Modal
        title={`售后审核 · ${target?.afterSaleNo ?? ''}`}
        open={!!target}
        onOk={submitAudit}
        onCancel={() => setTarget(undefined)}
        destroyOnClose
      >
        <Form form={form} layout="vertical" initialValues={{ agree: true }}>
          <Form.Item name="agree" label="审核结论" rules={[{ required: true }]}>
            <Radio.Group
              options={[
                { value: true, label: '同意退款' },
                { value: false, label: '拒绝' },
              ]}
              optionType="button"
              buttonStyle="solid"
            />
          </Form.Item>
          <Form.Item label="订单实付">{target ? `¥${target.refundAmount}` : ''}</Form.Item>
          {agree && (
            <Form.Item
              name="amount"
              label="退款金额（元，可调整）"
              rules={[{ required: true, message: '请填写退款金额' }]}
            >
              <InputNumber min={0.01} precision={2} style={{ width: '100%' }} addonAfter="元" />
            </Form.Item>
          )}
          <Form.Item
            name="note"
            label="备注"
            rules={agree ? [] : [{ required: true, message: '拒绝时必须填写备注' }]}
          >
            <Input.TextArea rows={3} placeholder={agree ? '选填' : '请填写拒绝原因（会员可见）'} />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
