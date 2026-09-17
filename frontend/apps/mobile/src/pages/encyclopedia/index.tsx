import { useCallback, useEffect, useState } from 'react'
import { Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { get } from '../../request'
import './index.css'

interface Breed { breedId: string; breedName: string; cover: string; articleCnt: number }
interface Article { title: string; content: string }

export default function Encyclopedia() {
  const [breeds, setBreeds] = useState<Breed[]>([])
  const [current, setCurrent] = useState<Breed>()
  const [articles, setArticles] = useState<Article[]>([])

  useEffect(() => {
    get<Breed[]>('/encyclopedia').then(setBreeds).catch(() => {})
  }, [])

  const open = useCallback((b: Breed) => {
    setCurrent(b)
    get<Article[]>('/encyclopedia/' + b.breedId).then(setArticles).catch(() => {})
  }, [])

  return (
    <View className='enc'>
      <View className='enc-title'>养宠百科</View>
      <View className='enc-chips'>
        {breeds.map((b) => (
          <View key={b.breedId} className={'enc-chip' + (current?.breedId === b.breedId ? ' enc-chip-on' : '')} onClick={() => open(b)}>
            {b.breedName}（{b.articleCnt}）
          </View>
        ))}
      </View>
      {articles.map((a) => (
        <View className='enc-card' key={a.title}>
          <Text className='enc-card-title'>{a.title}</Text>
          <Text className='enc-card-content'>{a.content}</Text>
        </View>
      ))}
      {!current && breeds.length > 0 && <View className='enc-tip'>选择品种查看养护知识</View>}
    </View>
  )
}
