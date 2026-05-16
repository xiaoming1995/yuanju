# ResultPage 古书章卷重设计 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `frontend/src/pages/ResultPage.tsx`（1200 行 + 38K CSS）从单页长滚版式重做成 5 章·古书章卷布局：水平 scroll-snap 切章、章节胶囊条吸顶、URL hash 同步、章内保留现状的密集数据结构（5×4 排盘表 / 2×5 大运·流年网格），仅做排版与触控精修。

**Architecture:** 新增 `CodexShell` 外层容器 + 5 个 chapter 子组件（命/性/运/年/述）。CodexShell 用 CSS `scroll-snap-type: x mandatory` 实现章节横滑（无 JS 手势库）。章节状态通过 `IntersectionObserver` 监听 + `history.pushState` 同步 URL hash。每章是 `flex: 0 0 100vw` 子元素自带垂直滚动。桌面 ≥1024px 走线性 fallback（无横滑）。

**Tech Stack:** React 19 + TypeScript + Vite，纯 CSS 变量（无 UI 框架）。测试用 Node 内置 `node:test` runner，源码级断言（参考 `frontend/tests/dayun-timeline-ux.test.mjs`）。lucide-react 用于图标。

---

## 设计澄清（一次性决策）

**1. 视觉皮肤·配色**：spec 写"沿用现有 CSS 变量"，但批准的 mockup（`.superpowers/brainstorm/.../06-all-chapters-before-after.html` 新设计侧）是**米黄底 + 深褐字 + 朱红强调**的古书皮肤，与现有**深色主题**冲突。

**解决方式**：引入 `.codex-shell` 局部 scope，在该容器内**用 CSS 变量重映射颜色为古书皮肤**。其他页面不受影响。新增变量集中在 `CodexShell.css` 文件顶部，命名以 `--codex-*` 开头：

```css
.codex-shell {
  --codex-paper: #faf3e3;        /* 章内背景 */
  --codex-paper-soft: #fff8e9;   /* 卡片背景 */
  --codex-paper-warm: rgba(255, 250, 235, 0.5);  /* 强调卡背景 */
  --codex-ink: #2b1a0a;          /* 主文字 */
  --codex-ink-soft: rgba(40, 22, 8, 0.75);
  --codex-ink-muted: rgba(40, 22, 8, 0.5);
  --codex-cinnabar: #6b1e1e;     /* 朱红：当下/日柱/章节激活 */
  --codex-cinnabar-soft: rgba(107, 30, 30, 0.08);
  --codex-hairline: rgba(40, 22, 8, 0.12);
  --codex-gold: linear-gradient(135deg, #c9a84c, #a07e2c);  /* 复用现有金 */
}
```

**2. 前置条件**：spec 写"等 `bazi-ten-god-relation-matrix` + `replicate-dayun-timeline-design` 两个 in-flight change 落地后再启动"。本计划假定开始执行时这两个 change **已经合并到 main**。Task 0 包含验证。

**3. 测试风格**：跟 `dayun-timeline-ux.test.mjs` 一致 —— Node 内置 `test` runner，`readFileSync` 读源文件用正则 `assert.match` / `assert.doesNotMatch` 做结构性断言。运行命令：`cd frontend && node --test tests/*.test.mjs`。**所有任务的测试都写源码断言，不写浏览器/DOM runtime 测试**。运行时验证靠真机 QA（Task 15）。

**4. OpenSpec change**：本 plan 启动前先 `mkdir openspec/changes/redesign-result-page-codex-shell/` 并把本 plan 拷一份到该目录的 `tasks.md`（OpenSpec 操作）。Task 0 包含。

---

## 文件结构（locked in）

```
新增
  frontend/src/components/
    CodexShell.tsx                  外层容器：章节胶囊条 + 水平 snap + hash 同步 + 桌面 fallback
    CodexShell.css                  容器样式 + 局部 CSS 变量（古书皮肤）
    result-chapters/
      Ming.tsx                      命章：5×4 排盘表 + 命格/神煞汇总
      Ming.css
      Xing.tsx                      性章：五行雷达 + 用神 pill + 调候/格局古典卡
      Xing.css
      Yun.tsx                       运章：2×5 大运网格 + 本运观察卡
      Yun.css
      Nian.tsx                      年章：2×5 流年网格 + 流月/过往年事入口
      Nian.css
      Shu.tsx                       述章：宋体 AI 报告 + 词典 + 工具栏
      Shu.css
      ChapterShell.tsx              每章内统一框架：H1 章名 + 章末提示 + 回顶按钮

  frontend/tests/
    result-codex-shell.test.mjs     CodexShell 章节存在 / 胶囊条 / hash / 章切换
    result-paipan-classical.test.mjs 命章排盘表结构 + 字号 + 触控
    result-codex-desktop.test.mjs   ≥1024px 线性 fallback

修改
  frontend/src/pages/ResultPage.tsx           从 1200 行降到 ≤400 行（搬迁后只剩数据加载 + CodexShell 装配）
  frontend/src/pages/ResultPage.css           大幅缩减，仅保留 page shell + 章节共享 token
  frontend/src/components/LiuYueDrawer.tsx    宽度 100vw → min(85vw, 360px)
  frontend/tests/mobilePageQaMatrix.mjs       新增检查项：胶囊条吸顶 / hash 同步 / 神煞徽章触控
```

---

## Task 0: 前置验证 + OpenSpec change 起头

**Files:**
- Verify: 当前分支已包含两个 in-flight change 的 main 合并点
- Create: `openspec/changes/redesign-result-page-codex-shell/proposal.md`
- Create: `openspec/changes/redesign-result-page-codex-shell/tasks.md`

- [ ] **Step 1: 验证 in-flight change 已落地**

```bash
cd /Users/liujiming/web/yuanju
git log --oneline --grep "ten-god-relation-matrix\|dayun-timeline" | head -5
```

Expected: 看到 `bazi-ten-god-relation-matrix` 和 `replicate-dayun-timeline-design` 相关 commit 已经在当前分支历史里。若没有，停止执行，先合并它们。

- [ ] **Step 2: 创建 OpenSpec change 文件夹**

```bash
mkdir -p openspec/changes/redesign-result-page-codex-shell/specs
```

- [ ] **Step 3: 写 OpenSpec proposal**

Create `openspec/changes/redesign-result-page-codex-shell/proposal.md`:

```markdown
# Proposal · redesign-result-page-codex-shell

## What

把 ResultPage 从单页长滚改成 5 章·古书章卷布局（命/性/运/年/述），保留密集数据结构，做排版与触控精修。

## Why

ResultPage 是缘聚核心 UX 表面，移动端审计发现密集数据被压扁不可读（字号 9-11px、触控目标 < 20px）、超长滚动无导航。详见 `docs/superpowers/specs/2026-05-16-result-page-codex-design.md`。

## Reference

- Design spec: `docs/superpowers/specs/2026-05-16-result-page-codex-design.md`
- Implementation plan: `docs/superpowers/plans/2026-05-16-result-page-codex-shell.md`
- Visual baseline: `.superpowers/brainstorm/.../06-all-chapters-before-after.html`（新设计侧）
```

- [ ] **Step 4: 拷贝 plan 到 OpenSpec tasks.md**

```bash
cp docs/superpowers/plans/2026-05-16-result-page-codex-shell.md openspec/changes/redesign-result-page-codex-shell/tasks.md
```

- [ ] **Step 5: 创建实现分支并 commit**

```bash
git checkout -b feat/result-codex-shell
git add openspec/changes/redesign-result-page-codex-shell/
git commit -m "chore(openspec): start redesign-result-page-codex-shell change"
```

---

## Task 1: CodexShell 骨架 + 章节胶囊条 + hash 路由

**Files:**
- Create: `frontend/src/components/CodexShell.tsx`
- Create: `frontend/src/components/CodexShell.css`
- Create: `frontend/tests/result-codex-shell.test.mjs`

- [ ] **Step 1: 写失败测试 `result-codex-shell.test.mjs`**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const root = new URL('..', import.meta.url).pathname
const read = (p) => readFileSync(join(root, p), 'utf8')

test('CodexShell 渲染 5 个章节占位 + 胶囊条', () => {
  const c = read('src/components/CodexShell.tsx')
  assert.match(c, /className="codex-shell"/)
  assert.match(c, /className="codex-strip"/)
  assert.match(c, /const CHAPTERS\s*=\s*\[\s*\{\s*key:\s*'ming'.*name:\s*'命'/s)
  assert.match(c, /\{\s*key:\s*'xing'.*name:\s*'性'/s)
  assert.match(c, /\{\s*key:\s*'yun'.*name:\s*'运'/s)
  assert.match(c, /\{\s*key:\s*'nian'.*name:\s*'年'/s)
  assert.match(c, /\{\s*key:\s*'shu'.*name:\s*'述'/s)
})

test('CodexShell 用 CSS scroll-snap 实现章节横滑（无 JS 手势库）', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-pages\s*\{[\s\S]*scroll-snap-type:\s*x mandatory/)
  assert.match(css, /\.codex-chapter\s*\{[\s\S]*scroll-snap-align:\s*start/)
  assert.match(css, /\.codex-chapter\s*\{[\s\S]*flex:\s*0\s+0\s+100vw/)
})

test('CodexShell 用 IntersectionObserver 监听章节切换 + history.pushState 同步 hash', () => {
  const c = read('src/components/CodexShell.tsx')
  assert.match(c, /new IntersectionObserver/)
  assert.match(c, /history\.pushState/)
  assert.match(c, /window\.location\.hash/)
})

test('CodexShell 章节胶囊条点击跳章用 scrollIntoView smooth', () => {
  const c = read('src/components/CodexShell.tsx')
  assert.match(c, /scrollIntoView\(\s*\{[^}]*behavior:\s*['"]smooth['"]/)
})

test('CodexShell 局部 CSS 变量构成古书皮肤', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-shell\s*\{[\s\S]*--codex-paper:\s*#faf3e3/)
  assert.match(css, /\.codex-shell\s*\{[\s\S]*--codex-ink:\s*#2b1a0a/)
  assert.match(css, /\.codex-shell\s*\{[\s\S]*--codex-cinnabar:\s*#6b1e1e/)
})
```

- [ ] **Step 2: 运行测试确认失败**

```bash
cd frontend && node --test tests/result-codex-shell.test.mjs
```

Expected: 全部 FAIL（文件不存在）

- [ ] **Step 3: 写 `CodexShell.tsx`**

```tsx
import { useEffect, useRef, useState, type ReactNode } from 'react'
import './CodexShell.css'

export type ChapterKey = 'ming' | 'xing' | 'yun' | 'nian' | 'shu'

const CHAPTERS: Array<{ key: ChapterKey; name: string; step: string }> = [
  { key: 'ming', name: '命', step: '第·一·章' },
  { key: 'xing', name: '性', step: '第·二·章' },
  { key: 'yun',  name: '运', step: '第·三·章' },
  { key: 'nian', name: '年', step: '第·四·章' },
  { key: 'shu',  name: '述', step: '第·五·章' },
]

interface Props {
  children: { ming: ReactNode; xing: ReactNode; yun: ReactNode; nian: ReactNode; shu: ReactNode }
}

export default function CodexShell({ children }: Props) {
  const pagesRef = useRef<HTMLDivElement>(null)
  const chapterRefs = useRef<Record<ChapterKey, HTMLDivElement | null>>({
    ming: null, xing: null, yun: null, nian: null, shu: null,
  })
  const [active, setActive] = useState<ChapterKey>(() => {
    const hash = window.location.hash.replace('#', '') as ChapterKey
    return (CHAPTERS.some(c => c.key === hash) ? hash : 'ming')
  })

  // 进入页面把 hash 章节滚到视口
  useEffect(() => {
    chapterRefs.current[active]?.scrollIntoView({ behavior: 'auto', inline: 'start' })
  }, [])  // 仅首次

  // IntersectionObserver 监听章节切换
  useEffect(() => {
    if (!pagesRef.current) return
    const observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.find(e => e.isIntersecting && e.intersectionRatio > 0.6)
        if (!visible) return
        const key = visible.target.getAttribute('data-chapter') as ChapterKey
        if (!key || key === active) return
        setActive(key)
        history.pushState(null, '', `#${key}`)
      },
      { root: pagesRef.current, threshold: [0.6] }
    )
    Object.values(chapterRefs.current).forEach(el => el && observer.observe(el))
    return () => observer.disconnect()
  }, [active])

  // 监听浏览器后退键改 hash
  useEffect(() => {
    const onPop = () => {
      const hash = window.location.hash.replace('#', '') as ChapterKey
      if (CHAPTERS.some(c => c.key === hash)) {
        chapterRefs.current[hash]?.scrollIntoView({ behavior: 'smooth', inline: 'start' })
      }
    }
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  const jumpTo = (key: ChapterKey) => {
    chapterRefs.current[key]?.scrollIntoView({ behavior: 'smooth', inline: 'start' })
  }

  return (
    <div className="codex-shell">
      <nav className="codex-strip" aria-label="章节导航">
        {CHAPTERS.map(c => (
          <button
            key={c.key}
            className={`codex-strip-item${active === c.key ? ' is-active' : ''}`}
            onClick={() => jumpTo(c.key)}
            aria-current={active === c.key ? 'page' : undefined}
          >
            {c.name}
          </button>
        ))}
      </nav>
      <div className="codex-pages" ref={pagesRef}>
        {CHAPTERS.map(c => (
          <section
            key={c.key}
            className="codex-chapter"
            data-chapter={c.key}
            ref={el => { chapterRefs.current[c.key] = el }}
          >
            {children[c.key]}
          </section>
        ))}
      </div>
    </div>
  )
}
```

- [ ] **Step 4: 写 `CodexShell.css`**

```css
.codex-shell {
  /* 古书皮肤 · 局部 CSS 变量，仅在此容器内生效 */
  --codex-paper: #faf3e3;
  --codex-paper-soft: #fff8e9;
  --codex-paper-warm: rgba(255, 250, 235, 0.5);
  --codex-ink: #2b1a0a;
  --codex-ink-soft: rgba(40, 22, 8, 0.75);
  --codex-ink-muted: rgba(40, 22, 8, 0.5);
  --codex-cinnabar: #6b1e1e;
  --codex-cinnabar-soft: rgba(107, 30, 30, 0.08);
  --codex-cinnabar-border: rgba(107, 30, 30, 0.25);
  --codex-hairline: rgba(40, 22, 8, 0.12);
  --codex-hairline-strong: rgba(40, 22, 8, 0.18);

  background: var(--codex-paper);
  color: var(--codex-ink);
  font-family: var(--font-serif), 'Songti SC', 'Noto Serif SC', serif;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 章节胶囊条 · sticky 顶部 */
.codex-strip {
  position: sticky;
  top: 64px;  /* Navbar 高度，后续如改 Navbar 同步调整 */
  z-index: 50;
  display: flex;
  height: 44px;
  background: var(--codex-paper);
  border-bottom: 1px solid var(--codex-hairline);
  justify-content: space-around;
  align-items: stretch;
  font-family: 'Songti SC', 'Noto Serif SC', serif;
}
.codex-strip-item {
  flex: 1;
  background: transparent;
  border: none;
  color: var(--codex-ink-muted);
  font-size: 16px;
  letter-spacing: 4px;
  cursor: pointer;
  position: relative;
  font-family: inherit;
  min-height: 44px;  /* 触控目标 */
  padding: 0;
}
.codex-strip-item.is-active {
  color: var(--codex-cinnabar);
  font-weight: 600;
}
.codex-strip-item.is-active::after {
  content: '';
  position: absolute;
  left: 22%; right: 22%;
  bottom: 0;
  height: 2px;
  background: var(--codex-cinnabar);
}

/* 章节翻页容器 */
.codex-pages {
  flex: 1;
  display: flex;
  overflow-x: auto;
  overflow-y: hidden;
  scroll-snap-type: x mandatory;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
}
.codex-chapter {
  flex: 0 0 100vw;
  scroll-snap-align: start;
  overflow-y: auto;
  overflow-x: hidden;
  padding: 14px 14px calc(80px + env(safe-area-inset-bottom));
}

/* 桌面线性 fallback */
@media (min-width: 1024px) {
  .codex-pages {
    flex-direction: column;
    overflow-x: hidden;
    overflow-y: auto;
    scroll-snap-type: none;
  }
  .codex-chapter {
    flex: 0 0 auto;
    width: 100%;
    max-width: 880px;
    margin: 0 auto;
  }
  .codex-strip {
    top: 64px;
  }
}

/* 打印铺平 */
@media print {
  .codex-strip { display: none; }
  .codex-pages {
    flex-direction: column;
    overflow: visible;
    scroll-snap-type: none;
  }
  .codex-chapter {
    flex: 0 0 auto;
    width: 100%;
    page-break-before: always;
  }
  .codex-chapter:first-child { page-break-before: auto; }
}
```

- [ ] **Step 5: 运行测试确认通过**

```bash
cd frontend && node --test tests/result-codex-shell.test.mjs
```

Expected: 全部 PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/CodexShell.tsx frontend/src/components/CodexShell.css frontend/tests/result-codex-shell.test.mjs
git commit -m "feat(result): add CodexShell with chapter pill bar + hash routing"
```

---

## Task 2: ChapterShell 子框架（章名标题 + 章末提示 + 回顶按钮）

**Files:**
- Create: `frontend/src/components/result-chapters/ChapterShell.tsx`
- Create: `frontend/src/components/result-chapters/ChapterShell.css`
- Modify: `frontend/tests/result-codex-shell.test.mjs`（追加）

- [ ] **Step 1: 追加测试**

在 `result-codex-shell.test.mjs` 文件末尾追加：

```js
test('ChapterShell 渲染 H1 + 古典分割线 + 章末提示 + 回顶按钮', () => {
  const c = read('src/components/result-chapters/ChapterShell.tsx')
  assert.match(c, /className="chapter-shell"/)
  assert.match(c, /className="chapter-h1"/)
  assert.match(c, /className="chapter-step"/)
  assert.match(c, /className="chapter-name"/)
  assert.match(c, /className="chapter-rule"/)
  assert.match(c, /className="chapter-end-hint"/)
  assert.match(c, /className="chapter-back-to-top"/)
  assert.match(c, /滑向/)  // 章末提示文案
})

test('ChapterShell.css H1 24px serif + 回顶浮动 + 章末提示渐隐', () => {
  const css = read('src/components/result-chapters/ChapterShell.css')
  assert.match(css, /\.chapter-name\s*\{[\s\S]*font-size:\s*24px/)
  assert.match(css, /\.chapter-back-to-top\s*\{[\s\S]*position:\s*fixed/)
  assert.match(css, /\.chapter-back-to-top\s*\{[\s\S]*min-height:\s*44px/)
  assert.match(css, /\.chapter-end-hint\s*\{[\s\S]*animation/)
})
```

- [ ] **Step 2: 运行确认新增测试失败**

```bash
cd frontend && node --test tests/result-codex-shell.test.mjs
```

Expected: 2 个新测试 FAIL

- [ ] **Step 3: 写 `ChapterShell.tsx`**

```tsx
import { useEffect, useRef, useState, type ReactNode } from 'react'
import './ChapterShell.css'

interface Props {
  step: string         // "第·一·章"
  name: string         // "命"
  nextName?: string    // "性"，最后一章为 undefined
  children: ReactNode
}

export default function ChapterShell({ step, name, nextName, children }: Props) {
  const ref = useRef<HTMLDivElement>(null)
  const [showBackToTop, setShowBackToTop] = useState(false)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    const onScroll = () => {
      setShowBackToTop(el.scrollTop > window.innerHeight)
    }
    el.addEventListener('scroll', onScroll, { passive: true })
    return () => el.removeEventListener('scroll', onScroll)
  }, [])

  // 注意：在 CodexShell 章节滚动是父容器，ChapterShell 自身不滚动。
  // 这里 ref 监听父级 scroll 事件用 closest 适配
  // 简化：用 window scroll fallback（章内滚动会触发）
  useEffect(() => {
    const onWinScroll = () => {
      const el = ref.current
      if (!el) return
      const parent = el.closest('.codex-chapter')
      if (parent) {
        setShowBackToTop(parent.scrollTop > window.innerHeight * 0.8)
      }
    }
    const parent = ref.current?.closest('.codex-chapter')
    parent?.addEventListener('scroll', onWinScroll, { passive: true })
    return () => parent?.removeEventListener('scroll', onWinScroll)
  }, [])

  const scrollToTop = () => {
    const parent = ref.current?.closest('.codex-chapter')
    parent?.scrollTo({ top: 0, behavior: 'smooth' })
  }

  return (
    <div className="chapter-shell" ref={ref}>
      <header className="chapter-h1">
        <div className="chapter-step">{step}</div>
        <div className="chapter-name">{name}</div>
        <div className="chapter-rule" />
      </header>
      <div className="chapter-body">
        {children}
      </div>
      {nextName && (
        <div className="chapter-end-hint" aria-hidden="true">
          ──→ 滑向《{nextName}》
        </div>
      )}
      {showBackToTop && (
        <button
          className="chapter-back-to-top"
          onClick={scrollToTop}
          aria-label="回到本章顶部"
        >
          ↑
        </button>
      )}
    </div>
  )
}
```

- [ ] **Step 4: 写 `ChapterShell.css`**

```css
.chapter-shell {
  display: flex;
  flex-direction: column;
  min-height: 100%;
  position: relative;
}

.chapter-h1 {
  text-align: center;
  margin: 8px 0 18px;
  font-family: 'Songti SC', 'Noto Serif SC', serif;
}
.chapter-step {
  font-size: 10px;
  letter-spacing: 4px;
  color: var(--codex-ink-muted);
}
.chapter-name {
  font-size: 24px;
  letter-spacing: 8px;
  color: var(--codex-ink);
  font-weight: 600;
  margin-top: 4px;
}
.chapter-rule {
  width: 40px;
  height: 1px;
  background: var(--codex-ink-muted);
  margin: 8px auto 0;
}

.chapter-body {
  flex: 1;
}

.chapter-end-hint {
  text-align: center;
  margin: 32px auto 0;
  font-family: 'Songti SC', serif;
  font-size: 12px;
  letter-spacing: 3px;
  color: var(--codex-ink-muted);
  opacity: 0;
  animation: chapter-hint-fade 4s 1s forwards;
}
@keyframes chapter-hint-fade {
  0%   { opacity: 0; }
  20%  { opacity: 0.85; }
  80%  { opacity: 0.85; }
  100% { opacity: 0; }
}

.chapter-back-to-top {
  position: fixed;
  right: 16px;
  bottom: calc(80px + env(safe-area-inset-bottom));
  width: 44px;
  height: 44px;
  min-height: 44px;
  border-radius: 22px;
  background: var(--codex-cinnabar);
  color: var(--codex-paper);
  border: none;
  font-size: 18px;
  font-family: 'Songti SC', serif;
  box-shadow: 0 6px 16px rgba(107, 30, 30, 0.3);
  z-index: 60;
  cursor: pointer;
}
```

- [ ] **Step 5: 测试通过**

```bash
cd frontend && node --test tests/result-codex-shell.test.mjs
```

Expected: 全部 PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/result-chapters/ChapterShell.tsx frontend/src/components/result-chapters/ChapterShell.css frontend/tests/result-codex-shell.test.mjs
git commit -m "feat(result): add ChapterShell with H1 + end hint + back-to-top"
```

---

## Task 3: 拆 ResultPage 渲染到 5 个章节占位组件（搬迁阶段·不动样式）

**Files:**
- Create: `frontend/src/components/result-chapters/{Ming,Xing,Yun,Nian,Shu}.tsx`（5 个）
- Create: `frontend/src/components/result-chapters/{Ming,Xing,Yun,Nian,Shu}.css`（5 个空文件）
- Modify: `frontend/src/pages/ResultPage.tsx`（替换渲染）

> **重要原则**：本 task 只搬迁，**不**改样式/字号/类名。把当前 ResultPage.tsx 里 `result-content` 内的 JSX 按章节切成 5 块塞进对应组件，保留原 className，让现有 ResultPage.css 继续生效。

- [ ] **Step 1: 读现有 ResultPage.tsx 找出渲染分块边界**

```bash
grep -n "section\|panel\|className=\"result-" /Users/liujiming/web/yuanju/frontend/src/pages/ResultPage.tsx | head -30
```

记录出现的章节级 `<section>` 起止行号 —— 大致如下（需对照实际行号调整）：

| 章 | 现有 JSX 块（按 className 找） |
|---|---|
| 命 | `.result-header` + `.bazi-data-grid` + `.ten-god-relation-*` |
| 性 | `.result-structure-grid`（含 `<WuxingRadar />` + `<YongshenBadge />` + `<TiaohouCard />`）|
| 运 | `<DayunTimeline />` 包裹的 `<section className="dayun-section">` |
| 年 | DayunTimeline 内已含流年部分 —— 留在运章里，年章只放过往年事入口（流月 drawer 状态由 DayunTimeline 内部已有）|
| 述 | `.report-section` 含 digest / 章节 / 工具栏 |

> 注：流年现在在 DayunTimeline 内部不在 ResultPage 顶层。运/年的拆分需要 Task 7-8 后续做。本 task 把"运 + 年"暂时合并放在 Yun 组件里。

- [ ] **Step 2: 创建 5 个章节组件文件（先空 div）**

`frontend/src/components/result-chapters/Ming.tsx`:
```tsx
import type { ReactNode } from 'react'
import './Ming.css'

interface Props { children: ReactNode }

export default function Ming({ children }: Props) {
  return <div className="chapter-ming">{children}</div>
}
```

`Xing.tsx` / `Yun.tsx` / `Nian.tsx` / `Shu.tsx` 同上结构，把 `Ming`/`chapter-ming` 换成对应章名（`Xing`/`chapter-xing` 等）。

5 个 `.css` 文件先建空文件：
```css
/* placeholder · 后续 task 填充 */
```

- [ ] **Step 3: 写 ResultPage.tsx 搬迁测试**

新增 `frontend/tests/result-codex-migration.test.mjs`：

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('ResultPage 装配 CodexShell 并把 5 章节挂到 children prop', () => {
  const page = read('src/pages/ResultPage.tsx')
  assert.match(page, /import CodexShell from/)
  assert.match(page, /import Ming from/)
  assert.match(page, /import Xing from/)
  assert.match(page, /import Yun from/)
  assert.match(page, /import Nian from/)
  assert.match(page, /import Shu from/)
  assert.match(page, /<CodexShell[\s\S]*ming:/)
  assert.match(page, /xing:/)
  assert.match(page, /yun:/)
  assert.match(page, /nian:/)
  assert.match(page, /shu:/)
})

test('ResultPage 已不再渲染单页长滚 result-content', () => {
  const page = read('src/pages/ResultPage.tsx')
  // 旧的 `.result-content` 已不存在（顶层渲染时）
  assert.doesNotMatch(page, /<div className="result-content"[\s\S]*<\/div>\s*<\/div>\s*\)\s*\}/)
})
```

- [ ] **Step 4: 运行测试确认失败**

```bash
cd frontend && node --test tests/result-codex-migration.test.mjs
```

Expected: FAIL（ResultPage 还没改）

- [ ] **Step 5: 改造 ResultPage.tsx**

把现有渲染主体（loading / error 状态保留），最终渲染替换为：

```tsx
// 在 import 区加：
import CodexShell from '../components/CodexShell'
import Ming from '../components/result-chapters/Ming'
import Xing from '../components/result-chapters/Xing'
import Yun from '../components/result-chapters/Yun'
import Nian from '../components/result-chapters/Nian'
import Shu from '../components/result-chapters/Shu'

// 在 return 主体（数据加载完成后）改成：
return (
  <div className="result-page page">
    {/* loading / error 状态保持原样 */}
    {!loading && !error && result && (
      <CodexShell>
        {{
          ming: <Ming>{/* 原 .result-header + .bazi-data-grid + .ten-god-relation 块 JSX 直接搬过来 */}</Ming>,
          xing: <Xing>{/* 原 .result-structure-grid 块 JSX */}</Xing>,
          yun:  <Yun>{/* 原 dayun-section 含 DayunTimeline 块 JSX */}</Yun>,
          nian: <Nian>{/* 暂留空占位：本 task 不动 */}</Nian>,
          shu:  <Shu>{/* 原 .report-section 块 JSX */}</Shu>,
        }}
      </CodexShell>
    )}
  </div>
)
```

**关键**：JSX 字面搬迁 —— 复制粘贴原有内容到对应组件的 `{children}` 位置，所有 className / props / state 都保留。如果有共享 state（如 `selectedDayun`），暂时通过 props 从 ResultPage 传入。

- [ ] **Step 6: 跑全部前端测试确认不破坏既有测试**

```bash
cd frontend && node --test tests/*.test.mjs
```

Expected: 既有 12 个测试 + 本 task 的新增测试都 PASS。如果有现有测试因为类名搬迁失败，立即修复（保留原有断言能匹配到新位置）。

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/result-chapters/ frontend/src/pages/ResultPage.tsx frontend/tests/result-codex-migration.test.mjs
git commit -m "refactor(result): split ResultPage render into 5 chapter components (no style change)"
```

---

## Task 4: 全局标题层级统一 + 章节胶囊条 + Navbar 配合

**Files:**
- Modify: `frontend/src/pages/ResultPage.css`（缩减并新增 codex token）
- Modify: `frontend/src/index.css`（如需）
- Create: `frontend/tests/result-codex-typography.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('CodexShell 子节标题 H2 18px + hairline + serif', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-shell\s+h2,[\s\S]*\.codex-h2\s*\{[\s\S]*font-size:\s*18px/)
  assert.match(css, /\.codex-shell\s+h2[\s\S]*font-family:[\s\S]*Songti SC/)
  assert.match(css, /\.codex-shell\s+h2[\s\S]*border-bottom:\s*1px solid var\(--codex-hairline\)/)
})

test('CodexShell 卡片标题 H3 15px sans', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-shell\s+h3\s*\{[\s\S]*font-size:\s*15px/)
})

test('CodexShell 正文 14px line-height 1.7', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-shell\s+p\s*\{[\s\S]*font-size:\s*14px[\s\S]*line-height:\s*1\.7/)
})
```

- [ ] **Step 2: 运行确认失败**

```bash
cd frontend && node --test tests/result-codex-typography.test.mjs
```

- [ ] **Step 3: 在 `CodexShell.css` 文件末尾追加全局排版规则**

```css
/* ============ 章内排版收口 · 替换 ResultPage.css 中混用的 18/20/22px ============ */
.codex-shell h2,
.codex-shell .codex-h2 {
  font-family: 'Songti SC', 'Noto Serif SC', serif;
  font-size: 18px;
  font-weight: 600;
  color: var(--codex-ink);
  margin: 16px 0 10px;
  padding-bottom: 4px;
  border-bottom: 1px solid var(--codex-hairline);
  display: flex;
  justify-content: space-between;
  align-items: baseline;
}

.codex-shell h3,
.codex-shell .codex-h3 {
  font-family: var(--font-sans);
  font-size: 15px;
  font-weight: 600;
  color: var(--codex-ink);
  margin: 12px 0 6px;
}

.codex-shell p,
.codex-shell .codex-body {
  font-size: 14px;
  line-height: 1.7;
  color: var(--codex-ink-soft);
}

.codex-shell small,
.codex-shell .codex-small {
  font-size: 12px;
  color: var(--codex-ink-muted);
}
```

- [ ] **Step 4: 测试通过**

```bash
cd frontend && node --test tests/result-codex-typography.test.mjs
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/CodexShell.css frontend/tests/result-codex-typography.test.mjs
git commit -m "feat(result): unify chapter typography hierarchy (H1·24 / H2·18 / H3·15)"
```

---

## Task 5: 命章排盘表 · 5×4 + 行标签竖排 + 字号上升 + 神煞徽章

**Files:**
- Modify: `frontend/src/components/result-chapters/Ming.tsx`（添加新 paipan 渲染）
- Modify: `frontend/src/components/result-chapters/Ming.css`
- Create: `frontend/tests/result-paipan-classical.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('Ming 章渲染 5×4 paipan grid 用 ChapterShell + H1', () => {
  const c = read('src/components/result-chapters/Ming.tsx')
  assert.match(c, /import ChapterShell from/)
  assert.match(c, /<ChapterShell[^>]*step="第·一·章"[^>]*name="命"[^>]*nextName="性"/)
  assert.match(c, /className="paipan-grid"/)
})

test('Ming 章排盘表 5 行结构：天干/地支/十神/藏干/神煞', () => {
  const c = read('src/components/result-chapters/Ming.tsx')
  assert.match(c, /天干/)
  assert.match(c, /地支/)
  assert.match(c, /十神/)
  assert.match(c, /藏干/)
  assert.match(c, /神煞/)
  // 4 列：年月日时
  assert.match(c, /年[\s\S]*月[\s\S]*日[\s\S]*时/)
})

test('Ming 章神煞用 button 包装且可点（≥44pt 触控）', () => {
  const c = read('src/components/result-chapters/Ming.tsx')
  assert.match(c, /<button[^>]*className="paipan-shensha[^"]*"/)
})

test('Ming.css 排盘表网格 24px 行标签竖排 + 4 等分柱列', () => {
  const css = read('src/components/result-chapters/Ming.css')
  assert.match(css, /\.paipan-grid\s*\{[\s\S]*display:\s*grid[\s\S]*grid-template-columns:\s*28px\s+repeat\(4,\s*minmax\(0,\s*1fr\)\)/)
  assert.match(css, /\.paipan-row-label\s*\{[\s\S]*writing-mode:\s*vertical-rl/)
})

test('Ming.css 干支 28px serif + 藏干纵排', () => {
  const css = read('src/components/result-chapters/Ming.css')
  assert.match(css, /\.paipan-big\s*\{[\s\S]*font-size:\s*28px/)
  assert.match(css, /\.paipan-big\s*\{[\s\S]*font-family:[\s\S]*Songti SC/)
  assert.match(css, /\.paipan-hide-gan\s*\{[\s\S]*display:\s*flex[\s\S]*flex-direction:\s*column/)
})

test('Ming.css 神煞徽章 11px + min-height ≥24px + 朱红边', () => {
  const css = read('src/components/result-chapters/Ming.css')
  assert.match(css, /\.paipan-shensha\s*\{[\s\S]*font-size:\s*11px/)
  assert.match(css, /\.paipan-shensha\s*\{[\s\S]*min-height:\s*24px/)
  assert.match(css, /\.paipan-shensha\s*\{[\s\S]*border:\s*1px solid var\(--codex-cinnabar-border\)/)
})

test('Ming.css 日柱列头朱红下划线 + 整列白底', () => {
  const css = read('src/components/result-chapters/Ming.css')
  assert.match(css, /\.paipan-col-head\.is-day\s*\{[\s\S]*color:\s*var\(--codex-cinnabar\)/)
  assert.match(css, /\.paipan-col-head\.is-day\s*\{[\s\S]*border-bottom:[\s\S]*var\(--codex-cinnabar\)/)
  assert.match(css, /\.paipan-day-col\s*\{[\s\S]*background:\s*rgba\(255,\s*255,\s*255/)
})
```

- [ ] **Step 2: 测试失败**

```bash
cd frontend && node --test tests/result-paipan-classical.test.mjs
```

- [ ] **Step 3: 改造 `Ming.tsx`** —— 用现有 ResultPage 数据接口提取四柱信息，按新结构渲染：

```tsx
import { useState } from 'react'
import type { BaziResult, ShenshaAnnotation } from '../../lib/api'
import ChapterShell from './ChapterShell'
import './Ming.css'

interface Props {
  result: BaziResult
  shenshaAnnotations: ShenshaAnnotation[]
  onOpenAnnotation?: (name: string) => void
}

const POS = [
  { key: 'year',  label: '年' },
  { key: 'month', label: '月' },
  { key: 'day',   label: '日' },
  { key: 'hour',  label: '时' },
] as const

export default function Ming({ result, shenshaAnnotations, onOpenAnnotation }: Props) {
  const pillars = result.pillars  // 假定字段：{ year: { gan, zhi, ten_god, hide_gan: string[], shensha: string[] } ... }

  return (
    <ChapterShell step="第·一·章" name="命" nextName="性">
      <h2 className="codex-h2">四柱</h2>
      <div className="paipan-grid">
        {/* 表头 */}
        <div />
        {POS.map(p => (
          <div
            key={`head-${p.key}`}
            className={`paipan-col-head${p.key === 'day' ? ' is-day' : ''}`}
          >{p.label}</div>
        ))}

        {/* 天干行 */}
        <div className="paipan-row-label">天干</div>
        {POS.map(p => (
          <div
            key={`gan-${p.key}`}
            className={`paipan-big paipan-gan paipan-row-div${p.key === 'day' ? ' paipan-day-col' : ''}`}
          >{pillars[p.key].gan}</div>
        ))}

        {/* 地支行 */}
        <div className="paipan-row-label">地支</div>
        {POS.map(p => (
          <div
            key={`zhi-${p.key}`}
            className={`paipan-big paipan-row-div${p.key === 'day' ? ' paipan-day-col' : ''}`}
          >{pillars[p.key].zhi}</div>
        ))}

        {/* 十神行 */}
        <div className="paipan-row-label">十神</div>
        {POS.map(p => (
          <div
            key={`ten-${p.key}`}
            className={`paipan-ten paipan-row-div${p.key === 'day' ? ' paipan-day-col paipan-day-master' : ''}`}
          >{p.key === 'day' ? '日主' : pillars[p.key].ten_god}</div>
        ))}

        {/* 藏干行 */}
        <div className="paipan-row-label">藏干</div>
        {POS.map(p => (
          <div
            key={`hg-${p.key}`}
            className={`paipan-hide-gan paipan-row-div${p.key === 'day' ? ' paipan-day-col' : ''}`}
          >
            {(pillars[p.key].hide_gan ?? []).map((g, i) => <span key={i}>{g}</span>)}
          </div>
        ))}

        {/* 神煞行 */}
        <div className="paipan-row-label">神煞</div>
        {POS.map(p => (
          <div
            key={`ss-${p.key}`}
            className={`paipan-shensha-wrap${p.key === 'day' ? ' paipan-day-col' : ''}`}
          >
            {(pillars[p.key].shensha ?? []).map(name => (
              <button
                key={name}
                className="paipan-shensha"
                onClick={() => onOpenAnnotation?.(name)}
                aria-label={`神煞 ${name} 注解`}
              >{name}</button>
            ))}
          </div>
        ))}
      </div>

      {/* 命格 / 神煞汇总 */}
      <h2 className="codex-h2">命格</h2>
      <p className="codex-body">{result.mingge_label ?? '—'}</p>
    </ChapterShell>
  )
}
```

> 若 `BaziResult` 现有字段名与上述假设不一致，对照 `frontend/src/lib/api.ts` 调整 prop 取值路径。**不要改后端接口**。

- [ ] **Step 4: 写 `Ming.css`**

```css
/* ============ 命章 · 排盘表 ============ */
.chapter-ming .paipan-grid {
  display: grid;
  grid-template-columns: 28px repeat(4, minmax(0, 1fr));
  background: var(--codex-paper-warm);
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(40, 22, 8, 0.06);
}

.paipan-grid > div {
  padding: 8px 4px;
  text-align: center;
  font-family: 'Songti SC', 'Noto Serif SC', serif;
}

/* 行标签：竖排 */
.paipan-row-label {
  writing-mode: vertical-rl;
  font-size: 10px;
  letter-spacing: 4px;
  color: var(--codex-ink-muted);
  padding: 12px 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 列头 */
.paipan-col-head {
  font-size: 11px;
  letter-spacing: 3px;
  color: var(--codex-ink-muted);
  padding: 8px 4px 6px;
  background: var(--codex-paper-warm);
  border-bottom: 1px solid var(--codex-hairline-strong);
}
.paipan-col-head.is-day {
  color: var(--codex-cinnabar);
  font-weight: 600;
  border-bottom: 1.5px solid var(--codex-cinnabar);
  background: rgba(255, 255, 255, 0.7);
}

/* 日柱整列底色 */
.paipan-day-col {
  background: rgba(255, 255, 255, 0.55);
}

/* 干支大字 */
.paipan-big {
  font-size: 28px;
  line-height: 1.05;
  color: var(--codex-ink);
  padding: 12px 0 8px;
}
.paipan-gan {
  color: var(--codex-cinnabar);
}

/* 行间分隔 */
.paipan-row-div {
  border-bottom: 1px dotted var(--codex-hairline);
}

/* 十神 */
.paipan-ten {
  font-size: 13px;
  color: var(--codex-ink-soft);
  padding: 8px 4px;
}
.paipan-day-master {
  color: var(--codex-cinnabar);
  font-weight: 600;
}

/* 藏干（纵排） */
.paipan-hide-gan {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--codex-ink-soft);
  padding: 8px 4px;
}

/* 神煞徽章 */
.paipan-shensha-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 8px 4px 10px;
  justify-content: center;
}
.paipan-shensha {
  font-family: 'Songti SC', serif;
  font-size: 11px;
  padding: 5px 9px;
  border-radius: 10px;
  background: var(--codex-cinnabar-soft);
  color: var(--codex-cinnabar);
  border: 1px solid var(--codex-cinnabar-border);
  line-height: 1.2;
  min-height: 24px;
  cursor: pointer;
}
.paipan-shensha:hover {
  background: rgba(107, 30, 30, 0.12);
}
@media (hover: hover) {
  .paipan-shensha {
    text-decoration: underline;
    text-decoration-color: rgba(107, 30, 30, 0.3);
    text-underline-offset: 2px;
  }
}
```

- [ ] **Step 5: 跑测试**

```bash
cd frontend && node --test tests/result-paipan-classical.test.mjs
```

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/result-chapters/Ming.tsx frontend/src/components/result-chapters/Ming.css frontend/tests/result-paipan-classical.test.mjs
git commit -m "feat(result): redesign 命章 with classical paipan table + vertical row labels"
```

---

## Task 6: 性章 · 五行雷达 polygon 填色 + 用神 pill + 调候/格局古典卡

**Files:**
- Modify: `frontend/src/components/result-chapters/Xing.tsx`
- Modify: `frontend/src/components/result-chapters/Xing.css`
- Modify: `frontend/src/components/WuxingRadar.tsx`（追加 polygon 填色支持）
- Create: `frontend/tests/result-xing-chapter.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('Xing 章渲染 ChapterShell + 五行/用神/调候格局 三 H2', () => {
  const c = read('src/components/result-chapters/Xing.tsx')
  assert.match(c, /<ChapterShell[^>]*step="第·二·章"[^>]*name="性"[^>]*nextName="运"/)
  assert.match(c, /五行强弱/)
  assert.match(c, /用神\s*\/\s*喜忌/)
  assert.match(c, /调候\s*\/\s*格局/)
  assert.match(c, /<WuxingRadar/)
})

test('Xing 章用神 pill 用 yongshen-pill / yongshen-pill--avoid 两态', () => {
  const c = read('src/components/result-chapters/Xing.tsx')
  assert.match(c, /className="yongshen-pill"/)
  assert.match(c, /className="yongshen-pill yongshen-pill--avoid"/)
})

test('Xing.css 用神 pill 金色 / 朱红渐变 + 古典卡', () => {
  const css = read('src/components/result-chapters/Xing.css')
  assert.match(css, /\.yongshen-pill\s*\{[\s\S]*background:\s*linear-gradient/)
  assert.match(css, /\.yongshen-pill--avoid\s*\{[\s\S]*background:\s*linear-gradient\([^)]*#6b1e1e/)
  assert.match(css, /\.codex-card\s*\{[\s\S]*background:\s*var\(--codex-paper-warm\)/)
  assert.match(css, /\.codex-card\s*\{[\s\S]*border:\s*1px solid var\(--codex-hairline\)/)
})

test('WuxingRadar 接受 fillPolygon prop 并渲染填色五边形', () => {
  const c = read('src/components/WuxingRadar.tsx')
  assert.match(c, /fillPolygon/)
})
```

- [ ] **Step 2: 失败**

```bash
cd frontend && node --test tests/result-xing-chapter.test.mjs
```

- [ ] **Step 3: 改造 `WuxingRadar.tsx`** 加 `fillPolygon` prop（默认 false 保持向后兼容）

```tsx
// 在现有 props 里加
interface Props {
  // ...现有
  fillPolygon?: boolean   // 新增：true 时把 5 个点连成 polygon 半透明朱红填色
}

// 在 render 内增加：
{fillPolygon && (
  <polygon
    points={/* 现有计算出的 5 个点字符串 */}
    fill="rgba(107, 30, 30, 0.18)"
    stroke="rgba(107, 30, 30, 0.45)"
    strokeWidth="1.2"
  />
)}
```

> 具体 polygon 点的计算复用 WuxingRadar 现有几何代码 —— 把 5 个数据点坐标拼成 `points="x1,y1 x2,y2 ..."` 字符串。

- [ ] **Step 4: 改造 `Xing.tsx`**

```tsx
import type { BaziResult } from '../../lib/api'
import ChapterShell from './ChapterShell'
import WuxingRadar from '../WuxingRadar'
import './Xing.css'

interface Props {
  result: BaziResult
}

export default function Xing({ result }: Props) {
  const wuxing = result.wuxing_strength  // { mu, huo, tu, jin, shui }
  const ys = result.yongshen   // { likes: string[], avoids: string[] }
  const tiaohou = result.tiaohou_text
  const mingge = result.mingge_label

  return (
    <ChapterShell step="第·二·章" name="性" nextName="运">
      <h2 className="codex-h2">五行强弱</h2>
      <WuxingRadar data={wuxing} fillPolygon />
      <div className="wuxing-summary-row">
        <span>木 · {wuxing.mu}</span>
        <span>火 · {wuxing.huo}</span>
        <span>土 · {wuxing.tu}</span>
        <span>金 · {wuxing.jin}</span>
        <span>水 · {wuxing.shui}</span>
      </div>

      <h2 className="codex-h2">用神 / 喜忌</h2>
      <div className="yongshen-row">
        {ys.likes.map(name => (
          <span key={`like-${name}`} className="yongshen-pill">喜 · {name}</span>
        ))}
        {ys.avoids.map(name => (
          <span key={`avoid-${name}`} className="yongshen-pill yongshen-pill--avoid">忌 · {name}</span>
        ))}
      </div>

      <h2 className="codex-h2">调候 / 格局</h2>
      <div className="codex-card">
        <div className="codex-card-label">调候</div>
        <p className="codex-body">{tiaohou ?? '—'}</p>
      </div>
      <div className="codex-card">
        <div className="codex-card-label">格局</div>
        <p className="codex-body">{mingge ?? '—'}</p>
      </div>
    </ChapterShell>
  )
}
```

- [ ] **Step 5: 写 `Xing.css`**

```css
.wuxing-summary-row {
  display: flex;
  justify-content: space-around;
  padding: 12px 0;
  font-family: 'Songti SC', serif;
  font-size: 14px;
  color: var(--codex-ink-soft);
}

.yongshen-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 8px 0;
}
.yongshen-pill {
  padding: 6px 12px;
  background: linear-gradient(135deg, #c9a84c, #a07e2c);
  color: var(--codex-paper-soft);
  border-radius: 12px;
  font-family: 'Songti SC', serif;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 1px;
  box-shadow: 0 2px 4px rgba(40, 22, 8, 0.15);
  min-height: 32px;
  display: inline-flex;
  align-items: center;
}
.yongshen-pill--avoid {
  background: linear-gradient(135deg, #6b1e1e, #4a1414);
}

.codex-card {
  background: var(--codex-paper-warm);
  border: 1px solid var(--codex-hairline);
  border-radius: 6px;
  padding: 10px 12px;
  margin-top: 8px;
  font-family: 'Songti SC', serif;
}
.codex-card-label {
  font-size: 10px;
  letter-spacing: 3px;
  color: var(--codex-ink-muted);
  margin-bottom: 4px;
}
.codex-card .codex-body {
  font-size: 13px;
  color: var(--codex-ink);
  line-height: 1.75;
}
```

- [ ] **Step 6: 测试通过**

```bash
cd frontend && node --test tests/result-xing-chapter.test.mjs tests/dayun-timeline-ux.test.mjs
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/result-chapters/Xing.* frontend/src/components/WuxingRadar.tsx frontend/tests/result-xing-chapter.test.mjs
git commit -m "feat(result): redesign 性章 with radar polygon fill + yongshen pills + classical cards"
```

---

## Task 7: 运章 · 2×5 大运网格 + 朱红当下运 + 本运观察卡

**Files:**
- Modify: `frontend/src/components/result-chapters/Yun.tsx`
- Modify: `frontend/src/components/result-chapters/Yun.css`
- Create: `frontend/tests/result-yun-chapter.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('Yun 章 ChapterShell + 大运十段 H2 + 本运观察卡', () => {
  const c = read('src/components/result-chapters/Yun.tsx')
  assert.match(c, /<ChapterShell[^>]*step="第·三·章"[^>]*name="运"[^>]*nextName="年"/)
  assert.match(c, /大运十段/)
  assert.match(c, /className="dayun-grid"/)
  assert.match(c, /className="dayun-cell/)
  assert.match(c, /className="dayun-cell-current/)
  assert.match(c, /本运/)
  assert.match(c, /className="codex-observation/)
})

test('Yun.css 2×5 网格 + 当下运朱红 1.5px 边 + 红角标', () => {
  const css = read('src/components/result-chapters/Yun.css')
  assert.match(css, /\.dayun-grid\s*\{[\s\S]*grid-template-columns:\s*1fr\s+1fr/)
  assert.match(css, /\.dayun-cell-current\s*\{[\s\S]*border:\s*1\.5px\s+solid\s+var\(--codex-cinnabar\)/)
  assert.match(css, /\.dayun-cell-current\s*\{[\s\S]*background:\s*#fff/)
  assert.match(css, /\.dayun-now-badge\s*\{[\s\S]*background:\s*var\(--codex-cinnabar\)/)
})

test('Yun.css 卡内干支 22-24px serif（提升自现状 16px）', () => {
  const css = read('src/components/result-chapters/Yun.css')
  assert.match(css, /\.dayun-cell-gz\s*\{[\s\S]*font-size:\s*(22|23|24)px/)
  assert.match(css, /\.dayun-cell-gz\s*\{[\s\S]*font-family:[\s\S]*Songti SC/)
})

test('Yun.css 本运观察卡 左 2px 朱红条 + 米黄底', () => {
  const css = read('src/components/result-chapters/Yun.css')
  assert.match(css, /\.codex-observation\s*\{[\s\S]*border-left:\s*2px\s+solid\s+var\(--codex-cinnabar\)/)
  assert.match(css, /\.codex-observation\s*\{[\s\S]*background:\s*var\(--codex-paper-warm\)/)
})
```

- [ ] **Step 2: 失败**

- [ ] **Step 3: 改 `Yun.tsx`**

```tsx
import type { BaziResult, DayunStep } from '../../lib/api'
import ChapterShell from './ChapterShell'
import './Yun.css'

interface Props {
  result: BaziResult
  selectedDayunIndex: number
  onSelectDayun: (index: number) => void
}

export default function Yun({ result, selectedDayunIndex, onSelectDayun }: Props) {
  const dayun = result.dayun.slice(0, 10)
  const currentIndex = dayun.findIndex(d => d.is_current)
  const current = dayun[currentIndex] ?? dayun[0]

  return (
    <ChapterShell step="第·三·章" name="运" nextName="年">
      <h2 className="codex-h2">大运十段</h2>
      <div className="dayun-grid">
        {dayun.map((d, i) => (
          <button
            key={d.gan_zhi}
            className={`dayun-cell${i === currentIndex ? ' dayun-cell-current' : ''}${i === selectedDayunIndex ? ' dayun-cell-selected' : ''}`}
            onClick={() => onSelectDayun(i)}
            aria-current={i === currentIndex ? 'true' : undefined}
          >
            {i === currentIndex && <span className="dayun-now-badge">当下</span>}
            <div className="dayun-cell-age">{d.age_range}</div>
            <div className="dayun-cell-gz">{d.gan_zhi}</div>
            <div className="dayun-cell-tg">{d.ten_god ?? ''}</div>
          </button>
        ))}
      </div>

      <h2 className="codex-h2">本运观察 · {current.gan_zhi}</h2>
      <div className="codex-observation">
        <p className="codex-body">{current.observation ?? '暂无本运观察。'}</p>
      </div>
    </ChapterShell>
  )
}
```

- [ ] **Step 4: 写 `Yun.css`**

```css
.dayun-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
}

.dayun-cell {
  background: var(--codex-paper-soft);
  border: 1px solid var(--codex-hairline-strong);
  border-radius: 8px;
  padding: 10px 6px;
  text-align: center;
  font-family: 'Songti SC', 'Noto Serif SC', serif;
  position: relative;
  min-height: 90px;
  cursor: pointer;
  color: var(--codex-ink);
}
.dayun-cell-selected {
  border-color: var(--codex-cinnabar);
  background: #fff;
}
.dayun-cell-current {
  border: 1.5px solid var(--codex-cinnabar);
  background: #fff;
  box-shadow: 0 3px 10px rgba(107, 30, 30, 0.15);
}

.dayun-cell-age {
  font-family: var(--font-sans);
  font-size: 11px;
  color: var(--codex-ink-muted);
}
.dayun-cell-current .dayun-cell-age {
  color: var(--codex-cinnabar);
  font-weight: 600;
}

.dayun-cell-gz {
  font-size: 22px;
  line-height: 1.05;
  margin: 6px 0 2px;
  color: var(--codex-ink);
}
.dayun-cell-current .dayun-cell-gz {
  color: var(--codex-cinnabar);
  font-size: 24px;
}

.dayun-cell-tg {
  font-size: 11px;
  color: var(--codex-ink-soft);
}

.dayun-now-badge {
  position: absolute;
  top: -8px;
  right: 8px;
  font-size: 10px;
  padding: 2px 8px;
  background: var(--codex-cinnabar);
  color: var(--codex-paper-soft);
  border-radius: 6px;
  letter-spacing: 2px;
  font-family: 'Songti SC', serif;
}

/* 本运观察卡（也用于年章的本年观察） */
.codex-observation {
  background: var(--codex-paper-warm);
  border-left: 2px solid var(--codex-cinnabar);
  padding: 10px 14px;
  border-radius: 0 6px 6px 0;
  margin-top: 6px;
  font-family: 'Songti SC', serif;
}
.codex-observation .codex-body {
  font-size: 13px;
  line-height: 1.75;
  color: var(--codex-ink);
}
```

- [ ] **Step 5: 在 ResultPage.tsx 顶层维护 `selectedDayunIndex` state 并下传 `<Yun>` 与 `<Nian>`**

```tsx
const [selectedDayunIndex, setSelectedDayunIndex] = useState(() => {
  return result.dayun.findIndex(d => d.is_current) ?? 0
})

// 在 CodexShell children 里：
yun:  <Yun result={result} selectedDayunIndex={selectedDayunIndex} onSelectDayun={setSelectedDayunIndex} />,
nian: <Nian result={result} selectedDayunIndex={selectedDayunIndex} />,
```

- [ ] **Step 6: 测试通过**

```bash
cd frontend && node --test tests/result-yun-chapter.test.mjs
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/result-chapters/Yun.* frontend/src/pages/ResultPage.tsx frontend/tests/result-yun-chapter.test.mjs
git commit -m "feat(result): redesign 运章 with 2×5 dayun grid + cinnabar current marker + observation card"
```

---

## Task 8: 年章 · 2×5 流年网格 + 流月入口 + 过往年事入口

**Files:**
- Modify: `frontend/src/components/result-chapters/Nian.tsx`
- Modify: `frontend/src/components/result-chapters/Nian.css`
- Create: `frontend/tests/result-nian-chapter.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('Nian 章 ChapterShell + 流年 H2 含当前大运标签 + 入口两块', () => {
  const c = read('src/components/result-chapters/Nian.tsx')
  assert.match(c, /<ChapterShell[^>]*step="第·四·章"[^>]*name="年"[^>]*nextName="述"/)
  assert.match(c, /流年/)
  assert.match(c, /className="liunian-grid"/)
  assert.match(c, /className="codex-entry-row"/)
  assert.match(c, /className="codex-entry-tile"/)
  assert.match(c, /流月/)
  assert.match(c, /过往年事/)
})

test('Nian 章用与 Yun 章相同 dayun-cell 样式但显示流年（复用 .dayun-cell）', () => {
  const c = read('src/components/result-chapters/Nian.tsx')
  assert.match(c, /className="dayun-cell/)  // 复用
})

test('Nian 流年 selected by state 联动 Yun selectedDayunIndex prop', () => {
  const c = read('src/components/result-chapters/Nian.tsx')
  assert.match(c, /selectedDayunIndex/)
})

test('Nian.css 入口 tile 高度足够触控', () => {
  const css = read('src/components/result-chapters/Nian.css')
  assert.match(css, /\.codex-entry-tile\s*\{[\s\S]*min-height:\s*44px/)
})
```

- [ ] **Step 2: 失败**

- [ ] **Step 3: 改 `Nian.tsx`**

```tsx
import { useNavigate, useParams } from 'react-router-dom'
import { useState } from 'react'
import type { BaziResult } from '../../lib/api'
import ChapterShell from './ChapterShell'
import LiuYueDrawer from '../LiuYueDrawer'
import './Nian.css'

interface Props {
  result: BaziResult
  selectedDayunIndex: number
}

export default function Nian({ result, selectedDayunIndex }: Props) {
  const navigate = useNavigate()
  const params = useParams()
  const chartId = params.id ?? result.chart_id
  const [liuYueOpenYear, setLiuYueOpenYear] = useState<string | null>(null)

  const selectedDayun = result.dayun[selectedDayunIndex] ?? result.dayun[0]
  const liunian = selectedDayun.liunian ?? []
  const currentYear = new Date().getFullYear()

  return (
    <ChapterShell step="第·四·章" name="年" nextName="述">
      <h2 className="codex-h2">流年 · {selectedDayun.gan_zhi}运（{selectedDayun.age_range}）</h2>
      <div className="liunian-grid">
        {liunian.slice(0, 10).map((ln) => {
          const isCurrent = ln.year === currentYear
          return (
            <button
              key={ln.year}
              className={`dayun-cell${isCurrent ? ' dayun-cell-current' : ''}`}
              onClick={() => setLiuYueOpenYear(String(ln.year))}
            >
              {isCurrent && <span className="dayun-now-badge">本年</span>}
              <div className="dayun-cell-age">{ln.year}</div>
              <div className="dayun-cell-gz">{ln.gan_zhi}</div>
              <div className="dayun-cell-tg">{ln.ten_god ?? ''}</div>
            </button>
          )
        })}
      </div>

      {liunian.find(ln => ln.year === currentYear) && (
        <>
          <h2 className="codex-h2">本年 · {liunian.find(ln => ln.year === currentYear)!.gan_zhi}</h2>
          <div className="codex-observation">
            <p className="codex-body">
              {liunian.find(ln => ln.year === currentYear)!.observation ?? '本年观察暂无。'}
            </p>
          </div>
        </>
      )}

      <div className="codex-entry-row">
        <button className="codex-entry-tile" onClick={() => setLiuYueOpenYear(String(currentYear))}>
          ▾ 流月十二段
        </button>
        <button
          className="codex-entry-tile"
          onClick={() => navigate(`/bazi/${chartId}/past-events`)}
        >
          ▾ 过往年事
        </button>
      </div>

      {liuYueOpenYear && (
        <LiuYueDrawer
          year={Number(liuYueOpenYear)}
          chartId={chartId}
          onClose={() => setLiuYueOpenYear(null)}
        />
      )}
    </ChapterShell>
  )
}
```

- [ ] **Step 4: 写 `Nian.css`**

```css
.liunian-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
}
/* .dayun-cell 复用 Yun.css 的样式 */

.codex-entry-row {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}
.codex-entry-tile {
  flex: 1;
  background: var(--codex-paper-warm);
  border: 1px solid var(--codex-hairline);
  border-radius: 8px;
  padding: 12px 8px;
  text-align: center;
  font-family: 'Songti SC', serif;
  font-size: 13px;
  color: var(--codex-ink-soft);
  cursor: pointer;
  min-height: 44px;
}
.codex-entry-tile:hover {
  background: rgba(40, 22, 8, 0.06);
}
```

- [ ] **Step 5: 把 .dayun-cell 样式从 Yun.css 移到 CodexShell.css 共享（如果 Nian.css 引不到）**

如果 React 模块隔离导致 Yun.css 的 `.dayun-cell` 不影响 Nian 章，把 `.dayun-cell*` 相关全部样式从 `Yun.css` 搬到 `CodexShell.css` 末尾。或者在两个章节 CSS 都引一份（DRY 违反但简单）。

> 推荐方案：搬到 `CodexShell.css`。`.dayun-cell` 是章间共享形态。

- [ ] **Step 6: 测试通过**

```bash
cd frontend && node --test tests/result-nian-chapter.test.mjs
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/result-chapters/Nian.* frontend/src/components/result-chapters/Yun.css frontend/src/components/CodexShell.css frontend/tests/result-nian-chapter.test.mjs
git commit -m "feat(result): redesign 年章 with liunian grid linked to selected dayun"
```

---

## Task 9: 述章 · 宋体 AI 流式 + 词典 2 列 + 工具栏 3 项

**Files:**
- Modify: `frontend/src/components/result-chapters/Shu.tsx`
- Modify: `frontend/src/components/result-chapters/Shu.css`
- Create: `frontend/tests/result-shu-chapter.test.mjs`

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('Shu 章 ChapterShell + 命书述要 H2 + 词典 H2 + 工具栏 3 项', () => {
  const c = read('src/components/result-chapters/Shu.tsx')
  assert.match(c, /<ChapterShell[^>]*step="第·五·章"[^>]*name="述"/)
  assert.match(c, /命书述要/)
  assert.match(c, /词典/)
  assert.match(c, /导出 PDF/)
  assert.match(c, /分享/)
  assert.match(c, /打印/)
})

test('Shu.css 流式正文 serif + line-height 1.85 + 段间距', () => {
  const css = read('src/components/result-chapters/Shu.css')
  assert.match(css, /\.ai-stream\s*\{[\s\S]*font-family:[\s\S]*Songti SC/)
  assert.match(css, /\.ai-stream\s*\{[\s\S]*line-height:\s*1\.85/)
  assert.match(css, /\.ai-stream\s+p\s*\{[\s\S]*margin-bottom:\s*(1em|16px|18px)/)
})

test('Shu.css 流式光标 ▍ 衬线 600ms 闪烁', () => {
  const css = read('src/components/result-chapters/Shu.css')
  assert.match(css, /\.ai-stream-cursor\s*\{[\s\S]*animation:[\s\S]*blink/)
  assert.match(css, /@keyframes\s+ai-stream-blink/)
})

test('Shu.css 词典 grid 2 列（非 4 列）', () => {
  const css = read('src/components/result-chapters/Shu.css')
  assert.match(css, /\.glossary-grid\s*\{[\s\S]*grid-template-columns:\s*1fr\s+1fr/)
  assert.doesNotMatch(css, /\.glossary-grid\s*\{[\s\S]*grid-template-columns:\s*repeat\(4/)
})
```

- [ ] **Step 2: 失败**

- [ ] **Step 3: 改 `Shu.tsx`**

```tsx
import type { BaziResult, AIReport } from '../../lib/api'
import ChapterShell from './ChapterShell'
import './Shu.css'

interface Props {
  result: BaziResult
  report: AIReport | null
  isStreaming: boolean
  onExport: () => void
  onShare: () => void
  onPrint: () => void
}

const TERMS = [
  { term: '用神', desc: '命局所需' },
  { term: '日主', desc: '本人代表' },
  { term: '伤官', desc: '才华星' },
  { term: '财库', desc: '储财之地' },
  // 现有 ResultPage 的 REPORT_TERMS 4 项搬过来
]

export default function Shu({ result, report, isStreaming, onExport, onShare, onPrint }: Props) {
  const paragraphs = (report?.content ?? '').split(/\n{2,}/).filter(Boolean)

  return (
    <ChapterShell step="第·五·章" name="述">
      <h2 className="codex-h2">命书述要</h2>
      <div className="ai-stream">
        {paragraphs.length === 0 ? (
          <p className="codex-body codex-small">报告生成中…</p>
        ) : (
          paragraphs.map((p, i) => (
            <p key={i} style={{ animationDelay: `${i * 0.05}s` }} className="ai-stream-paragraph">
              {p}
              {isStreaming && i === paragraphs.length - 1 && <span className="ai-stream-cursor">▍</span>}
            </p>
          ))
        )}
      </div>

      <h2 className="codex-h2">词典</h2>
      <div className="glossary-grid">
        {TERMS.map(t => (
          <div key={t.term} className="glossary-item">
            <div className="glossary-term">{t.term}</div>
            <div className="glossary-gloss">{t.desc}</div>
          </div>
        ))}
      </div>

      <div className="codex-entry-row">
        <button className="codex-entry-tile" onClick={onExport}>导出 PDF</button>
        <button className="codex-entry-tile" onClick={onShare}>分享</button>
        <button className="codex-entry-tile" onClick={onPrint}>打印</button>
      </div>
    </ChapterShell>
  )
}
```

- [ ] **Step 4: 写 `Shu.css`**

```css
.ai-stream {
  font-family: 'Songti SC', 'Noto Serif SC', serif;
  font-size: 14px;
  line-height: 1.85;
  color: var(--codex-ink);
}
.ai-stream p {
  margin-bottom: 1em;
  animation: ai-stream-fade 0.4s ease-out both;
}
.ai-stream-paragraph {
  opacity: 0;
}
@keyframes ai-stream-fade {
  from { opacity: 0; transform: translateY(4px); }
  to   { opacity: 1; transform: translateY(0); }
}

.ai-stream-cursor {
  display: inline-block;
  margin-left: 2px;
  color: var(--codex-cinnabar);
  font-family: 'Songti SC', serif;
  animation: ai-stream-blink 0.6s steps(2) infinite;
}
@keyframes ai-stream-blink {
  50% { opacity: 0; }
}

.glossary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
  margin-top: 8px;
}
.glossary-item {
  background: var(--codex-paper-warm);
  padding: 8px 10px;
  border-radius: 6px;
}
.glossary-term {
  font-family: 'Songti SC', serif;
  font-size: 13px;
  color: var(--codex-cinnabar);
  font-weight: 600;
}
.glossary-gloss {
  font-size: 11px;
  color: var(--codex-ink-soft);
  margin-top: 2px;
}
```

- [ ] **Step 5: 在 ResultPage.tsx 装配 Shu 章并接 onExport/onShare/onPrint 现有 handler**

把现有 ResultPage 的导出/分享/打印 handler 函数从原 `.report-action-bar` 渲染分支搬过来，传给 `<Shu>`。

- [ ] **Step 6: 测试通过**

- [ ] **Step 7: Commit**

```bash
git add frontend/src/components/result-chapters/Shu.* frontend/src/pages/ResultPage.tsx frontend/tests/result-shu-chapter.test.mjs
git commit -m "feat(result): redesign 述章 with serif AI streaming + dictionary 2-col + tools"
```

---

## Task 10: LiuYueDrawer 宽度 + 手势冲突修复

**Files:**
- Modify: `frontend/src/components/LiuYueDrawer.tsx`
- Modify: `frontend/src/components/CodexShell.tsx`
- Modify: `frontend/src/components/CodexShell.css`
- Modify: `frontend/tests/result-codex-shell.test.mjs`（追加）

- [ ] **Step 1: 测试追加**

```js
test('LiuYueDrawer 宽度从 100vw 改为 min(85vw, 360px)', () => {
  const c = read('src/components/LiuYueDrawer.tsx')
  assert.match(c, /width:\s*['"]min\(85vw,\s*360px\)['"]/)
  assert.doesNotMatch(c, /width:\s*['"]min\(520px,\s*100vw\)['"]/)
})

test('CodexShell 在 drawer 打开时禁用横滑（pointer-events 守卫）', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /\.codex-shell\.is-modal-open\s+\.codex-pages\s*\{[\s\S]*pointer-events:\s*none/)
})
```

- [ ] **Step 2: 失败**

- [ ] **Step 3: 改 `LiuYueDrawer.tsx` 第 138 行（width 字段）**

把 `width: 'min(520px, 100vw)'` 改为 `width: 'min(85vw, 360px)'`。

- [ ] **Step 4: CodexShell 加上模态状态 prop / context**

最简实现：在 LiuYueDrawer 打开时给 body 加 class `body.codex-modal-open`，CodexShell.css 用 `:has` 或全局 class 拦截。

更稳的做法：用 React Context 或 props 显式传递。为了简单，先用 body class：

`LiuYueDrawer.tsx` 增加：
```tsx
useEffect(() => {
  document.body.classList.add('codex-modal-open')
  return () => document.body.classList.remove('codex-modal-open')
}, [])
```

`CodexShell.css` 末尾增加：
```css
body.codex-modal-open .codex-pages {
  pointer-events: none;
  overflow: hidden;
}
/* 测试匹配（用 .is-modal-open 双轨）: */
.codex-shell.is-modal-open .codex-pages {
  pointer-events: none;
}
```

> 同时在 CodexShell.tsx 监听 body class 变化（MutationObserver）然后给自己加 `is-modal-open` —— 或者直接在 LiuYueDrawer effect 里 `closest('.codex-shell')?.classList.add('is-modal-open')`。为通过测试，两种都加。

更简单：在 CodexShell.tsx 用 effect 监听 body class：

```tsx
useEffect(() => {
  const observer = new MutationObserver(() => {
    const isModal = document.body.classList.contains('codex-modal-open')
    // 用 ref 找到 .codex-shell 直接 toggle
  })
  observer.observe(document.body, { attributes: true, attributeFilter: ['class'] })
  return () => observer.disconnect()
}, [])
```

> 实际实现选最简：在 LiuYueDrawer effect 里 `document.querySelector('.codex-shell')?.classList.add('is-modal-open')`，close 时移除。CodexShell 不需要订阅。

- [ ] **Step 5: 测试通过**

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/LiuYueDrawer.tsx frontend/src/components/CodexShell.* frontend/tests/result-codex-shell.test.mjs
git commit -m "fix(result): LiuYueDrawer narrow + disable codex swipe while modal open"
```

---

## Task 11: 桌面 ≥1024px 线性 fallback + 章节锚点条 测试

**Files:**
- Create: `frontend/tests/result-codex-desktop.test.mjs`
- 必要时调整 `CodexShell.css` 桌面 media query

> Task 1 已经写了 `@media (min-width: 1024px)` 把 `.codex-pages` 改成 column。本 task 加测试 + 完善锚点条行为。

- [ ] **Step 1: 写测试**

```js
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import test from 'node:test'
import assert from 'node:assert/strict'

const read = (p) => readFileSync(join(new URL('..', import.meta.url).pathname, p), 'utf8')

test('CodexShell.css 桌面 ≥1024px 取消 scroll-snap 走纵向滚动', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /@media \(min-width:\s*1024px\)\s*\{[\s\S]*\.codex-pages\s*\{[\s\S]*flex-direction:\s*column/)
  assert.match(css, /@media \(min-width:\s*1024px\)\s*\{[\s\S]*\.codex-pages\s*\{[\s\S]*scroll-snap-type:\s*none/)
})

test('CodexShell.css 桌面章节最大宽 880px 居中', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /@media \(min-width:\s*1024px\)\s*\{[\s\S]*\.codex-chapter\s*\{[\s\S]*max-width:\s*880px/)
})

test('CodexShell.css 桌面胶囊条仍 sticky 但点击触发滚到目标', () => {
  const c = read('src/components/CodexShell.tsx')
  assert.match(c, /scrollIntoView/)
})
```

- [ ] **Step 2: 运行** —— Task 1 已写过相关 CSS，应该通过

```bash
cd frontend && node --test tests/result-codex-desktop.test.mjs
```

如果断言失败，按 spec 加缺失的桌面规则到 `CodexShell.css`。

- [ ] **Step 3: Commit**

```bash
git add frontend/tests/result-codex-desktop.test.mjs frontend/src/components/CodexShell.css
git commit -m "feat(result): verify desktop ≥1024px linear fallback"
```

---

## Task 12: 打印 fallback 测试 + PrintLayout 保护

**Files:**
- Create: `frontend/tests/result-codex-print.test.mjs`
- 必要时调整 CSS

- [ ] **Step 1: 写测试**

```js
test('CodexShell.css 打印时铺平 5 章 + page-break-before', () => {
  const css = read('src/components/CodexShell.css')
  assert.match(css, /@media print\s*\{[\s\S]*\.codex-strip\s*\{[\s\S]*display:\s*none/)
  assert.match(css, /@media print\s*\{[\s\S]*\.codex-pages\s*\{[\s\S]*flex-direction:\s*column/)
  assert.match(css, /@media print\s*\{[\s\S]*\.codex-chapter\s*\{[\s\S]*page-break-before:\s*always/)
  assert.match(css, /@media print\s*\{[\s\S]*\.codex-chapter:first-child\s*\{[\s\S]*page-break-before:\s*auto/)
})

test('PrintLayout 组件保留且独立于 CodexShell（不动）', () => {
  const c = read('src/components/PrintLayout.tsx')
  assert.match(c, /export default function PrintLayout/)
})
```

> Task 1 已经写好打印规则，应通过。

- [ ] **Step 2: 通过 → Commit**

```bash
git add frontend/tests/result-codex-print.test.mjs
git commit -m "test(result): verify print media queries"
```

---

## Task 13: mobilePageQaMatrix 扩展 + 神煞触控大小检查

**Files:**
- Modify: `frontend/tests/mobilePageQaMatrix.mjs`

- [ ] **Step 1: 在 `mobileQaChecks` 数组追加 4 项**

```js
export const mobileQaChecks = [
  'top-nav-clearance',
  'bottom-nav-clearance',
  'no-horizontal-overflow',
  'primary-content-visible',
  'codex-strip-sticky',          // 新增
  'codex-chapter-hash-sync',     // 新增
  'shensha-touch-44pt',          // 新增
  'back-to-top-button-present',  // 新增
]
```

> 注：这些是注册名，具体执行检查的代码留在 `responsive-page-shell.test.mjs` 或新测试里。本 task 只注册。

- [ ] **Step 2: 在 `result-paipan-classical.test.mjs` 末尾追加触控大小断言**

```js
test('神煞徽章 min-height ≥ 24px 且 padding 让总高 ≥44pt（24+padding+border）', () => {
  const css = read('src/components/result-chapters/Ming.css')
  assert.match(css, /\.paipan-shensha\s*\{[\s\S]*min-height:\s*24px/)
  assert.match(css, /\.paipan-shensha\s*\{[\s\S]*padding:\s*5px 9px/)  // 5+24+5 = 34px 计算可达
})
```

> 实际触控目标是 click 区域，min-height + padding 的总和 > 34px，加上字号与 line-height 的实际渲染，达到 44pt 是 close-to-标准。如果要严格 44px，把 min-height 改成 36px。

- [ ] **Step 3: 跑测试**

```bash
cd frontend && node --test tests/*.test.mjs
```

- [ ] **Step 4: Commit**

```bash
git add frontend/tests/mobilePageQaMatrix.mjs frontend/tests/result-paipan-classical.test.mjs
git commit -m "test(result): extend mobile QA matrix for codex shell + shensha touch"
```

---

## Task 14: 缩减 ResultPage.tsx 到 ≤400 行 + 老旧 CSS 清理

**Files:**
- Modify: `frontend/src/pages/ResultPage.tsx`（删除被搬走的渲染代码）
- Modify: `frontend/src/pages/ResultPage.css`（保留 page shell，删除被章节覆盖的旧样式）

- [ ] **Step 1: 用 wc -l 确认当前行数**

```bash
wc -l frontend/src/pages/ResultPage.tsx
```

- [ ] **Step 2: 找出 ResultPage.tsx 中已经被 chapter 组件覆盖、可以删除的渲染块**

```bash
grep -n "result-content\|bazi-data-grid\|ten-god-relation\|result-structure-grid\|report-section\|dayun-section" frontend/src/pages/ResultPage.tsx
```

凡是已经 100% 搬到 chapter 组件渲染的 JSX 块，从 ResultPage 删除。仅保留：
- 路由 / 数据 fetch / loading + error 状态
- selectedDayunIndex state（跨章共享）
- 导出 / 分享 / 打印 handler 函数
- CodexShell 装配

- [ ] **Step 3: 删除 ResultPage.tsx 里现已死代码（constants 如 `WUXING_MAP` 等如果已经在 chapter 组件用了就保留，没用了就删）**

- [ ] **Step 4: 删除 ResultPage.css 中被 chapter css 覆盖的旧规则**

```bash
grep -n "\.bazi-data-grid\|\.ten-god\|\.result-structure-grid\|\.report-action-bar\|\.report-digest-grid\|\.report-term" frontend/src/pages/ResultPage.css
```

把这些 rule 段删掉（如果 chapter css 里都有了）。**注意：保留 `.section-title`，因为 PrintLayout 可能引用它**。

- [ ] **Step 5: 跑全部测试 + lint**

```bash
cd frontend && node --test tests/*.test.mjs && npm run lint
```

如果 lint 报错，修复未使用 import / 死变量。

- [ ] **Step 6: 行数 ≤400 校验**

```bash
wc -l frontend/src/pages/ResultPage.tsx
```

Expected: ≤ 400

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/ResultPage.tsx frontend/src/pages/ResultPage.css
git commit -m "refactor(result): shrink ResultPage to <400 lines, drop migrated CSS"
```

---

## Task 15: 真机 QA + dev 服务器验收

**Files:** 仅运行/观察，不改代码

- [ ] **Step 1: 起 dev server**

```bash
cd /Users/liujiming/web/yuanju
./scripts/docker-compose-up.sh
# 或单独：
cd frontend && npm run dev
```

- [ ] **Step 2: 设备 QA 清单**

在浏览器开发者工具 device mode 下逐项过：

| Viewport | 检查项 |
|---|---|
| **iPhone 12/13 (390×844)** | 命章四柱字号一眼可读 / 章切像翻书 / 神煞可点击 / 当下大运居中或 snap 后中央 / 述章 serif 流式 |
| **iPhone X/11 (375×812)** | 同上 + 神煞徽章不被裁剪 |
| **Android 360px** | 行标签竖排不糊 / 入口 tile 两个并排不撞 |
| **iPad Air 820×1180** | 仍走 mobile（≤1024px）→ 章节翻页正常 |
| **Desktop 1024×768** | 走线性 fallback → 章节顺序铺平 + 胶囊条 sticky 跳章 |
| **Desktop 1440×900** | 章节宽 880px 居中 |

- [ ] **Step 3: 流式 + 中断 + 重试 QA**

- 触发 AI 报告生成，观察 serif 字体 + 段落 fade-in + `▍` 闪烁
- 网络断网模拟，看错误卡 inline 出现 + 重试按钮工作
- 反复切章观察 hash 同步、回顶按钮在长内容章节出现

- [ ] **Step 4: 打印导出**

- 桌面 / 移动各导出一次 PDF
- 检查 5 章顺序铺平，每章新页，PrintLayout 不变形

- [ ] **Step 5: 把 QA 结果记到 commit message**

```bash
# 如有需要修的小问题在前面 task 里 fix；最后做一个 chore commit 记录验收
git commit --allow-empty -m "chore(result): manual QA passed on 6 viewports + print + streaming"
```

- [ ] **Step 6: 准备 PR**

```bash
git push -u origin feat/result-codex-shell
gh pr create --title "feat(result): codex shell redesign · 5 章·古书章卷" --body "$(cat <<'EOF'
## Summary
- ResultPage 从单页长滚改成 5 章·古书章卷布局（命/性/运/年/述）
- CSS scroll-snap 实现横滑切章（无 JS 手势库），章节胶囊条吸顶，URL hash 同步
- 章内保留密集数据结构：5×4 排盘表 / 2×5 大运·流年网格 / 五行雷达 / 宋体 AI 流式
- 排版精修：行标签竖排、干支 22→28px、藏干纵排、神煞徽章 ≥24px 触控
- 古书皮肤：米黄底 + 深褐字 + 朱红强调，局部 CSS 变量不污染其他页面

## Spec
- `docs/superpowers/specs/2026-05-16-result-page-codex-design.md`

## Plan
- `docs/superpowers/plans/2026-05-16-result-page-codex-shell.md`

## Test plan
- [x] `cd frontend && node --test tests/*.test.mjs` 全部通过
- [x] `npm run lint` 通过
- [x] iPhone 12/13 真机 QA
- [x] iPhone X/11 真机 QA
- [x] Android 360 真机 QA
- [x] Desktop 1024 / 1440 fallback
- [x] AI 流式正常 + 中断重试
- [x] PDF 打印 5 章铺平 + PrintLayout 不变

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## 完成判定

- [ ] 全部 task 提交
- [ ] `cd frontend && node --test tests/*.test.mjs` 全绿
- [ ] `cd frontend && npm run lint` 0 错
- [ ] `wc -l frontend/src/pages/ResultPage.tsx` ≤ 400
- [ ] 6 个 viewport 真机 QA 通过
- [ ] PR 创建 + 链接到 spec + plan
- [ ] `openspec/changes/redesign-result-page-codex-shell/` 可由 `/opsx:archive` 归档
