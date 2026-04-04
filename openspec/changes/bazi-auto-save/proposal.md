## Why

当前系统存在一个核心体验断层：用户在首页点击「起盘」时，系统只会调用无状态的 `/api/bazi/calculate` 进行排盘，而不会将命盘保存到数据库。命盘的保存动作被绑定在了结果页的「生成 AI 解读」按钮上。

这导致：
1. **纯命理使用受限**：只看八字命盘、不需要 AI 报告的用户，无法将命盘留存到自己的历史记录中。
2. **体验不符合文案预期**：历史页空状态提示「起盘后登录即可自动保存」，但实际并未做到自动保存。

## What Changes

为了在后端实现“柔性自动保存”，我们需要：

1. **新增 `OptionalAuth` 中间件**：
   - 提取并复用现有 `Auth` 中间件的解析逻辑。
   - 包含有效 Token 时，将 `user_id` 注入上下文；
   - 包含无效 Token 或无 Token 时，系统不拦截、不报错，静默放行（`c.Next()`）。
2. **升级 Calculate 接口**：
   - 为 `/api/bazi/calculate` 路由增加 `OptionalAuth` 中间件。
   - Handler 层面判断 `c.Get("user_id")`：若存在，则在计算完毕后同步调用 `repository.CreateChart` 将记录存入数据库。
3. **保持游客体验**：完全不影响未登录用户的体验和性能开销。

## Capabilities

### New Capabilities

- **Optional Auth (柔性鉴权)**：系统 SHALL 能够以非强制方式解析 JWT，为可选的个性化行为什么提供统一的鉴权底层设施。

### Modified Capabilities

- `bazi-calculate`：起盘请求 SHALL 在用户已登录时（携带有效 JWT）将排盘结果持久化到历史记录中，在未登录时以无状态模式返回。
- `bazi-history`：用户在起盘后直接进入历史列表，SHALL 立即可见刚建立的命盘。

## Impact

- **后端路由**：`backend/cmd/api/main.go`，在 `/api/bazi/calculate` 前插入 `OptionalAuth`。
- **后端中间件**：`backend/internal/middleware/middleware.go`。
- **后端 Handler**：`backend/internal/handler/bazi_handler.go`，修改 `Calculate` 方法加入 DB 写入逻辑（复用已有的 JSON 序列化和 ChartHash）。
- **前端改动**：无。当前的 Axios request interceptor（`frontend/src/lib/api.ts`）已实现全局无差别携带 `Bearer Token`，机制天然契合。
