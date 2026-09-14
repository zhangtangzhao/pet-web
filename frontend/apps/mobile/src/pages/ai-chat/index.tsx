import { useEffect, useRef, useState } from 'react'
import { Image, Input, ScrollView, Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, post } from '../../request'
import { AIAskResp, AIChatMessage, ProductDetail } from '../../types'
import './index.css'

interface ChatItem extends AIChatMessage {
  refs?: { title: string }[]
  demo?: boolean
}

const QUICK_QUESTIONS = ['它的性格适合我家吗', '日常怎么喂养照顾', '疫苗和健康情况怎么样', '这个品种有什么特点']

export default function AiChat() {
  const { params } = useRouter()
  const [d, setD] = useState<ProductDetail>()
  const [items, setItems] = useState<ChatItem[]>([])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [replyTo, setReplyTo] = useState('')
  const endRef = useRef('')

  useEffect(() => {
    get<ProductDetail>(`/products/${params.id}`)
      .then(setD)
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))
  }, [params.id])

  useEffect(() => {
    endRef.current = `msg-${items.length}`
  }, [items])

  const ask = async (q: string) => {
    const question = q.trim()
    if (!question || loading) return
    setInput('')
    setReplyTo(question)
    const history: AIChatMessage[] = items
      .filter((m) => m.role === 'user' || m.role === 'assistant')
      .slice(-6)
      .map((m) => ({ role: m.role, content: m.content }))
    setItems((prev) => [...prev, { role: 'user', content: question }])
    setLoading(true)
    try {
      const resp = await post<AIAskResp>('/ai/ask', { productId: params.id, question, history })
      setItems((prev) => [...prev, { role: 'assistant', content: resp.answer, refs: resp.refs, demo: resp.demo }])
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setLoading(false)
      setReplyTo('')
    }
  }

  return (
    <View className='ai'>
      {d && (
        <View className='ai-pet card'>
          {d.mainImage ? (
            <Image src={d.mainImage} mode='aspectFill' className='ai-pet-img' />
          ) : (
            <View className='ai-pet-img ai-pet-empty'>🐾</View>
          )}
          <View className='ai-pet-info'>
            <Text className='ai-pet-title'>{d.title}</Text>
            <Text className='ai-pet-sub'>
              {d.breedName} · ¥{d.price}
            </Text>
          </View>
        </View>
      )}

      <ScrollView scrollY className='ai-list' scrollIntoView={endRef.current} scrollWithAnimation>
        <View className='ai-tip'>我是 AI 宠物顾问「小宠」，可以基于平台知识库和这只宠物的档案为你解答～</View>

        {items.length === 0 && (
          <View className='ai-quick'>
            {QUICK_QUESTIONS.map((q) => (
              <View key={q} className='ai-chip' onClick={() => ask(q)}>
                {q}
              </View>
            ))}
          </View>
        )}

        {items.map((m, i) =>
          m.role === 'user' ? (
            <View key={i} className='row-user' id={`msg-${i}`}>
              <View className='bubble bubble-user'>{m.content}</View>
            </View>
          ) : (
            <View key={i} className='row-ai' id={`msg-${i}`}>
              <View className='bubble bubble-ai'>{m.content}</View>
              {m.refs && m.refs.length > 0 && (
                <View className='refs'>
                  参考：
                  {m.refs.map((r) => (
                    <Text key={r.title} className='ref-tag'>
                      {r.title}
                    </Text>
                  ))}
                </View>
              )}
            </View>
          ),
        )}

        {loading && (
          <View className='row-ai'>
            <View className='bubble bubble-ai bubble-typing'>
              {replyTo ? `正在分析「${replyTo}」…` : '思考中…'}
            </View>
          </View>
        )}
        <View className='ai-bottom-space' />
      </ScrollView>

      <View className='ai-input-bar'>
        <View className='ai-human' onClick={() => Taro.navigateTo({ url: '/pages/service-chat/index' })}>
          转人工
        </View>
        <Input
          className='ai-input'
          value={input}
          placeholder='想了解这只宠物的什么？'
          onInput={(e) => setInput(e.detail.value)}
          onConfirm={() => ask(input)}
          confirmType='send'
        />
        <View className={`ai-send ${input.trim() ? 'ai-send-on' : ''}`} onClick={() => ask(input)}>
          发送
        </View>
      </View>
    </View>
  )
}
