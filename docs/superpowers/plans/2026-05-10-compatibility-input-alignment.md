# Compatibility Input Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让合盘页与单人测算页共用同一套生辰输入交互，消除年月日时数字输入带来的摩擦。

**Architecture:** 抽出一个前端共享的出生信息输入组件与辅助逻辑，负责历法切换、年月日联动、闰月和十二时辰选择。`HomePage` 与 `CompatibilityPage` 只保留各自页面状态和提交逻辑，统一复用该输入组件。

**Tech Stack:** React 19, TypeScript, Vite, lunar-javascript, CSS Variables

---

### Task 1: 抽出生辰输入共享组件

**Files:**
- Create: `frontend/src/components/BirthProfileForm.tsx`

- [ ] **Step 1: 设计共享输入数据结构**

定义组件入参，覆盖单人页与合盘页共同需要的字段：

```ts
export interface BirthProfileFormValue {
  year: number
  month: number
  day: number
  hour: number
  gender: 'male' | 'female'
  calendarType: 'solar' | 'lunar'
  isLeapMonth: boolean
}
```

- [ ] **Step 2: 实现年月日、闰月、时辰联动逻辑**

在组件内复用现有单人页逻辑，输出：

```ts
const years = Array.from({ length: 120 }, (_, i) => currentYear - i)
const monthOptions = value.calendarType === 'solar'
  ? Array.from({ length: 12 }, (_, i) => ({ value: i + 1, isLeap: false, label: `${i + 1} 月` }))
  : buildLunarMonthOptions(value.year)
const days = Array.from({ length: maxDay }, (_, i) => i + 1)
const handleDateChange = (updates: { year?: number; month?: number; isLeapMonth?: boolean }) => {
  onChange(getNormalizedNextValue(value, updates))
}
```

- [ ] **Step 3: 实现统一 UI**

组件渲染以下控件：

```tsx
<div className="gender-selector">{(['male', 'female'] as const).map(...)}</div>
<div className="gender-selector">{(['solar', 'lunar'] as const).map(...)}</div>
<div className="date-row">
  <select className="form-select" value={value.year} onChange={...} />
  <select className="form-select" value={`${value.month}-${value.isLeapMonth}`} onChange={...} />
  <select className="form-select" value={value.day} onChange={...} />
</div>
<select className="form-select" value={value.hour} onChange={...}>
  {SHICHEN.map(h => <option key={h.value} value={h.value}>{h.label}</option>)}
</select>
```

标题使用可选 `title` 参数，便于合盘页分别显示“我的生辰”“对方的生辰”。

### Task 2: 接入单人测算页

**Files:**
- Modify: `frontend/src/pages/HomePage.tsx`

- [ ] **Step 1: 替换页面内联输入 UI**

删掉 `HomePage` 中重复的年月日时 UI 渲染，改为：

```tsx
<BirthProfileForm
  value={birthProfile}
  onChange={setBirthProfile}
/>
```

- [ ] **Step 2: 保留单人页专属字段**

`province` 与 `is_early_zishi` 继续由 `HomePage` 自己管理，只保留在单人页：

```tsx
{birthProfile.hour === 0 && (
  <label className="checkbox-label">
    <input type="checkbox" checked={form.is_early_zishi} onChange={...} />
    <span>早子时（23:00 前，日柱按前一天算）</span>
  </label>
)}
<select id="birth-province" className="form-select" value={form.province} onChange={...} />
```

- [ ] **Step 3: 校正提交映射**

提交时把共享表单值映射回现有 API 入参：

```ts
calendar_type: birthProfile.calendarType,
is_leap_month: birthProfile.isLeapMonth,
```

### Task 3: 接入合盘页

**Files:**
- Modify: `frontend/src/pages/CompatibilityPage.tsx`

- [ ] **Step 1: 将合盘页状态切到共享结构**

分别维护：

```ts
const [selfProfile, setSelfProfile] = useState<BirthProfileFormValue>(initialBirthProfile('male'))
const [partnerProfile, setPartnerProfile] = useState<BirthProfileFormValue>(initialBirthProfile('female'))
```

- [ ] **Step 2: 使用共享输入组件渲染双方表单**

```tsx
<BirthProfileForm title="我的生辰" value={selfProfile} onChange={setSelfProfile} />
<BirthProfileForm title="对方的生辰" value={partnerProfile} onChange={setPartnerProfile} />
```

- [ ] **Step 3: 映射回合盘 API 结构**

提交前转换字段名：

```ts
{
  year: value.year,
  month: value.month,
  day: value.day,
  hour: value.hour,
  gender: value.gender,
  calendar_type: value.calendarType,
  is_leap_month: value.isLeapMonth,
}
```

### Task 4: 验证与收尾

**Files:**
- Modify: `docs/superpowers/specs/2026-05-10-compatibility-input-alignment-design.md` (如需补充验收结果)

- [ ] **Step 1: 运行前端构建**

Run: `cd /Users/liujiming/web/yuanju/frontend && npm run build`  
Expected: PASS，允许保留大 chunk warning。

- [ ] **Step 2: 手工验证单人测算页**

检查：

```text
公历/农历切换正常
月份/闰月联动正常
日期不会越界
十二时辰可选
原有提交成功
```

- [ ] **Step 3: 手工验证合盘页**

检查：

```text
双方都使用相同输入体验
双方都可切换公历/农历
可正常提交到 /compatibility/:id
```
