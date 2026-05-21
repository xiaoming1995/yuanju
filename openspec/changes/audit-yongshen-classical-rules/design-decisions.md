# 设计决定（explore 阶段 4 问 4 答）

记录 2026-05-21 explore 模式中达成的设计选择，未来实施时直接采纳。

---

## 决定 1: 算法主轴

**选择**：扶抑主导 + 调候辅助（**渊海子平 + 滴天髓融合**）

**候选**：
- (a) 滴天髓 / 流通主导
- (b) 子平真诠 / 格局主导
- **(c) 渊海子平 / 扶抑主导**  ← 采纳
- (d) 多派加权投票

**理由**：
- 渊海子平是子平命理"地基"，扶抑是流派共识底层
- 工程上 `calcFuyiStrength` / `calcWeightedYongshen` 已就位，改造成本最低
- 解释性强（"身强→克泄"是命主能理解的逻辑）
- 调候作"补丁"叠加，保留穷通宝鉴价值
- 与用户提供的"调候为急、扶抑优先"规范一致（仅作微调：用户规范是"调候在至寒至热极端时优先"，
  其他情况扶抑，本质仍是扶抑主轴）

**最终采纳的算法形态**（综合用户规范）：

```
Step 0: 寒热判断 (是否至寒至热)
        ├─ 是 → Step 1A: 火水调候为主
        └─ 否 → Step 1B: 扶抑判断
                       ├─ 身强 → 克泄耗为用
                       ├─ 身弱 → 生扶为用
                       └─ 中和 → 调候补正
Step 2: 格局配合（取格 → 用神为格服务）
Step 3: 通关（流通顺/逆判断）
Step 4: 用药（针对特殊缺陷）
Step 5: 精修到天干（透干优先 / 调候字典优选）
```

---

## 决定 2: 字段表达

**选择**：主从级（primary_gan + wuxing_set + positions）

**候选**：
- (i) 五行级（"火"）
- (ii) 天干级（["丙"]）
- **(iii) 主从级**  ← 采纳

**字段设计**：

```go
type BaziResult struct {
    // 新字段
    YongshenPrimaryGan string   // 主用神天干，如 "丙"
    YongshenWuxingSet  string   // 大粒度兜底，如 "火"
    YongshenPositions  []string // 实际命中位置，如 ["年干","时干","月支藏"]
    JishenGan          []string // 主忌神天干，如 ["壬","癸"]
    JishenWuxingSet    string   // 大粒度兜底，如 "水"

    // 旧字段保留（向后兼容）
    Yongshen string  // = WuxingSet（旧逻辑读这个）
    Jishen   string  // = JishenWuxingSet
}
```

**bazi_charts 表新增**（migration）：

```sql
ALTER TABLE bazi_charts
  ADD COLUMN yongshen_primary_gan VARCHAR(2),
  ADD COLUMN yongshen_positions   JSONB,
  ADD COLUMN jishen_gan           JSONB;
-- 旧 yongshen / jishen 字段保留不动
```

**理由**：
- 精度对齐古法（丙 ≠ 丁）
- 旧字段保留 → 下游零改动
- migration 简单（3 个 nullable 字段）
- admin UI 能分层展示，debug 友好

---

## 决定 3: 缓存处理

**选择**：啥都不动

**候选**：
- (h) 升级保留 + 版本标签
- (i) 一次清空
- **啥都不动 + 加 admin 单盘重算按钮**  ← 采纳

**理由**：
- 当前数据库全是 dev 测试数据（13 用户，全 test@/codex-* 测试账号）
- 无真实生产用户，清理收益≈0
- AI 缓存只在访问时呈现，没访问就是死数据
- 用户测试的盘 < 5 个，需要时手动 DELETE 即可
- `bazi_charts.yongshen` 是反规范化缓存，`LoadOrCalculateResult` 每次会现算

**唯一可选辅助**：
- Admin UI 加 "重算此盘缓存" 按钮
- 后端实现：`DELETE FROM ai_dayun_summaries/ai_past_events/... WHERE chart_id=?`
  + `UPDATE bazi_charts SET yongshen=NULL, ... WHERE id=?`
- 估时 5 分钟

---

## 决定 4: 版本控制

**选择**：不加 algorithm_version 字段

**候选**：
- (j) 加 algorithm_version 字段
- **(k) 不加**  ← 采纳

**理由**：
- YAGNI——没有消费者的字段就是死字段
- `ai_dayun_summaries.algorithm_version` 字段已存在但**从未被读写**，证明这层概念现在用不到
- 未来真做 A/B 灰度时再 ADD COLUMN（PG 秒级操作）
- 与决定 3 "啥都不动" 基调一致

---

## 决定汇总

| # | 维度 | 选择 |
|---|------|------|
| 1 | 算法主轴 | 扶抑主导 + 调候辅助 (寒热判断分流) |
| 2 | 字段表达 | 主从级 (primary_gan + wuxing_set + positions) |
| 3 | 缓存处理 | 啥都不动，仅加 admin 单盘重算按钮 |
| 4 | 版本控制 | 不加 algorithm_version |

---

## 未决问题（实施时需再做选择）

1. **身强弱算法选型**：A 位置规则 / B 月令 40% 阈值 / C 五档评分
   - explore 阶段倾向 C，但用户暂停在此处未决
   - 当前 codebase 3 套互相矛盾，必须先选一个再继续
2. **精修选 gan 的优先级排序**：透干 > 通根 > 调候字典 > 月令气？需明确权重
3. **从格 / 化格 / 一气格特殊处理**：是否在 v1 即支持？
4. **大运整段定调判定**：用神/生助用神/忌神/冲克用神 的具体边界
5. **测试盘 fixture**：实施时需准备至少 5 个对照命局（身强/身弱/中和/极寒/极热）
