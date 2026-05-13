# 专业化文案与符号清理 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 去掉用户端所有「AI」字眼、emoji 及装饰性 Unicode 符号，使产品呈现为专业命理工具。

**Architecture:** 纯文本替换，涉及 5 个页面文件和 4 个组件文件，无结构变更、无新函数、无后端改动。每个文件独立一个 Task，互不依赖，可顺序执行。

**Tech Stack:** React 18 + TypeScript，直接编辑 TSX 文件。

---

## File Structure

只修改，不新建：

| 文件 | 改动数量 |
|------|---------|
| `frontend/src/pages/HomePage.tsx` | 6 处 |
| `frontend/src/pages/ResultPage.tsx` | 9 处 |
| `frontend/src/pages/PastEventsPage.tsx` | 5 处 |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 2 处 |
| `frontend/src/pages/CompatibilityPage.tsx` | 1 处 |
| `frontend/src/components/DayunTimeline.tsx` | 3 处 |
| `frontend/src/components/YongshenBadge.tsx` | 1 处 |
| `frontend/src/components/ShareCard.tsx` | 1 处 |
| `frontend/src/components/LiuYueDrawer.tsx` | 2 处 |

---

### Task 1: HomePage.tsx

**Files:**
- Modify: `frontend/src/pages/HomePage.tsx`

- [ ] **Step 1: 修改 hero-badge 文字**

将第 73 行：
```tsx
<div className="hero-badge serif">✦ 八字命理 · AI 解读 ✦</div>
```
改为：
```tsx
<div className="hero-badge serif">八字命理 · 命理解读</div>
```

- [ ] **Step 2: 修改 hero-desc 文字**

将第 78 行：
```tsx
融合传统八字命理与人工智能，为你解读命盘中的天赋与机遇
```
改为：
```tsx
融合传统八字命理与现代算法，为你解读命盘中的天赋与机遇
```

- [ ] **Step 3: 修改提交按钮文字**

将第 143 行：
```tsx
<>✦ 立即起盘</>
```
改为：
```tsx
立即起盘
```

- [ ] **Step 4: 修改 guest-hint 文字**

将第 149 行：
```tsx
<a href="/login">登录</a>后可保存记录并获得 AI 智能解读报告
```
改为：
```tsx
<a href="/login">登录</a>后可保存记录并获得完整解读报告
```

- [ ] **Step 5: 修改特性卡片数组**

将第 161–164 行的 features 数组：
```tsx
{ icon: '◉', title: '传统算法', desc: '基于 lunar-go 天文历法库，精确到秒级节气与真太阳时' },
{ icon: '✦', title: 'AI 智能解读', desc: '大模型结合命理知识，生成通俗易懂的个性报告' },
{ icon: '◈', title: '五行分析', desc: '可视化五行分布，直观了解命局特点' },
```
改为：
```tsx
{ icon: '', title: '传统算法', desc: '基于 lunar-go 天文历法库，精确到秒级节气与真太阳时' },
{ icon: '', title: '命理解读', desc: '结合命理知识，生成通俗易懂的个性报告' },
{ icon: '', title: '五行分析', desc: '可视化五行分布，直观了解命局特点' },
```

- [ ] **Step 6: 修改错误提示符号**

将第 129 行：
```tsx
{error && <p className="form-error">⚠ {error}</p>}
```
改为：
```tsx
{error && <p className="form-error">{error}</p>}
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/HomePage.tsx
git commit -m "chore(ui): HomePage 去掉 AI 字眼与装饰符号"
```

---

### Task 2: ResultPage.tsx

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`

- [ ] **Step 1: 修改喜用/忌神加载占位**

将第 349 行和第 352 行（出现两次）：
```tsx
reportLoading ? 'AI测算中...' : '待生成'
```
改为：
```tsx
reportLoading ? '测算中...' : '待生成'
```
（共 2 处，两行相同，均需修改）

- [ ] **Step 2: 修改命理头像章节标题**

将第 482 行（Feature Flag 区块内）：
```tsx
<h2 className="section-title serif">✦ 专属命理头像</h2>
```
改为：
```tsx
<h2 className="section-title serif">专属命理头像</h2>
```

- [ ] **Step 3: 修改命理解读章节标题**

将第 502 行：
```tsx
<h2 className="section-title serif">✦ AI 命理解读</h2>
```
改为：
```tsx
<h2 className="section-title serif">命理解读</h2>
```

- [ ] **Step 4: 删除 report-summary-icon span**

将第 554–556 行：
```tsx
<div className="report-summary">
  <span className="report-summary-icon">✦</span>
  <span>{structured.analysis.summary}</span>
</div>
```
改为：
```tsx
<div className="report-summary">
  <span>{structured.analysis.summary}</span>
</div>
```

- [ ] **Step 5: 修改免责声明**

将第 580 行：
```tsx
本报告由 AI 辅助生成，内容仅供参考，不构成任何决策建议。
```
改为：
```tsx
本报告内容仅供参考，不构成任何决策建议。
```

- [ ] **Step 6: 修改思考加载状态**

将第 603 行：
```tsx
🧠 AI 正在深度推理中...  已思考 {thinkingSeconds} 秒
```
改为：
```tsx
正在深度推演中...  已思考 {thinkingSeconds} 秒
```

- [ ] **Step 7: 修改错误提示符号**

将第 625 行：
```tsx
<p className="form-error" style={{ margin: '12px 0' }}>⚠ {reportError}</p>
```
改为：
```tsx
<p className="form-error" style={{ margin: '12px 0' }}>{reportError}</p>
```

- [ ] **Step 8: 修改 CTA 描述文字和按钮**

将第 634 行：
```tsx
点击下方按钮，AI 将根据你的命盘生成性格、感情、事业、健康四维解读
```
改为：
```tsx
点击下方按钮，生成性格、感情、事业、健康四维解读
```

将第 641 行：
```tsx
生成 AI 命理解读
```
改为：
```tsx
生成命理解读
```

- [ ] **Step 9: 修改未登录提示**

将第 646 行：
```tsx
<span>登录后可获得完整 AI 解读报告，并保存命盘记录</span>
```
改为：
```tsx
<span>登录后可获得完整解读报告，并保存命盘记录</span>
```

- [ ] **Step 10: Commit**

```bash
git add frontend/src/pages/ResultPage.tsx
git commit -m "chore(ui): ResultPage 去掉 AI 字眼、emoji 与装饰符号"
```

---

### Task 3: PastEventsPage.tsx

**Files:**
- Modify: `frontend/src/pages/PastEventsPage.tsx`

- [ ] **Step 1: 修改状态栏加载提示**

将第 173 行：
```tsx
'年份已就绪 · 大运 AI 总结正在后台生成'
```
改为：
```tsx
'年份已就绪 · 大运总结正在后台生成'
```

- [ ] **Step 2: 修改底部统计文字**

将第 208 行：
```tsx
共推算 {events.length} 个年份 · 算法即时生成 · 大运总结由 AI 后台生成
```
改为：
```tsx
共推算 {events.length} 个年份 · 算法即时生成 · 大运总结后台生成
```

- [ ] **Step 3: 修改大运总结加载占位**

将第 247 行：
```tsx
AI 正在生成本段大运总结……
```
改为：
```tsx
正在生成本段大运总结……
```

- [ ] **Step 4: 修改流中断错误提示**

将第 379 行：
```tsx
AI 大运总结流中断：{streamError}
```
改为：
```tsx
大运总结生成中断：{streamError}
```

- [ ] **Step 5: 修改免责声明**

将第 400 行：
```tsx
本推算基于八字命理算法与 AI 语言生成，仅供参考，不构成任何决策建议。
```
改为：
```tsx
本推算内容仅供参考，不构成任何决策建议。
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/pages/PastEventsPage.tsx
git commit -m "chore(ui): PastEventsPage 去掉 AI 字眼"
```

---

### Task 4: CompatibilityResultPage.tsx

**Files:**
- Modify: `frontend/src/pages/CompatibilityResultPage.tsx`

- [ ] **Step 1: 修改合盘解读章节标题**

将第 370 行：
```tsx
<h2 className="serif compatibility-section-title" style={{ margin: 0 }}>AI 合盘解读</h2>
```
改为：
```tsx
<h2 className="serif compatibility-section-title" style={{ margin: 0 }}>合盘解读</h2>
```

- [ ] **Step 2: 修改空状态文字**

将第 405 行：
```tsx
<p style={{ margin: 0, color: 'var(--text-muted)' }}>尚未生成 AI 合盘解读。</p>
```
改为：
```tsx
<p style={{ margin: 0, color: 'var(--text-muted)' }}>尚未生成合盘解读。</p>
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/pages/CompatibilityResultPage.tsx
git commit -m "chore(ui): CompatibilityResultPage 去掉 AI 字眼"
```

---

### Task 5: CompatibilityPage.tsx

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.tsx`

- [ ] **Step 1: 修改描述文字**

将第 57 行：
```tsx
第一版采用双盘结构化分析：吸引力、稳定度、沟通协同、现实磨合四维结果，再结合 AI 生成完整解读。
```
改为：
```tsx
第一版采用双盘结构化分析：吸引力、稳定度、沟通协同、现实磨合四维结果，生成完整解读。
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/pages/CompatibilityPage.tsx
git commit -m "chore(ui): CompatibilityPage 去掉 AI 字眼"
```

---

### Task 6: DayunTimeline.tsx

**Files:**
- Modify: `frontend/src/components/DayunTimeline.tsx`

- [ ] **Step 1: 修改交运时间标注**

将第 112 行：
```tsx
✦ 精确交运时间：{startYunSolar} ✦
```
改为：
```tsx
精确交运时间：{startYunSolar}
```

- [ ] **Step 2: 删除大运面板标题前的 ✦ span**

将第 221–224 行：
```tsx
<div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-accent)', marginBottom: 16, display: 'flex', alignItems: 'center', gap: 8 }}>
  <span>✦</span>
  <span>{activeDayun.gan}{activeDayun.zhi}大运流年</span>
</div>
```
改为：
```tsx
<div style={{ fontSize: 14, fontWeight: 600, color: 'var(--text-accent)', marginBottom: 16 }}>
  {activeDayun.gan}{activeDayun.zhi}大运流年
</div>
```

- [ ] **Step 3: 删除交脱日期前的 ⚡ span**

将第 277–279 行：
```tsx
<div style={{ 
  position: 'absolute', top: -8, background: 'var(--bg-elevated)', padding: '0 4px', fontSize: 9, 
  color: 'var(--wu-jin)', display: 'flex', alignItems: 'center', gap: 2, borderRadius: 2
}}>
  <span>⚡</span>{ln.trans_month}月{ln.trans_day}日交脱
</div>
```
改为：
```tsx
<div style={{ 
  position: 'absolute', top: -8, background: 'var(--bg-elevated)', padding: '0 4px', fontSize: 9, 
  color: 'var(--wu-jin)', borderRadius: 2
}}>
  {ln.trans_month}月{ln.trans_day}日交脱
</div>
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/DayunTimeline.tsx
git commit -m "chore(ui): DayunTimeline 去掉装饰符号"
```

---

### Task 7: YongshenBadge.tsx

**Files:**
- Modify: `frontend/src/components/YongshenBadge.tsx`

- [ ] **Step 1: 修改标题**

将第 70 行：
```tsx
<h3 className="yongshen-badge-title serif">✦ 命元特质</h3>
```
改为：
```tsx
<h3 className="yongshen-badge-title serif">命元特质</h3>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/YongshenBadge.tsx
git commit -m "chore(ui): YongshenBadge 去掉装饰符号"
```

---

### Task 8: ShareCard.tsx

**Files:**
- Modify: `frontend/src/components/ShareCard.tsx`

- [ ] **Step 1: 修改品牌标识文字**

将第 135 行：
```tsx
✦ 缘 聚 命 理 ✦
```
改为：
```tsx
缘 聚 命 理
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/ShareCard.tsx
git commit -m "chore(ui): ShareCard 去掉装饰符号"
```

---

### Task 9: LiuYueDrawer.tsx

**Files:**
- Modify: `frontend/src/components/LiuYueDrawer.tsx`

- [ ] **Step 1: 修改流年批断标题**

将第 219 行：
```tsx
✨ {year}年运势精批
```
改为：
```tsx
{year}年运势精批
```

- [ ] **Step 2: 修改描述文字**

将第 225 行：
```tsx
让 AI 结合你的原局喜忌，详细推演本年运势全景
```
改为：
```tsx
结合原局喜忌，详细推演本年运势全景
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/components/LiuYueDrawer.tsx
git commit -m "chore(ui): LiuYueDrawer 去掉 AI 字眼与 emoji"
```

---

### Task 10: 最终验证

- [ ] **Step 1: TypeScript 编译检查**

```bash
cd /Users/liujiming/web/yuanju/frontend && npx tsc --noEmit
```
预期输出：无任何报错（0 errors）

- [ ] **Step 2: 全局搜索剩余 AI 字眼（用户端）**

```bash
grep -rn "AI\|人工智能\|✦\|✨\|🧠\|⚡\|◉\|◈\|⚠" \
  frontend/src/pages/HomePage.tsx \
  frontend/src/pages/ResultPage.tsx \
  frontend/src/pages/PastEventsPage.tsx \
  frontend/src/pages/CompatibilityResultPage.tsx \
  frontend/src/pages/CompatibilityPage.tsx \
  frontend/src/components/DayunTimeline.tsx \
  frontend/src/components/YongshenBadge.tsx \
  frontend/src/components/ShareCard.tsx \
  frontend/src/components/LiuYueDrawer.tsx
```
预期：仅剩代码注释中的 `// AI 解读状态` 等（不影响用户端显示），以及类型名 `AIReport` 等——UI 文字中不应有任何匹配项。

- [ ] **Step 3: Final commit**

```bash
git add -A
git commit -m "chore(ui): 专业化文案清理完成，去掉用户端所有 AI 字眼与装饰符号"
```
