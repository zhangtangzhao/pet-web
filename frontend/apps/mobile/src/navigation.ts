import Taro from '@tarojs/taro'

export function goLogin(redirectPath: string) {
  const target = encodeURIComponent(redirectPath)
  return Taro.redirectTo({ url: `/pages/login/index?redirect=${target}` })
}

export function goLoginForCurrentPage() {
  const router = Taro.getCurrentInstance().router
  const search = new URLSearchParams()
  Object.entries(router?.params ?? {}).forEach(([key, value]) => {
    if (value) search.set(key, value)
  })
  const path = router?.path?.replace(/^\/?/, '/') || '/pages/index/index'
  const query = search.toString()
  return goLogin(query ? `${path}?${query}` : path)
}

export function canGoBack() {
  return Taro.getCurrentPages().length > 1
}

export function goBackOrRedirect(url: string) {
  if (canGoBack()) return Taro.navigateBack()
  return Taro.redirectTo({ url })
}
