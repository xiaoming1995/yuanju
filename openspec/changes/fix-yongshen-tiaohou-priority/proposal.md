## Why

经专业命理师反馈，当前命盘 yongshen/jishen 字段大量判错，下游所有"流年偏吉/偏凶"的判定（baseline polarity、财星 caiIsJi、AI 报告语气）都因此被污染。根因：现有 `inferNativeYongshen`（engine.go:554）只在「三冬无火 / 三夏无水」这种极端缺位场景才查《穷通宝鉴》调候字典，99% 命盘走"月令权重扶抑"二元阈值（>40% 翻转），而专业方法学的优先级是 **t0 调候用神主导，t1 扶抑作为 fallback**。

此外，调候字典 `tiaohou_dict.go` 已有完整 120 条（10 天干 × 12 月支）配《穷通宝鉴》原文，包含每个组合的具体调候用神**天干**（比五行更精确），但目前几乎没有被用作主路径。本变更让 t0 调候为绝大多数命盘提供 yongshen，仅当原局透干/藏干都无法呼应调候用神时回退至现有扶抑逻辑。

## What Changes

- **新增** `inferYongshenWithTiaohouPriority(natal, dayGan, monthZhi, allGans, allHideGans, stats) (yongshen, jishen, status string)` 取代 `inferNativeYongshen` 作为顶层入口
- **新增** 调候命中检测：从 4 干（年/月/日/时）+ 4 支藏干汇总所有出现的天干，与 `tiaohouDict[dayGan_monthZhi].Yongshen` 求交集
- **新增** `BaziResult.YongshenStatus` 字段（string）：取值 `"tiaohou_hit"` / `"tiaohou_miss_fallback_fuyi"` / `"tiaohou_dict_missing"` / `"fuyi"`，用于下游显示与诊断
- **修改** `inferNativeYongshen` 重命名为 `calcFuyiYongshen` 并保留为内部 t1 fallback；当前的"三冬/三夏急用"短路特殊化合并到主流程
- **修改** yongshen/jishen 的格式：从纯五行字符串（如"火木"）扩展为 `天干集` + `五行集`，前端可同时显示"调候用神：丙、癸（火、水）"
- **保留** `bazi_charts.yongshen/jishen` DB 字段；新算法仅在新建排盘时生效，已有命盘不主动迁移
- **不在范围**：格局检测（正官/七杀/食神/伤官/财格/印格）、大运动态身强弱评分、event_narrative.go 的偏凶 bug — 留至后续 OpenSpec
- **BREAKING（接口语义微变）**：yongshen 字段语义从"扶抑结果"变为"调候为主、扶抑为辅"。已有 AI 报告缓存仍可正常显示，但语气会因下次重建而变化

## Capabilities

### New Capabilities

- `bazi-yongshen-engine`：定义 yongshen/jishen 的优先级算法（t0 调候 → t1 扶抑），调候命中判定规则，藏干计入逻辑，以及缺位状态报告

### Modified Capabilities

- `bazi-precision-engine`：在已有"精准排盘"基础上补充 yongshen 算法的输出契约（YongshenStatus 字段、yongshen 字段语义变化）

## Impact

**后端（Go）**

- `pkg/bazi/engine.go` — `inferNativeYongshen` 重写为 `inferYongshenWithTiaohouPriority`，调用方 `Calculate()` 适配新签名（需要传 `allGans` + `allHideGans`）；`BaziResult` 增加 `YongshenStatus` 字段
- `pkg/bazi/tiaohou_dict.go` — 数据保持不变，但新增导出查询函数 `LookupTiaohouYongshen(dayGan, monthZhi) []string`
- `pkg/bazi/event_signals.go:getYongshenBaseline` — 读取 `natal.Yongshen`/`natal.Jishen` 的逻辑保持，但因字符串格式变化（含天干集）需确认 `strings.Contains` 仍能命中
- `pkg/bazi/event_signals.go:caiIsJi` — 同上，验证 `strings.Contains(natal.Jishen, caiWxCN)` 仍正确
- `internal/service/report_service.go:yongshenInfo` — 拼接逻辑校对，前缀文案补"调候用神"标签

**前端（React）**

- `frontend/src/components/YongshenBadge.tsx` — 适配新格式（天干 + 五行），增加"调候缺位"状态徽标
- 流年信号引擎和 AI 报告输出在 admin 调试页面显示 `YongshenStatus`，便于运营核对

**数据库**

- `bazi_charts.yongshen/jishen` 字段保留，schema 不变；新算法在新建排盘时写入新格式
- 不主动迁移历史命盘（旧值保留），新查询读旧值时 `getYongshenBaseline` 仍能工作（向后兼容）

**算法测试**

- 新增 `pkg/bazi/yongshen_test.go` 覆盖 t0 命中、t0 缺位 fallback、t0 字典缺失、藏干命中、调候多天干部分命中等场景

**风险**

- yongshen 字段语义变化导致 AI 报告下次重生成时语气可能偏离旧版（用户感知）；缓存层不主动失效
- 调候字典若有数据错漏，会被本次变更全量放大（120 条数据需在 admin 下增加校对工具，但本轮不实现）
