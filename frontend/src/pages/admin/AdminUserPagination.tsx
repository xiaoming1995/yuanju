interface Props {
  page: number
  pageSize: number
  total: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
}

export function AdminUserPagination({ page, pageSize, total, onPageChange, onPageSizeChange }: Props) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize))
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1
  const end = Math.min(page * pageSize, total)

  return (
    <nav className="admin-user-pagination" aria-label="用户列表分页">
      <span className="admin-user-pagination-range">显示 {start}–{end} 条，共 {total} 条</span>
      {onPageSizeChange && <label className="admin-user-page-size">每页
        <select value={pageSize} onChange={e => onPageSizeChange(Number(e.target.value))}>
          <option value={10}>10 条</option>
          <option value={20}>20 条</option>
          <option value={50}>50 条</option>
        </select>
      </label>}
      <div className="admin-user-pagination-pages">
        <button type="button" className="admin-btn admin-btn-ghost" disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}>上一页</button>
        <span aria-live="polite">{page} / {totalPages}</span>
        <button type="button" className="admin-btn admin-btn-ghost" disabled={page >= totalPages}
          onClick={() => onPageChange(page + 1)}>下一页</button>
      </div>
    </nav>
  )
}
