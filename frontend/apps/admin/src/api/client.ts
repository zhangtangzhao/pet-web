import axios from 'axios'

export const ACCESS_KEY = 'pet_admin_access'
export const REFRESH_KEY = 'pet_admin_refresh'

export const client = axios.create({ baseURL: '/api', timeout: 15000 })

client.interceptors.request.use((config) => {
  const token = localStorage.getItem(ACCESS_KEY)
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  (resp) => resp.data,
  (err) => {
    const body = err.response?.data
    if (err.response?.status === 401 && !location.pathname.startsWith('/login')) {
      localStorage.removeItem(ACCESS_KEY)
      location.href = '/login'
    }
    return Promise.reject(new Error(body?.msg || err.message || '请求失败'))
  },
)

export interface PageResp<T = unknown> {
  total: number
  list: T[]
}
