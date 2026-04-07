## 1. 数据结构重构

- [x] 1.1 在 `jin_bu_huan_dict.go` 中定义新的 `DayunRule` 结构体（GanXi/GanJi/ZhiXi/ZhiJi/Note/Verse 字段）
- [x] 1.2 创建 `dayunRuleDict` 字典变量（`map[string]DayunRule`，key 格式 `日干_月支`）
- [x] 1.3 移除旧的 `JinBuHuanEntry` 和 `jinBuHuanDict` 相关结构体与变量

## 2. 合表数据录入（文章一：甲乙丙丁戊）

- [x] 2.1 截取文章一所有表格完整截图（甲/乙/丙/丁/戊 各2个子表）
- [x] 2.2 录入甲木 12 月支数据
- [x] 2.3 录入乙木 12 月支数据
- [x] 2.4 录入丙火 12 月支数据
- [x] 2.5 录入丁火 12 月支数据
- [x] 2.6 录入戊土 12 月支数据

## 3. 合表数据录入（文章二：己庚辛壬癸）

- [x] 3.1 截取文章二所有表格完整截图（己/庚/辛/壬/癸 各2个子表）
- [x] 3.2 录入己土 12 月支数据
- [x] 3.3 录入庚金 12 月支数据
- [x] 3.4 录入辛金 12 月支数据
- [x] 3.5 录入壬水 12 月支数据
- [x] 3.6 录入癸水 12 月支数据

## 4. 评级算法重写

- [x] 4.1 重写 `CalcJinBuHuanDayun` 函数：前5年按 GanXi/GanJi 匹配天干，后5年按 ZhiXi/ZhiJi 匹配地支
- [x] 4.2 更新 `JinBuHuanResult` 结构体（保持 qian_level/hou_level/qian_desc/hou_desc/verse 字段）
- [x] 4.3 生成评级描述文本，包含天干/地支名称和喜忌原因

## 5. 测试验证

- [x] 5.1 单元测试：120 条数据完整覆盖验证（每个 key 均返回非空 DayunRule）
- [x] 5.2 单元测试：丙_戌 命盘（1995/10/12 午时）关键大运评级验证
- [x] 5.3 运行 `go test ./pkg/bazi/...` 确保所有测试通过
- [x] 5.4 启动后端服务，调用 API 验证返回格式正确
