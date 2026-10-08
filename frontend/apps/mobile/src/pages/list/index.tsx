import { useCallback, useEffect, useRef, useState } from 'react'
import { Image, Input, Text, View } from '@tarojs/components'
import Taro, { usePullDownRefresh, useReachBottom, useRouter } from '@tarojs/taro'
import { del, get, getToken, post } from '../../request'
import { CategoryItem, PageResp, ProductCard, SearchHotRow, StringListResp } from '../../types'
import './index.css'
import TabBar from '../../components/TabBar'
import { ErrorState, Skeleton } from '../../components/Feedback'
import ProductCardItem from '../../components/ProductCard'

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
  const [listError, setListError] = useState('')
  // 搜索增强：热搜 / 历史 / 联想
  const [hots, setHots] = useState<SearchHotRow[]>([])
  const [hist, setHist] = useState<string[]>([])
  const [sugs, setSugs] = useState<string[]>([])
  const [panel, setPanel] = useState(false)
  const blurTimer = useRef<ReturnType<typeof setTimeout>>()
  const sugTimer = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    get<CategoryItem[]>('/categories').then(setCats).catch(() => {})
    get<{ list: SearchHotRow[] }>('/search/hot').then((r) => setHots(r?.list ?? [])).catch(() => {})
    if (getToken()) get<StringListResp>('/search/history').then((r) => setHist(r?.list ?? [])).catch(() => {})
    return () => {
      clearTimeout(blurTimer.current)
      clearTimeout(sugTimer.current)
    }
  }, [])

  const refreshHist = () => {
    if (getToken()) get<StringListResp>('/search/history').then((r) => setHist(r?.list ?? [])).catch(() => {})
  }

  const doSearch = (kw: string) => {
    const k = kw.trim()
    if (!k) return
    setPanel(false)
    setKeyword(k)
    post('/search/trace', { keyword: k }).catch(() => {})
    refreshHist()
    reSearch(k, categoryId, sort)
  }

  const onKeywordInput = (v: string) => {
    setKeyword(v)
    clearTimeout(sugTimer.current)
    if (!v.trim()) {
      setSugs([])
      return
    }
    sugTimer.current = setTimeout(() => {
      get<StringListResp>(`/search/complete?prefix=${encodeURIComponent(v.trim())}`)
        .then((r) => setSugs((r?.list ?? []).slice(0, 8)))
        .catch(() => setSugs([]))
    }, 300)
  }

  const clearHist = () => {
    del('/search/history').then(() => setHist([])).catch(() => {})
  }

  const load = useCallback(
    async (nextPage: number, append: boolean, kw = keyword, cid = categoryId, st = sort) => {
      setLoading(true)
      setListError('')
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
        setListError(nextPage === 1 ? '商品加载失败' : '更多内容加载失败')
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
        <Image
          className='lst-search-icon'
          src='data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%23C64120" stroke-width="2" stroke-linecap="round"><circle cx="11" cy="11" r="7"/><path d="M20 20l-3.5-3.5"/></svg>'
        />
        <Input
          className='lst-input'
          value={keyword}
          placeholder='搜索心仪的宠物'
          confirmType='search'
          onFocus={() => {
            clearTimeout(blurTimer.current)
            setPanel(true)
          }}
          onBlur={() => {
            blurTimer.current = setTimeout(() => setPanel(false), 200)
          }}
          onInput={(e) => onKeywordInput(e.detail.value)}
          onConfirm={() => doSearch(keyword)}
        />
        {panel && (sugs.length > 0 || hots.length > 0 || hist.length > 0) && (
          <View className='lst-panel'>
            {sugs.length > 0
              ? sugs.map((s) => (
                  <View className='lst-sug' key={s} onClick={() => doSearch(s)}>
                    {s}
                  </View>
                ))
              : (
                <>
                  {hist.length > 0 && (
                    <View className='lst-panel-sec'>
                      <View className='lst-panel-head'>
                        <Text>搜索历史</Text>
                        <Text className='lst-panel-clear' onClick={clearHist}>
                          清空
                        </Text>
                      </View>
                      <View className='lst-chips'>
                        {hist.map((h) => (
                          <View className='lst-chip' key={h} onClick={() => doSearch(h)}>
                            {h}
                          </View>
                        ))}
                      </View>
                    </View>
                  )}
                  {hots.length > 0 && (
                    <View className='lst-panel-sec'>
                      <View className='lst-panel-head'>
                        <Text>热搜榜</Text>
                      </View>
                      <View className='lst-chips'>
                        {hots.map((h, i) => (
                          <View className='lst-chip' key={h.keyword} onClick={() => doSearch(h.keyword)}>
                            <Text className={`lst-chip-rank ${i < 3 ? 'lst-chip-rank-hot' : ''}`}>{i + 1}</Text>
                            {h.keyword}
                          </View>
                        ))}
                      </View>
                    </View>
                  )}
                </>
              )}
          </View>
        )}
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

      {!loading && !listError && (
        <View className='lst-count'>
          <Text>共 {total} 只</Text>
        </View>
      )}

      {loading && list.length === 0 && <Skeleton variant='list' />}
      {!loading && listError && list.length === 0 && (
        <ErrorState
          title={listError}
          description={`已保留“${keyword || '全部关键词'}”和当前筛选`}
          onRetry={() => load(1, false)}
        />
      )}
      {!loading && !listError && list.length === 0 && <View className='lst-empty'>没有找到合适的宠物，换个条件试试～</View>}
      <View className='lst-grid'>
        {list.map((p) => (
          <ProductCardItem key={p.id} product={p} onOpen={() => goDetail(p)} />
        ))}
      </View>
      {listError && list.length > 0 && (
        <ErrorState title={listError} description="已保留当前列表和筛选条件" onRetry={() => load(page + 1, true)} />
      )}
      {!listError && list.length < total && (
        <View className='lst-more'>{loading ? '加载中…' : '上拉加载更多'}</View>
      )}
    <TabBar active='category' />
    </View>
  )
}
