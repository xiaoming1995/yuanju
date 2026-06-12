// localStorage + TTL 的小缓存，用于神煞注解等准静态公开数据。
// 命中直接返回，未命中/过期/损坏则走 fetcher 并回写。

export function cachedFetch<T>(key: string, ttlMs: number, fetcher: () => Promise<T>): Promise<T> {
  try {
    const raw = localStorage.getItem(key)
    if (raw) {
      const { expires, data } = JSON.parse(raw) as { expires: number; data: T }
      if (Date.now() < expires) return Promise.resolve(data)
    }
  } catch {
    // 缓存损坏当作未命中
  }
  return fetcher().then(data => {
    try {
      localStorage.setItem(key, JSON.stringify({ expires: Date.now() + ttlMs, data }))
    } catch {
      // 存储满时放弃缓存，不影响请求结果
    }
    return data
  })
}
