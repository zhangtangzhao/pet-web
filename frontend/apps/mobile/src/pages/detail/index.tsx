import { useEffect, useState } from 'react'
import { Image, ScrollView, Swiper, SwiperItem, Text, Video, View } from '@tarojs/components'
import Taro, { useRouter, useShareAppMessage } from '@tarojs/taro'
import { del, get, getToken, post } from '../../request'
import { ErrorState, Skeleton } from '../../components/Feedback'
import { goBackOrRedirect } from '../../navigation'
import { PageResp, ProductCard, ProductDetail, ReviewListResp, ReviewSummaryResp, ReviewView } from '../../types'
import './index.css'

const HEART_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%235F6670" stroke-width="2" stroke-linecap="round"><path d="M12 20s-7-4.5-7-9.5A4 4 0 0 1 12 7a4 4 0 0 1 7 3.5C19 15.5 12 20 12 20z"/></svg>'
const HEART_ACTIVE_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%23C64120"><path d="M12 20s-7-4.5-7-9.5A4 4 0 0 1 12 7a4 4 0 0 1 7 3.5C19 15.5 12 20 12 20z"/></svg>'
const CHAT_ICON =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="%235F6670" stroke-width="2" stroke-linecap="round"><path d="M20 12a8 8 0 1 1-3-6.2L20 5l-.8 3.3A8 8 0 0 1 20 12z"/></svg>'
const EMPTY_PET_MEDIA =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 72"><rect width="96" height="72" fill="%23EFEBE5"/><g fill="%23CBC5BC"><ellipse cx="34" cy="26" rx="7" ry="9"/><ellipse cx="48" cy="22" rx="7" ry="9"/><ellipse cx="62" cy="26" rx="7" ry="9"/><path d="M48 34c10 0 18 8 18 17 0 8-8 13-18 13s-18-5-18-13c0-9 8-17 18-17z"/></g></svg>'

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
  const [activeId, setActiveId] = useState(params.id || '')
  const [d, setD] = useState<ProductDetail>()
  const [detailError, setDetailError] = useState('')
  const [favLoading, setFavLoading] = useState(false)
  const [summary, setSummary] = useState<ReviewSummaryResp>()
  const [rvOpen, setRvOpen] = useState(false)
  const [rvList, setRvList] = useState<ReviewView[]>([])
  const [rvCursor, setRvCursor] = useState('')
  const [rvHasMore, setRvHasMore] = useState(false)
  const [related, setRelated] = useState<ProductCard[]>([])
  const [skuId, setSkuId] = useState('')
  const [skuOpen, setSkuOpen] = useState(false)
  const [pendingSkuAction, setPendingSkuAction] = useState<'buy' | 'cart' | null>(null)

  const loadProduct = (id: string) => {
    setDetailError('')
    return get<ProductDetail>(`/products/${id}`)
      .then(setD)
      .catch(() => setDetailError('商品详情加载失败'))
  }

  useEffect(() => {
    if (!activeId) return
    loadProduct(activeId)
    get<ReviewSummaryResp>(`/products/${activeId}/review-summary`)
      .then(setSummary)
      .catch(() => {})
    get<PageResp<ProductCard>>(`/products/${activeId}/related`)
      .then((r) => setRelated(r?.list ?? []))
      .catch(() => {})
    // 缓存邀请码供分享带参（未登录静默跳过）
    if (getToken()) {
      get<{ inviteCode: string }>('/invite')
        .then((inv) => inv?.inviteCode && Taro.setStorageSync('pet_invite_code', inv.inviteCode))
        .catch(() => {})
    }
  }, [activeId])

  // 小程序分享：带上邀请码，好友注册自动归因
  useShareAppMessage(() => {
    const code = (Taro.getStorageSync('pet_invite_code') as string) || ''
    return {
      title: d?.title ? `宠物之家·${d.title}` : '宠物之家',
      path: `/pages/detail/index?id=${activeId}${code ? `&inviteCode=${code}` : ''}`,
      imageUrl: d?.mainImage || undefined,
    }
  })

  const loadReviews = (id: string, cursor: string) => {
    get<ReviewListResp>(`/products/${id}/reviews?limit=10${cursor ? `&cursor=${cursor}` : ''}`)
      .then((r) => {
        setRvList((prev) => (cursor ? [...prev, ...r.list] : r.list))
        setRvHasMore(r.hasMore)
        setRvCursor(r.list.length ? r.list[r.list.length - 1].id : cursor)
      })
      .catch(() => {})
  }

  const openReviews = () => {
    setRvOpen(true)
    if (rvList.length === 0) loadReviews(activeId, '')
  }

  const openRelated = (id: string) => {
    if (!id || id === activeId) return
    setD(undefined)
    setDetailError('')
    setSkuId('')
    setSkuOpen(false)
    setSummary(undefined)
    setRvOpen(false)
    setRvList([])
    setRvCursor('')
    setRvHasMore(false)
    setRelated([])
    setActiveId(id)
    Taro.pageScrollTo({ scrollTop: 0, duration: 200 }).catch(() => {})
  }

  const toggleFav = async () => {
    setFavLoading(true)
    try {
      if (d?.isFavorite) {
        await del(`/favorites/${activeId}`)
        setD({ ...d, isFavorite: false, favoriteCount: d.favoriteCount - 1 })
      } else {
        await post(`/favorites/${activeId}`)
        if (d) setD({ ...d, isFavorite: true, favoriteCount: d.favoriteCount + 1 })
      }
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setFavLoading(false)
    }
  }

  const buy = () => {
    if (d?.hasSku === 1 && !skuId) {
      setPendingSkuAction('buy')
      setSkuOpen(true)
      return
    }
    Taro.navigateTo({ url: `/pages/checkout/index?id=${activeId}${skuId ? `&skuId=${skuId}` : ''}` })
  }

  const addToCart = () => {
    if (d?.hasSku === 1 && !skuId) {
      setPendingSkuAction('cart')
      setSkuOpen(true)
      return
    }
    post('/cart', { productId: activeId, skuId })
      .then(() => Taro.showToast({ title: '已加入购物车', icon: 'success' }))
      .finally(() => closeSku())
      .catch((e: any) => Taro.showToast({ title: e.message, icon: 'none' }))
  }

  const closeSku = () => {
    setSkuOpen(false)
    setPendingSkuAction(null)
  }

  const renderNavbar = () => (
    <View className='detail-navbar'>
      <View className='detail-navbar-inner'>
        <View
          className='detail-back'
          onClick={() => goBackOrRedirect('/pages/index/index')}
        >
          <Text className='detail-back-icon'>←</Text>
          <Text className='detail-back-text'>返回</Text>
        </View>
      </View>
    </View>
  )

  const renderDesktopActions = (product: ProductDetail) => (
    <View className='detail-desktop-actions'>
      <View className='detail-secondary-actions'>
        <View className={`detail-secondary-action ${product.isFavorite ? 'fav-on' : ''}`} onClick={toggleFav}>
          <Image className='detail-action-icon' src={product.isFavorite ? HEART_ACTIVE_ICON : HEART_ICON} />
          <Text>{product.isFavorite ? '已收藏' : '收藏'}</Text>
        </View>
        <View className='detail-secondary-action' onClick={() => Taro.navigateTo({ url: '/pages/service-chat/index' })}>
          <Image className='detail-action-icon' src={CHAT_ICON} />
          <Text>客服</Text>
        </View>
      </View>
      <View className='btn-cart' onClick={addToCart}>加入购物车</View>
      <View className='btn-buy' onClick={buy}>立即购买</View>
    </View>
  )

  if (!d) {
    if (detailError) {
      return (
        <View className='detail'>
          {renderNavbar()}
          <View className='detail-feedback'>
            <ErrorState
              title="商品详情加载失败"
              description="已保留当前商品，可重试后继续选择规格"
              onRetry={() => loadProduct(activeId)}
            />
          </View>
        </View>
      )
    }
    return (
      <View className='detail'>
        {renderNavbar()}
        <ScrollView scrollY className='detail-scroll'>
          <Skeleton variant='detail' />
        </ScrollView>
      </View>
    )
  }

  const profile = d.petProfile
  const selectedSku = (d.skus ?? []).find((s) => s.id === skuId)
  const skuPrice = selectedSku ? selectedSku.price : d.price

  return (
    <View className='detail'>
      {renderNavbar()}
      <ScrollView scrollY className='detail-scroll'>
        <View className='detail-layout'>
          <View className='detail-gallery'>
            {d.images.length > 0 ? (
              <Swiper className='swp' indicatorDots circular>
                {d.images.map((u) => (
                  <SwiperItem key={u}>
                    <Image src={u} mode='aspectFill' className='swp-img' />
                  </SwiperItem>
                ))}
              </Swiper>
            ) : (
              <Image className='swp-img' src={EMPTY_PET_MEDIA} />
            )}
          </View>

          {!!d.videoUrl && (
            <View className='card video-card'>
              <Video src={d.videoUrl} poster={d.videoCover || undefined} className='video' controls />
            </View>
          )}

          <View className='detail-buy-panel'>
            <View className='head card'>
              <View className='price-row'>
                <Text className='price price-big'>¥{skuPrice}</Text>
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

            {d.hasSku === 1 && (d.skus ?? []).length > 0 && (
              <View className='card sku-card'>
                <View className='sku-title'>选择规格</View>
                <Text className='sku-selected'>{selectedSku ? `已选 ${selectedSku.specs}` : '请选择规格'}</Text>
                <View className='sku-chips'>
                  {(d.skus ?? []).map((s) => (
                    <View key={s.id} className={`sku-chip ${skuId === s.id ? 'sku-chip-on' : ''}`} onClick={() => setSkuId(s.id)}>
                      <Text>{s.specs}</Text>
                      <Text className='sku-chip-price'>¥{s.price}</Text>
                    </View>
                  ))}
                </View>
              </View>
            )}
            {renderDesktopActions(d)}
          </View>

          <View className='detail-main-content'>
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

            {(d.detailImages ?? []).length > 0 && (
              <View className='card detail-imgs'>
                {(d.detailImages ?? []).map((u) => (
                  <Image key={u} src={u} mode='widthFix' className='detail-img' onClick={() => Taro.previewImage({ urls: d.detailImages ?? [], current: u })} />
                ))}
              </View>
            )}

            {related.length > 0 && (
              <View className='related card'>
                <View className='related-title'>看了又看</View>
                <ScrollView scrollX className='related-scroll' enhanced showScrollbar={false}>
                  {related.map((p) => (
                    <View className='related-item' key={p.id} onClick={() => openRelated(p.id)}>
                      {p.mainImage ? (
                        <Image className='related-img' src={p.mainImage} mode='aspectFill' lazyLoad />
                      ) : (
                        <Image className='related-img' src={EMPTY_PET_MEDIA} />
                      )}
                      <Text className='related-name'>{p.title}</Text>
                      <Text className='related-price'>¥{p.price}</Text>
                    </View>
                  ))}
                </ScrollView>
              </View>
            )}
          </View>
        </View>
        <View className='detail-bottom-space' />
      </ScrollView>
      <View className='detail-actionbar'>
        <View className={`fav ${d.isFavorite ? 'fav-on' : ''}`} onClick={toggleFav}>
          <Image className='detail-action-icon' src={d.isFavorite ? HEART_ACTIVE_ICON : HEART_ICON} />
          <Text className='fav-text'>{d.isFavorite ? '已收藏' : '收藏'}</Text>
        </View>
        <View className='fav' onClick={() => Taro.navigateTo({ url: '/pages/service-chat/index' })}>
          <Image className='detail-action-icon' src={CHAT_ICON} />
          <Text className='fav-text'>客服</Text>
        </View>
        <View className='btn-cart' onClick={addToCart}>加购</View>
        <View className='btn-buy' onClick={buy}>立即购买</View>
      </View>
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
              onScrollToLower={() => rvHasMore && loadReviews(activeId, rvCursor)}
            >
              {rvList.map((r) => (
                <ReviewItem key={r.id} r={r} />
              ))}
              {rvHasMore ? (
                <View className='rvs-load' onClick={() => loadReviews(activeId, rvCursor)}>
                  加载更多
                </View>
              ) : (
                <View className='rvs-end'>没有更多了</View>
              )}
            </ScrollView>
          </View>
        </View>
      )}
      {skuOpen && (
        <View className='sku-sheet'>
          <View className='sku-sheet-mask' onClick={closeSku} />
          <View className='sku-sheet-body'>
            <View className='sku-sheet-head'>
              <Text className='sku-sheet-title'>选择规格</Text>
              <Text className='sku-sheet-close' onClick={closeSku}>关闭</Text>
            </View>
            <Text className='sku-sheet-price'>¥{skuPrice}</Text>
            <View className='sku-sheet-options'>
              {(d.skus ?? []).map((s) => (
                <View
                  key={s.id}
                  className={`sku-chip ${skuId === s.id ? 'sku-chip-on' : ''}`}
                  onClick={() => setSkuId(s.id)}
                >
                  <Text>{s.specs}</Text>
                  <Text className='sku-chip-price'>¥{s.price}</Text>
                </View>
              ))}
            </View>
            <View className='sku-sheet-confirm' onClick={() => {
              if (!skuId) {
                closeSku()
                return
              }
              const action = pendingSkuAction
              closeSku()
              if (action === 'buy') buy()
              else if (action === 'cart') addToCart()
            }}>
              {skuId ? '完成选择' : '暂不选择'}
            </View>
          </View>
        </View>
      )}
    </View>
  )
}
