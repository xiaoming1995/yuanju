# 实现任务清单 — shensha-annotations

## Phase 1（当前实现）

### 后端

- [x] **DB Migration**：创建 `shensha_annotations` 表（含 id/name/polarity/description/updated_at）
- [x] **Seed SQL**：预置 40 个神煞的默认文案（命理书级别解说）
- [x] **model/model.go**：添加 `ShenshaAnnotation` struct（含 JSON tags）
- [x] **repository/shensha_repo.go** [NEW]：`GetAll()` + `UpdateByName()` 方法
- [x] **handler/shensha_handler.go** [NEW]：
  - `GET /api/shensha/annotations` — 公开，返回全列表
  - `PUT /api/admin/shensha-annotations/:name` — Admin JWT 鉴权
- [x] **main.go**：注册以上两个路由

### 前端

- [x] **src/lib/api.ts**：添加 `ShenshaAnnotation` 类型 + `fetchShenshaAnnotations()` 函数
- [x] **ResultPage.tsx**：
  - 页面加载时调用 `fetchShenshaAnnotations()`，结果存入 state Map
  - 神煞标签 `<span>` 添加 `onClick` 事件
  - 实现浮层卡片组件（含背景蒙层、标题、描述内容、关闭按钮）
- [x] **ResultPage.css**：浮层卡片样式（居中、圆角、蒙层、极性色块、动画）

---

## Phase 2（后续实现，已记录）

- [x] **Admin UI**：管理后台「神煞注解」列表页，支持点击行内编辑
- [x] **分类字段**：启用 `category` 字段，前端卡片显示分类标签（贵人系/桃花系/凶煞系）
- [x] **桌面 hover 模式**：hover 300ms delay 显示预览，点击固定浮层不消失
- [x] **short_desc 字段**：一句话简介，显示在浮层卡片标题下方

---

## 验收标准

- [x] 所有已实现神煞均有注解文案（不为空）
- [x] 点击任意神煞标签 → 浮层卡片正确显示对应注解
- [x] 点击蒙层/关闭按钮 → 浮层正确关闭
- [x] 移动端点击功能正常
- [x] Admin PUT API 可正常更新文案
- [x] 更新后前端刷新页面能看到最新文案
