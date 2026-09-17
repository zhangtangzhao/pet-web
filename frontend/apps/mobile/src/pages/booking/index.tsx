import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post } from '../../request'
import './index.css'
export default function BookingPage() {
  const [list, setList] = useState<any[]>([])
  const [services, setServices] = useState<any[]>([])
  const [stores, setStores] = useState<any[]>([])
  const [slots] = useState(['09:00-11:00', '11:00-13:00', '13:00-15:00', '15:00-17:00'])
  const [showForm, setShowForm] = useState(false)
  const load = useCallback(() => {
    if (!getToken()) { Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fbooking%2Findex' }).catch(() => {}); return }
    get<any[]>('/bookings').then(setList).catch(() => {})
    get<any[]>('/services').then(setServices).catch(() => {})
    get<any[]>('/stores').then(setStores).catch(() => {})
  }, [])
  useEffect(() => { load() }, [load])
  const [form, setForm] = useState<any>({})
  const submit = () => {
    post('/bookings', form).then((r: any) => { Taro.showToast({ title: '预约成功，核销码 ' + r.verifyCode, icon: 'none' }); setShowForm(false); load() }).catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }
  const cancel = (no: string) => { post(`/bookings/${no}/cancel`).then(load).catch(() => {}) }
  return (
    <View className='bk'>
      <View className='bk-head'><Text className='bk-title'>服务预约</Text><View className='bk-add' onClick={() => setShowForm(!showForm)}>{showForm ? '收起' : '+ 预约'}</View></View>
      {showForm && (
        <View className='bk-form'>
          <View className='bk-field'><Text>服务</Text><View className='bk-chips'>{services.map((s) => <View key={s.id} className={`bk-chip ${form.serviceId === s.id ? 'bk-on' : ''}`} onClick={() => setForm({ ...form, serviceId: s.id })}>{s.name} ¥{s.price}</View>)}</View></View>
          <View className='bk-field'><Text>门店</Text><View className='bk-chips'>{stores.map((s) => <View key={s.id} className={`bk-chip ${form.storeId === s.id ? 'bk-on' : ''}`} onClick={() => setForm({ ...form, storeId: s.id })}>{s.name}</View>)}</View></View>
          <View className='bk-field'><Text>时段</Text><View className='bk-chips'>{slots.map((s) => <View key={s} className={`bk-chip ${form.slot === s ? 'bk-on' : ''}`} onClick={() => setForm({ ...form, slot: s })}>{s}</View>)}</View></View>
          <View className='bk-row'><Text>日期</Text><input className='bk-input' placeholder='yyyy-MM-dd' value={form.date || ''} onInput={(e: any) => setForm({ ...form, date: e.detail.value })} /></View>
          <View className='bk-row'><Text>联系人</Text><input className='bk-input' value={form.contact || ''} onInput={(e: any) => setForm({ ...form, contact: e.detail.value })} /></View>
          <View className='bk-row'><Text>手机号</Text><input className='bk-input' type='number' value={form.phone || ''} onInput={(e: any) => setForm({ ...form, phone: e.detail.value })} /></View>
          <View className='bk-save' onClick={submit}>提交预约</View>
        </View>
      )}
      {list.map((b) => (
        <View className='bk-card' key={b.bookingNo}>
          <View className='bk-row'><Text className='bk-svc'>{b.serviceName}</Text><Text className={`bk-st bk-st-${b.status}`}>{b.statusText}</Text></View>
          <Text className='bk-info'>{b.date} {b.slot} · {b.storeName} · {b.contact} {b.phone}</Text>
          {!!b.verifyCode && b.status === 0 && <Text className='bk-code'>核销码 {b.verifyCode}</Text>}
          {b.status === 0 && <Text className='bk-cancel' onClick={() => cancel(b.bookingNo)}>取消预约</Text>}
        </View>
      ))}
    </View>
  )
}
