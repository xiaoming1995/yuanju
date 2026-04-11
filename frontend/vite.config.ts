import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5200,
    proxy: {
      // SSE 流式端点：使用 selfHandleResponse 完全绕过 Vite 中间件管道的缓冲
      '/api/bazi/report-stream': {
        target: 'http://localhost:9002',
        changeOrigin: true,
        selfHandleResponse: true,
        timeout: 300000,
        configure: (proxy) => {
          proxy.on('proxyRes', (proxyRes, _req, res) => {
            // 直接将状态码和响应头写入客户端，不经过 Vite 中间件
            res.writeHead(proxyRes.statusCode || 200, proxyRes.headers)
            // 逐 chunk 手动传输，确保每个 SSE 数据块实时到达浏览器
            proxyRes.on('data', (chunk: Buffer) => {
              res.write(chunk)
              // 禁用 Nagle 算法，防止 TCP 层合并小包
              if (res.socket && !res.socket.destroyed) {
                res.socket.setNoDelay(true)
              }
            })
            proxyRes.on('end', () => {
              res.end()
            })
            proxyRes.on('error', () => {
              res.end()
            })
          })
        },
      },
      // 其他 API：常规代理
      '/api': {
        target: 'http://localhost:9002',
        changeOrigin: true,
        proxyTimeout: 300000,
        timeout: 300000,
      },
    },
  },
})

