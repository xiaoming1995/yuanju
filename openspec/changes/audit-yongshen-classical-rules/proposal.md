## Why

测试盘 1996-02-08 20:00 男 (chart 221c84e7) 的过往事件推算出现可信度问题：
癸巳大运段 AI 总结写"巳火忌神发力""癸水为调候喜神"等自相矛盾内容。

排查定位到根因：`pkg/bazi/yongshen.go::inferYongshenWithTiaohouPriority`
将"调候字典命中"作为用神判定的唯一拍板路径，跳过身强弱判断与寒热判断。
原局错标签 `yongshen="火水"  jishen="木土"` 沿全部下游 evidence/prompt
传播，导致 80 年所有大运/流年的喜忌定调全部反向。

命主提供了完整的命理用神算法规范（"调候为急、扶抑优先"……）。
本提案不实施代码改动，仅沉淀两项产物供未来实施时引用：

1. 用户算法 vs 当前代码的 10 条逐条审计
2. explore 模式下达成的 4 项设计决定（算法主轴/字段表达/缓存处理/版本控制）

## What Changes

**本提案不修改任何代码。**

仅产出两份审计/决策文档：

- `audit.md`：将用户算法拆为 10 条审计项，逐条与代码对照，标注 ✅/⚠️/❌
- `design-decisions.md`：记录 4 项设计选择及其理由，避免下次实施重复讨论

## Status

**Deferred（用户主动暂停）**

命主在 explore 阶段决定"先不改用神了"。本目录作为锚点保留，
未来恢复实施时直接读取本目录即可还原 explore 上下文。

## Out of Scope

- 不重写 `yongshen.go`
- 不新增任何 schema 字段
- 不清理 AI 缓存
- 不增加 algorithm_version 字段
- 不修改任何下游 `event_signals.go` 函数

## Impact

零代码改动。
仅在 `openspec/changes/audit-yongshen-classical-rules/` 新增 2 个 markdown 文件。
