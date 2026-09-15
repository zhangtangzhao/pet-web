import { useCallback, useEffect, useState } from 'react'
import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh, useReachBottom, useRouter } from '@tarojs/taro'
import { get } from '../../request'
import { CategoryItem, ProductCard, PageResp } from '../../types'
import './index.css'

const SORTS = [
  { v: '', label: '默认' },
  { v: 'sales', label: '销量' },
  { v: 'price_asc', label: '价格↑' },
  { v: 'price_desc', label: '价格↓' },
]

const PAGE_SIZE = 10

export default function List() {
  const { params } = useRouter()
  const [keyword, setKeyword] = useState(params.keyword || '')
  const [categoryId, setCategoryId] = useState(params.categoryId || '')
  const [cats, setCats] = useState<CategoryItem[]>([])
  const [sort, setSort] = useState('')
  const [list, setList] = useState<ProductCard[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    get<CategoryItem[]>('/categories').then(setCats).catch(() => {})
  }, [])

  const load = useCallback(
    async (nextPage: number, append: boolean, kw = keyword, cid = categoryId, st = sort) => {
      setLoading(true)
      try {
        const qs = new URLSearchParams({ page: String(nextPage), pageSize: String(PAGE_SIZE) })
        if (kw) qs.set('keyword', kw)
        if (cid) qs.set('categoryId', cid)
        if (st) qs.set('sort', st)
        const resp = await get<PageResp<ProductCard>>(`/products?${qs.toString()}`)
        setList((prev) => (append ? [...prev, ...resp.list] : resp.list))
        setTotal(resp.total)
        setPage(nextPage)
      } catch (e: any) {
        Taro.showToast({ title: e.message, icon: 'none' })
      } finally {
        setLoading(false)
      }
    },
    [keyword, categoryId, sort],
  )

  useEffect(() => {
    load(1, false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  usePullDownRefresh(() => {
    load(1, false).finally(() => Taro.stopPullDownRefresh())
  })

  useReachBottom(() => {
    if (loading || list.length >= total) return
    load(page + 1, true)
  })

  const reSearch = (kw: string, cid: string, st: string) => {
    setKeyword(kw)
    setCategoryId(cid)
    setSort(st)
    load(1, false, kw, cid, st)
  }

  const goDetail = (p: ProductCard) => Taro.navigateTo({ url: `/pages/detail/index?id=${p.id}` })

  return (
    <View className='lst'>
      <View className='lst-search'>
        <Text className='lst-search-icon'>🔍</Text>
        <Input
          className='lst-input'
          value={keyword}
          placeholder='搜索心仪的宠物'
          confirmType='search'
          onInput={(e) => setKeyword(e.detail.value)}
          onConfirm={() => reSearch(keyword, categoryId, sort)}
        />
      </View>

      <View className='lst-cats'>
        <View className={`lst-cat ${categoryId === '' ? 'lst-cat-on' : ''}`} onClick={() => reSearch(keyword, '', sort)}>
          全部
        </View>
        {cats.map((c) => (
          <View
            key={c.id}
            className={`lst-cat ${categoryId === c.id ? 'lst-cat-on' : ''}`}
            onClick={() => reSearch(keyword, c.id, sort)}
          >
            {c.name}
          </View>
        ))}
      </View>

      <View className='lst-sorts'>
        {SORTS.map((s) => (
          <View key={s.v} className={`lst-sort ${sort === s.v ? 'lst-sort-on' : ''}`} onClick={() => reSearch(keyword, categoryId, s.v)}>
            {s.label}
          </View>
        ))}
      </View>

      {list.length === 0 && !loading && <View className='lst-empty'>没有找到合适的宠物，换个条件试试～</View>}
      <View className='lst-grid'>
        {list.map((p) => (
          <View className='lst-card card' key={p.id} onClick={() => goDetail(p)}>
            {p.mainImage ? (
              <Image className='lst-img' src={p.mainImage} mode='aspectFill' lazyLoad />
            ) : (
              <View className='lst-img lst-img-empty'>🐾</View>
            )}
            <View className='lst-body'>
              <Text className='lst-title'>{p.title}</Text>
              <View className='lst-meta'>
                <Text className='price'>¥{p.price}</Text>
                <Text className='lst-sales'>已售{p.sales}</Text>
              </View>
            </View>
          </View>
        ))}
      </View>
      {list.length < total && (
        <View className='lst-more'>{loading ? '加载中…' : '上拉加载更多'}</View>
      )}
    </View>
  )
}
