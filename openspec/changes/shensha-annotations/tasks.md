# 实现任务清单 — shensha-annotations

## Phase 1（当前实现）

### 后端

- [ ] **DB Migration**：创建 `shensha_annotations` 表（含 id/name/polarity/description/updated_at）
- [ ] **Seed SQL**：预置 40+ 神煞的默认文案（命理书级别解说）
  - 贵人类：天乙贵人、文昌贵人、太极贵人、月德贵人、天德贵人、德秀贵人、天厨贵人、金舆贵人、福星贵人、国印贵人、三奇贵人、日德、将星、十灵日、词馆
  - 吉神类：禄神、天喜、天医、红艳、天赦
  - 格局类：阴差阳错、羊刃、飞刃
  - 驿动类：驿马、桃花、华盖
  - 凶煞类：天罗地网、地网、劫煞、亡神、灾煞、孤辰、寡宿、童子煞、流霞、吊客、墓门、魁罡
  - 其他：太极贵人、十灵日、词馆、墓门
- [ ] **model/model.go**：添加 `ShenshaAnnotation` struct（含 JSON tags）
- [ ] **repository/shensha_repo.go** [NEW]：`GetAll()` + `UpdateByName()` 方法
- [ ] **handler/shensha_handler.go** [NEW]：
  - `GET /api/shensha/annotations` — 公开，返回全列表
  - `PUT /api/admin/shensha-annotations/:name` — Admin JWT 鉴权
- [ ] **main.go**：注册以上两个路由

### 前端

- [ ] **src/lib/api.ts**：添加 `fetchShenshaAnnotations()` 函数
- [ ] **ResultPage.tsx**：
  - 页面加载时调用 `fetchShenshaAnnotations()`，结果存入 state Map
  - 神煞标签 `<span>` 添加 `onClick` 事件
  - 实现浮层卡片组件（含背景蒙层、标题、描述内容、关闭按钮）
- [ ] **ResultPage.css**：浮层卡片样式（居中、圆角、蒙层、极性色块）

---

## Phase 2（后续实现，已记录）

- [ ] **Admin UI**：管理后台「神煞注解」列表页，支持点击行内编辑
- [ ] **分类字段**：启用 `category` 字段，前端卡片显示分类标签（贵人系/桃花系/凶煞系）
- [ ] **桌面 hover 模式**：hover 300ms delay 显示预览，点击固定浮层不消失
- [ ] **short_desc 字段**：一句话简介，显示在浮层卡片标题下方

---

## 验收标准

- [ ] 所有已实现神煞均有注解文案（不为空）
- [ ] 点击任意神煞标签 → 浮层卡片正确显示对应注解
- [ ] 点击蒙层/关闭按钮 → 浮层正确关闭
- [ ] 移动端点击功能正常
- [ ] Admin PUT API 可正常更新文案
- [ ] 更新后前端刷新页面能看到最新文案
