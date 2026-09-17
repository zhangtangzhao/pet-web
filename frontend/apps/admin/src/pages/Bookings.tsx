import { useCallback, useEffect, useState } from 'react'
import { Button, Card, message, Modal, Table, Tag } from 'antd'
import { BookingRow, fetchBookings, verifyBooking } from '../api/play'
export default function Bookings() {
  const [list, setList] = useState<BookingRow[]>([])
  const load = useCallback(() => { fetchBookings({ page: 1, pageSize: 100 }).then((r) => setList(r.list)).catch((e: any) => message.error(e.message)) }, [])
  useEffect(() => { load() }, [load])
  const verify = (r: BookingRow) => {
    Modal.confirm({ title: `核销 ${r.bookingNo}`, content: '输入用户出示的核销码', okText: '核销',
      onOk: async () => { const el = document.querySelector<HTMLInputElement>('.ant-modal-confirm .ant-input'); try { await verifyBooking(r.bookingNo, el?.value ?? ''); message.success('核销成功'); load() } catch (e: any) { message.error(e.message) } } })
  }
  return (
    <Card title="服务预约核销">
      <Table rowKey="bookingNo" size="small" dataSource={list} pagination={false} columns={[
        { title: '预约号', dataIndex: 'bookingNo', width: 190 },
        { title: '服务', dataIndex: 'serviceName', width: 120 },
        { title: '门店', dataIndex: 'storeName', width: 120 },
        { title: '日期', dataIndex: 'date', width: 110 },
        { title: '时段', dataIndex: 'slot', width: 120 },
        { title: '联系人', width: 150, render: (_: unknown, r: BookingRow) => `${r.contact} ${r.phone}` },
        { title: '核销码', dataIndex: 'verifyCode', width: 90 },
        { title: '状态', dataIndex: 'statusText', width: 90, render: (v: string) => <Tag color={v === '待到店' ? 'orange' : v === '已完成' ? 'green' : 'default'}>{v}</Tag> },
        { title: '操作', width: 90, render: (_: unknown, r: BookingRow) => r.status === 0 ? <Button size="small" type="link" onClick={() => verify(r)}>核销</Button> : null },
      ] as never} />
    </Card>
  )
}
