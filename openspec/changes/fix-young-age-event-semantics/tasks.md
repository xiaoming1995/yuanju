## 1. 后端：常量与 Type 定义

- [x] 1.1 在 `pkg/bazi/event_signals.go` 顶部新增常量 `YoungAgeCutoff = 18`
- [x] 1.2 新增 7 个 Type 字符串常量

## 2. 后端：函数签名扩展

- [x] 2.1 修改 `GetYearEventSignals` 签名，加入 `age int` 参数
- [x] 2.2 修改 `GetAllYearSignals` 传 `ln.Age`；同步 `report_service.go` 与 `event_signals_test.go` 调用方
- [x] 2.3 函数体内 `isYoung := age > 0 && age < YoungAgeCutoff` 一次复用

## 3. 后端：少年期分支 — 财 / 比劫 / 食伤

- [x] 3.1 财星透干 → `学业_资源`（保留三档极性：忌神→凶 / 身弱→中性 / 否则→吉）
- [x] 3.2 少年期跳过"财官双叠"分支
- [x] 3.3 比劫透干 → `学业_竞争`（身弱→吉 / 身非弱→凶）
- [x] 3.4 食伤透干 → `学业_才艺`（身强→吉 / 身弱→凶）

## 4. 后端：少年期分支 — 官杀 / 印星

- [x] 4.1 官杀透干 → `学业_压力`（身弱→凶 / 身强→吉 / 中和→中性）；少年期不输出官杀双叠
- [x] 4.2 印星透干 → `学业_贵人`

## 5. 后端：少年期分支 — 合冲日支 / 桃花

- [x] 5.1 大运冲日支 → `性格_叛逆`
- [x] 5.2 大运合日支 → `性格_情谊`
- [x] 5.3 流年合冲日支已经走 `婚恋_合冲对消` → 少年期不会触发（因 婚恋_* 不再 emit），自动失效
- [x] 5.4 少年期跳过男财/女官 → `婚恋_合` 路径
- [x] 5.5 神煞 post-process：`婚恋_*` → `性格_情谊`、`财运_*` → `学业_资源`、`事业` → `学业_压力`/`学业_贵人`

## 6. 后端：三会引动的少年期重映射

- [x] 6.1 三会 / 三合引动的 `婚恋_合` 在少年期改为 `性格_情谊`

## 7. 后端：大运总结 prompt 按 youngRatio 动态调整

- [x] 7.1 计算 `youngCount/totalCount`
- [x] 7.2 `model.DayunSummaryTemplateData` 加 `LifeStageHint`
- [x] 7.3 prompt 模板加入 `{{if .LifeStageHint}}{{.LifeStageHint}}{{end}}`
- [x] 7.4 `buildLifeStageHint` 按三档生成提示词（全段读书期 / 跨界 / 成人期）

## 8. 前端：SIGNAL_LABEL 标签扩展

- [x] 8.1 加 7 个新 Type 映射（学业系木绿、性格系土黄、压力/叛逆红凶）
- [x] 8.2 视觉确认（chip 与现有 同款 height，未超宽）

## 9. 后端：测试覆盖

- [x] 9.1-9.10 `young_age_test.go` 11 个 case 全部通过（财非忌/财为忌/19 岁回归成人/合日支/冲日支/克日干/双叠/桃花神煞/youngRatio 三档）

## 10. 部署与验证

- [x] 10.1 `go build ./...` 编译通过
- [x] 10.2 `go test ./pkg/bazi/...` 全部通过（旧测试 + 新 11 个 case）
- [x] 10.3 `npm run build` 前端编译通过
- [x] 10.4 `docker compose up -d --build backend frontend` 部署完成
- [x] 10.5 1989-03-20 命盘（6 岁起运）→ 1994 (age 6) signals=['喜神临运','性格_情谊','综合变动','学业_压力','健康','学业_贵人']；2000 (age 12) signals=['学业_才艺','学业_贵人','性格_情谊',...]；2006 (age 18) 回归 ['婚恋_合','事业']；少年期 12 个年份 0 个成人期 type 残留
- [ ] 10.6 大运总结 stream 验证（需登录 + AI 调用，留作用户体验时观察）
- [ ] 10.7 手动清理 `ai_dayun_summaries`（用户上线时按需执行）

## 11. 文档与归档

- [x] 11.1 CLAUDE.md "Key Conventions" 补充年龄分段语义说明
- [ ] 11.2 完成所有任务后用 `/opsx:archive fix-young-age-event-semantics` 归档
