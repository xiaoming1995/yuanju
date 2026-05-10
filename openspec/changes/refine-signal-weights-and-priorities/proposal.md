## Why

流年事件信号目前采用平权叠加输出，所有信号（应期位置、十神宏观、神煞）没有主次关系，专业命理师审核后指出三类信号之间存在明确的优先级与权重规则，需要在算法中落实，避免产生误导性的矛盾输出。

## What Changes

- **应期位置信号（Layer 0）确立为主干**：当与十神宏观信号（Layer 4）方向冲突时，Layer 0 凶信号压制 Layer 4 吉信号；Layer 0 吉信号不压制 Layer 4 凶信号（凶的提醒保留）；方向一致时两层信号叠加输出；Layer 0 无信号时 Layer 4 正常输出
- **Layer 0 内部权重分级**：地支交互力度：冲=刑 > 穿 > 合而不化；天干克新增五行流通判断：原局天干存在中间五行使攻击能量最终转化为生用神时，信号降为中性
- **位置权重**：原局各柱位受冲克时，日柱 > 月柱 > 年柱/时柱
- **大运+流年双冲同一用神位**：合并为更强的叠加信号，不再各自独立输出两条
- **神煞内部分级**：区分重煞（羊刃、白虎、岁破、丧门、吊客、灾煞、劫煞、亡神）与轻煞；重煞出现时，同年轻煞吉信号标注「受重煞影响，仅作参考」；重煞不影响 Layer 4 十神信号
- **旬空减半适用范围确认**：旬空对所有信号（含重煞）一律标注力度减半

## Capabilities

### New Capabilities

- `layer0-layer4-interaction`: Layer 0 与 Layer 4 之间的优先级与压制规则
- `layer0-internal-weights`: Layer 0 内部的冲刑穿合权重分级与天干克流通判断
- `shensha-internal-weights`: 神煞内部重煞/轻煞分级及重煞对轻煞吉信号的压制规则

### Modified Capabilities

## Impact

- `backend/pkg/bazi/event_signals.go`：`GetYearEventSignals` 主函数逻辑、`collectYingqiSignals` 中天干克流通判断、大运+流年叠加合并逻辑
- 神煞白名单（`shenshaWhitelist`）需新增重煞标记字段
- 信号输出结构（`EventSignal`）可能需要新增权重/优先级字段
