import { useCallback, useEffect, useState } from 'react'
import { Input, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get, getToken, post, put, del } from '../../request'
import { AddressView } from '../../types'
import './index.css'

const EMPTY = { id: '', name: '', phone: '', address: '', isDefault: false }

export default function Address() {
  const [list, setList] = useState<AddressView[]>([])
  const [form, setForm] = useState<typeof EMPTY>(EMPTY)
  const [showForm, setShowForm] = useState(false)
  const [saving, setSaving] = useState(false)

  const load = useCallback(
    () =>
      get<AddressView[]>('/addresses')
        .then((l) => setList(l ?? []))
        .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' })),
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Faddress%2Findex' }).catch(() => {})
      return
    }
    load()
  }, [load])

  const edit = (a: AddressView) => {
    setForm({ id: a.id, name: a.name, phone: a.phone, address: a.address, isDefault: a.isDefault === 1 })
    setShowForm(true)
  }

  const save = async () => {
    if (saving) return
    if (!form.name.trim() || !form.address.trim()) {
      Taro.showToast({ title: '请填写联系人和地址', icon: 'none' })
      return
    }
    setSaving(true)
    try {
      await post('/addresses', {
        id: form.id || undefined,
        name: form.name.trim(),
        phone: form.phone.trim(),
        address: form.address.trim(),
        isDefault: form.isDefault,
      })
      Taro.showToast({ title: '已保存', icon: 'success' })
      setShowForm(false)
      setForm(EMPTY)
      load()
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setSaving(false)
    }
  }

  const setDefault = (a: AddressView) => {
    put(`/addresses/${a.id}/default`)
      .then(() => load())
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const remove = (a: AddressView) => {
    Taro.showModal({
      title: '删除地址',
      content: `确定删除「${a.name}」的地址吗？`,
      success: (r) => {
        if (!r.confirm) return
        del(`/addresses/${a.id}`)
          .then(() => {
            Taro.showToast({ title: '已删除', icon: 'success' })
            load()
          })
          .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
      },
    })
  }

  return (
    <View className='ab'>
      {list.length === 0 && !showForm && <View className='ab-empty'>还没有收货地址</View>}
      {list.map((a) => (
        <View className='ab-item' key={a.id} onClick={() => edit(a)}>
          <View className='ab-main'>
            <View className='ab-line1'>
              <Text className='ab-name'>{a.name}</Text>
              <Text className='ab-phone'>{a.phone}</Text>
              {a.isDefault === 1 && <Text className='ab-def'>默认</Text>}
            </View>
            <Text className='ab-addr'>{a.address}</Text>
          </View>
          <View className='ab-ops'>
            {a.isDefault !== 1 && (
              <Text
                className='ab-op'
                onClick={(e) => {
                  e.stopPropagation()
                  setDefault(a)
                }}
              >
                设为默认
              </Text>
            )}
            <Text
              className='ab-op ab-op-del'
              onClick={(e) => {
                e.stopPropagation()
                remove(a)
              }}
            >
              删除
            </Text>
          </View>
        </View>
      ))}

      {!showForm ? (
        <View
          className='ab-add'
          onClick={() => {
            setForm(EMPTY)
            setShowForm(true)
          }}
        >
          + 新增收货地址
        </View>
      ) : (
        <View className='ab-form'>
          <View className='ab-form-title'>{form.id ? '编辑地址' : '新增地址'}</View>
          <Input className='ab-input' placeholder='联系人姓名' value={form.name} onInput={(e) => setForm({ ...form, name: e.detail.value })} />
          <Input className='ab-input' placeholder='联系手机号' type='number' maxlength={11} value={form.phone} onInput={(e) => setForm({ ...form, phone: e.detail.value })} />
          <Input className='ab-input' placeholder='详细地址（含城市 / 到达机场或门牌号）' value={form.address} onInput={(e) => setForm({ ...form, address: e.detail.value })} />
          <View className='ab-form-default' onClick={() => setForm({ ...form, isDefault: !form.isDefault })}>
            <Text>设为默认地址</Text>
            <View className={`ab-switch ${form.isDefault ? 'ab-switch-on' : ''}`} />
          </View>
          <View className='ab-form-btns'>
            <View className='ab-btn' onClick={() => setShowForm(false)}>
              取消
            </View>
            <View className='ab-btn ab-btn-primary' onClick={save}>
              保存
            </View>
          </View>
        </View>
      )}
    </View>
  )
}
