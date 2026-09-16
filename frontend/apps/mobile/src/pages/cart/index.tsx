import { useCallback, useEffect, useState } from 'react'
import { Image, Text, View } from '@tarojs/components'
import Taro from '@tarojs/taro'
import { del, get, getToken, put } from '../../request'
import { CartItem, CartListResp } from '../../types'
import './index.css'

export default function Cart() {
  const [items, setItems] = useState<CartItem[]>([])
  const [loading, setLoading] = useState(true)

  const load = useCallback(
    () =>
      get<CartListResp>('/cart')
        .then((r) => setItems(r.list ?? []))
        .catch(() => {})
        .finally(() => setLoading(false)),
    [],
  )

  useEffect(() => {
    if (!getToken()) {
      Taro.redirectTo({ url: '/pages/login/index?redirect=%2Fpages%2Fcart%2Findex' }).catch(() => {})
      return
    }
    load()
  }, [load])

  const toggle = (it: CartItem) => {
    put(`/cart/${it.id}`, { checked: it.checked === 1 ? 0 : 1 }).then(load).catch(() => {})
  }

  const remove = (it: CartItem) => {
    del(`/cart/${it.id}`).then(load).catch(() => {})
  }

  const checkedItems = items.filter((i) => i.checked === 1 && i.onSale)
  const total = checkedItems.reduce((sum, i) => sum + Number(i.price), 0)

  const checkout = () => {
    if (checkedItems.length === 0) {
      Taro.showToast({ title: '请先勾选商品', icon: 'none' })
      return
    }
    Taro.setStorageSync('pet_cart_ids', checkedItems.map((i) => i.id))
    Taro.navigateTo({ url: '/pages/checkout/index?from=cart' }).catch(() => {})
  }

  const goDetail = (productId: string) =>
    Taro.navigateTo({ url: `/pages/detail/index?id=${productId}` }).catch(() => {})

  return (
    <View className='cart'>
      {loading && <View className='cart-empty'>加载中…</View>}
      {!loading && items.length === 0 && (
        <View className='cart-empty'>
          <Text className='cart-empty-icon'>🛒</Text>
          <Text>购物车还是空的</Text>
        </View>
      )}
      {items.map((it) => (
        <View className='cart-item' key={it.id}>
          <View className={`cart-check ${it.checked === 1 ? 'cart-check-on' : ''} ${it.onSale ? '' : 'cart-check-off'}`} onClick={() => it.onSale && toggle(it)}>
            ✓
          </View>
          <Image className='cart-img' src={it.productImage} onClick={() => goDetail(it.productId)} />
          <View className='cart-main'>
            <View className='cart-title' onClick={() => goDetail(it.productId)}>{it.productTitle}</View>
            {!!it.skuSpecs && <View className='cart-specs'>{it.skuSpecs}</View>}
            {!it.onSale && <View className='cart-offtag'>已下架</View>}
            <View className='cart-bottom'>
              <Text className='cart-price'>¥{it.price}</Text>
              <Text className='cart-del' onClick={() => remove(it)}>删除</Text>
            </View>
          </View>
        </View>
      ))}
      {items.length > 0 && (
        <View className='cart-footer'>
          <View className='cart-total'>
            合计 <Text className='cart-total-num'>¥{total.toFixed(2)}</Text>
          </View>
          <View className={`cart-btn ${checkedItems.length === 0 ? 'cart-btn-off' : ''}`} onClick={checkout}>
            去结算({checkedItems.length})
          </View>
        </View>
      )}
    </View>
  )
}
