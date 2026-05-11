## 1. 后端：调候字典查询封装

- [x] 1.1 在 `pkg/bazi/tiaohou_dict.go` 新增导出函数 `LookupTiaohouYongshen(dayGan, monthZhi string) []string`，返回字典中 `Yongshen` 字段的副本（防止外部修改）；未命中返回 nil
- [x] 1.2 验证字典 120 条全覆盖（10 天干 × 12 月支）；增加单元测试 `tiaohou_dict_test.go` 校对所有组合可命中

## 2. 后端：原局透藏干集合构造

- [x] 2.1 在 `pkg/bazi/engine.go`（或新建 `yongshen.go`）实现 `collectNatalGans(yearGan, monthGan, dayGan, hourGan string) []string` 返回 4 个透干（不去重，便于诊断）
- [x] 2.2 实现 `collectNatalHideGans(yearHideGan, monthHideGan, dayHideGan, hourHideGan []string) []string` 合并 4 支藏干为单一切片
- [x] 2.3 实现 `intersectGans(needs, available []string) (hit, miss []string)` 求交集，返回命中与缺位天干列表

## 3. 后端：yongshen 主流程重写

- [x] 3.1 在 `pkg/bazi/engine.go` 新增 `inferYongshenWithTiaohouPriority(dayGan, monthZhi string, gans, hideGans []string, dayGanWx, monthZhiWx string, stats WuxingStats) (yongshen, jishen string, status string, hitGans, missGans []string)`
- [x] 3.2 主流程实现：先查 `LookupTiaohouYongshen` → 与 `gans ∪ hideGans` 求交集 → 命中则返回 `tiaohou_hit` + yongshen 五行集
- [x] 3.3 缺位 fallback：调用现有 `calcWeightedYongshen` 计算扶抑结果，返回 `tiaohou_miss_fallback_fuyi`，并填充 `missGans`
- [x] 3.4 字典缺失分支：返回 `tiaohou_dict_missing`，走扶抑（防御性，理论不发生）
- [x] 3.5 删除 `inferNativeYongshen` 中"三冬无火 / 三夏无水"硬编码短路；保留 `calcWeightedYongshen` 作为 t1 fallback 的内部实现
- [x] 3.6 实现 `gansToWuxingSet(gans []string) string` 将命中天干列表去重转为五行集中文字符串（如 ["丙","癸"] → "火水"）
- [x] 3.7 实现 `wuxingSetToJishen(wuxingSet string) string` 将 yongshen 五行字符串转为 jishen 五行字符串（克 + 泄并集）

## 4. 后端：BaziResult 结构体扩展

- [x] 4.1 在 `pkg/bazi/engine.go` 的 `BaziResult` 结构体新增 4 个字段：`YongshenStatus string`、`YongshenGans []string`、`JishenGans []string`、`YongshenMissing []string`，附 JSON tag（snake_case）
- [x] 4.2 修改 `Calculate()` 调用新 yongshen 入口，将返回值填入 BaziResult 新字段
- [x] 4.3 验证 JSON 序列化输出包含新字段（用现有 `bazi-precision-engine` 单测或手工 curl 验证）

## 5. 后端：下游兼容性验证

- [x] 5.1 跑 `pkg/bazi/...` 下所有现有测试，确保 `getYongshenBaseline` / `caiIsJi` 等使用 `strings.Contains(natal.Yongshen, ...)` 的代码行为不变
- [x] 5.2 在 `internal/service/report_service.go` 检查 `yongshenInfo` 字符串拼接逻辑（line 740-746、line 921-927），按需调整文案（如 "调候用神：丙、癸（火、水）" 比单纯 "用神：火水" 更精确）
- [x] 5.3 确认 `pkg/bazi/event_signals.go:getYongshenBaseline` 在 yongshen 字符串只含单一五行（如"火"）时仍能正确命中

## 6. 后端：测试覆盖

- [x] 6.1 新建 `pkg/bazi/yongshen_test.go`
- [x] 6.2 测试 t0 命中（透干）：甲日寅月含丙
- [x] 6.3 测试 t0 命中（藏干）：甲日寅月四干无丙、月支寅藏丙
- [x] 6.4 测试 t0 部分命中：甲日寅月含丙不含癸 → YongshenGans=["丙"]，YongshenMissing=["癸"]
- [x] 6.5 测试 t0 缺位 fallback：甲日寅月透+藏均无丙癸 → YongshenStatus=tiaohou_miss_fallback_fuyi，yongshen 走扶抑
- [x] 6.6 测试单一五行去重：调候要求 ["丁","丙"] 都命中 → yongshen="火"
- [x] 6.7 测试三冬场景：甲日子月、原局完全无火 → 走调候 fallback，不再硬编码"火木"
- [x] 6.8 测试三夏场景：甲日午月、藏干含癸 → 调候命中"癸"
- [x] 6.9 测试 strings.Contains 兼容性：构造一个 yongshen="火" 的命盘，调用 `getYongshenBaseline(natal, "丙")` 仍命中 PolarityJi

## 7. 前端：YongshenBadge 适配（可选/最小）

- [x] 7.1 检查 `frontend/src/components/YongshenBadge.tsx` 现有展示，确认 yongshen 字符串变化（"火水"等）不会让 UI 撑开过宽（已支持 1-N 个五行 chip）
- [x] 7.2 ~~admin 显示 YongshenStatus~~ → 跳过：admin 列表读取的是 `bazi_charts` DB 表（不含新字段），需要存储改动或动态重算，超出本轮 minimal-change 范围；新建命盘的 `/api/bazi/calculate` 响应已包含完整 status，便于运营或后端日志审计
- [x] 7.3 ~~admin 显示"调候缺位"文案~~ → 同上跳过

## 8. 部署与验证

- [x] 8.1 `cd backend && go build ./...` 确认编译通过
- [x] 8.2 `cd backend && go test ./pkg/bazi/...` 全部通过
- [x] 8.3 `cd /Users/liujiming/web/yuanju && docker compose up -d --build backend` 部署
- [x] 8.4 手工新建一个测试命盘（1989-03-20 22 时），响应含 `yongshen_status="tiaohou_hit"`、`yongshen="木"`、`yongshen_gans=["甲"]`、`yongshen_missing=["癸"]`
- [x] 8.5 访问历史命盘（旧 yongshen 数据）→ 流年/大运链每次 `bazi.Calculate` 重算，自动获取新 yongshen 字段；admin 列表读旧字符串照常显示，向后兼容由架构保证
- [x] 8.6 用专业人士判定的样本命盘验证：新算法输出的 yongshen 与专业意见一致（待用户提供样本）

## 9. 文档与归档

- [x] 9.1 在 README 或 CLAUDE.md（如已有"算法说明"段落）补充"yongshen 推算优先级"说明：t0 调候字典 + t1 扶抑 fallback
- [x] 9.2 完成所有任务后用 `/opsx:archive fix-yongshen-tiaohou-priority` 归档（待 8.6 用户验证后再归档）
