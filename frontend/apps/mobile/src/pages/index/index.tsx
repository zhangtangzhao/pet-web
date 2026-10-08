import { useEffect, useState } from 'react'
import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh } from '@tarojs/taro'
import { get } from '../../request'
import { CategoryItem, ProductCard } from '../../types'
import TabBar from '../../components/TabBar'
import { ErrorState, Skeleton } from '../../components/Feedback'
import ProductCardItem from '../../components/ProductCard'
import './index.css'

const SEARCH_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%23C64120" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/></svg>'
const PROFILE_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%232B6A8C" stroke-width="2" stroke-linecap="round"><path d="M12 4l7 3v5c0 4-3 7-7 8-4-1-7-4-7-8V7z"/></svg>'
const CONTRACT_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%232B6A8C" stroke-width="2" stroke-linecap="round"><path d="M7 3h10v18H7z"/><path d="M10 8h4M10 12h4M10 16h4"/></svg>'
const ORDER_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%232B6A8C" stroke-width="2" stroke-linecap="round"><path d="M5 5h14l-1 14H6z"/><path d="M9 9h6M9 13h6"/></svg>'

export default function Home() {
  const [cats, setCats] = useState<CategoryItem[]>([])
  const [catId, setCatId] = useState('')
  const [products, setProducts] = useState<ProductCard[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const load = (category?: string) => {
    setLoading(true)
    setError('')
    const qs = category ? `&categoryId=${encodeURIComponent(category)}` : ''
    get<{ list: ProductCard[] }>(`/products?pageSize=20${qs}`)
      .then((r) => setProducts(r?.list ?? []))
      .catch(() => setError('商品加载失败'))
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    get<CategoryItem[]>('/categories')
      .then((list) => setCats((list ?? []).filter((item) => item.id && item.name)))
      .catch(() => setCats([]))
  }, [])

  useEffect(() => { load(catId) }, [catId])

  usePullDownRefresh(() => { load(catId); Taro.stopPullDownRefresh() })

  const goDetail = (id: string) => Taro.navigateTo({ url: `/pages/detail/index?id=${id}` }).catch(() => {})
  const goSearch = () => Taro.navigateTo({ url: '/pages/list/index' }).catch(() => {})

  // Split into two columns for waterfall
  const colL = products.filter((_, i) => i % 2 === 0)
  const colR = products.filter((_, i) => i % 2 === 1)

  return (
    <View className='home'>
      <View className='home-search' onClick={goSearch}>
        <Image className='home-search-icon' src={SEARCH_ICON} />
        <Input className='home-search-input' placeholder='搜索心仪的宠物' disabled />
      </View>

      <View className='home-assurance'>
        <View className='assurance-item'><Image className='assurance-icon' src={PROFILE_ICON} /><Text>宠物档案</Text></View>
        <View className='assurance-item'><Image className='assurance-icon' src={CONTRACT_ICON} /><Text>交易协议</Text></View>
        <View className='assurance-item'><Image className='assurance-icon' src={ORDER_ICON} /><Text>订单状态</Text></View>
      </View>

      <View className='home-section-head'>
        <Text className='home-section-title'>精选在售宠物</Text>
        <Text className='home-view-all' onClick={goSearch}>查看全部</Text>
      </View>

      <View className='home-cats'>
        {[{ id: '', name: '全部' }, ...cats].map((c) => (
          <View key={c.id || 'all'} className={`home-cat ${catId === c.id ? 'home-cat-on' : ''}`} onClick={() => setCatId(c.id)}>
            <Text>{c.name}</Text>
          </View>
        ))}
      </View>

      {loading ? (
        <Skeleton variant='home' />
      ) : error ? (
        <ErrorState
          title='商品加载失败'
          description="请检查网络后重试，已保留当前分类"
          onRetry={() => load(catId)}
        />
      ) : (
        <View className='home-grid'>
          <View className='home-col'>{colL.map((p) => <ProductCardItem key={p.id} product={p} tone='home' onOpen={() => goDetail(p.id)} />)}</View>
          <View className='home-col'>{colR.map((p) => <ProductCardItem key={p.id} product={p} tone='home' onOpen={() => goDetail(p.id)} />)}</View>
        </View>
      )}
      {!loading && !error && products.length === 0 && <View className='home-state'><Text>暂无商品</Text></View>}

      <TabBar active='home' />
    </View>
  )
}
