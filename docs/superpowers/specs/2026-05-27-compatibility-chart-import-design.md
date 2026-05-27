# 合盘从单盘导入与命盘称呼设计

## 背景

当前合盘页 `/compatibility` 需要用户分别填写“我的生辰”和“对方生辰”。用户如果已经在单盘测算里保存过命盘，仍要重复录入年月日时、性别、历法和闰月信息，输入摩擦较高。

现有系统已经具备可复用基础：

- 单盘历史接口保存并返回 `birth_year`、`birth_month`、`birth_day`、`birth_hour`、`gender`、`calendar_type`、`is_leap_month`。
- 合盘页使用同一个 `BirthProfileForm` 出生信息表单组件。
- 合盘创建接口已经支持手填双方出生资料，后端创建 `compatibility_participants` 并保存 `display_name`，但目前固定为“我”和“对方”。

本次优化目标是让用户可以从已保存的单盘命盘导入到合盘，并补齐命盘档案称呼，方便识别“我”“小王”“1998 女生”等常见关系对象。

## 目标

- 合盘页支持显式导入最近命盘或从命盘档案选择，避免重复填写。
- 命盘历史列表和命盘详情页支持编辑命盘称呼。
- 从历史列表或命盘详情页可直接发起合盘，并选择该命盘作为“我”或“对方”。
- 合盘结果中的参与人名称优先使用导入命盘的称呼。
- 保留手动填写能力，不改变合盘算法、评分、AI prompt 和报告结构。

## 非目标

- 不做命盘搜索、筛选、分页或批量管理。
- 不建立长期关系档案、关系追踪或对象管理系统。
- 不让合盘创建接口只依赖 `chart_id`；本期仍提交出生资料，保持手填和导入路径一致。
- 不改命盘 hash、AI 报告缓存、八字算法计算逻辑。

## 方案选择

采用“后端补命盘称呼，前端做导入”的方案。

备选方案包括：

1. 纯前端导入：最快，但命盘称呼无法持久化，历史、详情、合盘结果名称仍会割裂。
2. 后端补命盘称呼，前端把命盘转为合盘表单：改动克制，能满足当前导入和称呼需求。
3. 完整命盘档案引用模型：合盘创建请求传 `self_chart_id` / `partner_chart_id`，数据血缘清楚，但会扩大接口和服务层改动。

推荐并采用第 2 种。本次需求核心是降低填写摩擦，不是重做合盘数据模型。后续如果要做长期关系档案，可再升级到第 3 种。

## 后端设计

### 命盘称呼字段

给 `bazi_charts` 增加可选字段：

```sql
display_name TEXT
```

字段语义：

- 用户给命盘档案的称呼或备注。
- 可为空。
- 不参与命盘 hash、八字算法、AI 报告缓存和任何命理计算。
- 历史列表、详情页、个人中心最近命盘都应返回该字段。

校验规则：

- 更新时 trim 首尾空白。
- 空字符串允许保存，表示清空称呼。
- 非空时限制在 20 个 Unicode 字符以内。
- 只有命盘所属用户可修改。

### 称呼更新接口

新增接口：

```http
PATCH /api/bazi/history/:id/display-name
Content-Type: application/json

{ "display_name": "小王" }
```

成功响应：

```json
{
  "data": {
    "id": "chart-id",
    "display_name": "小王"
  }
}
```

错误语义：

- 401：未登录。
- 403：命盘不属于当前用户。
- 404：命盘不存在。
- 400：称呼超长或请求体格式错误。

### 历史接口扩展

以下响应增加 `display_name`：

- `GET /api/bazi/history`
- `GET /api/bazi/history/:id`
- `GET /api/user/profile` 中的 `recent_charts`

保持向后兼容：旧数据没有称呼时返回空字符串或省略都可以，但前端统一按空值处理。

### 合盘创建名称扩展

现有合盘创建请求保留 `self`、`partner`、`relationship_stage`、`primary_question`。新增两个可选字段：

```json
{
  "self_display_name": "我",
  "partner_display_name": "小王"
}
```

后端创建 `compatibility_participants` 时：

- `self_display_name` 非空则作为 `self` 参与人名称，否则使用“我”。
- `partner_display_name` 非空则作为 `partner` 参与人名称，否则使用“对方”。
- 名称同样 trim，非空限制 20 个 Unicode 字符以内。
- 不要求 display name 与命盘称呼强绑定；前端导入时传入即可。

## 前端设计

### 通用转换 helper

新增或复用一个前端 helper：

```ts
chartToBirthProfile(chart): BirthProfileFormValue
```

它把历史命盘字段转换为 `BirthProfileForm` 使用的表单值：

- `birth_year` -> `year`
- `birth_month` -> `month`
- `birth_day` -> `day`
- `birth_hour` -> `hour`
- `gender` -> `gender`
- `calendar_type || 'solar'` -> `calendarType`
- `is_leap_month || false` -> `isLeapMonth`

合盘页维护两类状态：

- `selfProfile` / `partnerProfile`：实际提交的出生资料。
- `selfImportSource` / `partnerImportSource`：导入来源信息，如 `chartId`、`displayName`、原始 profile。

当用户导入后再手动改动表单，如果当前表单与原始 profile 不一致，来源提示展示为“已基于小王修改”。

### 合盘页导入入口

`/compatibility` 不自动填最近命盘，避免静默导错人。两个出生资料面板顶部各增加导入工具条：

- “导入最近命盘”
- “从命盘档案选择”

导入后显示来源提示：

```text
已导入：小王 · 女命 · 1998年3月12日 午时
```

若用户手动修改：

```text
已基于小王修改
```

未登录用户点击导入动作时跳转登录。

“从命盘档案选择”打开轻量弹层，列出最近 20 条命盘：

- 称呼，空值时回退生日/性别。
- 性别。
- 生日与时辰。
- 四柱。
- 保存日期。

本期不做搜索和分页。历史为空时显示空状态，引导先起盘。

### 从单盘发起合盘

命盘历史 `/history`：

- 卡片展示称呼；没有称呼时展示现有生日/四柱信息。
- 每张卡片提供“编辑称呼”和“用此命盘合盘”。
- 点击“用此命盘合盘”弹出角色选择：“作为我”或“作为对方”。
- 选择后跳转：

```text
/compatibility?importChart=<chart_id>&role=self
/compatibility?importChart=<chart_id>&role=partner
```

命盘详情 `/history/:id`：

- 在页面上方增加“档案称呼”编辑入口。
- 提供“用此命盘发起合盘”。
- 同样先选择“作为我”或“作为对方”，再跳合盘页。

合盘页读取 query 后：

- 调用命盘详情接口确认权限并获取完整命盘信息。
- 将命盘导入指定角色。
- 如果失败，显示错误提示，不覆盖当前表单。

### 命盘称呼编辑

历史列表支持快速编辑称呼。命盘详情也支持编辑称呼。

交互规则：

- 保存成功后立即更新当前页面展示。
- 保存失败时保留编辑态并显示错误。
- 清空称呼是合法操作。
- 列表编辑只更新对应卡片，不强制刷新整页。

### 合盘提交

提交时仍发送现有出生资料，同时附带可选显示名：

```json
{
  "self": {
    "year": 1990,
    "month": 1,
    "day": 1,
    "hour": 12,
    "gender": "male",
    "calendar_type": "solar",
    "is_leap_month": false
  },
  "partner": {
    "year": 1998,
    "month": 3,
    "day": 12,
    "hour": 12,
    "gender": "female",
    "calendar_type": "solar",
    "is_leap_month": false
  },
  "self_display_name": "我",
  "partner_display_name": "小王",
  "relationship_stage": "ambiguous",
  "primary_question": "continue_investment"
}
```

显示名来源优先级：

1. 导入来源的 `display_name`。
2. 默认“我”或“对方”。

如果用户导入后手动修改出生资料，仍沿用来源称呼，除非称呼为空。

## 错误处理

- 称呼保存失败：保留编辑框和用户输入，显示错误，不更新本地展示。
- 称呼超长：前端即时提示，后端仍做最终校验。
- 导入命盘不存在或无权限：提示“命盘不存在或无权访问”，不覆盖当前表单。
- 历史为空：导入弹层显示空状态，引导“先新建命盘”。
- query 导入失败：停留合盘页，顶部显示错误，用户仍可手填。
- 合盘创建不带 display name：后端回退“我/对方”，兼容旧请求。

## 测试计划

后端测试：

- 迁移包含 `display_name` 字段。
- `PATCH /api/bazi/history/:id/display-name` 支持保存、清空、超长拦截。
- 用户不能修改他人命盘称呼。
- 历史列表、详情、个人中心最近命盘返回 `display_name`。
- 合盘创建请求带 `self_display_name` / `partner_display_name` 时，参与人名称按请求保存。
- 旧合盘创建请求不带名称时仍保存“我/对方”。

前端测试：

- 历史卡展示称呼并支持编辑成功/失败状态。
- 命盘详情展示并支持编辑称呼。
- 历史卡和详情页点击“用此命盘合盘”会先选择角色，再跳转正确 query。
- 合盘页“导入最近命盘”填入对应表单并显示来源提示。
- 合盘页“从命盘档案选择”可把选中命盘导入当前角色。
- query 导入按 `role=self|partner` 填入正确面板。
- 导入后手动修改表单会显示“已基于某称呼修改”。
- 提交合盘时带上可选 display name。

回归验证：

- 单盘起盘仍正常。
- 历史详情仍能查看完整命盘。
- 合盘手填创建仍正常。
- 合盘结果页继续展示参与人名称。
- 前端 lint/build 通过，后端相关测试通过。

## 实施边界

本次实现应局限在：

- 数据库迁移与 `BaziChart` / 用户资料命盘摘要模型字段扩展。
- 命盘称呼更新 API。
- 合盘创建请求 display name 兼容扩展。
- 历史页、结果页、合盘页导入相关 UI。
- 必要的 API 类型和测试。

不做与合盘算法、AI prompt、结果页内容层级无关的重构。
