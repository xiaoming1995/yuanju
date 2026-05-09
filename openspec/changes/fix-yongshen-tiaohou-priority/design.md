## Context

**当前状态**

`backend/pkg/bazi/engine.go:546-570` 的 `inferNativeYongshen` 是 yongshen 计算的唯一入口：

```
inferNativeYongshen(dayGanWx, monthZhiWx, monthZhi, stats):
  if monthZhi ∈ {亥,子,丑} && stats.火 == 0:
    return ("火木", "水金土")          # 极寒急用
  if monthZhi ∈ {巳,午,未} && stats.水 == 0:
    return ("水金", "火木土")          # 极热急用
  return calcWeightedYongshen(...)     # 90%+ 命盘走这里：月令权重扶抑
```

`tiaohou_dict.go` 已包含 120 条《穷通宝鉴》调候用神配方（每条给 2-4 个具体调候用神**天干**），但仅在 `tiaohou.go:CalcTiaohou` 里被读取用于在前端展示"调候建议"，**没有进入 yongshen 主计算路径**。

**专业方法学（用户提供）**

- t0：调候用神 — 永远先查字典；命中即作为主用神
- t1：扶抑用神 + 格局用神 — 仅当 t0 缺位时回退（格局留至下一个 OpenSpec）

**约束**

- 不重写命盘字段 schema（`bazi_charts.yongshen/jishen` 仍是 string）；保持向后兼容
- 不主动迁移历史命盘
- 下游 `getYongshenBaseline` 用 `strings.Contains(natal.Yongshen, lnWxCN)` 匹配五行名，新算法输出必须保留五行名以维持兼容
- 单文件 ≤ 500 行（CLAUDE.md 约束）

## Goals / Non-Goals

**Goals**

1. 让 99% 命盘走调候字典 → yongshen 反映《穷通宝鉴》经典而非月令权重扶抑
2. 当调候用神在原局完全缺位（透干 ∪ 藏干都没有）时，明确报告"命有病：缺 X"，并 fallback 到现有扶抑逻辑
3. 输出 `YongshenStatus` 状态字段供下游与运营诊断
4. yongshen 字段含五行名 → 维持现有 `getYongshenBaseline` 与 `caiIsJi` 兼容
5. 新增独立测试覆盖 t0 命中 / t0 部分命中 / t0 缺位 fallback / 字典缺失（不应发生）

**Non-Goals**

- 格局用神检测（正官/七杀/食神/伤官/财格/印格）— 下个 OpenSpec
- 大运动态身强弱（effective_strength = natal + dayun）— 下个 OpenSpec
- event_narrative.go 的"偏凶 bug"（结尾比较与 evidence 排序）— 下个 OpenSpec
- 主动迁移历史命盘 yongshen 数据
- 调候字典数据校对 / admin 编辑 UI

## Decisions

### D1：t0 命中条件 = 调候用神天干 ∈ (4 干透 ∪ 4 支藏干)

**Why**：经典命理认为透干为强、藏干为弱、不透不藏为缺。本变更采用"宽松命中"——只要至少一个调候用神天干出现在透或藏，即视为 t0 成立。这能让最大比例命盘走调候路径，同时保留"完全缺位"的清晰诊断。

**Alternatives considered**

- A1（严格）：必须透干才算命中 — 太严格，藏干在调候上的影响在子平实战中是被认可的，会导致大量命盘错误回退到扶抑
- A2（加权）：透 = 1.0、藏 = 0.5、缺 = 0.0，按总分 ≥ 0.5 视为命中 — 第一阶段过度复杂，权重需经验证；先用宽松命中，未来可加权
- **A3（采用）**：透 ∪ 藏 都视为命中 — 简单可解释，符合"调候为纲"的方法学

### D2：调候用神字典返回的天干列表为"或"关系（任一命中即 t0 成立）

**Why**：tiaohou_dict 的每条 `Yongshen` 数组（如 `["丙","癸"]`）在《穷通宝鉴》原文里通常是"取其一二"或"两全则佳"。第一阶段简化为"任一命中即足够"，后续可升级为"全部命中→t0 高强度，部分命中→t0 中等"。

**Alternatives considered**

- B1（合）：必须全部命中 — 过严，很多命盘只能命中一个调候用神，会错失大量正确判定
- B2（采用）：任一命中即 t0 成立 — 与 D1 同向宽松
- B3（区分等级）：全命中 → strong；部分命中 → partial；都不中 → miss — 后续优化方向，第一版不引入

### D3：t0 命中时，输出 yongshen = 命中天干对应的五行集合

**Why**：维持下游 `strings.Contains(natal.Yongshen, lnWxCN)` 兼容。例：调候字典给 `["丙","癸"]`，原局透"丙"、藏"癸"，则两者都命中 → yongshen 输出"火水"（丙=火、癸=水），jishen = 克泄之（"金土" / "土火"）。

实现细节：
- yongshen 字符串使用现有 wuxing 中文名（火/水/土/金/木）
- 重复五行去重（如调候要求 `["丁","丙"]` 都是火 → yongshen = "火"）
- jishen 取 yongshen 五行的"克 + 泄"对（与现有逻辑一致）

**Alternatives considered**

- C1：天干字符串而非五行（如 yongshen = "丙癸"）— 下游所有依赖五行匹配的代码全部需要重写，超出本变更范围
- **C2（采用）**：yongshen 仍是五行名 — 兼容
- C3：yongshen 同时返回天干和五行（如新增 `YongshenGans []string` 字段）— 比 C2 更精确，前端可展示"丙癸（火水）"。**采用 C3**，新增字段。

更新决策：**C3 采用** — `BaziResult.YongshenGans []string` + `BaziResult.JishenGans []string` 新增字段（前端可读，但下游算法仍以 `Yongshen` 五行字符串为主）。

### D4：t0 缺位（一个天干都不在原局）→ fallback 至现有 `calcWeightedYongshen`

**Why**：保留现有扶抑逻辑作为 t1 fallback，符合用户选定的方案。fallback 时 `YongshenStatus = "tiaohou_miss_fallback_fuyi"`，前端可显示"调候缺位（缺 X、Y），现行用神按扶抑：..."。

`YongshenGans` 字段在 fallback 时保持为空数组（扶抑层级没有具体天干概念）。

### D5：合并现有"急用短路"到主流程

**Why**：当前 `inferNativeYongshen` 的"三冬无火 / 三夏无水"短路实际上是不完整的调候特化（仅覆盖 2 个月支边界）。新算法的调候字典 120 条已经全量覆盖月令冷热燥湿，这两个短路完全冗余且不一致（短路写死"火木"/"水金"，与字典对应月支的具体调候用神天干可能不同）。**删除短路，全部由字典驱动**。

### D6：BaziResult 新增字段命名与语义

```go
type BaziResult struct {
    // ... 现有字段
    Yongshen        string   `json:"yongshen"`           // 五行集合，如"火水"。保持向后兼容
    Jishen          string   `json:"jishen"`             // 五行集合，如"金土"
    YongshenGans    []string `json:"yongshen_gans"`      // 新增：调候命中的具体天干，如["丙","癸"]
    JishenGans      []string `json:"jishen_gans"`        // 新增：与 YongshenGans 对应的克/泄天干集
    YongshenStatus  string   `json:"yongshen_status"`    // 新增：tiaohou_hit / tiaohou_miss_fallback_fuyi / tiaohou_dict_missing / fuyi
    YongshenMissing []string `json:"yongshen_missing"`   // 新增：t0 缺位时记录"命缺哪些调候用神天干"，如["丙"]；命中或不可用时为空
}
```

### D7：YongshenStatus 取值与含义

| 状态 | 含义 | 触发条件 |
|---|---|---|
| `tiaohou_hit` | 调候命中 | 字典命中 + 至少一个调候用神天干在原局透/藏 |
| `tiaohou_miss_fallback_fuyi` | 调候缺位，回退扶抑 | 字典命中但调候用神天干在原局完全缺位 |
| `tiaohou_dict_missing` | 字典缺该日干_月支组合（理论不发生） | 字典查不到，直接走扶抑 |
| `fuyi` | 历史命盘 / 未来扩展专用 | 旧 inferNativeYongshen 的短路场景已并入主流程，本字段保留以备它用 |

### D8：测试覆盖矩阵

| 场景 | 输入 | 预期输出 |
|---|---|---|
| t0 命中（透干）| 甲日寅月、原局含丙 | yongshen 含火、status=tiaohou_hit、YongshenGans 含丙 |
| t0 命中（藏干）| 甲日寅月、原局无丙但含寅藏丙 | yongshen 含火、status=tiaohou_hit、YongshenGans 含丙 |
| t0 部分命中 | 甲日寅月、原局含丙不含癸 | yongshen 含火、status=tiaohou_hit、YongshenGans=["丙"]，YongshenMissing=["癸"] |
| t0 缺位 | 甲日寅月、原局完全无丙癸（透+藏） | status=tiaohou_miss_fallback_fuyi、yongshen 走扶抑结果、YongshenMissing=["丙","癸"] |
| 字典缺失（不应发生）| 不存在的 dayGan_monthZhi | status=tiaohou_dict_missing、yongshen 走扶抑结果 |

### D9：迁移与缓存策略

- yongshen/jishen 字段语义微变（含义从"扶抑"为主转为"调候+扶抑"），但格式（五行名字符串）兼容
- 已有 `bazi_charts` 表中的 yongshen 保留旧值；新建排盘时使用新算法
- AI 报告缓存 `ai_reports` / `ai_dayun_summaries` 不主动失效；下次重新生成时拿到的将是新 yongshen
- 不提供 admin 一键迁移按钮（防止误操作；运营可通过 chart 重建间接刷新）

## Risks / Trade-offs

| 风险 | 缓解 |
|---|---|
| 调候字典数据不准（120 条某些条目偏《穷通宝鉴》摘录可能有误差）| 本轮不校对字典；增加 `YongshenGans` 字段让运营可对比命盘原文与字典；如发现错误，下个变更专门做字典校对 |
| 旧命盘 yongshen 与新命盘不一致 | 设计上接受。前端 admin 面板显示 `YongshenStatus`，老命盘 status 缺失时降级显示"旧版用神"标签 |
| `getYongshenBaseline` 等下游依赖五行名匹配 — 如调候命中只产生单一五行（去重后）→ 五行字符串变短 → `strings.Contains` 仍匹配 ✓ | 在 yongshen_test.go 增加针对 `strings.Contains` 的回归 case |
| 调候用神被识别为忌神冲突（极少数命局，调候要求火、扶抑也判忌神含火）| t0 命中时 jishen 由 `wxKe[yongshen 五行] + wxXie[yongshen 五行]` 派生，与扶抑无关。不会出现冲突 |
| 现有"三冬急用 / 三夏急用"短路被删除导致行为差异 | 删除短路后，原属于该路径的命盘改走字典：字典对寒月的处理比"硬编码火木"更精细。增加测试覆盖 子月日干各档调候用神 |
| AI 报告内容因 yongshen 变化语气改变，用户体感"为什么我的命盘解释变了" | 前端在 yongshen 变化时不显式提示；老报告缓存继续提供旧解释。本轮不引入"算法版本号"字段（见 D9） |

## Migration Plan

1. 后端代码上线（go build + docker compose up -d --build backend）
2. 验证：手工新建几个测试命盘，确认 yongshen/jishen/YongshenStatus/YongshenGans 字段正确返回
3. 验证：访问历史命盘，确认旧 yongshen 仍能渲染（向后兼容）
4. 前端 YongshenBadge 上线（如改动 UI），确认"调候缺位"状态正确展示
5. 不需要 DB 迁移、不需要数据回填

**Rollback**：直接 `git revert` 后重新部署，已存量命盘的 yongshen 不受影响（因为字段格式兼容）。

## Open Questions

- 是否需要在 admin 面板暴露"调候字典查看 / 编辑"工具？（本轮不实现，留作后续 OpenSpec）
- `YongshenStatus = "tiaohou_miss_fallback_fuyi"` 的命盘，AI 报告 prompt 是否需要专门提示"命有病"？— 当前 prompt 模板由 admin 维护，本变更不强制改 prompt；运营可在 yongshen_status 字段稳定后自行决定是否在模板中引用
