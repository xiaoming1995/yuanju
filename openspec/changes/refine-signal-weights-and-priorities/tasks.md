## 1. 神煞白名单重煞分级

- [x] 1.1 在 `event_signals.go` 的 `shenshaMeta` struct 中增加 `IsHeavy bool` 字段
- [x] 1.2 在 `shenshaWhitelist` 中为羊刃、白虎、岁破、丧门、吊客、灾煞、劫煞、亡神设置 `IsHeavy: true`，其余为 `false`

## 2. Layer 0 地支信号力度分级（Evidence 措辞）

- [x] 2.1 在 `checkZhi` 的六冲/六刑分支中，Evidence 文字添加「应期力度强」标注
- [x] 2.2 在 `checkZhi` 的六害（穿）分支中，Evidence 文字添加「应期力度中」标注
- [x] 2.3 在 `checkZhi` 的六合而不化分支中，Evidence 文字添加「应期力度弱」标注

## 3. Layer 0 柱位权重标注（Evidence 措辞）

- [x] 3.1 在 `collectYingqiSignals` 信号生成时，判断被冲克位是否为日柱（日干/日支），若是则在 Evidence 末尾追加「（日柱宫位，权重较重）」
- [x] 3.2 判断被冲克位是否为月柱（月干/月支），若是则追加「（月柱宫位，权重次之）」；年柱/时柱不额外标注

## 4. 天干克五行流通判断

- [x] 4.1 在 `collectYingqiSignals` 的 `checkGan` 闭包中，检测到「流年/大运天干克用神天干位」后，额外执行流通检查：取攻击五行 A、用神五行 B，在原局天干中查找 M 使得 A生M 且 M生B
- [x] 4.2 若存在满足条件的中间五行 M，将该克信号 Polarity 改为中性，Evidence 注明「五行流通，克势转化，力度减弱」
- [x] 4.3 六害/六冲/六刑等地支交互不执行流通判断（已通过 spec 确认）

## 5. 大运+流年双冲同一用神位合并

- [x] 5.1 在 `GetYearEventSignals` 收集完大运信号和流年信号后，新增后处理：检测流年信号与大运信号是否同类型（如均为六冲）且指向同一原局位置
- [x] 5.2 若发现双重命中，合并为一条信号，Evidence 标注「大运流年双重命中，力度倍增」，Polarity=凶；不同类型的交互独立保留

## 6. Layer 0 vs Layer 4 压制规则

- [x] 6.1 在 `GetYearEventSignals` 末尾，收集该年所有 Layer 0 信号（`collectYingqiSignals` 返回值）的极性
- [x] 6.2 若 Layer 0 存在凶信号，过滤掉同年 Source=柱位互动 且 Polarity=吉 的 Layer 4 信号
- [x] 6.3 若 Layer 0 仅有吉信号或无信号，Layer 4 凶信号正常保留（不压制）

## 7. 神煞重煞压制轻煞吉信号

- [x] 7.1 在 `GetYearEventSignals` 末尾，扫描同年神煞信号，检测是否存在重煞（`IsHeavy=true`）凶信号
- [x] 7.2 若存在重煞凶信号，对同年 `IsHeavy=false` 的神煞吉信号 Evidence 末尾追加「（本年有重煞，此信号仅作参考）」；Polarity 不改变
- [x] 7.3 确认重煞压制不影响 Layer 4（Source=柱位互动）信号

## 8. 验证与测试

- [x] 8.1 为天干克流通判断新增单元测试（有流通→中性，无流通→凶）
- [x] 8.2 为 Layer 0 vs Layer 4 压制规则新增测试（凶压制吉，吉不压制凶）
- [x] 8.3 为双冲合并逻辑新增测试（同类型合并，不同类型不合并）
- [x] 8.4 为神煞重煞压制轻煞吉信号新增测试
- [x] 8.5 运行 `go test ./pkg/bazi/...` 确认所有测试通过
