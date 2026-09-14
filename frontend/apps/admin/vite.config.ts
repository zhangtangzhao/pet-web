import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => ({
  // 生产产物部署在 nginx 的 /admin 子路径下（见 scripts/deploy/nginx.conf）
  base: mode === 'production' ? '/admin/' : '/',
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8888',
        changeOrigin: true,
        // 客服 WebSocket（/api/ws/cs）需要升级连接
        ws: true,
      },
    },
  },
}))
