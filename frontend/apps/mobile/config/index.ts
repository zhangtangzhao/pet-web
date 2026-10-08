import { defineConfig } from '@tarojs/cli'

const outputRoot = process.env.TARO_OUTPUT_ROOT || 'dist'

export default defineConfig({
  projectName: 'pet-mobile',
  date: '2026-9-10',
  designWidth: 750,
  deviceRatio: { 640: 2.34 / 2, 750: 1, 828: 1.81 / 2 },
  sourceRoot: 'src',
  outputRoot,
  framework: 'react',
  compiler: 'webpack5',
  h5: {
    publicPath: '/',
    staticDirectory: 'static',
    router: {
      mode: 'browser',
    },
    devServer: {
      port: 10086,
      proxy: {
        '/api': {
          target: 'http://127.0.0.1:8888',
          changeOrigin: true,
        },
      },
    },
  },
})
