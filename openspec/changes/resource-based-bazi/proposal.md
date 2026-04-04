## Why

系统目前面临两个与数据留存模型相关的核心痛点：
1. **AI 报告过度横向共享**：目前的逻辑使用 `(chart_hash, user_id)` 对八字信息进行严格去重，相同的排盘在历史列表仅能留存唯独一条。这就导致无论该用户为几位同生辰的不同实体（客户或亲朋）算命，AI 缓存都会死锁互斥，无法基于各次点击生成相互独立或更新版本的报告。
2. **前后端交互断层导致 Bug**：目前从“起盘历史”列表跳往结果卡片页（`/history/:id`）时，若该历史条目先前并未生成 AI 报告，用户点按 **「生成 AI 解读」** 是完全失效不作为的。这是因为目前报告的入参强依赖纯前端的暂存表单输入 `input` 对象，而未真正走向后端状态持久化所代表的资源化存取（即通过已生生成的 `chart_id` 下发工作）。

## What Changes

从“计算器模式”全面翻新演变为“标准 RESTful 资源生命周期模式”：
1. **解除生辰去重限制**：剔除 `bazi_charts` 表中关于 `chart_hash` 与 `user_id` 的强唯一性干预，允许用户针对完全一样的生辰多频率起盘。每一次起盘保存均视作独立的生命快照对象。
2. **后端 Calculate 资源化**：首页起盘接口触发时，一旦带有 `user_id` 将通过 Blind Insert 自动建立独立资源并分配专属 uuid 给到前端状态库（而不仅是计算结果），前端依此 uuid 绑定视图与历史跳转。
3. **AI 触发接口改造为面向资源**：原本利用表单参数投喂进行 GenerateAIReport (POST `/api/bazi/report`) 的途径将重做，直接由前端呼叫新的生成接口 `POST /api/bazi/report/:chart_id` 触发基于某个既定存在的生命快照独立生成报告。通过分离命盘主体和附属的分析能力，化解前述断层 Bug 故障。

## Capabilities

### Modified Capabilities

- **bazi-history**：同生辰反复算将会独立形成多条历史堆栈；缓存锁不再由生辰特征持有，而是针对独一无二的 history 条目（chart_id）发生。
- **ai-generation-fallback**：对于无 AI 报告的历史条目，能够正常激活生成解析调用并获得最终留底反馈。

## Impact

1. **数据库 DDL**：在 Migration 区块彻底删去前置增加的 `UNIQUE (chart_hash, user_id)` 复合约束锁。
2. **Repository**：撤除 `CreateChart` 中因应去重的 `ON CONFLICT DO UPDATE` 功能，变为简单无障碍直接 `INSERT` 堆积写入式。
3. **API 路由重定义**：抛弃 POST `/api/bazi/report` 中的冗余起盘步骤，构建新型接口 `/api/bazi/report/:chart_id` 单体指派报告生产任务。
4. **前端 ResultPage**：消除原有按需提交 Input 回包的数据依赖结构，使用目前所在的 `chart_id` 提交请求。从而同时修复跨页面导航无法追认 AI 点击操作的硬伤流转错乱。
