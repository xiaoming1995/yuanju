## 验证结果

- `go test ./internal/service ./internal/handler -count=1`：通过。
- `node --test miniapp/tools/*.test.js`：87 项通过。新增验证覆盖真实日期超出演示年份、事项来源、个人方向、快速切日旧响应隔离和详情数据一致性。
- 后端受控用例覆盖 2026-11-02 的庚辰日、一项中性/双中性/双喜用/双忌神/方向相反/调候冲突、出行通用宜忌、日支相冲、依据不足及财务文案边界。未读取用户私有命盘来猜测截图中的个人喜忌。
- `openspec validate clarify-miniapp-personal-day-direction --strict` 与 `git diff --check`：通过。
- 微信开发者工具 CLI 成功打开 `miniapp/` 项目；桌面截图命令返回 `could not create image from display`。本轮未取得模拟器或真机画面，窄屏下四个事项的实际换行和可见区域仍需在可见开发者工具或真机上确认。

## 传统日历式文案补充验证

- `go test ./internal/service -run '^TestPersonalDay' -count=1`：通过；覆盖截图日期的“宜／平／未列”、通用出行“宜／忌”与相冲“慎”。
- `node --test miniapp/tools/*.test.js`：87 项通过；详情依据展开状态通过。
- `openspec validate clarify-miniapp-personal-day-direction --strict` 与 `git diff --check`：通过。
- 已重建并重启本机 Docker 后端；运行中的 `9002` 服务镜像含新文案并监听请求。未借用用户登录态验证其个人响应，也未取得模拟器画面。

## 完整通用黄历宜忌补充验证

- `go test ./internal/service ./internal/handler -count=1`：通过；公开日期接口的最终字段另以 `go test ./internal/handler -run '^TestCalendarDayPublicYiJi$' -count=1` 验证通过。
- `node --test miniapp/tools/*.test.js`：89 项通过，覆盖未登录读取、完整宜忌、快速切日旧响应隔离和本人事项筛选。
- 运行中的本机 `9002` 公开接口返回 2026-11-02 的完整列表：宜“祭祀、沐浴、解除、破屋、坏垣、馀事勿取”，忌“开市、嫁娶”；无效日期返回 HTTP 400。
- `openspec validate clarify-miniapp-personal-day-direction --strict`、`git diff --check` 和小程序页面 JS 语法检查通过；微信开发者工具项目已重新打开。未取得模拟器或真机截图，实际视觉排版尚未人工确认。

## 单项中性状态文案补充验证

- `go test ./internal/service -run '^TestPersonalDay' -count=1`：通过；覆盖日干／日支单项喜用、单项忌神、双中性与原有双向、冲突分支。
- `node --test miniapp/tools/personal_chart_live.test.js`：4 项通过；详情使用新标签。OpenSpec strict 校验和 `git diff --check` 通过。
- 本机 Docker 后端已重建并重启，运行程序包含“日干偏合”“均无明显喜忌”，`/health` 返回正常。未使用用户登录态读取其个人结果；模拟器画面仍未取得。
