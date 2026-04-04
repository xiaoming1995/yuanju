## Why

命理市场需求旺盛，但现有产品普遍存在解读晦涩、体验割裂的问题——专业用户嫌不够深度，普通用户看不懂术语。缘聚通过"算法精算 + AI 自然语言解读"的双层模式，同时满足两类用户，MVP 阶段聚焦八字，免费获客，验证核心体验。

## What Changes

- 新建 Go 后端服务，提供用户认证、八字计算与 AI 报告生成 API
- 新建 React 前端应用，实现命盘展示与 AI 解读的双层视图
- 集成大模型 API（DeepSeek / OpenAI）生成自然语言八字报告
- 实现阳历 → 农历 → 四柱天干地支的完整计算链路
- 支持用户历史记录保存与查看

## Capabilities

### New Capabilities

- `user-auth`: 用户注册、登录、JWT 鉴权、个人信息管理
- `bazi-engine`: 阳历转换四柱（年柱/月柱/日柱/时柱）、五行分布计算、用神喜忌推算
- `bazi-chart`: 八字命盘可视化展示，包含天干地支、五行雷达图、大运时间轴
- `ai-report`: 调用大模型生成通俗易懂的自然语言八字解读报告
- `history`: 用户历史命盘记录保存与查看

### Modified Capabilities

（无，MVP 为全新项目）

## Impact

- **后端**：全新 Go 项目，需搭建基础框架（路由/中间件/数据库/AI 接入）
- **前端**：全新 React 项目，需建立设计系统与组件库
- **数据库**：PostgreSQL，需设计 users、bazi_charts、ai_reports 等核心表
- **外部依赖**：大模型 API（DeepSeek 优先，OpenAI 备选）、Redis（会话缓存）
- **部署**：Docker 容器化，支持本地开发与生产部署
