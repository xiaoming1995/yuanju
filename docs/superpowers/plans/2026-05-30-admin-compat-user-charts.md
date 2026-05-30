# 后台合盘查看 + 用户/起盘管理补齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为后台新增「合盘明细」只读页，并补齐用户管理（分页/重置密码/禁用/删除）与起盘明细搜索。

**Architecture:** 完全对标既有「起盘明细」实现（`AdminListCharts` + `repository.ListBaziCharts` + `AdminChartsPage.tsx`）。后端 Gin + PostgreSQL，admin 路由组在 `backend/cmd/api/main.go`。前端 React + Vite，admin 页面在 `frontend/src/pages/admin/`，导航在 `AdminLayout.tsx`，路由在 `App.tsx`。

**Tech Stack:** Go 1.25 / Gin / PostgreSQL（database/sql）/ bcrypt / JWT；React 19 + TypeScript + Vite + axios + lucide-react。

**测试约定（重要，已核实）：**
- 后端 admin 端点的既有测试约定是 **auth-gating 测试**（`httptest` 验证未授权返回 401，不连数据库）——见 `backend/internal/handler/admin_registration_control_test.go`。新端点按此写「先失败测试」。
- 涉及数据库的查询/登录逻辑：用 `go build ./...` + `go vet ./...` 编译校验 + 手动运行验证（本仓库未为 admin 查询逻辑建 DB 单测，**不**新增 DB 测试框架，避免范围蔓延）。
- **前端无测试运行器**（`package.json` 无 vitest/jest）。前端任务用 `npm run build`（tsc 类型检查 + 构建）+ 手动浏览器验证，**不**引入测试框架。

**执行顺序：** Phase 1 → 2 → 3，每个 Phase 可独立交付验证。

---

## Phase 1 — 合盘明细页（只读，新建）

### 文件结构

- 修改 `backend/internal/model/admin.go` — 新增 `AdminCompatListItem` 列表行模型。
- 修改 `backend/internal/repository/compatibility_repository.go` — 新增 `AdminListCompatibilityReadings` + `GetCompatibilityReadingUserEmail`。
- 修改 `backend/internal/handler/admin_handler.go` — 新增 `AdminListCompatReadings` + `AdminGetCompatReadingDetail`。
- 修改 `backend/cmd/api/main.go` — 注册 2 条路由。
- 新增 `backend/internal/handler/admin_compat_test.go` — auth-gating 测试。
- 修改 `frontend/src/lib/adminApi.ts` — 新增 `adminCompatAPI`。
- 新增 `frontend/src/pages/admin/AdminCompatPage.tsx` — 列表 + 详情页。
- 修改 `frontend/src/components/AdminLayout.tsx` — 新增导航项。
- 修改 `frontend/src/App.tsx` — 注册 `/admin/compatibility` 路由。

---

### Task 1: 后端 — 合盘列表仓储函数 + 模型

**Files:**
- Modify: `backend/internal/model/admin.go`（文件末尾追加）
- Modify: `backend/internal/repository/compatibility_repository.go`（文件末尾追加）

- [ ] **Step 1: 在 `admin.go` 末尾追加列表行模型**

```go
// AdminCompatListItem 后台合盘明细列表行
type AdminCompatListItem struct {
	ID                string    `json:"id"`
	UserEmail         *string   `json:"user_email"`
	SelfName          string    `json:"self_name"`
	PartnerName       string    `json:"partner_name"`
	OverallScore      int       `json:"overall_score"`
	OverallLevel      string    `json:"overall_level"`
	RelationshipStage string    `json:"relationship_stage"`
	PrimaryQuestion   string    `json:"primary_question"`
	AnalysisVersion   string    `json:"analysis_version"`
	CreatedAt         time.Time `json:"created_at"`
}
```

确认 `admin.go` 顶部已 `import "time"`；若未导入则加入。

- [ ] **Step 2: 在 `compatibility_repository.go` 末尾追加列表与邮箱查询函数**

```go
// AdminListCompatibilityReadings 后台全量合盘列表（分页，含创建者邮箱）
func AdminListCompatibilityReadings(page, pageSize int) ([]model.AdminCompatListItem, int, error) {
	offset := (page - 1) * pageSize

	var total int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM compatibility_readings`).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.AdminCompatListItem{}, 0, nil
	}

	rows, err := database.DB.Query(`
		SELECT r.id, u.email, r.overall_score, r.overall_level,
		       r.relationship_stage, r.primary_question, r.analysis_version, r.created_at,
		       COALESCE((SELECT display_name FROM compatibility_participants WHERE reading_id=r.id AND role='self'  LIMIT 1), '') AS self_name,
		       COALESCE((SELECT display_name FROM compatibility_participants WHERE reading_id=r.id AND role='partner' LIMIT 1), '') AS partner_name
		FROM compatibility_readings r
		LEFT JOIN users u ON r.user_id = u.id
		ORDER BY r.created_at DESC
		LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := []model.AdminCompatListItem{}
	for rows.Next() {
		var it model.AdminCompatListItem
		if err := rows.Scan(&it.ID, &it.UserEmail, &it.OverallScore, &it.OverallLevel,
			&it.RelationshipStage, &it.PrimaryQuestion, &it.AnalysisVersion, &it.CreatedAt,
			&it.SelfName, &it.PartnerName); err != nil {
			return nil, 0, err
		}
		out = append(out, it)
	}
	return out, total, rows.Err()
}

// GetCompatibilityReadingUserEmail 取某条合盘创建者邮箱（游客或已删用户返回空串）
func GetCompatibilityReadingUserEmail(readingID string) (string, error) {
	var email sql.NullString
	err := database.DB.QueryRow(
		`SELECT u.email FROM compatibility_readings r LEFT JOIN users u ON r.user_id=u.id WHERE r.id=$1`,
		readingID,
	).Scan(&email)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return email.String, err
}
```

确认该文件已 `import "database/sql"`（既有 `GetCompatibilityReadingOwner` 已用到 `sql.ErrNoRows`，应已导入）。

- [ ] **Step 3: 编译校验**

Run: `cd backend && go build ./...`
Expected: 无报错。

- [ ] **Step 4: Commit**

```bash
git add backend/internal/model/admin.go backend/internal/repository/compatibility_repository.go
git commit -m "feat(admin): add compat reading list repository + model"
```

---

### Task 2: 后端 — 合盘 handler + 路由 + auth-gating 测试

**Files:**
- Modify: `backend/internal/handler/admin_handler.go`（文件末尾追加，参照 `AdminListCharts`）
- Modify: `backend/cmd/api/main.go:206` 附近（admin 路由组）
- Create: `backend/internal/handler/admin_compat_test.go`

- [ ] **Step 1: 写先失败的 auth-gating 测试**

`backend/internal/handler/admin_compat_test.go`：

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yuanju/configs"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestAdminListCompatReadingsRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/compatibility/readings", AdminListCompatReadings)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/compatibility/readings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated compat list, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminGetCompatReadingDetailRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.GET("/compatibility/readings/:id", AdminGetCompatReadingDetail)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/compatibility/readings/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated compat detail, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: 运行测试确认失败（handler 未定义）**

Run: `cd backend && go test ./internal/handler/ -run TestAdminListCompatReadingsRequiresAdminAuth -v`
Expected: 编译失败 `undefined: AdminListCompatReadings`。

- [ ] **Step 3: 在 `admin_handler.go` 末尾追加 handler**

```go
// AdminListCompatReadings 后台全量合盘明细（分页，只读）
func AdminListCompatReadings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := repository.AdminListCompatibilityReadings(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘明细失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "total": total, "page": page})
}

// AdminGetCompatReadingDetail 后台合盘详情（只读）
func AdminGetCompatReadingDetail(c *gin.Context) {
	id := c.Param("id")
	detail, err := repository.GetCompatibilityDetail(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取合盘详情失败"})
		return
	}
	if detail == nil || detail.Reading == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "合盘记录不存在"})
		return
	}
	email, _ := repository.GetCompatibilityReadingUserEmail(id)
	c.JSON(http.StatusOK, gin.H{"data": detail, "user_email": email})
}
```

确认 `admin_handler.go` 已 import `strconv`、`net/http`、`repository`（既有 `AdminListCharts` 已用，应齐全）。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd backend && go test ./internal/handler/ -run 'TestAdminListCompatReadings|TestAdminGetCompatReadingDetail' -v`
Expected: PASS（两条均 401）。

- [ ] **Step 5: 在 `main.go` admin 路由组注册路由**

在 `backend/cmd/api/main.go` 第 206 行 `adminAuth.GET("/charts", handler.AdminListCharts)` 之后插入：

```go
				adminAuth.GET("/compatibility/readings", handler.AdminListCompatReadings)
				adminAuth.GET("/compatibility/readings/:id", handler.AdminGetCompatReadingDetail)
```

- [ ] **Step 6: 编译校验**

Run: `cd backend && go build ./... && go vet ./internal/handler/`
Expected: 无报错。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/handler/admin_handler.go backend/internal/handler/admin_compat_test.go backend/cmd/api/main.go
git commit -m "feat(admin): add compat reading list/detail endpoints + auth tests"
```

---

### Task 3: 前端 — adminCompatAPI

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`（在 `adminChartsAPI` 之后追加）

- [ ] **Step 1: 追加 API 客户端**

```ts
export const adminCompatAPI = {
  list: (page: number = 1, pageSize: number = 20) =>
    adminApi.get(`/api/admin/compatibility/readings?page=${page}&pageSize=${pageSize}`),
  detail: (id: string) =>
    adminApi.get(`/api/admin/compatibility/readings/${id}`),
}
```

- [ ] **Step 2: 类型检查**

Run: `cd frontend && npx tsc --noEmit`
Expected: 无新增类型错误。

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/adminApi.ts
git commit -m "feat(admin): add adminCompatAPI client"
```

---

### Task 4: 前端 — AdminCompatPage 页面 + 导航 + 路由

**Files:**
- Create: `frontend/src/pages/admin/AdminCompatPage.tsx`
- Modify: `frontend/src/components/AdminLayout.tsx`
- Modify: `frontend/src/App.tsx`

- [ ] **Step 1: 新建 `AdminCompatPage.tsx`（列表 + 懒加载详情，对标 AdminChartsPage）**

```tsx
import React, { useEffect, useState } from 'react'
import { Heart, User, RefreshCw } from 'lucide-react'
import { adminCompatAPI } from '../../lib/adminApi'

interface CompatListItem {
  id: string
  user_email?: string | null
  self_name: string
  partner_name: string
  overall_score: number
  overall_level: string
  relationship_stage: string
  primary_question: string
  analysis_version: string
  created_at: string
}

interface CompatDetail {
  reading: {
    overall_score: number
    overall_level: string
    relationship_stage: string
    primary_question: string
    analysis_version: string
    dimension_scores: Record<string, unknown>
    duration_assessment: { summary?: string; overall_band?: string }
    consulting_assessment: Record<string, unknown>
    summary_tags: string[]
  }
  participants: Array<{
    role: string
    display_name: string
    birth_profile: { year: number; month: number; day: number; hour: number; gender: string }
    chart_snapshot?: {
      year_gan?: string; year_zhi?: string; month_gan?: string; month_zhi?: string
      day_gan?: string; day_zhi?: string; hour_gan?: string; hour_zhi?: string
    } | null
  }>
  evidences: Array<{ id: string; dimension: string; polarity: string; title: string; detail: string }>
  latest_report?: { content: string; model: string; created_at: string } | null
}

const levelLabel = (l: string) => l === 'high' ? '高' : l === 'low' ? '低' : '中'
const stageLabel: Record<string, string> = {
  ambiguous: '暧昧', dating: '热恋', long_distance: '异地', reconciliation: '复合',
  marriage_or_engagement: '婚姻/订婚', crush: '单恋',
}

export default function AdminCompatPage() {
  const [items, setItems] = useState<CompatListItem[]>([])
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [loading, setLoading] = useState(true)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [details, setDetails] = useState<Record<string, CompatDetail>>({})
  const [detailLoading, setDetailLoading] = useState<Record<string, boolean>>({})
  const pageSize = 20

  const fetchList = async (pageNum: number) => {
    try {
      setLoading(true)
      const res = await adminCompatAPI.list(pageNum, pageSize)
      setItems(res.data?.data || [])
      setTotal(res.data?.total || 0)
    } catch (err) {
      console.error('获取合盘明细失败:', err)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { fetchList(page) }, [page])

  useEffect(() => {
    if (expandedId && !details[expandedId]) {
      setDetailLoading(prev => ({ ...prev, [expandedId]: true }))
      adminCompatAPI.detail(expandedId)
        .then(res => setDetails(prev => ({ ...prev, [expandedId]: res.data?.data })))
        .catch(err => console.error(err))
        .finally(() => setDetailLoading(prev => ({ ...prev, [expandedId]: false })))
    }
  }, [expandedId]) // eslint-disable-line react-hooks/exhaustive-deps

  const totalPages = Math.ceil((total || 0) / pageSize) || 1

  const pillars = (s?: CompatDetail['participants'][0]['chart_snapshot']) => [
    { label: '年', gan: s?.year_gan, zhi: s?.year_zhi },
    { label: '月', gan: s?.month_gan, zhi: s?.month_zhi },
    { label: '日', gan: s?.day_gan, zhi: s?.day_zhi },
    { label: '时', gan: s?.hour_gan, zhi: s?.hour_zhi },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <h1 className="admin-page-title" style={{ display: 'flex', alignItems: 'center', gap: 8, margin: 0 }}>
          <Heart size={24} /> 全站合盘明细
        </h1>
        <button onClick={() => fetchList(page)} style={{ display: 'flex', alignItems: 'center', gap: 6, padding: '6px 14px', borderRadius: 8, border: '1px solid #444', background: '#2a2a3a', color: '#ccc', cursor: 'pointer', fontSize: 13 }}>
          <RefreshCw size={14} /> 刷新
        </button>
      </div>

      <div className="admin-card">
        <div style={{ marginBottom: 16, fontSize: 13, color: '#888' }}>
          记录平台上每一次合盘测算，共 {total} 条记录。
        </div>

        {loading ? (
          <div className="admin-loading">加载中...</div>
        ) : items.length === 0 ? (
          <div style={{ textAlign: 'center', color: '#555', padding: '40px 0' }}>
            <Heart size={48} color="#333" style={{ margin: '0 auto 16px' }} />
            <p>暂无合盘记录</p>
          </div>
        ) : (
          <table className="admin-table">
            <thead>
              <tr><th>排盘用户</th><th>排盘时间</th><th>双方命主</th><th>总分/等级</th><th>关系阶段</th><th>操作</th></tr>
            </thead>
            <tbody>
              {items.map(it => (
                <React.Fragment key={it.id}>
                  <tr style={{ background: expandedId === it.id ? 'rgba(255,255,255,0.02)' : 'transparent' }}>
                    <td>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                        <User size={14} color="#888" />
                        {it.user_email
                          ? <span style={{ color: '#ccc' }}>{it.user_email}</span>
                          : <span style={{ color: '#666', fontStyle: 'italic' }}>未知</span>}
                      </div>
                    </td>
                    <td style={{ fontSize: 12, color: '#888' }}>{new Date(it.created_at).toLocaleString('zh-CN')}</td>
                    <td style={{ color: '#ccc' }}>{it.self_name || '—'} × {it.partner_name || '—'}</td>
                    <td>
                      <span style={{ color: '#f472b6', fontWeight: 600 }}>{it.overall_score}</span>
                      <span style={{ color: '#888', fontSize: 12 }}> / {levelLabel(it.overall_level)}</span>
                    </td>
                    <td style={{ fontSize: 12, color: '#aaa' }}>{stageLabel[it.relationship_stage] || it.relationship_stage}</td>
                    <td>
                      <button onClick={() => setExpandedId(expandedId === it.id ? null : it.id)} style={{ padding: '4px 12px', fontSize: 12, background: expandedId === it.id ? '#a78bfa' : '#2a2a3a', border: expandedId === it.id ? 'none' : '1px solid #444', borderRadius: 6, color: expandedId === it.id ? '#fff' : '#ccc', cursor: 'pointer' }}>
                        {expandedId === it.id ? '收起详情' : '查看详情'}
                      </button>
                    </td>
                  </tr>

                  {expandedId === it.id && (
                    <tr style={{ background: 'rgba(0,0,0,0.2)' }}>
                      <td colSpan={6} style={{ padding: '20px 24px', borderLeft: '3px solid #f472b6' }}>
                        {detailLoading[it.id] && <div style={{ fontSize: 12, color: '#666' }}>加载中...</div>}
                        {!detailLoading[it.id] && details[it.id] && (() => {
                          const d = details[it.id]
                          return (
                            <div>
                              {/* 双方四柱 */}
                              <div style={{ display: 'flex', gap: 40, flexWrap: 'wrap', marginBottom: 20 }}>
                                {d.participants.map((p, pi) => (
                                  <div key={pi}>
                                    <div style={{ fontSize: 13, color: '#888', marginBottom: 12 }}>
                                      {p.role === 'self' ? '命主' : '对方'}：{p.display_name || '—'}（{p.birth_profile.gender === 'male' ? '男' : '女'} {p.birth_profile.year}-{p.birth_profile.month}-{p.birth_profile.day}）
                                    </div>
                                    <div style={{ display: 'flex', gap: 12 }}>
                                      {pillars(p.chart_snapshot).map((col, ci) => (
                                        <div key={ci} style={{ textAlign: 'center', background: '#222', padding: '8px 14px', borderRadius: 8, border: '1px solid #333' }}>
                                          <div style={{ fontSize: 11, color: '#666', marginBottom: 8 }}>{col.label}柱</div>
                                          <div style={{ fontSize: 16, fontWeight: 600, color: '#ccc' }}>{col.gan || '—'}</div>
                                          <div style={{ fontSize: 16, fontWeight: 600, color: '#ccc' }}>{col.zhi || ''}</div>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                ))}
                              </div>

                              {/* 期限评估 */}
                              {d.reading.duration_assessment?.summary && (
                                <div style={{ background: '#222', padding: '12px 16px', borderRadius: 8, border: '1px solid #333', marginBottom: 16 }}>
                                  <div style={{ fontSize: 13, color: '#888', marginBottom: 6 }}>期限评估</div>
                                  <div style={{ fontSize: 13, color: '#ccc', lineHeight: 1.7 }}>{d.reading.duration_assessment.summary}</div>
                                </div>
                              )}

                              {/* 证据列表 */}
                              <div style={{ fontSize: 13, color: '#888', marginBottom: 8 }}>证据 ({d.evidences.length} 条)</div>
                              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginBottom: 16 }}>
                                {d.evidences.map(ev => (
                                  <div key={ev.id} style={{ background: 'rgba(0,0,0,0.2)', padding: '8px 12px', borderRadius: 6 }}>
                                    <div style={{ fontSize: 11, color: ev.polarity === 'positive' ? '#34d399' : ev.polarity === 'negative' ? '#ff6b6b' : '#aaa', marginBottom: 4, fontWeight: 600 }}>
                                      [{ev.dimension}] {ev.title}
                                    </div>
                                    <div style={{ fontSize: 12, color: '#ccc', lineHeight: 1.6 }}>{ev.detail}</div>
                                  </div>
                                ))}
                              </div>

                              {/* AI 报告 */}
                              <div style={{ fontSize: 13, color: '#888', marginBottom: 8 }}>AI 报告</div>
                              {d.latest_report
                                ? <div style={{ background: 'rgba(244,114,182,0.05)', padding: 16, borderRadius: 8, border: '1px solid rgba(244,114,182,0.2)', fontSize: 12, color: '#ccc', lineHeight: 1.8, whiteSpace: 'pre-wrap', maxHeight: 300, overflowY: 'auto' }}>{d.latest_report.content}</div>
                                : <div style={{ background: '#222', padding: 16, borderRadius: 8, border: '1px dashed #444', color: '#666', fontSize: 13 }}>此合盘尚未生成 AI 报告。</div>}
                            </div>
                          )
                        })()}
                      </td>
                    </tr>
                  )}
                </React.Fragment>
              ))}
            </tbody>
          </table>
        )}

        {totalPages > 1 && (
          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 20 }}>
            <button disabled={page === 1} onClick={() => setPage(p => p - 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page === 1 ? '#1a1a2e' : '#2a2a3a', color: page === 1 ? '#555' : '#ccc', cursor: page === 1 ? 'not-allowed' : 'pointer' }}>上一页</button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: '#666', margin: '0 8px' }}>第 {page} / {totalPages} 页</span>
            <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page >= totalPages ? '#1a1a2e' : '#2a2a3a', color: page >= totalPages ? '#555' : '#ccc', cursor: page >= totalPages ? 'not-allowed' : 'pointer' }}>下一页</button>
          </div>
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 在 `AdminLayout.tsx` 导航中新增「合盘明细」**

在 import 行把 `Heart` 加入 lucide 引入：找到第 2 行 `import { Hexagon, ... Trash2 } from 'lucide-react'`，在其中加入 `Heart`。
在「起盘明细」NavLink（`to="/admin/charts"`）之后插入：

```tsx
          <NavLink to="/admin/compatibility" className={({isActive}) => isActive ? 'admin-nav-item active' : 'admin-nav-item'}>
            <span style={{ display: 'flex', alignItems: 'center', marginRight: 10 }}><Heart size={18} /></span> 合盘明细
          </NavLink>
```

- [ ] **Step 3: 在 `App.tsx` 注册路由**

在第 19 行 `import AdminChartsPage ...` 之后新增 import：

```tsx
import AdminCompatPage from './pages/admin/AdminCompatPage'
```

admin 子路由是 `/admin` 下的嵌套 `<Route>`（见第 65-80 行）。在第 73 行 `<Route path="charts" element={<AdminChartsPage />} />` 之后新增（解析为 `/admin/compatibility`，与 NavLink 对齐）：

```tsx
              <Route path="compatibility" element={<AdminCompatPage />} />
```

- [ ] **Step 4: 类型检查 + 构建**

Run: `cd frontend && npm run build`
Expected: 构建成功，无类型错误。

- [ ] **Step 5: 手动验证**

启动前后端，登录后台 → 侧边栏点「合盘明细」→ 确认：列表显示创建者邮箱/双方姓名/总分/关系阶段；翻页可用；展开任一条能看到双方四柱、期限评估、证据、AI 报告（无报告时显示占位）。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/admin/AdminCompatPage.tsx frontend/src/components/AdminLayout.tsx frontend/src/App.tsx
git commit -m "feat(admin): add 合盘明细 read-only page + nav + route"
```

---

## Phase 2 — 用户管理补齐

### 文件结构

- 修改 `backend/internal/repository/repository.go` — `GetUserByEmail` 增 `disabled_at`；新增 `UpdateUserPassword` / `SetUserDisabled` / `DeleteUser`。
- 修改 `backend/internal/model/model.go` — `User` 增 `DisabledAt`。
- 修改 `backend/internal/service/auth_service.go` — 登录拦截禁用用户；新增 `ResetUserPassword`。
- 新增 `backend/pkg/database/migrations/00014_add_user_disabled_at.sql`。
- 修改 `backend/internal/handler/admin_handler.go` — `AdminGetUsers` 增 `disabled_at`/`compat_count`；新增重置密码/禁用/删除 handler。
- 修改 `backend/cmd/api/main.go` — 注册 3 条路由。
- 新增 `backend/internal/handler/admin_user_mgmt_test.go` — auth-gating 测试。
- 修改 `frontend/src/lib/adminApi.ts` — `adminUsersAPI` 增方法。
- 修改 `frontend/src/pages/admin/AdminUsersPage.tsx` — 分页 + 行操作（重置/禁用/删除）。

---

### Task 5: 前端 — 用户列表分页修复

**Files:**
- Modify: `frontend/src/pages/admin/AdminUsersPage.tsx`

- [ ] **Step 1: 改造分页状态与加载**

把 `load` 改为接受 `page`，并新增 `page` 状态。将第 28-33 行的 `load` 替换为：

```tsx
  const [page, setPage] = useState(1)

  const load = useCallback((q: string, pageNum: number) => {
    setLoading(true)
    adminStatsAPI.users(pageNum, q)
      .then(r => { setUsers(r.data.users || []); setTotal(r.data.total || 0) })
      .finally(() => setLoading(false))
  }, [])
```

更新所有 `load` 调用点：
- `useEffect`（第 41-44 行）：`load('', 1)`，依赖数组保持 `[load]`。
- `handleSearch`：`setPage(1); load(query, 1)`。
- 搜索「清除」按钮：`setPage(1); load('', 1)`。
- `handleCreate` 成功后：`load(query, page)`。
- 新增 `useEffect(() => { load(query, page) }, [page])`（仅翻页时触发；为避免与搜索重复，可在翻页按钮内直接 `setPage`）。

- [ ] **Step 2: 在表格 `</table>` 容器后新增分页器**

在表格 `admin-card` 之后（第 171 行 `)}` 之前合适位置）新增，与 `AdminChartsPage` 一致：

```tsx
      {(() => {
        const totalPages = Math.ceil((total || 0) / 20) || 1
        if (totalPages <= 1) return null
        return (
          <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 20 }}>
            <button disabled={page === 1} onClick={() => setPage(p => p - 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page === 1 ? '#1a1a2e' : '#2a2a3a', color: page === 1 ? '#555' : '#ccc', cursor: page === 1 ? 'not-allowed' : 'pointer' }}>上一页</button>
            <span style={{ lineHeight: '32px', fontSize: 13, color: '#666', margin: '0 8px' }}>第 {page} / {totalPages} 页</span>
            <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)} style={{ padding: '6px 14px', borderRadius: 8, border: 'none', background: page >= totalPages ? '#1a1a2e' : '#2a2a3a', color: page >= totalPages ? '#555' : '#ccc', cursor: page >= totalPages ? 'not-allowed' : 'pointer' }}>下一页</button>
          </div>
        )
      })()}
```

- [ ] **Step 3: 类型检查 + 构建**

Run: `cd frontend && npm run build`
Expected: 构建成功。

- [ ] **Step 4: 手动验证**

构造 >20 用户（或临时把 pageSize 当作 20 验证），确认能翻到第 2 页且数据变化；搜索后回到第 1 页。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/pages/admin/AdminUsersPage.tsx
git commit -m "fix(admin): paginate user list (was stuck on page 1)"
```

---

### Task 6: 后端 — 重置用户密码

**Files:**
- Modify: `backend/internal/repository/repository.go`（用户区追加）
- Modify: `backend/internal/service/auth_service.go`（追加）
- Modify: `backend/internal/handler/admin_handler.go`（追加）
- Modify: `backend/cmd/api/main.go`
- Create: `backend/internal/handler/admin_user_mgmt_test.go`

- [ ] **Step 1: 写先失败的 auth-gating 测试**

`backend/internal/handler/admin_user_mgmt_test.go`：

```go
package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yuanju/configs"
	"yuanju/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestAdminResetUserPasswordRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.POST("/users/:id/reset-password", AdminResetUserPassword)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/abc/reset-password", strings.NewReader(`{"password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd backend && go test ./internal/handler/ -run TestAdminResetUserPasswordRequiresAdminAuth -v`
Expected: 编译失败 `undefined: AdminResetUserPassword`。

- [ ] **Step 3: 仓储新增 `UpdateUserPassword`**

在 `repository.go` 用户区（`GetUserByID` 之后）追加：

```go
func UpdateUserPassword(userID, passwordHash string) error {
	_, err := database.DB.Exec(`UPDATE users SET password_hash=$1 WHERE id=$2`, passwordHash, userID)
	return err
}
```

- [ ] **Step 4: service 新增 `ResetUserPassword`**

在 `auth_service.go` 追加：

```go
func ResetUserPassword(userID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return repository.UpdateUserPassword(userID, string(hash))
}
```

- [ ] **Step 5: handler 新增 `AdminResetUserPassword`**

在 `admin_handler.go` 追加：

```go
// AdminResetUserPassword 后台重置指定用户密码
func AdminResetUserPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "新密码至少需要8位"})
		return
	}
	if err := service.ResetUserPassword(id, req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "重置密码失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已重置"})
}
```

确认 `admin_handler.go` 已 import `service`（既有 `AdminCreateUser` 已用）。

- [ ] **Step 6: 运行测试确认通过**

Run: `cd backend && go test ./internal/handler/ -run TestAdminResetUserPasswordRequiresAdminAuth -v`
Expected: PASS。

- [ ] **Step 7: 注册路由**

在 `main.go` admin 路由组 `adminAuth.POST("/users", handler.AdminCreateUser)` 之后插入：

```go
				adminAuth.POST("/users/:id/reset-password", handler.AdminResetUserPassword)
```

- [ ] **Step 8: 编译校验 + Commit**

Run: `cd backend && go build ./... && go test ./internal/handler/ -run TestAdminResetUserPassword`
Expected: 通过。

```bash
git add backend/internal/repository/repository.go backend/internal/service/auth_service.go backend/internal/handler/admin_handler.go backend/internal/handler/admin_user_mgmt_test.go backend/cmd/api/main.go
git commit -m "feat(admin): reset user password endpoint"
```

---

### Task 7: 后端 — 禁用用户（迁移 + 模型 + 登录拦截 + 端点）

**Files:**
- Create: `backend/pkg/database/migrations/00014_add_user_disabled_at.sql`
- Modify: `backend/internal/model/model.go`
- Modify: `backend/internal/repository/repository.go`
- Modify: `backend/internal/service/auth_service.go`
- Modify: `backend/internal/handler/admin_handler.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/internal/handler/admin_user_mgmt_test.go`

- [ ] **Step 1: 新建迁移**

`backend/pkg/database/migrations/00014_add_user_disabled_at.sql`：

```sql
-- 用户禁用标记：NULL = 正常，非空 = 已禁用（禁用时间）
ALTER TABLE users ADD COLUMN IF NOT EXISTS disabled_at TIMESTAMPTZ;
```

- [ ] **Step 2: `User` 模型增字段**

`model.go` 第 8-15 行 `User` 结构体在 `CreatedAt` 之后加：

```go
	DisabledAt   *time.Time `json:"disabled_at,omitempty"`
```

- [ ] **Step 3: `GetUserByEmail` 同步取 `disabled_at`**

把 `repository.go` 第 27-37 行 `GetUserByEmail` 改为：

```go
func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	err := database.DB.QueryRow(
		`SELECT id, email, password_hash, nickname, COALESCE(source, 'self_registered'), created_at, disabled_at FROM users WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nickname, &user.Source, &user.CreatedAt, &user.DisabledAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}
```

- [ ] **Step 4: 登录拦截禁用用户**

把 `auth_service.go` `Login` 第 103-105 行（bcrypt 比对块）之后、生成 token 之前插入：

```go
	if user.DisabledAt != nil {
		return nil, "", errors.New("该账号已被禁用")
	}
```

- [ ] **Step 5: 仓储新增 `SetUserDisabled`**

`repository.go` 追加：

```go
func SetUserDisabled(userID string, disabled bool) error {
	if disabled {
		_, err := database.DB.Exec(`UPDATE users SET disabled_at=NOW() WHERE id=$1`, userID)
		return err
	}
	_, err := database.DB.Exec(`UPDATE users SET disabled_at=NULL WHERE id=$1`, userID)
	return err
}
```

- [ ] **Step 6: handler 新增 `AdminSetUserDisabled`**

`admin_handler.go` 追加：

```go
// AdminSetUserDisabled 禁用/解禁用户
func AdminSetUserDisabled(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Disabled *bool `json:"disabled" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Disabled == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "disabled 必填"})
		return
	}
	if err := repository.SetUserDisabled(id, *req.Disabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"disabled": *req.Disabled})
}
```

- [ ] **Step 7: `AdminGetUsers` 返回 `disabled_at`**

修改 `admin_handler.go` `AdminGetUsers`（第 305-340 行）：
- SQL 的 SELECT 增加 `u.disabled_at`，并加入 GROUP BY：

```go
	query := `
		SELECT u.id, u.email, u.nickname, COALESCE(u.source, 'self_registered'), u.created_at, u.disabled_at,
		       COUNT(b.id) as chart_count,
		       (SELECT COUNT(*) FROM compatibility_readings cr WHERE cr.user_id = u.id) as compat_count
		FROM users u
		LEFT JOIN bazi_charts b ON b.user_id = u.id
		WHERE ($1 = '' OR u.email ILIKE '%' || $1 || '%')
		GROUP BY u.id, u.email, u.nickname, u.source, u.created_at, u.disabled_at
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3`
```

- `UserRow` 结构体增加字段并调整 Scan：

```go
	type UserRow struct {
		ID          string     `json:"id"`
		Email       string     `json:"email"`
		Nickname    string     `json:"nickname"`
		Source      string     `json:"source"`
		CreatedAt   string     `json:"created_at"`
		DisabledAt  *time.Time `json:"disabled_at"`
		ChartCount  int        `json:"chart_count"`
		CompatCount int        `json:"compat_count"`
	}
	var users []UserRow
	for rows.Next() {
		var u UserRow
		rows.Scan(&u.ID, &u.Email, &u.Nickname, &u.Source, &u.CreatedAt, &u.DisabledAt, &u.ChartCount, &u.CompatCount)
		users = append(users, u)
	}
```

确认 `admin_handler.go` 已 import `time`；若未导入则加入。

- [ ] **Step 8: 追加 auth-gating 测试**

在 `admin_user_mgmt_test.go` 追加：

```go
func TestAdminSetUserDisabledRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.PUT("/users/:id/disable", AdminSetUserDisabled)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/users/abc/disable", strings.NewReader(`{"disabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 9: 注册路由**

`main.go` admin 路由组追加：

```go
				adminAuth.PUT("/users/:id/disable", handler.AdminSetUserDisabled)
```

- [ ] **Step 10: 编译 + 测试**

Run: `cd backend && go build ./... && go test ./internal/handler/ -run 'TestAdminSetUserDisabled' -v`
Expected: 通过。

- [ ] **Step 11: 手动验证迁移与登录拦截**

应用迁移后：后台禁用某用户 → 该用户登录返回「该账号已被禁用」；解禁后可正常登录。

- [ ] **Step 12: Commit**

```bash
git add backend/pkg/database/migrations/00014_add_user_disabled_at.sql backend/internal/model/model.go backend/internal/repository/repository.go backend/internal/service/auth_service.go backend/internal/handler/admin_handler.go backend/internal/handler/admin_user_mgmt_test.go backend/cmd/api/main.go
git commit -m "feat(admin): soft-disable users (disabled_at + login block)"
```

---

### Task 8: 后端 — 删除用户（硬删除）

**Files:**
- Modify: `backend/internal/repository/repository.go`
- Modify: `backend/internal/handler/admin_handler.go`
- Modify: `backend/cmd/api/main.go`
- Modify: `backend/internal/handler/admin_user_mgmt_test.go`

- [ ] **Step 1: 追加 auth-gating 测试**

`admin_user_mgmt_test.go` 追加：

```go
func TestAdminDeleteUserRequiresAdminAuth(t *testing.T) {
	configs.AppConfig.AdminJWTSecret = "test-admin-secret"
	r := gin.New()
	admin := r.Group("/api/admin", middleware.AdminAuth())
	admin.DELETE("/users/:id", AdminDeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/api/admin/users/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `cd backend && go test ./internal/handler/ -run TestAdminDeleteUserRequiresAdminAuth -v`
Expected: 编译失败 `undefined: AdminDeleteUser`。

- [ ] **Step 3: 仓储新增 `DeleteUser`**

`repository.go` 追加（依赖既有外键：合盘 CASCADE、八字 SET NULL）：

```go
func DeleteUser(userID string) error {
	_, err := database.DB.Exec(`DELETE FROM users WHERE id=$1`, userID)
	return err
}
```

- [ ] **Step 4: handler 新增 `AdminDeleteUser`**

`admin_handler.go` 追加：

```go
// AdminDeleteUser 硬删除用户（合盘记录级联删除，八字命盘转游客保留）
func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := repository.DeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}
```

- [ ] **Step 5: 注册路由**

`main.go` admin 路由组追加：

```go
				adminAuth.DELETE("/users/:id", handler.AdminDeleteUser)
```

- [ ] **Step 6: 编译 + 测试**

Run: `cd backend && go build ./... && go test ./internal/handler/ -run TestAdminDeleteUser -v`
Expected: 通过。

- [ ] **Step 7: Commit**

```bash
git add backend/internal/repository/repository.go backend/internal/handler/admin_handler.go backend/internal/handler/admin_user_mgmt_test.go backend/cmd/api/main.go
git commit -m "feat(admin): hard-delete user endpoint"
```

---

### Task 9: 前端 — 用户行操作（重置密码 / 禁用 / 删除）

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`
- Modify: `frontend/src/pages/admin/AdminUsersPage.tsx`

- [ ] **Step 1: 扩展 `adminUsersAPI`**

把 `adminApi.ts` 第 58-61 行 `adminUsersAPI` 替换为：

```ts
export const adminUsersAPI = {
  create: (data: { email: string; password: string; nickname?: string }) =>
    adminApi.post('/api/admin/users', data),
  resetPassword: (id: string, password: string) =>
    adminApi.post(`/api/admin/users/${id}/reset-password`, { password }),
  setDisabled: (id: string, disabled: boolean) =>
    adminApi.put(`/api/admin/users/${id}/disable`, { disabled }),
  remove: (id: string) =>
    adminApi.delete(`/api/admin/users/${id}`),
}
```

- [ ] **Step 2: `User` 接口增字段**

`AdminUsersPage.tsx` 第 4-11 行 `User` 接口加：

```tsx
  disabled_at?: string | null
  compat_count?: number
```

- [ ] **Step 3: 表格新增「状态」列与「操作」列**

把表头（第 141 行）改为：

```tsx
              <tr><th>邮箱</th><th>昵称</th><th>来源</th><th>命盘数</th><th>状态</th><th>注册时间</th><th>操作</th></tr>
```

空数据行 `colSpan` 由 5 改为 7。在每行「注册时间」`<td>` 之后新增状态列（放在来源后亦可，这里放注册时间前）与操作列。最简：在注册时间 `<td>` 前插入状态列，在其后插入操作列：

```tsx
                  <td>
                    {u.disabled_at
                      ? <span style={{ color: '#ff6b6b', fontSize: 12 }}>已禁用</span>
                      : <span style={{ color: '#22c55e', fontSize: 12 }}>正常</span>}
                  </td>
```

操作列（注册时间 `<td>` 之后）：

```tsx
                  <td style={{ whiteSpace: 'nowrap' }}>
                    <button className="admin-btn admin-btn-ghost" style={{ marginRight: 6 }}
                      onClick={() => openReset(u)}>重置密码</button>
                    <button className="admin-btn admin-btn-ghost" style={{ marginRight: 6 }}
                      onClick={() => toggleDisabled(u)}>{u.disabled_at ? '解禁' : '禁用'}</button>
                    <button className="admin-btn admin-btn-ghost" style={{ color: '#ff6b6b' }}
                      onClick={() => removeUser(u)}>删除</button>
                  </td>
```

- [ ] **Step 4: 新增操作处理函数与重置密码弹窗状态**

在组件内（`handleCreate` 之后）新增：

```tsx
  const [resetTarget, setResetTarget] = useState<User | null>(null)
  const [resetPwd, setResetPwd] = useState('')
  const [resetErr, setResetErr] = useState('')
  const [resetSaving, setResetSaving] = useState(false)

  const openReset = (u: User) => { setResetTarget(u); setResetPwd(''); setResetErr('') }

  const submitReset = async () => {
    if (resetPwd.length < 8) { setResetErr('新密码至少需要8位'); return }
    setResetSaving(true)
    try {
      await adminUsersAPI.resetPassword(resetTarget!.id, resetPwd)
      setResetTarget(null)
      alert('已重置，请通过安全渠道告知用户新密码。')
    } catch (e: unknown) {
      setResetErr(e instanceof Error ? e.message : '重置失败')
    } finally {
      setResetSaving(false)
    }
  }

  const toggleDisabled = async (u: User) => {
    const next = !u.disabled_at
    if (!window.confirm(next ? `确认禁用用户 ${u.email}？该用户将无法登录。` : `确认解禁用户 ${u.email}？`)) return
    try {
      await adminUsersAPI.setDisabled(u.id, next)
      load(query, page)
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '操作失败')
    }
  }

  const removeUser = async (u: User) => {
    const n = u.compat_count || 0
    const warning = `确认删除用户 ${u.email}？\n\n` +
      `· 将连带删除其 ${n} 条合盘记录（不可恢复）\n` +
      `· 其八字命盘将转为游客记录保留\n\n此操作不可撤销。`
    if (!window.confirm(warning)) return
    if (!window.confirm(`再次确认：删除 ${u.email} 及其 ${n} 条合盘记录？`)) return
    try {
      await adminUsersAPI.remove(u.id)
      load(query, page)
    } catch (e: unknown) {
      alert(e instanceof Error ? e.message : '删除失败')
    }
  }
```

确认 `adminUsersAPI` 已在文件顶部 import（第 2 行）。

- [ ] **Step 5: 新增重置密码弹窗（复用创建弹窗样式）**

在创建用户弹窗（`{showModal && ...}`）之后新增：

```tsx
      {resetTarget && (
        <div className="admin-modal-overlay" onClick={e => e.target === e.currentTarget && setResetTarget(null)}>
          <div className="admin-modal">
            <div className="admin-modal-title">重置密码 — {resetTarget.email}</div>
            {resetErr && <div className="admin-error">{resetErr}</div>}
            <div className="admin-form-group">
              <label className="admin-form-label">新密码（至少8位）</label>
              <input className="admin-form-input" type="password" value={resetPwd}
                onChange={e => setResetPwd(e.target.value)} />
            </div>
            <div className="admin-modal-actions">
              <button className="admin-btn admin-btn-ghost" onClick={() => setResetTarget(null)}>取消</button>
              <button className="admin-btn admin-btn-primary" onClick={submitReset} disabled={resetSaving}>
                {resetSaving ? '重置中...' : '确认重置'}
              </button>
            </div>
          </div>
        </div>
      )}
```

- [ ] **Step 6: 类型检查 + 构建**

Run: `cd frontend && npm run build`
Expected: 构建成功。

- [ ] **Step 7: 手动验证**

后台用户列表：重置密码弹窗可用并生效；禁用后状态变「已禁用」、该用户登录被拒、解禁恢复；删除弹两次确认且文案含「N 条合盘记录」，删除后列表移除。

- [ ] **Step 8: Commit**

```bash
git add frontend/src/lib/adminApi.ts frontend/src/pages/admin/AdminUsersPage.tsx
git commit -m "feat(admin): user row actions (reset password / disable / delete)"
```

---

## Phase 3 — 起盘明细搜索

### 文件结构

- 修改 `backend/internal/repository/admin_repository.go` — `ListBaziCharts` 增 `q/from/to` 过滤。
- 修改 `backend/internal/handler/admin_handler.go` — `AdminListCharts` 解析新参数。
- 修改 `frontend/src/lib/adminApi.ts` — `adminChartsAPI.list` 增参数。
- 修改 `frontend/src/pages/admin/AdminChartsPage.tsx` — 搜索栏。

---

### Task 10: 后端 — 起盘列表按邮箱/排盘时间过滤

**Files:**
- Modify: `backend/internal/repository/admin_repository.go:278-333`
- Modify: `backend/internal/handler/admin_handler.go`（`AdminListCharts`，第 427-449 行）

- [ ] **Step 1: 改造 `ListBaziCharts` 签名与查询**

把 `admin_repository.go` 第 278-333 行 `ListBaziCharts` 替换为（新增 `q, from, to string` 参数，动态拼 WHERE；有 `q` 时需 JOIN users 且排除游客）：

```go
func ListBaziCharts(page, pageSize int, q, from, to string) ([]model.AdminChartRecord, int, error) {
	offset := (page - 1) * pageSize

	// 动态过滤条件（$1=q, $2=from, $3=to）
	where := `WHERE ($1 = '' OR u.email ILIKE '%' || $1 || '%')
		AND ($2 = '' OR c.created_at >= $2::timestamptz)
		AND ($3 = '' OR c.created_at < ($3::timestamptz + interval '1 day'))`
	if q != "" {
		// 按邮箱搜索时排除无归属用户的游客记录
		where += ` AND c.user_id IS NOT NULL`
	}

	var total int
	err := database.DB.QueryRow(`SELECT COUNT(*) FROM bazi_charts c LEFT JOIN users u ON c.user_id = u.id `+where, q, from, to).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.AdminChartRecord{}, 0, nil
	}

	rows, err := database.DB.Query(`
		SELECT 
			c.id, c.user_id, u.email as user_email, 
			c.birth_year, c.birth_month, c.birth_day, c.birth_hour, 
			c.gender, c.year_gan, c.year_zhi, 
			c.month_gan, c.month_zhi, c.day_gan, c.day_zhi, c.hour_gan, c.hour_zhi,
			COALESCE(c.yongshen, '') as yongshen, 
			COALESCE(c.jishen, '') as jishen, 
			(SELECT content FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result,
			(SELECT content_structured FROM ai_reports WHERE chart_id=c.id ORDER BY created_at DESC LIMIT 1) as ai_result_structured,
			c.created_at
		FROM bazi_charts c
		LEFT JOIN users u ON c.user_id = u.id
		`+where+`
		ORDER BY c.created_at DESC
		LIMIT $4 OFFSET $5
	`, q, from, to, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var charts []model.AdminChartRecord
	for rows.Next() {
		var r model.AdminChartRecord
		if err := rows.Scan(
			&r.ID, &r.UserID, &r.UserEmail,
			&r.BirthYear, &r.BirthMonth, &r.BirthDay, &r.BirthHour,
			&r.Gender, &r.YearGan, &r.YearZhi,
			&r.MonthGan, &r.MonthZhi, &r.DayGan, &r.DayZhi, &r.HourGan, &r.HourZhi,
			&r.Yongshen, &r.Jishen,
			&r.AIResult,
			&r.AIResultStructured,
			&r.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		charts = append(charts, r)
	}

	return charts, total, nil
}
```

- [ ] **Step 2: `AdminListCharts` 传入新参数**

把 `admin_handler.go` `AdminListCharts`（第 438 行）调用改为：

```go
	q := c.Query("q")
	from := c.Query("from")
	to := c.Query("to")
	charts, total, err := repository.ListBaziCharts(page, pageSize, q, from, to)
```

（在第 438 行 `charts, total, err := repository.ListBaziCharts(page, pageSize)` 上方加 3 行取参数，并替换该调用。）

- [ ] **Step 3: 全局检查其它调用点**

Run: `cd backend && grep -rn "ListBaziCharts(" --include=*.go .`
Expected: 仅 `AdminListCharts` 一处调用；若有其它（如测试）一并更新为 5 参数。

- [ ] **Step 4: 编译校验**

Run: `cd backend && go build ./...`
Expected: 无报错。

- [ ] **Step 5: Commit**

```bash
git add backend/internal/repository/admin_repository.go backend/internal/handler/admin_handler.go
git commit -m "feat(admin): filter charts by email and cast-time range"
```

---

### Task 11: 前端 — 起盘明细搜索栏

**Files:**
- Modify: `frontend/src/lib/adminApi.ts`
- Modify: `frontend/src/pages/admin/AdminChartsPage.tsx`

- [ ] **Step 1: 扩展 `adminChartsAPI.list`**

把 `adminApi.ts` 第 70-71 行 `list` 改为：

```ts
  list: (page: number = 1, pageSize: number = 20, q = '', from = '', to = '') =>
    adminApi.get(`/api/admin/charts?page=${page}&pageSize=${pageSize}&q=${encodeURIComponent(q)}&from=${from}&to=${to}`),
```

- [ ] **Step 2: 页面增加搜索状态并传入 fetch**

`AdminChartsPage.tsx`：新增状态

```tsx
  const [q, setQ] = useState('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
```

把 `fetchCharts`（第 72-83 行）的 `adminChartsAPI.list(pageNum, pageSize)` 改为 `adminChartsAPI.list(pageNum, pageSize, q, from, to)`。

- [ ] **Step 3: 在表格上方（`admin-card` 内、第 134 行说明文字之后）新增搜索栏**

```tsx
        <form className="admin-search-bar" onSubmit={e => { e.preventDefault(); setPage(1); fetchCharts(1) }}>
          <input className="admin-search-input" value={q} onChange={e => setQ(e.target.value)} placeholder="按邮箱搜索..." />
          <input type="date" className="admin-search-input" value={from} onChange={e => setFrom(e.target.value)} title="排盘起始日期" />
          <input type="date" className="admin-search-input" value={to} onChange={e => setTo(e.target.value)} title="排盘截止日期" />
          <button type="submit" className="admin-btn admin-btn-primary">搜索</button>
          {(q || from || to) && (
            <button type="button" className="admin-btn admin-btn-ghost"
              onClick={() => { setQ(''); setFrom(''); setTo(''); setPage(1); adminChartsAPI.list(1, pageSize).then(res => { setCharts(res.data?.data || []); setTotal(res.data?.total || 0) }) }}>清除</button>
          )}
        </form>
```

注：清除时直接用无过滤参数重新拉取，避免依赖 state 异步更新。

- [ ] **Step 4: 类型检查 + 构建**

Run: `cd frontend && npm run build`
Expected: 构建成功。

- [ ] **Step 5: 手动验证**

起盘明细页：输入邮箱可筛出该用户起盘（游客被排除）；选排盘起止日期可限定时间段；清除恢复全量。

- [ ] **Step 6: Commit**

```bash
git add frontend/src/lib/adminApi.ts frontend/src/pages/admin/AdminChartsPage.tsx
git commit -m "feat(admin): search charts by email and cast-time range"
```

---

## 自检对照（spec 覆盖）

- 合盘只读列表+详情 → Task 1-4 ✓
- 用户列表分页 Bug → Task 5 ✓
- 重置用户密码 → Task 6 + Task 9 ✓
- 禁用用户（disabled_at + 登录拦截）→ Task 7 + Task 9 ✓
- 删除用户（硬删除 + 强警告）→ Task 8 + Task 9 ✓
- 起盘按邮箱/排盘时间搜索 → Task 10-11 ✓
- 不在范围（合盘删除/搜索、角色权限、起盘删除/按出生日期）→ 未引入 ✓
