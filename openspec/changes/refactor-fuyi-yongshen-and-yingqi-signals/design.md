## Context

当前 `yongshen.go` 的 `calcWeightedYongshen()` 以"帮扶五行占比>40%"为阈值判断身强弱，缺乏对"根深浅/比劫/印"优先级的建模。`event_signals.go` 的 `getYongshenBaseline()` 对全年所有信号染一个底色极性，原理上等价于"流年天干五行属用神→全年偏吉"，未检测原局各干支位置是否被刑冲克合穿破。

现有数据基础充足：`BaziResult.{Year/Month/Day/Hour}HideGan` 已由 lunar-go 按主气/中气/余气顺序填充；`event_signals.go` 已有六冲、六合、三刑等关系表；`tiaohou_dict.go` 已完整覆盖120条，调候路径不需变动。

## Goals / Non-Goals

**Goals:**
- 用有根/无根→比劫≥3→印旺三段优先级取代五行比例阈值，输出同样的 yongshen/jishen 五行集字符串
- 以"刑冲克合穿破原局用神/忌神位"替换全年底色，每条交互独立输出 EventSignal
- 补充六害（穿破）关系表

**Non-Goals:**
- 不变动调候路径（tiaohou_dict.go、inferYongshenWithTiaohouPriority）
- 不变动十神透干、神煞、伏吟反吟、空亡等其他信号层
- 不触碰格局判断

## Decisions

### D1：藏干取前两项（主气+中气），余气不计

lunar-go 的 HideGan 数组按 [主气, 中气, 余气] 顺序排列（部分地支只有1-2项）。有根/比劫多/印旺三个条件均只计入 `hideGan[:min(len,2)]`，余气不参与计数。这与专业命理师确认的规则一致，也避免余气误判。

### D2：calcFuyiStrength 作为独立函数，与 calcWeightedYongshen 并存过渡

新建 `calcFuyiStrength(natal *BaziResult) (isStrong bool, reason string)` 实现新规则。`inferYongshenWithTiaohouPriority` 内部将 fallback 从 `calcWeightedYongshen` 切换到通过 `calcFuyiStrength` 派生用神/忌神。旧函数保留但不再被主路径调用，方便 A/B 对比与回滚。

### D3：应期信号函数 collectYingqiSignals 独立于 GetYearEventSignals 主循环

新建 `collectYingqiSignals(natal, lnGan, lnZhi, dyGan, dyZhi string) []EventSignal`，返回所有刑冲克合穿破位置信号。在 `GetYearEventSignals` 开头调用，替换 `getYongshenBaseline` 的位置，不再生成全年底色。其他信号层的 `addP()/applyPolarity()` 不再接收 baseline 参数（清零为空字符串），各信号极性由 signalSelf 独立决定。

### D4：六合化出五行取用神/忌神极性

地支六合是否化取决于原局根气，此处简化处理：直接按"合化后五行"（已在 ganWuhe 表里）判断属用神还是忌神。天干五合同理。合而不化（无根气）暂按"被合住=锁定用神=凶/锁定忌神=吉"处理，evidence 中注明"合而不化"。

## Risks / Trade-offs

- [风险] 扶抑规则变化导致少量命盘的 yongshen/jishen 与旧版不同，影响已缓存 AI 报告语气 → 缓存报告不主动刷新，下次用户重新生成时以新结果为准
- [风险] 去掉全年底色后，部分流年的信号极性分布比旧版更分散（无统一基调） → AI 提示词已能处理多条信号的综合解读，可接受
- [Trade-off] calcWeightedYongshen 保留为死代码，暂不删除，方便出问题时快速回滚

## Migration Plan

1. 实现并单测 `calcFuyiStrength`，对比旧版结果
2. 实现并单测 `collectYingqiSignals`
3. 切换主路径，移除 `getYongshenBaseline` 调用，清空 baseline 传参
4. 运行全量测试，重点验证1995-10-12午时的2024/2025年信号
5. 部署后观察 AI 报告质量，无问题后再清理旧函数
