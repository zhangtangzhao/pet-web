import { useEffect, useState } from 'react'
import { Image, Text, Textarea, View } from '@tarojs/components'
import Taro, { useRouter } from '@tarojs/taro'
import { get, post } from '../../request'
import { OrderView, UploadTokenResp } from '../../types'
import './index.css'

const MAX_IMAGES = 9

export default function OrderReview() {
  const { params } = useRouter()
  const [order, setOrder] = useState<OrderView>()
  const [rating, setRating] = useState(5)
  const [content, setContent] = useState('')
  const [images, setImages] = useState<string[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [healthScore, setHealthScore] = useState(5)
  const [lookScore, setLookScore] = useState(5)
  const [serviceScore, setServiceScore] = useState(5)

  useEffect(() => {
    get<OrderView>(`/orders/${params.orderNo}`).then(setOrder).catch(() => {})
  }, [params.orderNo])

  const uploadImage = async (filePath: string): Promise<string> => {
    const t = await post<UploadTokenResp>('/upload-token', { dir: 'chat', contentType: 'image/jpeg' })
    if (process.env.TARO_ENV === 'h5') {
      const blob = await fetch(filePath).then((r) => r.blob())
      const resp = await fetch(t.uploadUrl, { method: 'PUT', body: blob })
      if (!resp.ok) throw new Error('图片上传失败')
    } else {
      const buf = Taro.getFileSystemManager().readFileSync(filePath) as ArrayBuffer
      const resp = await Taro.request({
        url: t.uploadUrl,
        method: 'PUT',
        data: buf,
        header: { 'Content-Type': 'image/jpeg' },
      })
      if (resp.statusCode >= 300) throw new Error('图片上传失败')
    }
    return t.fileUrl
  }

  const chooseImage = () => {
    if (images.length >= MAX_IMAGES) {
      Taro.showToast({ title: `最多 ${MAX_IMAGES} 张图片`, icon: 'none' })
      return
    }
    Taro.chooseImage({ count: MAX_IMAGES - images.length })
      .then((res) => {
        const paths = res.tempFilePaths ?? []
        Promise.all(paths.map((p) => uploadImage(p)))
          .then((urls) => setImages((prev) => [...prev, ...urls]))
          .catch((e: Error) => Taro.showToast({ title: e.message || '图片上传失败', icon: 'none' }))
      })
      .catch(() => {})
  }

  const submit = async () => {
    if (submitting) return
    if (content.trim() === '' && images.length === 0) {
      Taro.showToast({ title: '写下评价或晒图吧', icon: 'none' })
      return
    }
    setSubmitting(true)
    try {
      await post(`/orders/${params.orderNo}/review`, {
        rating,
        content: content.trim(),
        images,
        healthScore,
        lookScore,
        serviceScore,
      })
      Taro.showToast({ title: '评价成功', icon: 'success' })
      setTimeout(() => Taro.navigateBack().catch(() => {}), 800)
    } catch (e: any) {
      Taro.showToast({ title: e.message, icon: 'none' })
    } finally {
      setSubmitting(false)
    }
  }

  const item = order?.items?.[0]

  return (
    <View className='rv'>
      {item && (
        <View className='rv-prod card'>
          {item.productImage ? (
            <Image className='rv-img' src={item.productImage} mode='aspectFill' />
          ) : (
            <View className='rv-img'>🐾</View>
          )}
          <View className='rv-prod-main'>
            <Text className='rv-title'>{item.productTitle}</Text>
            <Text className='rv-no'>{order?.orderNo}</Text>
          </View>
        </View>
      )}

      <View className='rv-card card'>
        <View className='rv-sec'>宝贝满足你的期待吗？</View>
        <View className='rv-stars'>
          {[1, 2, 3, 4, 5].map((n) => (
            <Text key={n} className={`rv-star ${n <= rating ? 'rv-star-on' : ''}`} onClick={() => setRating(n)}>
              ★
            </Text>
          ))}
          <Text className='rv-star-text'>{['', '很差', '较差', '一般', '满意', '超赞'][rating]}</Text>
        </View>
        <View className='rv-dims'>
          {([
            ['健康程度', healthScore, setHealthScore],
            ['品相外观', lookScore, setLookScore],
            ['服务体验', serviceScore, setServiceScore],
          ] as [string, number, (n: number) => void][]).map(([label, val, set]) => (
            <View className='rv-dim' key={label}>
              <Text className='rv-dim-label'>{label}</Text>
              <View className='rv-dim-stars'>
                {[1, 2, 3, 4, 5].map((n) => (
                  <Text key={n} className={`rv-dim-star ${n <= val ? 'rv-star-on' : ''}`} onClick={() => set(n)}>★</Text>
                ))}
              </View>
            </View>
          ))}
        </View>
        <Textarea
          className='rv-textarea'
          maxlength={500}
          placeholder='说说你的养宠初体验吧～'
          value={content}
          onInput={(e) => setContent(e.detail.value)}
        />
        <View className='rv-imgs'>
          {images.map((u) => (
            <Image key={u} className='rv-img-item' src={u} mode='aspectFill' onClick={() => Taro.previewImage({ urls: images, current: u })} />
          ))}
          {images.length < MAX_IMAGES && <View className='rv-img-add' onClick={chooseImage}>＋</View>}
        </View>
      </View>

      <View className={`rv-submit ${submitting ? 'rv-submit-off' : ''}`} onClick={submit}>
        发布评价
      </View>
    </View>
  )
}
