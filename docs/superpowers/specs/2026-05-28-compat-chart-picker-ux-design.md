# 合盘命盘档案选择器 UX 优化设计

> **Status:** Draft · brainstorming output
> **Author:** Claude (Opus 4.7) + 用户
> **Date:** 2026-05-28
> **Related artifacts:** `frontend/src/pages/CompatibilityPage.tsx`, `frontend/src/pages/CompatibilityPage.css`, `frontend/src/pages/HistoryPage.tsx`

## 1. 背景与问题

合盘页（CompatibilityPage）的"从命盘档案选择"功能通过一个 modal picker 实现。当前实现见 `CompatibilityPage.tsx:347-380`，每条记录只展示：

- 主标题：`display_name` 或 fallback `${year_gan}${year_zhi} ${day_gan}${day_zhi}`
- 副标题：`YYYY-M-D H:00`

实际使用中存在以下用户体验问题：

| # | 问题 | 影响 |
|---|------|------|
| P1 | 未命名命盘的 fallback 只有两对干支（如"庚午 庚子"），普通用户无法识别 | 必须靠生日二次确认，选错风险高 |
| P2 | 列表不显示性别 | 挑"对方"时无法快速过滤 |
| P3 | 没有"已选中"标记 | 重开 picker 不知道上次选了哪个，容易重复选 |
| P4 | 没有搜索/筛选 | 档案 10+ 张时只能上下扫 |
| P5 | 空状态只有文字"先新建命盘"，无 CTA | 用户必须主动想到去 / 起盘 |
| P6 | 没有同盘冲突提示 | 可静默把同一张盘塞给"我"和"对方" |
| P7 | 手机端 panel 固定 520px，小屏挤 | 移动端不易点击 |

目标：打开 picker → **3 秒内**识别正确命盘并选中，即使档案有 10+ 张。

## 2. 范围

### 2.1 In scope

- `CompatibilityPage.tsx` 中的 picker JSX 改造
- `CompatibilityPage.css` 新增 picker 相关样式
- 新建 `frontend/src/lib/chartLabel.ts`：抽出 `chartDisplayName / chartFallbackName / formatPillars / genderText / formatBirth` 五个纯函数，由 HistoryPage + CompatibilityPage 共享
- HistoryPage.tsx 改用 `chartLabel.ts` 提供的函数（替换 11-29 行的本地函数定义）

### 2.2 Out of scope

- 命盘分组（按时段、按家族标签等）
- 批量选择/删除
- 自定义排序（按出生日期、按性别）
- 跨设备同步状态指示
- BirthProfileForm 自身的改动

## 3. 设计方案

### 3.1 架构

复用现有 modal 容器（`.compatibility-chart-picker`），picker 内部从"标题 + 列表"两层升级为"标题 + 工具栏 + 列表"三层。工具栏与列表过滤/排序逻辑放在 `CompatibilityPage` 函数体内的 useState + useMemo（不抽组件文件——picker JSX 升级后约 80-100 行，留在原文件即可；将来真复杂了再拆）。

### 3.2 共享工具 `frontend/src/lib/chartLabel.ts`

```ts
import type { BaziHistoryChart } from './api'

export function genderText(gender: string): string {
  return gender === 'female' ? '女命' : '男命'
}

export function chartFallbackName(chart: BaziHistoryChart): string {
  return `${genderText(chart.gender)} · ${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日`
}

export function chartDisplayName(chart: BaziHistoryChart): string {
  return chart.display_name?.trim() || chartFallbackName(chart)
}

export function formatPillars(chart: BaziHistoryChart): string {
  return `${chart.year_gan}${chart.year_zhi} · ${chart.month_gan}${chart.month_zhi} · ${chart.day_gan}${chart.day_zhi} · ${chart.hour_gan}${chart.hour_zhi}`
}

export function formatBirth(chart: BaziHistoryChart): string {
  return `${chart.birth_year}年${chart.birth_month}月${chart.birth_day}日 ${chart.birth_hour}时`
}
```

依赖：`BaziHistoryChart` 类型（已存在于 `lib/api.ts`）。无新增依赖。

### 3.3 Picker 内部状态

新增 picker 局部 state（与 picker lifecycle 绑定，关闭即重置）：

```ts
const [pickerSearch, setPickerSearch] = useState('')
const [pickerGenderFilter, setPickerGenderFilter] = useState<'all' | 'male' | 'female'>('all')
```

每次 `openChartPicker(role)` 被调用时：
- 重置 `pickerSearch = ''`
- 重置 `pickerGenderFilter`：
  - `role === 'partner'` → 设为 self 的异性（`selfProfile.gender === 'male' ? 'female' : 'male'`）
  - `role === 'self'` → 设为 `selfProfile.gender`（与自己同性）
  - selfProfile/partnerProfile 默认都有 gender（来自 `initialBirthProfile('male' | 'female')`），不存在未知分支

### 3.4 列表派生（useMemo）

```ts
const pickerCharts = useMemo(() => {
  const sameSideId = pickerRole === 'self' ? selfImportSource?.chartId : partnerImportSource?.chartId
  const otherSideId = pickerRole === 'self' ? partnerImportSource?.chartId : selfImportSource?.chartId
  const q = pickerSearch.trim().toLowerCase()

  return historyCharts
    .filter(c => pickerGenderFilter === 'all' || c.gender === pickerGenderFilter)
    .filter(c => !q || (c.display_name || '').toLowerCase().includes(q))
    .map(c => ({
      chart: c,
      isSelectedHere: c.id === sameSideId,
      isUsedByOtherSide: c.id === otherSideId,
    }))
}, [historyCharts, pickerGenderFilter, pickerSearch, pickerRole, selfImportSource, partnerImportSource])
```

排序不在 useMemo 内做额外处理——`historyCharts` 已按 `created_at desc` 返回，初始默认就是想要的顺序。仅在 `pickerGenderFilter` 默认设置上反映"异性优先"的偏好，不做混合排序（避免实现复杂度，YAGNI）。

### 3.5 Picker 渲染（JSX 结构）

```
<div role="dialog" aria-modal="true" aria-label="选择命盘档案">
  <div ref={chartPickerRef} tabIndex={-1}>
    <header>
      <strong>选择命盘档案</strong>
      <button>关闭</button>
    </header>
    <div className="...-toolbar">
      <input
        type="search"
        placeholder="搜索命盘称呼"
        value={pickerSearch}
        onChange={e => setPickerSearch(e.target.value)}
        autoFocus
        aria-label="搜索命盘称呼"
      />
      <div role="tablist" aria-label="性别筛选">
        {['all','male','female'].map(g => (
          <button
            role="tab"
            aria-selected={pickerGenderFilter === g}
            onClick={() => setPickerGenderFilter(g)}
          >
            {g === 'all' ? '全部' : g === 'male' ? '男命' : '女命'}
          </button>
        ))}
      </div>
    </div>
    {pickerCharts.length === 0 ? (
      historyCharts.length === 0
        ? <EmptyArchive />     // 档案为空
        : <EmptyFilter />      // 档案非空但过滤后无命中
    ) : (
      <div role="listbox" className="...-list">
        {pickerCharts.map(({ chart, isSelectedHere, isUsedByOtherSide }) => (
          <button
            role="option"
            aria-selected={isSelectedHere}
            onClick={() => { applyImportedChart(pickerRole, chart); setPickerRole(null) }}
          >
            <div className="row-head">
              <span className="name">{chartDisplayName(chart)}</span>
              <span className="badges">
                <span className="badge badge-gender">{genderText(chart.gender)}</span>
                {isSelectedHere && <span className="badge badge-selected">已选中</span>}
                {!isSelectedHere && isUsedByOtherSide && (
                  <span className="badge badge-conflict">
                    已作为{pickerRole === 'self' ? '对方' : '我'}
                  </span>
                )}
              </span>
            </div>
            <div className="row-sub">{formatBirth(chart)}</div>
            <div className="row-pillars serif">{formatPillars(chart)}</div>
          </button>
        ))}
      </div>
    )}
  </div>
</div>
```

#### 空状态分支

- **EmptyArchive**（`historyCharts.length === 0`）：
  ```
  还没有命盘档案
  [立即新建命盘]   ← <Link to="/" className="btn btn-primary">
  ```
- **EmptyFilter**（有档案但当前筛选 0 命中）：
  ```
  没有匹配的命盘
  [清除筛选]   ← onClick: setPickerSearch(''); setPickerGenderFilter('all')
  ```

### 3.6 行为细节

#### 已选中（badge-selected）
仅当 chart.id 等于**当前 picker role 对应**的 importSource.chartId 时显示。
- 例：picker 打开作为 'self'，且 selfImportSource.chartId === c.id → 显示"已选中"
- 不显示对方那侧的"已选中"——避免混淆

#### 已被另一侧使用（badge-conflict）
当 chart.id 已被对方 importSource 使用：
- 显示禁用感的小徽章 `已作为${otherRole}`（实际 CSS 用低饱和度色）
- **仍然可点击**——用户可能就是想交换角色
- 不弹确认框（信任用户操作，避免拦截感）

#### 排序
不主动重排序。`historyCharts` 已经是 `created_at desc`，通过性别筛选 default 选异性即可达到"对方挑选异性优先"的体感。

### 3.7 CSS（新增样式）

新增的 CSS class 集中放在 `CompatibilityPage.css` 内 picker 现有规则之后（约 269 行后）：

```
.compatibility-chart-picker-toolbar { ... }       // 搜索框 + 筛选 chip 容器
.compatibility-chart-picker-search { ... }        // 搜索 input
.compatibility-chart-picker-filter { ... }        // chip 组容器（role=tablist）
.compatibility-chart-picker-filter button { ... } // 单个 chip
.compatibility-chart-picker-filter button[aria-selected="true"] { ... }
.compatibility-chart-picker-item-head { ... }     // 主标题 + badge 行
.compatibility-chart-picker-item-badges { ... }   // 右侧 badge 容器
.compatibility-chart-picker-badge-gender { ... }
.compatibility-chart-picker-badge-selected { ... }    // 高亮金色
.compatibility-chart-picker-badge-conflict { ... }    // 低饱和度警示色
.compatibility-chart-picker-item-pillars { ... }      // 四柱小字
.compatibility-chart-picker-empty-cta { ... }         // 空态 CTA 按钮
```

手机端 ≤640px：使用 flex column 让 head/toolbar/list 自然堆叠，list 区域 `flex: 1 1 auto` 自己处理滚动，避免计算 sticky 偏移。

```css
@media (max-width: 640px) {
  .compatibility-chart-picker { padding: 0; }
  .compatibility-chart-picker-panel {
    width: 100%;
    height: 100vh;
    max-height: 100vh;
    border-radius: 0;
    border: none;
    display: flex;
    flex-direction: column;
  }
  .compatibility-chart-picker-head { flex: 0 0 auto; }
  .compatibility-chart-picker-toolbar { flex: 0 0 auto; }
  .compatibility-chart-picker-list { flex: 1 1 auto; max-height: none; }
}
```

### 3.8 A11y

- picker root 已有 `role="dialog" aria-modal="true"`
- 列表升级为 `role="listbox"`，单项为 `role="option" aria-selected={isSelectedHere}`
- 性别筛选用 `role="tablist"` + `role="tab" aria-selected`
- 搜索框 `autoFocus`（仅在 picker 打开时；现有 useEffect 已 focus panel，需改成 focus 搜索框）
- Esc 关闭已有，保留

## 4. 数据流

```
openChartPicker(role)
  └─ setPickerRole(role)
  └─ setPickerSearch('')
  └─ setPickerGenderFilter(default-by-role)
  └─ loadHistoryCharts()  (已存在，缓存命中则秒返)
       ↓
  user types in search / clicks gender chip
       ↓
  pickerCharts useMemo 重算
       ↓
  user clicks row
       ↓
  applyImportedChart(role, chart)  (已存在)
  setPickerRole(null)              // 关闭 picker，state 自然消失
```

无新增 API 调用，无新增父组件状态契约（`*ImportSource` 接口不变）。

## 5. 错误处理

- `loadHistoryCharts()` 失败：维持现有行为（通过 `setError` 显示在主页错误区）
- 搜索/筛选导致 0 命中：走 `EmptyFilter` 分支，提供"清除筛选"
- 同盘冲突：badge 提示即可，不拦截

## 6. 验证标准

完成后需通过以下手动验证：

1. **未命名命盘可识别**：选一张没有 display_name 的盘，列表项主标题显示为"男命 · 1996年2月8日"形式（非"庚午 庚子"）
2. **性别筛选**：点击"男命" chip，只显示男性命盘；点"女命"反之；点"全部"恢复
3. **partner 角色默认异性**：先填好 self（男），点对方的"从命盘档案选择" → picker 打开时 gender filter 默认在"女命"
4. **已选中标记**：导入命盘 A 作为 self → 重开 picker → A 显示"已选中"徽章
5. **同盘冲突徽章**：A 已作为 self → 打开 partner picker → A 显示"已作为我"徽章；点击 A 仍可成功导入为 partner（self 那侧仍保留 A，因为没有清除——确认这是预期，见 §7）
6. **空档案 CTA**：删光命盘（或新账号）→ 打开 picker → 看到"还没有命盘档案 / [立即新建命盘]"，点击跳转 `/`
7. **搜索无命中**：随便输入"xxxx" → 显示"没有匹配的命盘 / [清除筛选]"，点击清除后回到完整列表
8. **手机端**：DevTools 切到 iPhone SE 尺寸（375px）→ picker 全屏显示，head 和 toolbar sticky
9. **Esc 关闭**：picker 打开后按 Esc 关闭，焦点返回触发按钮

## 7. 已确认的设计决策

| 决策 | 选择 | 理由 |
|------|------|------|
| 是否抽组件文件 | 不抽 | picker JSX ~80-100 行，留在 CompatibilityPage 内；超过 ~150 行再拆 |
| 搜索范围 | 仅 display_name | 未命名命盘靠性别 chip + fallback 显示性别+生日即可定位；拼音/全文搜索 YAGNI |
| 是否引入虚拟列表 | 不引入 | 档案量级估计 < 50，普通 overflow:auto 够用 |
| 同盘冲突 | 仅徽章提示，不拦截 | 信任用户操作，避免确认框打断流程 |
| 同盘冲突时点击行为 | 不自动清除对方那侧 | 用户可能就是想让两侧都引用同一张盘（罕见但有效）；自动清除会"偷"状态。徽章已足够提示。 |
| 排序策略 | 不主动重排序，靠 partner 角色 default 选女命 | 实现简单；用户偏好性别过滤而非混合排序 |

## 8. Out of Scope（明确不做）

- 拼音/姓名首字母搜索
- 按出生年份/月份分组
- 命盘标签（家族、好友等元数据）
- 多选导入
- 命盘排序自定义
- Picker 历史使用频次统计
