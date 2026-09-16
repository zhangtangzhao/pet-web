import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { del, get, getToken, post } from '../../request'
import { UserPet } from '../../types'
import './index.css'

const EMPTY = { id: '', name: '', breedName: '', gender: 0, birthday: '', weight: '', vaccineAt: '', nextVaccineDate: '', nextDewormDate: '', remark: '' }

export default function PetProfiles() {
  const [pets, setPets] = useState<UserPet[]>([])
  const [form, setForm] = useState<typeof EMPTY>(EMPTY)
  const [showForm, setShowForm] = useState(false)

  const load = useCallback(
    () =>
      get<{ list: UserPet[] }>('/pets')
        .then((r) => setPets(r.list ?? []))
        .catch(() => {}),
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fpet%2Findex' }).catch(() => {})
      return
    }
    load()
  }, [load])

  const save = () => {
    if (!form.name.trim()) {
      Taro.showToast({ title: '请填写昵称', icon: 'none' })
      return
    }
    post('/pets', form)
      .then(() => {
        Taro.showToast({ title: '已保存', icon: 'success' })
        setForm(EMPTY)
        setShowForm(false)
        load()
      })
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const remove = (id: string) => {
    del(`/pets/${id}`).then(load).catch(() => {})
  }

  const edit = (p: UserPet) => {
    setForm({
      id: p.id, name: p.name, breedName: p.breedName, gender: p.gender,
      birthday: p.birthday ?? '', weight: p.weight ?? '', vaccineAt: p.vaccineAt ?? '',
      nextVaccineDate: p.nextVaccineDate ?? '', nextDewormDate: p.nextDewormDate ?? '',
      remark: p.remark ?? '',
    })
    setShowForm(true)
  }

  const field = (label: string, key: keyof typeof EMPTY, placeholder: string, type = 'text') => (
    <View className='pet-field'>
      <Text className='pet-label'>{label}</Text>
      <input
        className='pet-input'
        type={type}
        value={form[key] as string}
        placeholder={placeholder}
        onInput={(e: any) => setForm({ ...form, [key]: e.detail.value })}
      />
    </View>
  )

  const genderPick = (g: number, label: string) => (
    <View className={`pet-gender-item ${form.gender === g ? 'pet-gender-on' : ''}`} onClick={() => setForm({ ...form, gender: g })}>
      {label}
    </View>
  )

  return (
    <View className='pet'>
      <View className='pet-head'>
        <Text className='pet-head-title'>我的宠物</Text>
        <View className='pet-add' onClick={() => { setForm(EMPTY); setShowForm(!showForm) }}>
          {showForm ? '收起' : '+ 添加'}
        </View>
      </View>

      {showForm && (
        <View className='pet-form'>
          {field('昵称', 'name', '毛孩子的名字')}
          {field('品种', 'breedName', '如 金毛')}
          <View className='pet-field'>
            <Text className='pet-label'>性别</Text>
            <View className='pet-genders'>
              {genderPick(1, '公')}
              {genderPick(2, '母')}
            </View>
          </View>
          {field('生日', 'birthday', 'yyyy-MM-dd')}
          {field('体重(kg)', 'weight', '如 4.5')}
          {field('最近疫苗', 'vaccineAt', 'yyyy-MM-dd')}
          {field('下次疫苗', 'nextVaccineDate', 'yyyy-MM-dd')}
          {field('下次驱虫', 'nextDewormDate', 'yyyy-MM-dd')}
          {field('备注', 'remark', '过敏史/性格等')}
          <View className='pet-save' onClick={save}>保存</View>
        </View>
      )}

      {pets.length === 0 && !showForm && <View className='pet-empty'>还没有宠物档案，点右上角添加</View>}
      {pets.map((p) => (
        <View className='pet-card' key={p.id}>
          <View className='pet-card-head'>
            <Text className='pet-name'>{p.name}</Text>
            {!!p.breedName && <Text className='pet-breed'>{p.breedName}</Text>}
            {p.gender === 1 && <Text className='pet-gender'>♂</Text>}
            {p.gender === 2 && <Text className='pet-gender pet-gender-f'>♀</Text>}
          </View>
          <View className='pet-meta'>
            {!!p.birthday && <Text>生日 {p.birthday}</Text>}
            {!!p.weight && <Text>{p.weight}kg</Text>}
          </View>
          {(!!p.nextVaccineDate || !!p.nextDewormDate) && (
            <View className='pet-care'>
              {!!p.nextVaccineDate && <Text className='pet-care-item'>💉 疫苗 {p.nextVaccineDate}</Text>}
              {!!p.nextDewormDate && <Text className='pet-care-item'>🐛 驱虫 {p.nextDewormDate}</Text>}
            </View>
          )}
          {!!p.remark && <View className='pet-remark'>{p.remark}</View>}
          <View className='pet-actions'>
            <Text onClick={() => edit(p)}>编辑</Text>
            <Text className='pet-del' onClick={() => remove(p.id)}>删除</Text>
          </View>
        </View>
      ))}
    </View>
  )
}
