## 1. 后端：流月计算引擎

- [x] 1.1 在 `pkg/bazi/engine.go` 新增 `LiuYueItem` 结构体（含 index、month_name、gan_zhi、gan_shishen、zhi_shishen、jie_qi_name、start_date、end_date）
- [x] 1.2 在 `pkg/bazi/engine.go` 新增 `CalcLiuYue(year int, dayGan string) ([]LiuYueItem, int, error)` 函数：调用 lunar-go 的 `GetLiuYue()` 获取 12 个流月干支，通过 `GetJieQiTable()` 查表得到节气起止日期，并计算当前流月 index
- [x] 1.3 实现丑月（index=11）跨年结束日期逻辑：取次年立春日期 - 1 天
- [x] 1.4 为 `CalcLiuYue` 编写单元测试：覆盖正常年份、当前流月 index 准确性、丑月跨年日期三个 case

## 2. 后端：API 接口

- [x] 2.1 在 `internal/handler/bazi_handler.go` 新增 `HandleLiuYue` 处理函数：解析请求体 `{ liu_nian_year, day_gan }`，调用 `CalcLiuYue`，返回 `{ liu_yue: [...], current_month_index: N }`
- [x] 2.2 在 `cmd/api/main.go`（或路由注册处）添加路由 `POST /api/bazi/liu-yue`，绑定 `HandleLiuYue`，无需鉴权中间件
- [x] 2.3 验证参数校验：`liu_nian_year` 为合法年份（1900-2200），`day_gan` 为 10 个天干之一，否则返回 HTTP 400

## 3. 前端：流月抽屉组件

- [x] 3.1 新建 `frontend/src/components/LiuYueDrawer.tsx`：实现右侧抽屉（含遮罩层、关闭按钮、滑入动画），移动端（宽度 < 768px）改为底部全屏 sheet
- [x] 3.2 实现抽屉顶部年份切换器（← 年份 →），默认年份为点击的流年公历年，切换时重新请求 `/api/bazi/liu-yue`
- [x] 3.3 实现 12 个流月格网格布局（4列×3行 或 3列×4行），每格展示节气名、起止日期、干支（天干按五行配色）、天干十神、地支十神
- [x] 3.4 实现当前流月高亮逻辑：基于后端返回的 `current_month_index` 给对应格加金色边框 + "当前"徽章
- [x] 3.5 实现加载骨架屏（12 格占位）和请求失败错误提示（含重试按钮）
- [x] 3.6 在 `frontend/src/lib/api.ts` 新增 `fetchLiuYue(year: number, dayGan: string)` 请求函数

## 4. 前端：大运时间线集成

- [x] 4.1 修改 `DayunTimeline.tsx`：流年格接收 `dayGan`（日主天干）prop，点击流年格时打开 `LiuYueDrawer`，并传递流年年份和日主天干
- [x] 4.2 在 `LiuYueDrawer` 中维护抽屉开关状态，确保同时只有一个流年的抽屉打开

## 5. 验证

- [x] 5.1 本地启动后端，请求 `POST /api/bazi/liu-yue` 验证：2026 年甲日主返回正确干支（寅月=庚寅）、清明节气日期、当前 index=2 ✓
- [x] 5.2 前端本地开发环境中，点击流年格确认抽屉弹出、12 个流月格数据正确、当前流月金色高亮 ✓
- [x] 5.3 切换年份后确认数据刷新、干支计算正确 ✓（切换至2025年，干支重新计算）
- [x] 5.4 移动端（375px 宽）验证抽屉改为底部 sheet 展示（CSS 媒体查询已配置）
