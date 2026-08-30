## Why

当前评估虽然已识别日干调候用神，并在天透地藏时给予可用性分数，但它仍只是总分中的一项。与此同时，主格的通用根气阈值可能把“天透地藏”的盘标记为“部分成立”。这与产品采用的判定规则不一致：日干调候用神天透地藏即为调候成格，应明确呈现为高格基础。

制化配合、扶抑用神和寒热调候仍然反映原局的不同维度，不能被调候成格覆盖或混写成同一结论。

## What Changes

- 新增日干调候成格层：日干调候所需五行或干支同时在天干透出、地支或藏干出现时，判定为“天透地藏成格”，并标记为“高格基础”。
- 将“日干调候是否成格”与“可用性得分”及“格局制化配合”分离输出，避免由主格根气阈值否定该成格结论。
- 调整命盘座驾的评级解释：高格基础应作为独立的基础加成和用户可见依据；扶抑、寒热调候、制化配合及关键破损仍独立影响最终承载与等级。
- 在命盘结果页和 AI 解读上下文中展示并传递日干调候成格的证据（所需用神、天干、地支/藏干），同时保留制化配合的原有说明。
- 补充覆盖“天透地藏成格”“仅透/仅藏部分到位”“成格但其他维度受限”的后端与前端测试。

## Impact

- Affected specs: `bazi-advanced-data`, `yongshen-driven-flow-year-judgment`
- Affected code: `backend/pkg/bazi/natal_assessment.go`, `backend/pkg/bazi/vehicle_profile.go`, `backend/internal/service/report_service.go`, `frontend/src/pages/ResultPage.tsx`, 调候与座驾相关组件及测试。
