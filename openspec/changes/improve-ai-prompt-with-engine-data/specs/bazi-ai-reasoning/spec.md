## MODIFIED Requirements

### Requirement: Analytical LLM Prompts
后端系统 SHALL 向 LLM 提供完整的引擎精算数据（四柱、十神、十二长生、神煞、大运），并通过重构的 Chain of Thought Prompt 引导 LLM 基于这些精算数据进行综合判断，而非让 LLM 自行执行排盘推算。

LLM 的 CoT 第一步 SHALL 改为：综合引擎提供的十神关系、日主星运、神煞特质，整合评估日主得令情况，参考引擎初步用神确认或微调用神/忌神，归纳相关神煞特质——上述推理过程不在报告中输出。

#### Scenario: Engine-grounded interpretation generation
- **WHEN** 后端请求 LLM 生成报告
- **THEN** LLM 基于 Prompt 中注入的十神、长生、神煞等精算数据进行解读
- **THEN** LLM 可参考引擎初步用神，结合十神全貌确认或微调最终用神/忌神输出
- **THEN** LLM 不需要从原始干支数据重新推算十神或日主强弱

### Requirement: Engine Preliminary Yongshen as Reference
引擎 SHALL 通过基础算法预算初步用神/忌神，并将其作为参考提示注入 Prompt，辅助 LLM 进行更准确的综合判断。

**（本条款替代原「Removal of Code-based Yongshen Algorithm」条款）**

#### Scenario: Preliminary yongshen assists LLM
- **WHEN** 八字排盘完成后生成命盘
- **THEN** 引擎运行 `inferNativeYongshen` 得出初步用神/忌神
- **THEN** 该初步结论作为参考传入 Prompt，标注为「引擎初步推算，供 LLM 参考」
- **WHEN** LLM 完成推理
- **THEN** LLM 输出的最终用神/忌神可能与引擎初步结论一致或不同，均视为有效

## REMOVED Requirements

### Requirement: Removal of Code-based Yongshen Algorithm
**Reason**: 原条款全面禁止代码预推用神，但实践证明引擎初步推算对 LLM 有辅助参考价值，全面禁止反而限制了系统准确性。新版本允许引擎提供初步用神作为 LLM 的参考输入。
**Migration**: 由上方「Engine Preliminary Yongshen as Reference」条款取代，引擎代码 `inferNativeYongshen` 保留并继续运行。
