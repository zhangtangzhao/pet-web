import { Text, View } from '@tarojs/components'
import './Feedback.css'

type SkeletonVariant = 'home' | 'list' | 'detail' | 'cart' | 'checkout' | 'order'

function SkeletonLines({ widths }: { widths: string[] }) {
  return (
    <View className='fb-lines'>
      {widths.map((width, index) => (
        <View className='fb-line' key={index} style={{ width }} />
      ))}
    </View>
  )
}

export function Skeleton({ variant }: { variant: SkeletonVariant }) {
  if (variant === 'home' || variant === 'list') {
    return (
      <View className={`fb-grid ${variant === 'home' ? 'fb-grid-home' : 'fb-grid-list'}`}>
        {Array.from({ length: variant === 'home' ? 4 : 6 }).map((_, index) => (
          <View className='fb-card' key={index}>
            <View className='fb-media' />
            <SkeletonLines widths={['88%', '64%', '44%']} />
          </View>
        ))}
      </View>
    )
  }

  if (variant === 'detail') {
    return (
      <View className='fb-detail'>
        <View className='fb-hero' />
        <View className='fb-card fb-panel'>
          <SkeletonLines widths={['42%', '84%', '48%', '52%']} />
        </View>
        <View className='fb-card fb-panel'>
          <View className='fb-grid-2'>
            {Array.from({ length: 4 }).map((_, index) => (
              <SkeletonLines key={index} widths={['32%', '64%']} />
            ))}
          </View>
        </View>
      </View>
    )
  }

  if (variant === 'cart') {
    return (
      <View className='fb-list-rows'>
        {Array.from({ length: 3 }).map((_, index) => (
          <View className='fb-card fb-row' key={index}>
            <View className='fb-thumb' />
            <SkeletonLines widths={['76%', '48%', '32%']} />
          </View>
        ))}
      </View>
    )
  }

  return (
    <View className='fb-form'>
      {Array.from({ length: 3 }).map((_, index) => (
        <View className='fb-card fb-panel' key={index}>
          <SkeletonLines widths={['32%', index % 2 ? '58%' : '84%', '44%']} />
        </View>
      ))}
    </View>
  )
}

export function ErrorState({
  title = '内容加载失败',
  description = '网络连接不稳定，请稍后重试',
  retryText = '重试',
  onRetry,
}: {
  title?: string
  description?: string
  retryText?: string
  onRetry: () => void
}) {
  return (
    <View className='fb-error'>
      <Text className='fb-error-title'>{title}</Text>
      <Text className='fb-error-desc'>{description}</Text>
      <View className='fb-retry' onClick={onRetry}>{retryText}</View>
    </View>
  )
}
