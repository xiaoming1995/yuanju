# Token 用量按模型拆分展示设计文档

**日期：** 2026-05-12
**范围：** 在现有 token 用量汇总页，将每用户一行改为按 (用户 × 模型) 拆分多行展示，并在明细抽屉加模型筛选器。

---

## 目标

管理员可清晰看到每个用户在不同模型（Flash / Pro）上分别消耗了多少 token 和费用，不再是合并后的黑盒数字。

---

## 架构

**核心思路：** 后端返回平铺的 per-(user, model) 数组，前端按 user_id 分组渲染，合计行由前端计算。

```
GetTokenUsageSummary → 平铺 per-(user,model) 数组
                              │
                              ▼
              前端按 user_id 分组
              ├── 第一行：email + 第一个模型数据
              ├── 后续行：空白 + 其他模型数据
              └── 合计行：空白 + 跨模型求和 + 明细按钮

GetTokenUsageDetail + model 参数 → 明细抽屉模型筛选
```

**不改动：** `CreateTokenUsageLog`、DB 表结构、其他 Admin 接口。

---

## 后端改动

### 1. `internal/repository/token_usage_repository.go`

**`TokenUsageSummaryRow` 新增 `Model` 字段：**
```go
type TokenUsageSummaryRow struct {
    UserID           string  `json:"user_id"`
    Email            string  `json:"email"`
    Nickname         string  `json:"nickname"`
    Model            string  `json:"model"`
    RequestCount     int     `json:"request_count"`
    PromptTokens     int     `json:"prompt_tokens"`
    CompletionTokens int     `json:"completion_tokens"`
    TotalTokens      int     `json:"total_tokens"`
    EstimatedCostCny float64 `json:"estimated_cost_cny"`
}
```

**`GetTokenUsageSummary` 逻辑变化：**
- SQL 保持 `GROUP BY u.id, u.email, u.nickname, t.model`（已有）
- 移除现有 Go 侧聚合回单用户行的逻辑
- 新逻辑：Go 侧计算每个用户的总 token 用于排序（map[userID]int），按总 token 降序排列用户顺序，再按此顺序展开为平铺数组
- 同一用户内多个模型行按该模型 total_tokens 降序排列

**`GetTokenUsageDetail` 新增 `model` 参数：**
```go
func GetTokenUsageDetail(userID string, from, to time.Time, page, limit int,
    model string, costFn func(string, int, int) float64) (total int, items []TokenUsageDetailRow, err error)
```
WHERE 子句加：`AND ($6 = '' OR COALESCE(model, '') = $6)`

### 2. `internal/handler/token_usage_handler.go`

**`AdminGetTokenUsageDetail`** 从 query param 读 `model`（默认空字符串），传给 repository：
```go
model := c.DefaultQuery("model", "")
total, items, err := repository.GetTokenUsageDetail(userID, from, to, page, limit, model, service.CalcCost)
```

---

## 前端改动（`TokenUsagePage.tsx`）

### Interface 更新

`SummaryRow` 新增 `model: string`：
```tsx
interface SummaryRow {
  user_id: string
  email: string
  nickname: string
  model: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  estimated_cost_cny: number
}
```

### 汇总表分组渲染

`summary: SummaryRow[]` 保持平铺数组，渲染时按 `user_id` 分组：

```
用户邮箱          | 昵称 | 模型              | 请求次数 | 输入    | 输出    | 总tokens  | 预估费用
user@email.com   | 张三 | deepseek-v4-flash |   8    | 80,000  | 20,000  | 100,000   | ¥ 0.049
（空白）          |（空）| deepseek-v4-pro   |   2    | 10,000  |  5,000  |  15,000   | ¥ 0.120
（合计行）        |（空）| 合计              |  10    | 90,000  | 25,000  | 115,000   | ¥ 0.169  [明细]
```

**渲染规则：**
- 每组第一行：显示 `email`、`nickname`、模型数据
- 后续行：`email` 和 `nickname` 显示空白（`''`），只显示模型和 token 数据
- 合计行（前端计算）：`model` 显示"合计"，token/cost 求和，背景色 `#1e1e40`，加粗，明细按钮在此行
- 每个用户组之间加 `<tr>` 高度为 4px 的分隔行（`background: #0d0d1a`）

### State 新增

```tsx
const [drawerModel, setDrawerModel] = useState('')
```
打开新用户抽屉时重置 `drawerModel = ''`。

### 明细抽屉：模型筛选 tabs

抽屉顶部、标题下方加模型筛选区域：

```tsx
// 可用模型列表从 summary 中该用户的行推导
const userModels = summary.filter(r => r.user_id === drawerUser.user_id).map(r => r.model)

// 渲染 tabs
['', ...userModels].map(m => (
  <button
    key={m}
    onClick={() => { setDrawerModel(m); openDetail(drawerUser, 1, m) }}
    style={{
      background: drawerModel === m ? '#a78bfa' : 'transparent',
      color: drawerModel === m ? '#fff' : '#888',
      border: '1px solid #333',
      borderRadius: 4, padding: '4px 12px', fontSize: 12, cursor: 'pointer'
    }}
  >
    {m === '' ? '全部' : m}
  </button>
))
```

`openDetail` 函数签名扩展为接受可选 `model` 参数，detail API 请求加 `model` query param：
```tsx
const openDetail = async (row: SummaryRow, page = 1, model = drawerModel) => { ... }
// API call:
adminTokenUsageAPI.detail(row.user_id, from, to, page, detailLimit, model)
```

`adminTokenUsageAPI.detail` 签名更新：
```ts
detail: (userId: string, from: string, to: string, page: number, limit: number, model = '') =>
  adminApi.get(`/api/admin/token-usage/detail?user_id=${userId}&from=${from}&to=${to}&page=${page}&limit=${limit}${model ? `&model=${encodeURIComponent(model)}` : ''}`)
```

---

## 文件改动清单

| 文件 | 改动类型 |
|------|---------|
| `backend/internal/repository/token_usage_repository.go` | 改 struct + 改 `GetTokenUsageSummary` + 改 `GetTokenUsageDetail` 签名 |
| `backend/internal/handler/token_usage_handler.go` | `AdminGetTokenUsageDetail` 读 model param |
| `frontend/src/lib/adminApi.ts` | `adminTokenUsageAPI.detail` 加 model 参数 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | interface 更新 + 分组渲染 + 明细 tabs |

---

## 不在本次范围

- 汇总表导出
- 跨用户的模型费用汇总（全局 by-model 视图）
- 明细抽屉内按 call_type 筛选
