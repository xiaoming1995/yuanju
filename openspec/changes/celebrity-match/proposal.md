## Why

在目前的八字算命产品中，普通的命理分析由于过于专业/抽象，普通用户缺乏代入感。引入“名人八字相似度匹配”可以给用户一个直观、有趣的参照系统，同时大幅提升测算结果的社交分享属性与趣味性。

## What Changes

1.  **新建精选名人库**：在后台及数据库中建立并管理一套经过验证的、具有代表性的历史/当代名人八字库（包含其姓名、核心特征）。
2.  **融合 AI 报告生成流程**：不再单独新建页面，而是在现有的 AI 报告请求中，将名人库作为上下文一并传入，请求大模型在常规的（性格、感情、事业、健康）章节之外，新增第五章节：“命理相似名人”，从而让 AI 发挥类比与解释的能力。
3.  **后台管理**：在 Admin 后台增加相应的增删改查前端页面和后端接口，方便人工维护名人素材。

## Capabilities

### New Capabilities
- `celebrity-directory`: 名人库的增删改查后台管理与数据库存储模块。

### Modified Capabilities
- `bazi-report`: 原有的 AI 报告生成能力（注入名人上下文，输出名人匹配解读）。

## Impact

- 数据库需新增表：`celebrities` (或 `celebrity_charts`)。
- Admin API / 路由：需扩展 `/api/admin/celebrities` 相关的增删改查。
- Backend 服务层：`report.go` / `prompt.go` 中，在组装 prompt 前需先加载数据库内的名人并合并到上下文。
- Frontend Admin：需新增 `AdminCelebritiesPage` 及路由。
- 前端普通用户页面几乎零改动，只需要报告本身正常渲染 Markdown 即可自适应。
