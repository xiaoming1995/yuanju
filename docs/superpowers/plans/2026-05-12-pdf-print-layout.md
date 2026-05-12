# PDF 打印布局优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增专属打印布局组件 `PrintLayout`，替换失效的 `@media print` CSS 变量覆写方案，使 PDF 导出格式正确、美观。

**Architecture:** 新建 `PrintLayout.tsx`（全内联样式，不依赖 CSS 变量），在 ResultPage 中将屏幕内容包裹在 `.screen-only` div，并将 `<PrintLayout>` 作为兄弟节点追加——打印时两者切换显示。导出 PDF 按钮改为不依赖 `report` 存在，始终在桌面端显示。

**Tech Stack:** React 18 + TypeScript，纯 CSS `@media print / @media screen`，无额外依赖

---

## 文件结构

| 文件 | 操作 | 说明 |
|------|------|------|
| `frontend/src/components/PrintLayout.tsx` | **新建** | 打印专属布局，全内联样式 |
| `frontend/src/pages/ResultPage.tsx` | **修改** | 引入 PrintLayout，包裹 screen-only，调整按钮条件 |
| `frontend/src/pages/ResultPage.css` | **修改** | 替换旧 `@media print` 块，添加 screen-only / print-only 规则 |

---

### Task 1: 新建 PrintLayout 组件

**Files:**
- Create: `frontend/src/components/PrintLayout.tsx`

- [ ] **Step 1: 创建文件，写入完整组件**

```tsx
import type { ShenshaAnnotation } from '../lib/api'

const WX_COLOR: Record<string, string> = {
  木: '#3d7a55', 火: '#c0392b', 土: '#8b6940', 金: '#7a6830', 水: '#2c5282',
}
function wxColor(wxStr: string): string {
  for (const [k, v] of Object.entries(WX_COLOR)) {
    if (wxStr?.startsWith(k)) return v
  }
  return '#333'
}

const SHA_COLOR = { ji: '#3d6b4f', xiong: '#b52525', zhong: '#666' } as const
const SHA_BG   = { ji: '#f0f7f2', xiong: '#fff0f0', zhong: '#f5f5f5' } as const

interface Pillar {
  label: string
  gan: string; zhi: string
  ganWx: string; zhiWx: string
  ganShiShen: string
  zhiShiShen: string[]
  diShi: string
  xunKong: string
  shenSha: string[]
}

interface DayunItem {
  index: number
  gan: string; zhi: string
  start_age: number; start_year: number; end_year: number
  gan_shishen: string; zhi_shishen: string; di_shi: string
}

interface ReportChapter {
  title: string; detail: string; brief: string
}

interface PrintLayoutProps {
  birthYear: number; birthMonth: number; birthDay: number; birthHour: number; gender: string
  yongshen: string; jishen: string
  pillars: Pillar[]
  dayun: DayunItem[]
  structured: { chapters: ReportChapter[]; analysis?: { logic: string; summary: string } | null } | null
  shenshaMap: Map<string, ShenshaAnnotation>
}

const thStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 12, fontWeight: 700, color: '#5a3a1a',
  border: '1px solid #e8dcc8', background: '#faf5eb',
}
const tdStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 13, color: '#333', border: '1px solid #eee',
}
const tdLabelStyle: React.CSSProperties = {
  padding: '8px 10px', textAlign: 'center',
  fontSize: 13, border: '1px solid #eee',
  fontWeight: 600, color: '#5a3a1a',
  background: '#faf5eb', whiteSpace: 'nowrap',
}
const sectionTitleStyle: React.CSSProperties = {
  fontSize: 14, fontWeight: 700, color: '#7a5c3a',
  letterSpacing: 2, marginBottom: 12,
  borderLeft: '3px solid #c9a84c', paddingLeft: 10,
}

export default function PrintLayout({
  birthYear, birthMonth, birthDay, birthHour, gender,
  yongshen, jishen, pillars, dayun, structured, shenshaMap,
}: PrintLayoutProps) {
  const chapters = structured?.chapters ?? []

  // 收集全部神煞（含注解）
  const allShensha: Array<{
    pillarLabel: string
    name: string
    annotation: ShenshaAnnotation | undefined
    polarity: 'ji' | 'xiong' | 'zhong'
  }> = []
  pillars.forEach(p => {
    p.shenSha.forEach(name => {
      const annotation = shenshaMap.get(name)
      allShensha.push({
        pillarLabel: p.label,
        name,
        annotation,
        polarity: annotation?.polarity ?? 'zhong',
      })
    })
  })

  return (
    <div
      className="print-only"
      style={{
        fontFamily: '"Noto Serif SC", "SimSun", serif',
        color: '#1a1a1a', background: '#fff',
        padding: '24px 32px', maxWidth: 800, margin: '0 auto',
      }}
    >
      {/* 品牌头部 */}
      <div style={{ textAlign: 'center', borderBottom: '2px solid #c9a84c', paddingBottom: 16, marginBottom: 24 }}>
        <div style={{ fontSize: 26, fontWeight: 700, letterSpacing: 8, color: '#3a2416', marginBottom: 6 }}>
          缘 聚 命 理
        </div>
        <div style={{ fontSize: 13, color: '#666', letterSpacing: 1 }}>
          {birthYear}年{birthMonth}月{birthDay}日 {birthHour}时
          &nbsp;·&nbsp;{gender === 'male' ? '男命' : '女命'}
        </div>
      </div>

      {/* 四柱 */}
      <div style={{ marginBottom: 28 }}>
        <div style={sectionTitleStyle}>四 柱</div>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={thStyle}>柱</th>
              {pillars.map(p => (
                <th key={p.label} style={{ ...thStyle, color: p.label === '日柱' ? '#c9a84c' : '#3a2416' }}>
                  {p.label}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            <tr>
              <td style={tdLabelStyle}>主星</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.ganShiShen}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>天干</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdStyle, fontWeight: 700, fontSize: 22, color: wxColor(p.ganWx) }}>
                  {p.gan}
                  <span style={{ fontSize: 10, color: '#999', marginLeft: 2 }}>({p.ganWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelStyle}>地支</td>
              {pillars.map(p => (
                <td key={p.label} style={{ ...tdStyle, fontWeight: 700, fontSize: 22, color: wxColor(p.zhiWx) }}>
                  {p.zhi}
                  <span style={{ fontSize: 10, color: '#999', marginLeft: 2 }}>({p.zhiWx})</span>
                </td>
              ))}
            </tr>
            <tr>
              <td style={tdLabelStyle}>副星</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.zhiShiShen.join(' / ')}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>星运</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.diShi}</td>)}
            </tr>
            <tr>
              <td style={tdLabelStyle}>空亡</td>
              {pillars.map(p => <td key={p.label} style={tdStyle}>{p.xunKong}</td>)}
            </tr>
          </tbody>
        </table>
      </div>

      {/* 喜用神 / 忌神 */}
      {(yongshen || jishen) && (
        <div style={{ marginBottom: 28, display: 'flex', gap: 20, flexWrap: 'wrap' }}>
          {yongshen && (
            <div style={{ padding: '8px 16px', border: '1px solid #b5d6c3', borderRadius: 6, background: '#f4fbf7', fontSize: 13 }}>
              <span style={{ color: '#666', marginRight: 6 }}>喜用神</span>
              <span style={{ fontWeight: 700, color: '#3d6b4f' }}>{yongshen}</span>
            </div>
          )}
          {jishen && (
            <div style={{ padding: '8px 16px', border: '1px solid #f5c6c6', borderRadius: 6, background: '#fff5f5', fontSize: 13 }}>
              <span style={{ color: '#666', marginRight: 6 }}>忌神</span>
              <span style={{ fontWeight: 700, color: '#b52525' }}>{jishen}</span>
            </div>
          )}
        </div>
      )}

      {/* 神煞（内联注解） */}
      {allShensha.length > 0 && (
        <div style={{ marginBottom: 28 }}>
          <div style={sectionTitleStyle}>神 煞</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {allShensha.map((sha, i) => (
              <div
                key={i}
                style={{
                  display: 'flex', alignItems: 'flex-start', gap: 10,
                  padding: '8px 12px',
                  background: SHA_BG[sha.polarity],
                  borderRadius: 6,
                  borderLeft: `3px solid ${SHA_COLOR[sha.polarity]}`,
                  pageBreakInside: 'avoid', breakInside: 'avoid',
                }}
              >
                <span style={{ fontSize: 11, color: '#888', whiteSpace: 'nowrap', paddingTop: 2, minWidth: 28 }}>
                  {sha.pillarLabel}
                </span>
                <span style={{ fontWeight: 700, color: SHA_COLOR[sha.polarity], whiteSpace: 'nowrap', minWidth: 64 }}>
                  {sha.name}
                </span>
                {sha.annotation?.short_desc && (
                  <span style={{ fontSize: 12, color: '#444', lineHeight: 1.6 }}>
                    {sha.annotation.short_desc}
                  </span>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 命理解读 */}
      <div style={{ marginBottom: 28, pageBreakBefore: 'always', breakBefore: 'page' }}>
        <div style={sectionTitleStyle}>命 理 解 读</div>
        {chapters.length === 0 ? (
          <div style={{ fontSize: 13, color: '#999', fontStyle: 'italic', padding: '16px 0' }}>
            命理解读尚未生成
          </div>
        ) : (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
            {chapters.map((ch, i) => (
              <div key={i} style={{ pageBreakInside: 'avoid', breakInside: 'avoid' }}>
                <div style={{
                  fontSize: 14, fontWeight: 700, color: '#3a2416',
                  marginBottom: 8, paddingBottom: 6,
                  borderBottom: '1px dashed #e0cca0',
                }}>
                  {ch.title}
                </div>
                <div style={{
                  fontSize: 13, color: '#333', lineHeight: 1.9,
                  fontFamily: '"Noto Sans SC", "PingFang SC", sans-serif',
                }}>
                  {ch.detail || ch.brief}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* 大运总览（竖向列表） */}
      <div style={{ marginBottom: 28, pageBreakBefore: 'always', breakBefore: 'page' }}>
        <div style={sectionTitleStyle}>大 运 总 览</div>
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 12 }}>
          <thead>
            <tr>
              <th style={thStyle}>段</th>
              <th style={thStyle}>起止岁</th>
              <th style={thStyle}>起止年</th>
              <th style={thStyle}>干支</th>
              <th style={thStyle}>天干十神</th>
              <th style={thStyle}>地支十神</th>
              <th style={thStyle}>星运</th>
            </tr>
          </thead>
          <tbody>
            {dayun.map(d => (
              <tr key={d.index} style={{ borderBottom: '1px solid #eee' }}>
                <td style={tdStyle}>第{d.index + 1}运</td>
                <td style={tdStyle}>{d.start_age}—{d.start_age + 9}岁</td>
                <td style={tdStyle}>{d.start_year}—{d.end_year}</td>
                <td style={{ ...tdStyle, fontWeight: 700, fontSize: 18 }}>{d.gan}{d.zhi}</td>
                <td style={tdStyle}>{d.gan_shishen}</td>
                <td style={tdStyle}>{d.zhi_shishen}</td>
                <td style={tdStyle}>{d.di_shi}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* 品牌落款 */}
      <div style={{
        borderTop: '1px solid #e0cca0', paddingTop: 12,
        display: 'flex', justifyContent: 'space-between',
        fontSize: 11, color: '#999',
      }}>
        <span>本报告内容仅供参考，不构成任何决策建议。</span>
        <span style={{ color: '#c9a84c' }}>yuanju.com</span>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 验证 TypeScript 编译**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -30
```

期望：无报错（或仅有与新文件无关的已存在警告）

- [ ] **Step 3: 提交**

```bash
git add frontend/src/components/PrintLayout.tsx
git commit -m "feat(ui): 新增 PrintLayout 打印专属布局组件"
```

---

### Task 2: 修改 ResultPage.tsx

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`

- [ ] **Step 1: 引入 PrintLayout**

在文件顶部 import 区域，在 `import ShareCard` 下方添加：

```tsx
import PrintLayout from '../components/PrintLayout'
```

- [ ] **Step 2: 修改 return 语句，用 Fragment 包裹，添加 screen-only**

将 `return (` 后的最外层 div 改为 Fragment：

原：
```tsx
  return (
    <div className="result-page page">
      <div className="container">
```

改为：
```tsx
  return (
    <>
      <div className="result-page page screen-only">
        <div className="container">
```

对应末尾，原：
```tsx
    </div>
  )
}
```

改为（在最后一个 `</div>` 后、`}` 前插入 PrintLayout 并闭合 Fragment）：
```tsx
      </div>
      <PrintLayout
        birthYear={result.birth_year}
        birthMonth={result.birth_month}
        birthDay={result.birth_day}
        birthHour={result.birth_hour}
        gender={result.gender}
        yongshen={result.yongshen || ''}
        jishen={result.jishen || ''}
        pillars={pillars}
        dayun={result.dayun}
        structured={structured}
        shenshaMap={shenshaMap}
      />
    </>
  )
}
```

注意：ResultPage 末尾当前结构（约 729-731 行）：
```tsx
      )}
    </div>   ← 这是 .result-page page 的关闭
  )
}
```
在这个 `</div>` 后、`)` 前插入 `<PrintLayout ... />` 并加上 `</>` 替换原来的 `)`。

- [ ] **Step 3: 导出 PDF 按钮改为始终显示（不依赖 report）**

找到约 499-524 行的报告区域头部：

原：
```tsx
          <div className="report-section-header">
            <h2 className="section-title serif">命理解读</h2>
            {report && (
              <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                <button
                  id="save-card-btn"
                  className="btn btn-ghost btn-sm"
                  onClick={handleSaveImage}
                  disabled={savingImage}
                >
                  {savingImage ? '生成中...' : '保存分享图'}
                </button>
                {/* PDF 导出移动端不支持，仅桂面展示 */}
                {!isMobileDevice && (
                  <button
                    id="export-report-btn"
                    className="btn btn-ghost btn-sm"
                    onClick={() => window.print()}
                  >
                    导出 PDF
                  </button>
                )}
              </div>
            )}
          </div>
```

改为：
```tsx
          <div className="report-section-header">
            <h2 className="section-title serif">命理解读</h2>
            <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
              {report && (
                <button
                  id="save-card-btn"
                  className="btn btn-ghost btn-sm"
                  onClick={handleSaveImage}
                  disabled={savingImage}
                >
                  {savingImage ? '生成中...' : '保存分享图'}
                </button>
              )}
              {!isMobileDevice && (
                <button
                  id="export-report-btn"
                  className="btn btn-ghost btn-sm"
                  onClick={() => window.print()}
                >
                  导出 PDF
                </button>
              )}
            </div>
          </div>
```

- [ ] **Step 4: 验证 TypeScript 编译**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -30
```

期望：无报错

- [ ] **Step 5: 提交**

```bash
git add frontend/src/pages/ResultPage.tsx
git commit -m "feat(ui): ResultPage 引入 PrintLayout，导出按钮始终显示"
```

---

### Task 3: 替换 ResultPage.css 中的打印 CSS

**Files:**
- Modify: `frontend/src/pages/ResultPage.css`

- [ ] **Step 1: 替换旧 @media print 块**

找到约 638-728 行的整个旧打印 CSS 块：

```css
/* ===== 打印/导出 PDF 样式 ===== */
@media print {
  ...（整个块，约 90 行）...
}
```

将其替换为以下新规则：

```css
/* ===== 打印 / PDF 样式 ===== */

/* 屏幕视图：隐藏打印布局 */
@media screen {
  .print-only {
    display: none !important;
  }
}

/* 打印时：隐藏屏幕内容，仅渲染打印布局 */
@media print {
  .screen-only {
    display: none !important;
  }
}
```

- [ ] **Step 2: 验证 CSS 文件末尾结构正常**

```bash
tail -30 frontend/src/pages/ResultPage.css
```

期望：新规则在文件末尾，`@keyframes blink` 等原有动画 keyframes 保持在原位（不受影响）

- [ ] **Step 3: 提交**

```bash
git add frontend/src/pages/ResultPage.css
git commit -m "fix(ui): 替换失效的打印 CSS 变量覆写为 screen-only/print-only 隔离方案"
```

---

### Task 4: 构建验证

**Files:** 无新文件

- [ ] **Step 1: 运行完整前端构建**

```bash
cd frontend && npm run build 2>&1 | tail -20
```

期望：
```
✓ built in Xs
```
无 TypeScript 错误，无 Vite 错误。

- [ ] **Step 2: 启动开发服务，手动验证打印效果**

```bash
cd frontend && npm run dev
```

打开 `http://localhost:5200`，起盘后进入结果页，按 `Cmd+P`（Mac）或 `Ctrl+P`（Windows）呼出打印对话框：

验证清单：
- [ ] 打印预览中显示「缘 聚 命 理」品牌头
- [ ] 四柱表格清晰，白底黑字，天干地支有五行颜色
- [ ] 喜用神/忌神显示正确
- [ ] 神煞行含注解说明文字
- [ ] 大运总览为竖向表格
- [ ] 品牌落款文案为「本报告内容仅供参考，不构成任何决策建议。」
- [ ] 屏幕端页面无变化（导航、按钮、交互元素正常）
- [ ] 「导出 PDF」按钮在无报告时也可见

- [ ] **Step 3: 提交收尾并推送**

```bash
git push
```
