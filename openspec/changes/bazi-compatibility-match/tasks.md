## 1. 数据模型与迁移

- [x] 1.1 在 `backend/pkg/database/database.go` 增加合盘相关表的增量迁移：`compatibility_readings`、`compatibility_participants`、`compatibility_evidences`、`ai_compatibility_reports`
- [x] 1.2 为合盘表补充必要索引与约束（`user_id`、`reading_id`、`role`、时间倒序查询）
- [x] 1.3 在 `ai_prompts` 初始化逻辑中新增 `compatibility` 默认 Prompt 模块
- [x] 1.4 在 `backend/internal/model/` 中定义合盘阅读结果、参与者、证据、AI 报告的结构体

## 2. 后端基础存储层

- [x] 2.1 新建合盘 repository 文件，支持创建 reading、写入 participants、写入 evidences、读取 detail、读取 history
- [x] 2.2 在 repository 中实现“按 reading_id 读取最新 AI compatibility report”
- [x] 2.3 在 repository 中实现“按 user_id 分页读取 compatibility history”
- [x] 2.4 在 repository 中实现 owner 权限校验所需的 detail 查询接口

## 3. 双盘分析引擎

- [x] 3.1 在 `backend/pkg/bazi/` 新建 compatibility signal engine 入口，接收双方 `BaziResult`
- [x] 3.2 实现四维分数框架：`attraction`、`stability`、`communication`、`practicality`
- [x] 3.3 实现第一批证据来源：日主关系、五行互补/偏枯、日支（夫妻宫）合冲刑害
- [x] 3.4 实现第一批证据来源：财星/官星等配偶相关互动、干支合冲刑害、关键神煞辅助
- [x] 3.5 实现 evidence 归一化结构（dimension/type/polarity/source/title/detail/weight）
- [x] 3.6 实现总体等级 `overall_level` 的聚合规则（`high` / `medium` / `low`）
- [x] 3.7 为合盘分析结果补充摘要标签或短结论字段，便于历史列表展示

## 4. 合盘服务与 AI 报告链路

- [x] 4.1 在 service 层实现“根据双方出生信息生成两份 chart snapshot”的合盘创建流程
- [x] 4.2 在 service 层实现 compatibility reading 的保存与 detail 聚合返回
- [x] 4.3 在 service 层实现 compatibility AI 报告生成，输入为结构化分数与 evidences，而非原始表单
- [x] 4.4 为 compatibility 报告定义 `content_structured` 结构（summary / dimensions / risks / advice）
- [x] 4.5 实现 compatibility report 缓存复用逻辑，避免重复生成多条同 reading 报告

## 5. API 与鉴权

- [x] 5.1 在 `backend/internal/handler/` 新增 compatibility handler
- [x] 5.2 注册创建合盘、获取合盘详情、生成合盘报告、获取合盘历史等路由
- [x] 5.3 所有合盘接口接入用户鉴权，并校验 reading owner
- [x] 5.4 设计并返回前端可直接消费的响应结构，包含 participants、scores、evidences、latest report

## 6. 前端页面与交互

- [x] 6.1 在 `frontend/src/lib/api.ts` 增加 compatibility API 封装
- [x] 6.2 新增合盘输入页，支持填写双方出生信息并提交创建 reading
- [x] 6.3 新增合盘结果页，展示总评、四维结果、关键证据与 AI 解读
- [x] 6.4 新增合盘历史页与详情跳转入口
- [x] 6.5 将合盘历史与普通命盘历史在导航和 UI 上明确区分，避免用户误解

## 7. 测试与验收

- [x] 7.1 为 compatibility signal engine 编写单元测试，覆盖强匹配、强冲突、正负证据并存等场景
- [ ] 7.2 为 repository / service / handler 编写后端测试，覆盖创建、详情、历史、越权访问
- [ ] 7.3 为前端核心页面补充交互测试，覆盖创建成功、历史进入详情、无报告/有报告状态
- [ ] 7.4 手工验收：同一用户创建多条合盘记录，确认对象 B 不会出现在普通 `/history`
- [ ] 7.5 手工验收：生成 compatibility 报告后再次打开详情，确认命中缓存且展示一致

## 8. 文档与收尾

- [ ] 8.1 在项目说明或 AGENTS/CLAUDE 级别文档中补充“合盘能力”概述与边界
- [ ] 8.2 与运营/提示词维护方确认默认 compatibility Prompt 文案
- [ ] 8.3 完成后使用 `/opsx:apply bazi-compatibility-match` 或进入实现阶段执行该 change
