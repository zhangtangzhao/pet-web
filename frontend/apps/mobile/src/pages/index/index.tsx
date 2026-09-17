import { useEffect, useState } from 'react'
import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get } from '../../request'
import { ProductCard } from '../../types'
import TabBar from '../../components/TabBar'
import './index.css'

const CATS = ['全部', '猫', '狗', '鸟', '异宠']
const TABS = [
  { key: 'home', icon: '🏠', label: '首页' },
  { key: 'category', icon: '📋', label: '分类' },
  { key: 'message', icon: '💬', label: '消息' },
  { key: 'me', icon: '👤', label: '我的' },
]

export default function Home() {
  const [cat, setCat] = useState('全部')
  const [products, setProducts] = useState<ProductCard[]>([])
  const [loading, setLoading] = useState(false)

  const load = (category?: string) => {
    setLoading(true)
    const qs = category && category !== '全部' ? `&categoryId=` : ''
    get<{ list: ProductCard[] }>(`/products?pageSize=20${qs}`)
      .then((r) => setProducts(r?.list ?? []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }

  useEffect(() => { load(cat) }, [cat])

  usePullDownRefresh(() => { load(cat); Taro.stopPullDownRefresh() })

  const goDetail = (id: string) => Taro.navigateTo({ url: `/pages/detail/index?id=${id}` }).catch(() => {})
  const goSearch = () => Taro.switchTab({ url: '/pages/list/index' }).catch(() => {})
  const goTab = (key: string) => {
    const map: Record<string, string> = {
      category: '/pages/list/index',
      message: '/pages/notify/index',
      me: '/pages/me/index',
    }
    if (map[key]) Taro.navigateTo({ url: map[key] }).catch(() => {})
  }

  // Split into two columns for waterfall
  const colL = products.filter((_, i) => i % 2 === 0)
  const colR = products.filter((_, i) => i % 2 === 1)

  const Card = (p: ProductCard) => (
    <View className='pc-card' onClick={() => goDetail(p.id)}>
      <View className='pc-img-wrap'>
        {p.mainImage ? <Image className='pc-img' src={p.mainImage} mode='aspectFill' lazyLoad /> : <View className='pc-img pc-img-empty'>🐾</View>}
      </View>
      <View className='pc-body'>
        <Text className='pc-title'>{p.title}</Text>
        <View className='pc-bottom'>
          <Text className='pc-price'>¥{Number(p.price).toFixed(2)}</Text>
          <Text className='pc-sales'>已售{p.sales}</Text>
        </View>
      </View>
    </View>
  )

  return (
    <View className='home'>
      {/* 搜索框 */}
      <View className='home-search' onClick={goSearch}>
        <Text className='home-search-icon'>🔍</Text>
        <Input className='home-search-input' placeholder='搜索心仪的宠物' disabled />
      </View>

      {/* 分类标签 */}
      <View className='home-cats'>
        {CATS.map((c) => (
          <View key={c} className={`home-cat ${cat === c ? 'home-cat-on' : ''}`} onClick={() => setCat(c)}>
            <Text>{c}</Text>
          </View>
        ))}
      </View>

      {/* 商品瀑布流 */}
      {loading ? (
        <View className='home-loading'><Text>加载中…</Text></View>
      ) : (
        <View className='home-grid'>
          <View className='home-col'>{colL.map(Card)}</View>
          <View className='home-col'>{colR.map(Card)}</View>
        </View>
      )}
      {!loading && products.length === 0 && <View className='home-empty'><Text>暂无商品</Text></View>}

      <TabBar active='home' />
    </View>
  )
}