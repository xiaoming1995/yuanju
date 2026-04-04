## 1. 柔性鉴权中间件开发

- [x] 1.1 在 `backend/internal/middleware/middleware.go` 中新增 `OptionalAuth()` 中间件函数。
- [x] 1.2 在 `OptionalAuth()` 内实现逻辑：检查 `Authorization` Bearer 令牌是否存在且前缀正确；如果不满足，直接 `c.Next()` 并 return，不执行 `c.Abort()`。
- [x] 1.3 若格式正确，解析 JWT 并验证（复用目前的 `jwt.Parse` 和 `configs.AppConfig.JWTSecret`）。
- [x] 1.4 如果 JWT 无效或过期，跳过错误仅触发 `c.Next()` 放行。
- [x] 1.5 若解析成功且包含 `claims["user_id"]`，将其注入 Gin Context（`c.Set("user_id", ...)`），然后调用 `c.Next()` 放行。

## 2. Calculate Handler 支持带状态落库

- [x] 2.1 修改 `backend/internal/handler/bazi_handler.go` 中的 `Calculate` 函数。
- [x] 2.2 在原有完成八字计算操作后（得到 `result`），增加尝试从上下文提取 ID 的判断语句：`userID, exists := c.Get("user_id")`。
- [x] 2.3 如果 `exists` 为真，构造 `model.BaziChart` 数据结构并填充所有所需的字段上下文，包括 `Gender`、四柱天干地支、五行等与原先一致参数映射。
- [x] 2.4 调用 `repository.CreateChart(chart)` 同步实现存储命令压库，将计算出的结果（无论全新创建还是更新复用）安全留底。如果执行出错则可静默吞并错误以维护对端纯前端只读状态的顺利响应而不发生致命熔断。

## 3. 路由与功能组装验证

- [x] 3.1 修改 `backend/cmd/api/main.go` 中对 `POST("/api/bazi/calculate")` 的路由挂载，在其链条中前置编排注册 `middleware.OptionalAuth()` 中间件拦截器。
- [ ] 3.2 手动使用游客不带任何身份操作进行起盘。确认返回正常计算数据，验证控制台及数据库未新增表列项产生；确保不受 401 身份拒绝异常骚扰。
- [ ] 3.3 使用登陆态界面起盘。操作响应后点击导航查验"本账户排盘历史"并可见此套入座记录顺利诞生存在。
- [ ] 3.4 登录后连续二次、三次对相同输入参数再次发力计算请求操作，以验证 DB 冲突机制的稳定及避免生成无效无尽重复冗杂账单数据积累；确认列表仍仅有一列最新纪录代表呈现即表明验证彻底结束。
