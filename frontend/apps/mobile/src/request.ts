import Taro from '@tarojs/taro'

const BASE_URL = process.env.TARO_ENV === 'h5' ? '/api' : 'https://your-domain.com/api'

const ACCESS_KEY = 'pet_access'
const REFRESH_KEY = 'pet_refresh'

export function getToken() {
  return Taro.getStorageSync(ACCESS_KEY) as string
}

export function setTokens(access: string, refresh: string) {
  Taro.setStorageSync(ACCESS_KEY, access)
  Taro.setStorageSync(REFRESH_KEY, refresh)
}

export function clearTokens() {
  Taro.removeStorageSync(ACCESS_KEY)
  Taro.removeStorageSync(REFRESH_KEY)
}

interface Body<T> {
  code: number
  msg: string
  data: T
}

export async function request<T>(method: 'GET' | 'POST' | 'PUT' | 'DELETE', path: string, data?: unknown): Promise<T> {
  const res = await Taro.request<Body<T>>({
    url: BASE_URL + path,
    method,
    data: data as never,
    header: {
      'Content-Type': 'application/json',
      Authorization: getToken() ? `Bearer ${getToken()}` : '',
    },
  })
  const body = res.data
  if (body.code !== 0) {
    if (body.code === 40100 || body.code === 40101) {
      clearTokens()
      Taro.navigateTo({ url: '/pages/login/index' }).catch(() => {})
    }
    throw new Error(body.msg)
  }
  return body.data
}

export const get = <T,>(path: string) => request<T>('GET', path)
export const post = <T,>(path: string, data?: unknown) => request<T>('POST', path, data)
export const put = <T,>(path: string, data?: unknown) => request<T>('PUT', path, data)
export const del = <T,>(path: string) => request<T>('DELETE', path)
