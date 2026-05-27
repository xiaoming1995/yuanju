# v3.1 Relax-Threshold Design

详见 `docs/superpowers/specs/2026-05-28-compatibility-scoring-v3.1-relax-design.md`。

本 change 的关键设计点：
1. 同时构成「五行同」与「六冲」的辰戌、丑未仍判中档 30（与「纯加分制」一致）
2. 合日柱下档 3 不区分干关系（下档已是安慰分）
3. 归一化 `(sum × 2 + 1) / 3` 不变，sum 仍 ∈ [0, 30]
4. 新增 helper：`branchSameElement` / `branchShengElement`，复用 `event_signals.go` 的 `zhiWuxing` / `wxSheng`
