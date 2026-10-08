import { Image, Text, View } from '@tarojs/components'
import './ProductCard.css'
import type { ProductCard as ProductCardData } from '../types'

const EMPTY_PET =
  'data:image/svg+xml;utf8,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 96 72"><rect width="96" height="72" fill="%23EFEBE5"/><g fill="%23CBC5BC"><ellipse cx="34" cy="26" rx="7" ry="9"/><ellipse cx="48" cy="22" rx="7" ry="9"/><ellipse cx="62" cy="26" rx="7" ry="9"/><path d="M48 34c10 0 18 8 18 17 0 8-8 13-18 13s-18-5-18-13c0-9 8-17 18-17z"/></g></svg>'

const genderText = (value: number) => (value === 1 ? '公' : value === 2 ? '母' : '')

export default function ProductCardItem({
  product,
  onOpen,
  tone = 'list',
}: {
  product: ProductCardData
  onOpen: () => void
  tone?: 'home' | 'list'
}) {
  const gender = genderText(product.petGender)
  const showOriginal = Number(product.originalPrice) > Number(product.price)

  return (
    <View className={`pet-card ${tone === 'home' ? 'pet-card-home' : ''}`} onClick={onOpen}>
      <View className='pet-media'>
        <Image
          className='pet-media-img'
          src={product.mainImage || EMPTY_PET}
          mode='aspectFill'
          lazyLoad
        />
      </View>
      <View className='pet-body'>
        <Text className='pet-title'>{product.title}</Text>
        <View className='pet-tags'>
          {product.breedName && <Text className='pet-tag'>{product.breedName}</Text>}
          {gender && <Text className='pet-tag'>{gender}</Text>}
        </View>
        <View className='pet-price-row'>
          <Text className='pet-price'>¥{Number(product.price).toFixed(2)}</Text>
          {showOriginal && <Text className='pet-original'>¥{Number(product.originalPrice).toFixed(2)}</Text>}
        </View>
        <View className='pet-meta'>
          {product.sales > 0 && <Text>已售 {product.sales}</Text>}
          {product.favoriteCount > 0 && <Text>{product.favoriteCount} 人收藏</Text>}
        </View>
      </View>
    </View>
  )
}
