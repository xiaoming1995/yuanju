import { defineConfig } from 'vitest/config'

export default defineConfig({
  test: {
    // 只跑 src 下的 vitest 单测；tests/ 下是 node:test 静态断言（npm run test:static）
    include: ['src/**/*.test.{ts,tsx}'],
  },
})
