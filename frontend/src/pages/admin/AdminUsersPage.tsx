import { useEffect, useState } from 'react'
import { adminStatsAPI } from '../../lib/adminApi'

interface User {
  id: string; email: string; nickname: string
  created_at: string; chart_count: number
}

export default function AdminUsersPage() {
  const [users, setUsers] = useState<User[]>([])
  const [total, setTotal] = useState(0)
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)

  const load = (q = query) => {
    setLoading(true)
    adminStatsAPI.users(1, q)
      .then(r => { setUsers(r.data.users || []); setTotal(r.data.total || 0) })
      .finally(() => setLoading(false))
  }

  useEffect(() => {
    adminStatsAPI.users(1, '')
      .then(r => { setUsers(r.data.users || []); setTotal(r.data.total || 0) })
      .finally(() => setLoading(false))
  }, [])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    load(query)
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ margin: 0 }}>👥 用户列表</h1>
        <span style={{ color: '#888', fontSize: 14 }}>共 {total} 名用户</span>
      </div>

      <form onSubmit={handleSearch} className="admin-search-bar">
        <input
          id="admin-user-search"
          className="admin-search-input"
          value={query}
          onChange={e => setQuery(e.target.value)}
          placeholder="搜索邮箱..."
        />
        <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
        {query && (
          <button type="button" className="admin-btn admin-btn-ghost"
            onClick={() => { setQuery(''); load('') }}>清除</button>
        )}
      </form>

      {loading ? <div className="admin-loading">加载中...</div> : (
        <div className="admin-card" style={{ padding: 0, overflow: 'hidden' }}>
          <table className="admin-table">
            <thead>
              <tr><th>邮箱</th><th>昵称</th><th>命盘数</th><th>注册时间</th></tr>
            </thead>
            <tbody>
              {!users?.length && (
                <tr><td colSpan={4} style={{ textAlign: 'center', color: '#666', padding: 40 }}>
                  暂无用户数据
                </td></tr>
              )}
              {users?.map(u => (
                <tr key={u.id}>
                  <td style={{ color: '#e8e8e8' }}>{u.email}</td>
                  <td style={{ color: '#aaa' }}>{u.nickname || '—'}</td>
                  <td>
                    <span style={{
                      background: u.chart_count > 0 ? 'rgba(201,162,39,0.15)' : 'transparent',
                      color: u.chart_count > 0 ? '#c9a227' : '#666',
                      padding: '2px 8px', borderRadius: 12, fontSize: 13
                    }}>{u.chart_count}</span>
                  </td>
                  <td style={{ color: '#666', fontSize: 13 }}>
                    {new Date(u.created_at).toLocaleDateString('zh-CN')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
