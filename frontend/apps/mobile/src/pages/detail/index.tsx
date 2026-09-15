import { useEffect, useState } from 'react'
import { Image, ScrollView, Swiper, SwiperItem, Text, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { del, get, post } from '../../request'
import { ProductDetail, ReviewListResp, ReviewSummaryResp, ReviewView } from '../../types'
import './index.css'

function ReviewItem({ r }: { r: ReviewView }) {
  return (
    <View className='rvs-item'>
      <View className='rvs-item-head'>
        <Image src={r.avatar} className='rvs-avatar' />
        <Text className='rvs-name'>{r.nickname || '匿名用户'}</Text>
        <Text className='rvs-item-stars'>{'★'.repeat(r.rating)}</Text>
      </View>
      {r.content && <Text className='rvs-content'>{r.content}</Text>}
      {r.images.length > 0 && (
        <View className='rvs-imgs'>
          {r.images.map((u) => (
            <Image
              key={u}
              src={u}
              mode='aspectFill'
              className='rvs-img'
              onClick={() => Taro.previewImage({ urls: r.images, current: u })}
            />
          ))}
        </View>
      )}
      {r.reply && (
        <View className='rvs-reply'>
          <Text className='rvs-reply-tag'>商家回复</Text>
          <Text className='rvs-reply-text'>{r.reply}</Text>
        </View>
      )}
      <Text className='rvs-time'>{r.createdAt.slice(0, 16).replace('T', ' ')}</Text>
    </View>
  )
}

export default function Detail() {
  const { params } = useRouter()
  const [d, setD] = useState<ProductDetail>()
  const [favLoading, setFavLoading] = useState(false)
  const [summary, setSummary] = useState<ReviewSummaryResp>()
  const [rvOpen, setRvOpen] = useState(false)
  const [rvList, setRvList] = useState<ReviewView[]>([])
  const [rvCursor, setRvCursor] = useState('')
  const [rvHasMore, setRvHasMore] = useState(false)

  const load = () =>
    get<ProductDetail>(`/products/${params.id}`)
      .then(setD)
      .catch((e) => Taro.showToast({ title: e.message, icon: 'none' }))

  useEffect(() => {
    load()
    get<ReviewSummaryResp>(`/products/${params.id}/review-summary`)
      .then(setSummary)
      .catch(() => {})
  }, [params.id])

  const openReviews = () => {
    setRvOpen(true)
    if (rvList.length === 0) loadReviews('')
  }

  const loadReviews = (cursor: string) => {
    get<ReviewListResp>(`/products/${params.id}/reviews?limit=10${cursor ? `&cursor=${cursor}` : ''}`)
      .then((r) => {
        setRvList((prev) => (cursor ? [...prev, ...r.list] : r.list))
        setRvHasMore(r.hasMore)
        setRvCursor(r.list.length ? r.list[r.list.length - 1].id : cursor)
      })
      .catch(() => {})
  }

  const toggleFav = async () => {
    setFavLoading(true)
    try {
      if (d?.isFavorite) {
        await del(`/favorites/${params.id}`)
        setD({ ...d, isFavorite: false, favoriteCount: d.favoriteCount - 1 })
      } else {
        await post(`/favorites/${params.id}`)
        if (d) setD({ ...d, isFavorite: true, favoriteCount: d.favoriteCount + 1 })
      }
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setFavLoading(false)
    }
  }

  const buy = () => {
    Taro.navigateTo({ url: `/pages/checkout/index?id=${params.id}` })
  }

  if (!d) return <View className='detail-loading'>加载中…</View>

  const profile = d.petProfile

  return (
    <View className='detail'>
      <ScrollView scrollY className='detail-scroll'>
        {d.images.length > 0 ? (
          <Swiper className='swp' indicatorDots circular>
            {d.images.map((u) => (
              <SwiperItem key={u}>
                <Image src={u} mode='aspectFill' className='swp-img' />
              </SwiperItem>
            ))}
          </Swiper>
        ) : (
          <View className='swp swp-empty'>🐾</View>
        )}

        <View className='head card'>
          <View className='price-row'>
            <Text className='price price-big'>¥{d.price}</Text>
            {Number(d.originalPrice) > Number(d.price) && (
              <Text className='orig'>¥{d.originalPrice}</Text>
            )}
          </View>
          <Text className='title'>{d.title}</Text>
          <View className='stats'>
            <Text>已售 {d.sales}</Text>
            <Text>{d.favoriteCount} 人收藏</Text>
          </View>
        </View>

        <View className='profile card'>
          <View className='profile-title'>宠物档案</View>
          <View className='profile-grid'>
            <View className='p-item'>
              <Text className='p-label'>性别</Text>
              <Text>{profile.genderText}</Text>
            </View>
            <View className='p-item'>
              <Text className='p-label'>年龄</Text>
              <Text>{profile.ageText || '未知'}</Text>
            </View>
            <View className='p-item'>
              <Text className='p-label'>体型</Text>
              <Text>{profile.bodyType || '未知'}</Text>
            </View>
            <View className='p-item'>
              <Text className='p-label'>毛色</Text>
              <Text>{profile.coatColor || '未知'}</Text>
            </View>
            <View className='p-item'>
              <Text className='p-label'>疫苗</Text>
              <Text>{profile.vaccineDesc || '未知'}</Text>
            </View>
            <View className='p-item'>
              <Text className='p-label'>驱虫</Text>
              <Text>{profile.dewormDesc || '未知'}</Text>
            </View>
          </View>
          {profile.personality && (
            <View className='p-line'>
              <Text className='p-label'>性格：</Text>
              <Text>{profile.personality}</Text>
            </View>
          )}
          {d.aiEnabled && (
            <View className='ai-entry' onClick={() => Taro.navigateTo({ url: `/pages/ai-chat/index?id=${params.id}` })}>
              <View className='ai-entry-main'>
                <Text className='ai-entry-title'>AI 智能分析</Text>
                <Text className='ai-entry-sub'>向 AI 顾问了解这只宠物的品种习性与养护要点</Text>
              </View>
              <Text className='ai-entry-arrow'>→</Text>
            </View>
          )}
        </View>

        {summary && summary.total > 0 && (
          <View className='reviews card'>
            <View className='rvs-head' onClick={openReviews}>
              <Text className='rvs-title'>用户评价（{summary.total}）</Text>
              <Text className='rvs-more'>全部评价 ›</Text>
            </View>
            <View className='rvs-score'>
              <Text className='rvs-avg'>{summary.avgRating}</Text>
              <Text className='rvs-stars'>{'★'.repeat(Math.round(Number(summary.avgRating)))}</Text>
            </View>
            {summary.latest.map((r) => (
              <ReviewItem key={r.id} r={r} />
            ))}
          </View>
        )}

        {d.detailHtml && <View className='rich card'>{d.detailHtml}</View>}
        <View className='detail-bottom-space' />
      </ScrollView>

      {rvOpen && (
        <View className='rvs-sheet'>
          <View className='rvs-sheet-mask' onClick={() => setRvOpen(false)} />
          <View className='rvs-sheet-body'>
            <View className='rvs-sheet-head'>
              <Text className='rvs-sheet-title'>全部评价（{summary?.total ?? 0}）</Text>
              <Text className='rvs-sheet-close' onClick={() => setRvOpen(false)}>
                ✕
              </Text>
            </View>
            <ScrollView
              scrollY
              className='rvs-sheet-scroll'
              onScrollToLower={() => rvHasMore && loadReviews(rvCursor)}
            >
              {rvList.map((r) => (
                <ReviewItem key={r.id} r={r} />
              ))}
              {rvHasMore ? (
                <View className='rvs-load' onClick={() => loadReviews(rvCursor)}>
                  加载更多
                </View>
              ) : (
                <View className='rvs-end'>没有更多了</View>
              )}
            </ScrollView>
          </View>
        </View>
      )}

      <View className='footer'>
        <View className={`fav ${d.isFavorite ? 'fav-on' : ''}`} onClick={toggleFav}>
          <Text>{d.isFavorite ? '♥' : '♡'}</Text>
          <Text className='fav-text'>收藏</Text>
        </View>
        <View className='fav' onClick={() => Taro.navigateTo({ url: '/pages/service-chat/index' })}>
          <Text>💬</Text>
          <Text className='fav-text'>客服</Text>
        </View>
        <View className='btn-buy' onClick={!favLoading ? buy : undefined}>
          立即购买
        </View>
      </View>
    </View>
  )
}
