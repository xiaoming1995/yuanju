import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5200,
    proxy: {
      '/api': {
        target: 'http://localhost:9002',
        changeOrigin: true,
        proxyTimeout: 300000,  // 300s，适配推理模型
        timeout: 300000,
      },
    },
  },
})
