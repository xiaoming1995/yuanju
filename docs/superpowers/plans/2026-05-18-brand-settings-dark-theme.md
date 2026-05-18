# Brand Settings 深色主题对齐 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把 `/settings/brand` 页面与 `LogoCropModal` 弹窗的所有局部浅色硬编码替换成全站深色 + 金色主题 token，保留 `BrandPreviewCard` 浅色样张外观。

**Architecture:** 纯样式 PR。两个 CSS 文件整体重写，两个 TSX 文件各加 / 改 2 处按钮 `className`。`BrandPreviewCard` 不动。新增一个静态正则测试守住反向回归（CSS 文件不能再含旧色卡硬编码，TSX 必须使用全局 `.btn` 类）。

**Tech Stack:** React 19 + TypeScript + Vite，CSS Variables（无 CSS-in-JS / 无 Tailwind / 无 UI 框架），Node `node:test` 静态正则测试。

**Repo paths:**
- Repo root: `/Users/liujiming/web/yuanju`
- Frontend cwd: `/Users/liujiming/web/yuanju/frontend`
- Git 命令使用 `git -C /Users/liujiming/web/yuanju ...`
- 当前分支：`feat/export-brand-customization`，起始 HEAD：`ac7cc01`

**Spec:** `docs/superpowers/specs/2026-05-18-brand-settings-dark-theme-design.md`

---

## File Structure

| 文件 | 改动 | 责任 |
|------|------|------|
| `frontend/tests/brand-settings-dark-theme.test.mjs` | **新建** | 反向回归守卫：断言两个 CSS 不再含旧浅色硬编码、两个 TSX 使用全局 `.btn` 类 |
| `frontend/src/pages/BrandSettingsPage.css` | **整体重写** | 全部局部色 → 主题 token；移除 `.brand-logo-actions button` 规则（按钮交给 `.btn`） |
| `frontend/src/pages/BrandSettingsPage.tsx` | **2 处 className 加** | logo 上传/删除按钮加 `className="btn btn-ghost btn-sm"` |
| `frontend/src/components/LogoCropModal.css` | **整体重写** | 全部局部色 → 主题 token；删除所有 `.logo-crop-btn-*` 规则 |
| `frontend/src/components/LogoCropModal.tsx` | **2 处 className 替换** | 取消 → `className="btn btn-ghost"`，确认 → `className="btn btn-primary"` |

不动文件：`BrandPreviewCard.tsx` / `ShareCard.tsx` / `PrintLayout.tsx` / `ResultPage.*` / backend / database / api。

---

## Task 0: 确认基线

**Files:** 无修改

- [ ] **Step 1: 确认分支与 HEAD**

```bash
git -C /Users/liujiming/web/yuanju status
git -C /Users/liujiming/web/yuanju log --oneline -1
```

Expected:
- Branch: `feat/export-brand-customization`
- 工作树干净（除可能的未提交 frontend/.. 文件外）
- HEAD: `ac7cc01 docs(spec): brand settings dark theme alignment`

如果工作树不干净 / 不在该分支上，stop 并报告 BLOCKED。

- [ ] **Step 2: 确认全局 `.btn` 与 token 都存在（健全性 sanity check）**

```bash
grep -nE '^\.btn\b|^\.btn-' /Users/liujiming/web/yuanju/frontend/src/index.css | head -10
grep -nE '--bg-card|--bg-elevated|--text-primary|--text-secondary|--text-muted|--text-accent|--border-default|--border-subtle|--border-accent|--radius-md|--radius-sm' /Users/liujiming/web/yuanju/frontend/src/index.css | wc -l
```

Expected:
- `.btn` / `.btn-primary` / `.btn-secondary` / `.btn-ghost` / `.btn-sm` / `.btn-lg` 至少各 1 行
- token grep 返回 ≥ 10 行（11 个 token 各至少一处声明 + 引用）

如果有 token 缺失，停下来汇报；这意味着 spec 与代码现状不符。

---

## Task 1: RED — 加反向回归静态测试

**Files:**
- Create: `frontend/tests/brand-settings-dark-theme.test.mjs`

- [ ] **Step 1: 写测试文件**

```javascript
// frontend/tests/brand-settings-dark-theme.test.mjs
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { fileURLToPath } from 'node:url'
import { dirname, resolve } from 'node:path'

const __dirname = dirname(fileURLToPath(import.meta.url))
const REPO_ROOT = resolve(__dirname, '..')

// 旧主题硬编码 — 改造后 BrandSettingsPage.css / LogoCropModal.css 中均不应再出现。
// （这些色值仅作为本次重构的反向守卫；它们在 PrintLayout.tsx 等导出产物中仍可合理出现，
//  因此我们只在这两个 CSS 文件中做断言。）
const LEGACY_LIGHT_HEXES = [
  '#fdf9f2',  // cream surface
  '#e0cca0',  // tan border
  '#2a1a0a',  // dark-brown text on light bg
  '#5a3a1a',  // mid-brown muted text
]

const CSS_FILES_TO_CHECK = [
  'src/pages/BrandSettingsPage.css',
  'src/components/LogoCropModal.css',
]

for (const relPath of CSS_FILES_TO_CHECK) {
  test(`${relPath} 不再包含原 light-theme 硬编码色`, async () => {
    const text = (await readFile(resolve(REPO_ROOT, relPath), 'utf8')).toLowerCase()
    for (const hex of LEGACY_LIGHT_HEXES) {
      assert.ok(
        !text.includes(hex.toLowerCase()),
        `${relPath} 仍包含浅色硬编码 ${hex}，应改用 var(--*) token`,
      )
    }
  })
}

test('BrandSettingsPage.tsx logo 上传/删除按钮使用全局 .btn-ghost-sm', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/pages/BrandSettingsPage.tsx'),
    'utf8',
  )
  // logo-actions 区域内至少需出现一次 `className="btn btn-ghost btn-sm"`
  assert.match(
    text,
    /className="btn btn-ghost btn-sm"/,
    'BrandSettingsPage.tsx 的 logo 按钮应使用 className="btn btn-ghost btn-sm"',
  )
})

test('LogoCropModal.tsx 取消/确认按钮使用全局 .btn 类', async () => {
  const text = await readFile(
    resolve(REPO_ROOT, 'src/components/LogoCropModal.tsx'),
    'utf8',
  )
  assert.match(
    text,
    /className="btn btn-ghost"/,
    'LogoCropModal.tsx 的取消按钮应使用 className="btn btn-ghost"',
  )
  assert.match(
    text,
    /className="btn btn-primary"/,
    'LogoCropModal.tsx 的确认按钮应使用 className="btn btn-primary"',
  )
  assert.doesNotMatch(
    text,
    /className="logo-crop-btn-(ghost|primary)"/,
    'LogoCropModal.tsx 不应再使用本地 logo-crop-btn-* 类（已废弃）',
  )
})
```

- [ ] **Step 2: 运行测试，确认 RED**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected: 4 tests fail with messages：
- `BrandSettingsPage.css 仍包含浅色硬编码 #fdf9f2`
- `LogoCropModal.css 仍包含浅色硬编码 #fdf9f2`
- `BrandSettingsPage.tsx 的 logo 按钮应使用 className="btn btn-ghost btn-sm"`
- `LogoCropModal.tsx 不应再使用本地 logo-crop-btn-* 类`

至少这 4 个测试必须 FAIL；如果有任何一个 PASS，调查为什么（可能是预期之外的代码状态），别继续。

- [ ] **Step 3: 提交 RED**

```bash
git -C /Users/liujiming/web/yuanju add frontend/tests/brand-settings-dark-theme.test.mjs
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "test(brand-settings): add dark-theme regression guard (RED)"
```

---

## Task 2: GREEN — 整体重写 BrandSettingsPage.css

**Files:**
- Modify: `frontend/src/pages/BrandSettingsPage.css`

- [ ] **Step 1: 用下面这版替换整个文件**

```css
.brand-page {
  max-width: 720px;
  margin: 0 auto;
  padding-bottom: 80px;
}
.brand-page-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 24px;
}
.brand-page-header h1 {
  font-size: 20px;
  margin: 0;
  color: var(--text-primary);
}
.brand-back {
  background: none;
  border: 1px solid var(--border-default);
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: var(--text-secondary);
  transition: border-color 0.18s, color 0.18s;
}
.brand-back:hover {
  border-color: var(--border-accent);
  color: var(--text-primary);
}
.brand-loading {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
}
.brand-error {
  background: rgba(192, 57, 43, 0.12);
  border: 1px solid rgba(192, 57, 43, 0.4);
  color: #e87171;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-size: 13px;
}
.brand-unsaved {
  background: rgba(212, 184, 150, 0.10);
  border: 1px solid var(--border-accent);
  color: var(--text-accent);
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-size: 13px;
}
.brand-success {
  background: rgba(108, 191, 122, 0.12);
  border: 1px solid rgba(108, 191, 122, 0.4);
  color: #7dd87f;
  padding: 8px 14px;
  border-radius: var(--radius-sm);
  margin-bottom: 16px;
  font-size: 13px;
  animation: brand-success-fade 2.5s ease-out forwards;
}
@keyframes brand-success-fade {
  0%   { opacity: 0; transform: translateY(-4px); }
  10%  { opacity: 1; transform: translateY(0); }
  80%  { opacity: 1; }
  100% { opacity: 0; }
}
.brand-section {
  background: var(--bg-card);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: 18px 20px;
  margin-bottom: 18px;
}
.brand-section h2 {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 14px;
  letter-spacing: 1px;
}
.brand-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}
.brand-field span {
  font-size: 13px;
  color: var(--text-secondary);
}
.brand-field input[type="text"] {
  padding: 8px 12px;
  border: 1px solid var(--border-default);
  border-radius: var(--radius-sm);
  font-size: 14px;
  background: var(--bg-elevated);
  color: var(--text-primary);
  transition: border-color 0.18s;
}
.brand-field input[type="text"]:focus {
  outline: none;
  border-color: var(--border-accent);
}
.brand-field input[type="text"]:disabled {
  background: var(--bg-card);
  color: var(--text-muted);
}
.brand-field small {
  align-self: flex-end;
  font-size: 11px;
  color: var(--text-muted);
}
.brand-logo-row {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-top: 8px;
}
.brand-logo-preview {
  width: 80px;
  height: 80px;
  border: 1px dashed var(--border-default);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-elevated);
  overflow: hidden;
}
.brand-logo-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.brand-logo-empty {
  font-size: 11px;
  color: var(--text-muted);
}
.brand-logo-actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.brand-logo-actions small {
  font-size: 11px;
  color: var(--text-muted);
}
.brand-radio-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  margin-bottom: 12px;
  font-size: 13px;
  color: var(--text-primary);
}
.brand-radio-row label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  cursor: pointer;
}
.brand-preview-section {
  display: flex;
  flex-direction: column;
  align-items: center;
}
.brand-footer-actions {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-top: 8px;
}
.brand-tip {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  margin-top: 16px;
}
.brand-bottom-link {
  display: block;
  text-align: center;
  margin-top: 24px;
  color: var(--text-muted);
  font-size: 13px;
}
```

- [ ] **Step 2: 跑 RED 测试，确认两条 BrandSettingsPage.css 断言转为 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected:
- `BrandSettingsPage.css 不再包含原 light-theme 硬编码色` → **PASS**
- `LogoCropModal.css 不再包含原 light-theme 硬编码色` → FAIL（下一任务处理）
- 两个 TSX 断言 → FAIL（任务 3、5 处理）

- [ ] **Step 3: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/BrandSettingsPage.css
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "style(brand-settings): swap page CSS to dark theme tokens"
```

---

## Task 3: BrandSettingsPage.tsx — 2 处按钮 className 加全局 .btn

**Files:**
- Modify: `frontend/src/pages/BrandSettingsPage.tsx`（当前 194–205 行附近的 logo-actions 区域）

- [ ] **Step 1: 找到现状代码**

```bash
grep -n 'brand-logo-actions' /Users/liujiming/web/yuanju/frontend/src/pages/BrandSettingsPage.tsx
```

Expected: 输出 `194:<div className="brand-logo-actions">` 这一行（前后可能略有偏移），其下两个 `<button>` 需要改。

- [ ] **Step 2: 将这段 JSX**

```tsx
          <div className="brand-logo-actions">
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? '上传中...' : (serverState.logo_url ? '更换' : '上传')}
            </button>
            {serverState.logo_url && (
              <button type="button" onClick={onLogoDelete} disabled={uploading}>删除</button>
            )}
            <small>PNG / JPG / WebP，≤ 2MB</small>
          </div>
```

**替换为：**

```tsx
          <div className="brand-logo-actions">
            <button
              type="button"
              className="btn btn-ghost btn-sm"
              onClick={() => fileInputRef.current?.click()}
              disabled={uploading}
            >
              {uploading ? '上传中...' : (serverState.logo_url ? '更换' : '上传')}
            </button>
            {serverState.logo_url && (
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={onLogoDelete}
                disabled={uploading}
              >
                删除
              </button>
            )}
            <small>PNG / JPG / WebP，≤ 2MB</small>
          </div>
```

只有两个变化：两个 `<button>` 各加了一个 `className="btn btn-ghost btn-sm"`。其他属性和 logic 完全不变。

- [ ] **Step 3: 跑 RED 测试，确认 BrandSettingsPage.tsx 那条断言 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected:
- `BrandSettingsPage.tsx logo 上传/删除按钮使用全局 .btn-ghost-sm` → **PASS**
- LogoCropModal 的两条仍 FAIL（任务 4、5 处理）

- [ ] **Step 4: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/pages/BrandSettingsPage.tsx
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "style(brand-settings): use global .btn-ghost-sm for logo action buttons"
```

---

## Task 4: GREEN — 整体重写 LogoCropModal.css

**Files:**
- Modify: `frontend/src/components/LogoCropModal.css`

- [ ] **Step 1: 用下面这版替换整个文件**

```css
.logo-crop-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}

.logo-crop-modal {
  background: var(--bg-card);
  border: 1px solid var(--border-default);
  border-radius: var(--radius-md);
  width: min(440px, 100%);
  display: flex;
  flex-direction: column;
  padding: 18px 20px 16px;
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.45);
}

.logo-crop-title {
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: 1px;
}

.logo-crop-canvas-area {
  position: relative;
  width: 100%;
  aspect-ratio: 1 / 1;
  background: var(--bg-base);
  border-radius: var(--radius-sm);
  overflow: hidden;
  margin-bottom: 12px;
}

.logo-crop-loading {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-muted);
  font-size: 13px;
}

.logo-crop-zoom-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.logo-crop-zoom-row input[type="range"] {
  flex: 1;
}

.logo-crop-note {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  margin-bottom: 14px;
}

.logo-crop-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
```

注意：所有 `.logo-crop-btn-ghost` / `.logo-crop-btn-primary` 规则已**整体删除**，因为下一步 TSX 改用全局 `.btn`。

- [ ] **Step 2: 跑 RED 测试，确认 LogoCropModal.css 那条断言 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected:
- `LogoCropModal.css 不再包含原 light-theme 硬编码色` → **PASS**
- `LogoCropModal.tsx 取消/确认按钮使用全局 .btn 类` → 仍 FAIL（任务 5 处理）

- [ ] **Step 3: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/LogoCropModal.css
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "style(logo-crop-modal): swap modal CSS to dark theme tokens"
```

---

## Task 5: LogoCropModal.tsx — 2 处按钮 className 替换

**Files:**
- Modify: `frontend/src/components/LogoCropModal.tsx`（当前 101–113 行）

- [ ] **Step 1: 将这段 JSX**

```tsx
        <div className="logo-crop-actions">
          <button type="button" className="logo-crop-btn-ghost" onClick={onCancel} disabled={processing}>
            取消
          </button>
          <button
            type="button"
            className="logo-crop-btn-primary"
            onClick={handleConfirm}
            disabled={processing || !areaPx || !imgUrl}
          >
            {processing ? '处理中...' : '确认裁剪'}
          </button>
        </div>
```

**替换为：**

```tsx
        <div className="logo-crop-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel} disabled={processing}>
            取消
          </button>
          <button
            type="button"
            className="btn btn-primary"
            onClick={handleConfirm}
            disabled={processing || !areaPx || !imgUrl}
          >
            {processing ? '处理中...' : '确认裁剪'}
          </button>
        </div>
```

只有两个变化：`logo-crop-btn-ghost` → `btn btn-ghost`，`logo-crop-btn-primary` → `btn btn-primary`。其他不变。

- [ ] **Step 2: 跑 RED 测试，确认所有断言 PASS**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected: **all 4 tests PASS**。

- [ ] **Step 3: 提交**

```bash
git -C /Users/liujiming/web/yuanju add frontend/src/components/LogoCropModal.tsx
git -C /Users/liujiming/web/yuanju -c commit.gpgsign=false commit -m "style(logo-crop-modal): use global .btn classes for action buttons"
```

---

## Task 6: 编译 / lint / 既有测试全套验证

**Files:** 无修改

- [ ] **Step 1: 跑既有 brand-settings 静态测试，确认未被破坏**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings.test.mjs
```

Expected: 全绿（既有 6 个断言通过）。这些断言只覆盖文件存在与上传流程，跟样式无关，应不受本次改动影响。

- [ ] **Step 2: 跑本次新加的回归守卫测试**

```bash
cd /Users/liujiming/web/yuanju/frontend && node --test tests/brand-settings-dark-theme.test.mjs
```

Expected: 4 tests pass。

- [ ] **Step 3: lint**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run lint
```

Expected: 干净 — 不应有新的 ESLint 错误 / 警告。CSS 文件不被 ESLint 扫，只检查两处 TSX 改动，应当干净。

- [ ] **Step 4: TypeScript 编译 + Vite 构建**

```bash
cd /Users/liujiming/web/yuanju/frontend && npm run build
```

Expected: 干净 — TSC 无错，Vite 输出 `dist/`。两处 `className="btn ..."` 是字符串字面量，无类型变化，应当无错。

如果有任何步骤失败，停下来查 root cause，**不要**直接 `--no-verify` 或绕过。

- [ ] **Step 5: （仅在本任务全步骤通过后）提交一笔空一致性 marker**

不需要单独的 commit 文件 — 验证步骤都不改动文件。如果 npm run build 产生了任何 build artifact 文件（如 `frontend/dist`），它们应当已经在 `.gitignore`。运行：

```bash
git -C /Users/liujiming/web/yuanju status
```

Expected: 工作树干净，本次改动共 5 个 commit（任务 1 RED + 任务 2-5 各 1 commit）。

---

## Task 7: 手动视觉验收清单（用户/QA）

**Files:** 无修改 — 这一步交给用户在 docker compose 环境下浏览验证。

- [ ] **Step 1: 重建 frontend 容器**

```bash
docker compose -f /Users/liujiming/web/yuanju/docker-compose.yml up -d --build frontend
```

Expected: `Container yuanju_frontend Started`。

- [ ] **Step 2: 浏览器逐项核对**

打开 http://localhost:3000/settings/brand（需登录）。依次核对：

1. **整体氛围**：页面背景与卡片均为深色（`#0d0f14` / `#1a1f2e`），与 `/profile`、`/history` 视觉一致；不应再有奶油色浮岛感
2. **顶部"返回 ← / 导出品牌设置"标题区**：返回按钮 hover 边框转金色；标题文字 off-white
3. **3 个 section 卡片**（顶部品牌 / 底部品牌 / 水印）：bg = `#1a1f2e`，边框 = 半透明白
4. **输入框 focus**：边框转金色 `rgba(201,168,76,0.3)`；输入字色 off-white；占位字色灰
5. **Logo 上传按钮**：使用全局 ghost-sm 风格，hover 与其它 ghost 按钮一致
6. **Logo 预览框**：80×80 虚线边框、深色 elevated 底色，未上传时灰色"未上传"提示
7. **水印模式三个 radio**：label 文字 off-white；原生 radio 控件可见
8. **中央"预览" section**：BrandPreviewCard **仍是奶油色样张**（这是 *特意* 的，不要误改）
9. **"重置默认 / 保存" 底部按钮**：分别用 ghost / primary，保存可用时是金色渐变
10. **3 个状态条**：
    - 错误（红色）→ 暗红半透明底 + 浅红字，深色下不刺眼
    - 成功（绿色，保存 / 上传后短暂出现 2.5s 自动淡出）→ 暗绿半透明底 + 浅绿字
    - 未保存（金色提示）→ 金色字 + 极淡金色底
11. **裁剪 Modal**（点上传 → 选图）：
    - 蒙层中性 `rgba(0,0,0,0.6)`，不再有暖色调
    - 弹窗面板深色卡 + 浅描边
    - canvas 区域 `#0d0f14` 底
    - 缩放滑块 label 浅灰
    - "取消"按钮 ghost、"确认裁剪"按钮金色 primary（确认时变金色渐变）

12. **回归冒烟**（确认功能未被破坏）：
    - 改一个字段 → "有未保存的修改" 金色条出现
    - 点保存 → 绿色"已保存"条出现 2.5s 后淡出 → 金色条消失
    - 上传一张图 → 裁剪 Modal 出现 → 拖动 + 缩放 → 点确认 → 预览框出现新 logo
    - 点删除 → 弹原生 confirm → 确认后 logo 消失，绿色"Logo 已删除"条出现
    - 点重置默认 → 弹原生 confirm → 确认后所有字段清空且 logo 删除

任何一项不符，回头检查对应任务的 commit。

---

## Self-Review Notes

**Spec coverage 核对：**

| Spec §  | 章节标题 | 对应任务 |
|---------|---------|----------|
| §1      | 背景与问题 | 任务 0 baseline 已隐式确认现状 |
| §2      | 目标 | 任务 2 + 4 实现 |
| §3      | 文件改动清单 | 任务 2-5 一一对应 |
| §4      | Token 映射规则 | 任务 2（页面）+ 任务 4（弹窗）整体重写时实现 |
| §5      | 行为不变量 | 任务 3 + 5（只追加 className，不改 props/逻辑）+ 任务 6 跑既有测试 |
| §6      | 测试策略 | 任务 1（新回归守卫）+ 任务 6（既有测试 + lint + build） |
| §7      | 风险与回滚 | 任务 6 跑既有测试守住行为；每任务一个 commit 便于精准 revert |
| §8      | 不在范围 | 任务清单不含 BrandPreviewCard / backend / 其他页面 |

**Placeholder 扫描：** 已扫描全文，无 TBD / TODO / "fill in details" / "similar to Task N"。所有任务都附完整代码 / 命令 / 期望输出。

**Type / 命名一致性：**
- `className="btn btn-ghost btn-sm"`（BrandSettingsPage logo 按钮 — 任务 3 与测试 — 任务 1） ✓
- `className="btn btn-ghost"`（LogoCropModal 取消 — 任务 5 与测试 — 任务 1） ✓
- `className="btn btn-primary"`（LogoCropModal 确认 — 任务 5 与测试 — 任务 1） ✓
- token 名称（`--bg-card` / `--bg-elevated` / `--text-primary` 等）— 任务 0 step 2 已 sanity-check 它们存在 ✓
- 测试文件路径 `frontend/tests/brand-settings-dark-theme.test.mjs` — 在任务 1 / 6 中一致 ✓
