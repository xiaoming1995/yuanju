## Implementation verification

- 后端定向测试覆盖：北京时间当天由服务端确定、绑定用户与未绑定用户隔离、客户端伪造日期/命盘字段不影响结果、模拟 Provider 的完整接口调用、AI 证据 ID 校验、混合信号、数字概率与录用保证拦截、无个人依据返回「依据不足」、立春/流月/大运交界及通用黄历来源隔离。
- 小程序测试覆盖：首页卡片位于日期与通用宜忌之间、当日提问只发送事项与问题、绑定变化和跨日旧响应不展示、未绑定不调用问事接口、长问题拒绝。
- `node --test miniapp/tools/*.test.js`：133 项通过，包含今日问事入口、绑定变化、跨日、长问题和失败重试。
- `go test ./internal/service ./internal/handler ./pkg/prompt ./cmd/api -run 'TestTodayQuestion|TestSyncCanonical' -count=1`：通过；模拟 LLM Provider 使用隔离的 PostgreSQL 测试容器。
- `openspec validate miniapp-today-question --strict` 与 `git diff --check`：通过。

## Runtime limits

- 本次联调使用测试容器内的本人命盘和模拟 LLM Provider，验证了真实日期接口路径与响应契约；未调用外部真实 LLM，也未产生外部模型费用。
- 当前运行中的 Docker 后端镜像创建于本次修改之前，未重建或部署；小程序尚未在微信开发者工具模拟器或真机做视觉验收。首页新增卡片及提问页已按现有样式实现，实际设备排版仍需查看。
- 工作区原有首页鼓励语、相关测试及归档文件的未提交改动均保留，未清理或提交。
