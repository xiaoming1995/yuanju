## API & Database Design

### 1. Database Migrations (DDL)

- **移除唯一约束 `idx_bazi_charts_hash_user` 和 `bazi_charts_chart_hash_user_id_key`**：
  我们需要通过增量 Migration 方式，切断原先 `UNIQUE (chart_hash, user_id)` 对同参排盘拦截的操作，使每次执行均可无限落定。

### 2. API Schema Changes

- **修改 `POST /api/bazi/calculate` Response**:
  - 返回对象需附加新产生的 `chart_id`: `string`，将创建的命盘主体赋予具象标识。若是未登录游客则该值为空。

- **新增/替换 `POST /api/bazi/report/:chart_id` Request**:
  - **路径变量**：直接读取 `:chart_id` 以定位命盘资源。
  - **废除 Input 提交**：前置不需要承接收集 `ReportInput` 等原始表征表单。
  - **实现过程**：通过给定 `Chart_ID` `repository.GetChartByID(chart_id)`，以此调起内部运算 `Calculate` 再交予大模型推理。
  - **鉴权要求**：需增加比目前强的主属校验：确保调用的操作者具备该命盘数据的读写主权 `if chart.UserID != userId { abort() }`。

### 3. Repository Data Layer

- **`repository.CreateChart`**:
  将原来的 `ON CONFLICT DO UPDATE` 分支彻底洗牌，回归原教旨简单直接的 `INSERT INTO ... RETURNING id`。

- **新增 `repository.GetChartByID`**:
  供新的 Generate Report 等接口通过 `id` 定位使用所需。

### 4. Frontend Component Behavior 

- **Result 绑定 `chart_id` / State 分离联动**:
  当获得底层命局应答并成功注册资源化后（`res.data.chart_id`），若处在未请求 AI 之前，页面中的“生成 AI 报告按钮”只将此 ID 向上穿透回掉发往端倪，不再进行本地冗余 payload 回顾。从而确保页面跳转来源无论经由 /Home 亦或 /History 均能顺利通过 ID 合约完成响应循环。
