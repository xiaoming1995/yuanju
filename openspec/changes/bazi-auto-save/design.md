## Context

当前系统仅在用户点击「生成 AI 解读」时才会持久化命盘记录（`BaziChart`）。直接通过首页「起盘」排盘时（`/api/bazi/calculate`），请求虽计算完详尽的八字结构图，但不进行数据库落盘。

为满足用户“登录即可自动保存排盘以便日后查看”的诉求，必须消除无状态 Calculate 和重度有状态 Report 之间的落差，同时保留游客随时可零摩擦使用 Calculate 的初衷。

由于前端现有框架在起盘发出的 XHR 中已经内置了 JWT（如果存在），后端需从鉴权边界的改造切入。

## Goals / Non-Goals

**Goals:**
- 已登录用户只需执行排盘动作，命盘即安全、透明地落入数据库，后续立即可查。
- 未登录游客执行排盘，体验应当纯粹无感、性能如初，不受额外持久化逻辑拖累。
- 采用复用度高、不侵入核心鉴权逻辑的架构解法。

**Non-Goals:**
- 将历史详情页的 AI 报告入口与纯命盘数据深度解耦（现有 ResultPage 前端已有优雅的回落降级处理，不存报告即可工作，无需额外设计）。
- 为了本次需求强推完全客户端侧的离线命盘存储支持（仍依赖云端隔离）。

## Decisions

### D1：基于 Gin 扩展 `OptionalAuth` 中间件

在 `middleware.go` 中，新增 `OptionalAuth()`：
- 提取并复用现有 `Auth()` 中关于解析字符串、校验及提取 `claims["user_id"]` 的逻辑。
- 当 `AuthHeader` 不合规、为空、过期或签名异常时，均仅做静默处理，直通 `c.Next()` 而不阻断请求。

由于核心校验体量极小且已有可靠的 `github.com/golang-jwt/jwt/v5`，冗余成本极低且安全可控。

### D2：`Calculate()` Handler 支持有状态分叉

在 `/api/bazi/calculate` 中，排盘库 `bazi.Calculate()` 计算得出 `BaziResult` 之后：
- 通过 `userID, exists := c.Get("user_id")` 获取可能注入的上下文。
- 当 `exists` 为 `true` 时，用现有结构体（结合 `result`）实例化 `model.BaziChart` 并发出 `repository.CreateChart` 插入操作。

得益于已修复的 `(chart_hash, user_id)` 复合约束及合理容错，相同的盘在同用户的二次点击下，UPSERT 可平滑无损去重，不会带来无脑冗余堆积隐患。

### D3：路由分组架构的剥离组合

在 `main.go` 针对路由：
```go
// ... 游客级别路由 ...
api.POST("/bazi/calculate", middleware.OptionalAuth(), handler.Calculate)
```
使得在 API 链路上该节点具有独占的兼顾二者的弹性和语义。

## Risks / Trade-offs

- **[无感写入延迟]**：Calculate 原本是不接触 DB 的纯内存 O(1) 短期运算。挂载 OptionalAuth 和异步落库可能拖慢约 5ms~20ms 响应响应时间。当前作为 MVP 可直接同步调用落库；后续若并发流量拉起，只需将其放入 `goroutine` 即可完全消解感知。
- **[Token 脏读漏洞]**：若客户端上送了恶意拼装且失效的过期 token，OptionalAuth 会忽略它导致 fallthrough 至游客模式处理不落地。该行为恰好匹配预期，无越权风险。
