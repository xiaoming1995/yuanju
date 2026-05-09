## Why

`past-year-events` 上线后，命主反馈推算结果存在四类同时出现的问题：(a) 漏掉重大事件年份、(b) 事件类型对不上、(c) 大运整体基调偏差、(d) 吉凶判断颠倒。代码体检发现 `pkg/bazi/event_signals.go` 在算法层覆盖维度严重不足——`shensha.go`(30K) 与 `shensha_dayun.go`(13K) 已实现的神煞引擎没有被调用，且缺少用神/忌神基底色、年/月/时柱互动、空亡、三会、三刑全局、伏吟反吟、大运合化、原局格局、加权身强弱等十类核心命理参数。AI 拿到的信号矩阵不全，自然写不准。

## What Changes

- **新增**：`event_signals.go` 接入 `pkg/bazi/shensha_dayun.go` 已有的神煞计算，输出天乙贵人/羊刃/华盖/白虎/丧门/吊客/天月德/红艳/将星/勾绞 等流年神煞信号
- **新增**：`EventSignal` 增加 `Polarity`（吉/凶/中性）与 `Source`（神煞/柱位互动/合化/空亡/刑/会/伏吟）字段
- **新增**：用神/忌神比对作为每个流年信号的吉凶基底色——流年五行 vs `BaziResult.Yongshen` / `Jishen`
- **新增**：流年 × 年柱/月柱/时柱 的合冲互动检测（当前只扫日柱与大运）
- **新增**：流年 × 四柱+大运 的伏吟（柱完全相同）与反吟（天克地冲）检测
- **新增**：空亡检测（按日柱旬空），流年/大运地支落空亡时输出降权信号
- **新增**：三会局检测（寅卯辰会木、巳午未会火、申酉戌会金、亥子丑会水）
- **新增**：三刑全局检测（寅巳申、丑未戌），与既有二刑互补
- **新增**：大运天干合化日干检测（如丁壬合木），合化条件成立时整段大运打上"性向偏转"标签
- **修改**：`dayMasterStrength` 从粗粒度（月支+三天干 → strong/weak/neutral）重写为加权评分（得令×5/得地×3/得势×2，含藏干透出权重，输出极强/强/中和/弱/极弱 5 档）
- **新增**：`PastEventsTemplateData` 注入 `GejuSummary`（原局格局描述）、`YongshenInfo`（用忌神描述）、`DayunHuahe`（大运合化标签）、`StrengthDetail`（身强弱评分明细）
- **修改**：升级 `defaultPastEventsPrompt`，让 AI 利用 polarity / 格局 / 合化 / 神煞 写更准确的批断；通过升级 SQL 让旧 prompt 失效重灌
- **新增**：5–10 例真实命例的回归测试集，覆盖大事年信号是否被算法捕获

## Capabilities

### New Capabilities

（无，本次为既有 capability 的能力扩展，不引入新的能力域）

### Modified Capabilities

- `past-events-signal-engine`：扩展信号检测维度，新增 polarity 字段、神煞接入、用神基底色、柱位互动、空亡/三会/三刑/伏吟/反吟/大运合化检测、加权身强弱
- `past-events-report`：Prompt 模板升级，AI 输入新增格局/用忌神/大运合化/身强弱明细等上下文字段，要求 AI 综合 polarity 与格局写出更精准的吉凶判断

## Impact

- `backend/pkg/bazi/event_signals.go`：核心改造，规则集大幅扩充
- `backend/pkg/bazi/event_signals_test.go`：新增回归测试文件
- `backend/pkg/bazi/shensha_dayun.go`：被新调用（无需改动其本身）
- `backend/internal/model/model.go`：`EventSignal` 与 `PastEventsTemplateData` 字段扩展
- `backend/internal/service/report_service.go`：`GeneratePastEventsStream` 注入新字段
- `backend/pkg/database/database.go`：`defaultPastEventsPrompt` 升级 + 升级 SQL 触发旧 prompt 失效重灌
- 缓存影响：用户已生成的过往事件报告会因 prompt 更新而显得"旧"，但缓存按 chart_id 保留；如需替换，命主可通过 Admin "清缓存"按钮重新生成
- 无 breaking change：API 路径、请求/响应结构（顶层 JSON 字段）保持兼容；仅 `signals` 数组中的信号项形态扩展（增加可选字段）
