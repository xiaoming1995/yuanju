# 神煞注解系统 — 技术设计

## 数据库

### 新表：shensha_annotations

```sql
CREATE TABLE shensha_annotations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(20)  NOT NULL UNIQUE,  -- 神煞名，如"天乙贵人"
    polarity   VARCHAR(10)  NOT NULL DEFAULT 'zhong', -- ji / xiong / zhong
    category   VARCHAR(20),                  -- Phase 2: 贵人系/桃花系等
    short_desc VARCHAR(50),                  -- Phase 2: 一句话摘要
    description TEXT         NOT NULL,       -- 详细说明（100-200字）
    updated_at TIMESTAMPTZ  DEFAULT now()
);
```

**预置数据**：在数据库迁移中使用 INSERT 预置所有已知神煞文案（约 40+ 条）。

## 后端 API

### 公开接口（无需鉴权）

```
GET /api/shensha/annotations
```
- 返回全部神煞注解列表
- 无分页（数据量小，一次返回全部）
- 响应格式：

```json
{
  "data": [
    {
      "name": "天乙贵人",
      "polarity": "ji",
      "description": "天乙贵人为诸吉神之首..."
    }
  ]
}
```

### Admin 接口（需 Admin JWT）

```
PUT /api/admin/shensha-annotations/:name
Body: { "description": "..." }
```
- 更新指定神煞的描述文案
- Phase 1 仅提供 API，无 Admin UI

## 文件改动

### 后端
- `internal/model/model.go` — 新增 `ShenshaAnnotation` 结构体
- `internal/repository/shensha_repo.go` — [NEW] 数据库 CRUD
- `internal/handler/shensha_handler.go` — [NEW] GET + Admin PUT
- `cmd/api/main.go` — 注册新路由
- `db/migrations/` — [NEW] 建表 + 预置 seed SQL

### 前端
- `src/lib/api.ts` — 新增 `fetchShenshaAnnotations()` 函数
- `src/pages/ResultPage.tsx` — 神煞标签加点击事件 + 浮层卡片组件
- `src/pages/ResultPage.css` — 浮层卡片样式

## 浮层卡片 UI 设计

```
点击神煞标签后，全局显示浮层（居中）：

┌──────────────────────────────────────┐
│  [极性色块] 天乙贵人           [× ]  │
├──────────────────────────────────────┤
│                                      │
│  天乙贵人为诸吉神之首，主逢凶化吉、  │
│  贵人相助。命带此星之人，往往能在关  │
│  键时刻得贵人提携...                 │
│                                      │
└──────────────────────────────────────┘

- 背景蒙层，点击蒙层关闭
- 卡片最大宽度 400px
- 标题行：极性颜色圆点 + 神煞名 + 关闭按钮
- 内容区：description 文本，行高 1.7
```

## 前端数据策略

- 进入结果页时**一次性预加载**所有注解
- 存入 React state/Map，key 为 name
- 点击标签时 O(1) 查找，零延迟显示

## Phase 2 记录（待实现）

1. **Admin UI**：在管理后台添加「神煞注解」列表页，支持在线文案编辑
2. **分类系统**：在表中启用 `category` 字段，前端卡片展示分类标签
3. **桌面 hover 模式**：hover 300ms delay 显示预览，点击固定浮层
4. **short_desc 摘要**：在浮层标题旁展示一句话简介
