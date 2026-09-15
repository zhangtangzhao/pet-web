import { useEffect, useMemo, useState } from 'react'
import { Image, Text, Textarea, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { getToken, post } from '../../request'
import { AIRecommendResp, RecommendItem } from '../../types'
import './index.css'

// 快捷标签：选中后以「；」拼入 requirement 一起提交
const CHIP_GROUPS: { label: string; chips: string[] }[] = [
  { label: '预算', chips: ['预算3000以内', '预算3000-8000', '预算8000-15000', '预算不限'] },
  { label: '家庭', chips: ['住公寓', '家里有院', '有小孩', '有老人'] },
  { label: '作息', chips: ['上班族', '在家时间多', '经常出差'] },
  { label: '经验', chips: ['第一次养宠', '有养宠经验'] },
]

export default function RecommendPage() {
  const [text, setText] = useState('')
  const [picked, setPicked] = useState<string[]>([])
  const [items, setItems] = useState<RecommendItem[]>()
  const [source, setSource] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({
        url: `/pages/login/index?redirect=${encodeURIComponent('/pages/recommend/index')}`,
      }).catch(() => {})
    }
  }, [])

  const requirement = useMemo(() => {
    const parts = [...picked, ...[text.trim()]].filter((s) => s !== '')
    return parts.join('；')
  }, [picked, text])

  const toggleChip = (chip: string) => {
    setPicked((prev) => (prev.includes(chip) ? prev.filter((c) => c !== chip) : [...prev, chip]))
  }

  const submit = async () => {
    if (loading) return
    if (requirement === '') {
      Taro.showToast({ title: '说说你的预算或养宠条件吧', icon: 'none' })
      return
    }
    setLoading(true)
    try {
      const resp = await post<AIRecommendResp>('/ai/recommend', { requirement })
      setItems(resp.items)
      setSource(resp.source)
      if (!resp.items || resp.items.length === 0) {
        Taro.showToast({ title: '暂时没有匹配的在售宠物', icon: 'none' })
      }
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setLoading(false)
    }
  }

  const goDetail = (id: string) => Taro.navigateTo({ url: `/pages/detail/index?id=${id}` })

  return (
    <View className='rcm'>
      <View className='rcm-form card'>
        <View className='rcm-title'>智能选宠</View>
        <View className='rcm-sub'>描述你的预算、家庭和生活习惯，为你挑 3 只最合适的</View>
        <View className='rcm-textarea-wrap'>
          <Textarea
            className='rcm-textarea'
            maxlength={500}
            placeholder='例如：预算5000以内，住公寓，上班族，第一次养宠，想要性格温顺的'
            value={text}
            onInput={(e) => setText(e.detail.value)}
          />
        </View>
        {CHIP_GROUPS.map((g) => (
          <View className='rcm-group' key={g.label}>
            <Text className='rcm-group-label'>{g.label}</Text>
            <View className='rcm-chips'>
              {g.chips.map((chip) => (
                <View
                  key={chip}
                  className={`rcm-chip ${picked.includes(chip) ? 'rcm-chip-on' : ''}`}
                  onClick={() => toggleChip(chip)}
                >
                  {chip}
                </View>
              ))}
            </View>
          </View>
        ))}
        <View className={`rcm-submit ${loading ? 'rcm-submit-off' : ''}`} onClick={submit}>
          {loading ? '正在为你挑选…' : '开始推荐'}
        </View>
      </View>

      {items && items.length > 0 && (
        <View className='rcm-results'>
          <View className='rcm-source'>{source === 'ai' ? '✦ AI 智能推荐' : '智能匹配（本地规则）'}</View>
          {items.map((it) => (
            <View className='rcm-item card' key={it.id} onClick={() => goDetail(it.id)}>
              <View className='rcm-score'>
                <Text className='rcm-score-num'>{it.score}</Text>
                <Text className='rcm-score-unit'>分</Text>
              </View>
              {it.mainImage ? (
                <Image className='rcm-img' src={it.mainImage} mode='aspectFill' />
              ) : (
                <View className='rcm-img'>🐾</View>
              )}
              <View className='rcm-main'>
                <Text className='rcm-pet-title'>{it.title}</Text>
                <Text className='rcm-breed'>
                  {it.breedName || '宠物'} · {it.petGender === 1 ? '公' : it.petGender === 2 ? '母' : '未知'}
                </Text>
                <Text className='rcm-reason'>{it.reason}</Text>
              </View>
              <Text className='price'>¥{it.price}</Text>
            </View>
          ))}
        </View>
      )}
    </View>
  )
}
