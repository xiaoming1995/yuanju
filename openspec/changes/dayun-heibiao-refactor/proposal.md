## Why

当前大运评级引擎（金不换）使用自定义的「方向映射」逻辑（GoodDirections / BadDirections）来判断大运吉凶，这套逻辑存在根本性缺陷：

1. **数据源不权威**：方向映射是从诗句中人工推导的近似值，与原始文献对不上
2. **调候用神与金不换脱钩**：两套字典（`tiaohouDict` 和 `jinBuHuanDict`）各自独立，存在互相矛盾的情况（如丙_戌：调候喜壬水，金不换忌北方水）
3. **精度不够**：方向映射把地支归为"东南西北"四方，丢失了地支级别的精细判定

梁湘润《子平教材讲义》第二级次中的「金不换大运——调候用神合表」是权威数据源，直接给出了每个日干月令下的精确喜忌天干和地支，无需任何推导。

## What Changes

- **替换** `jin_bu_huan_dict.go` 中的数据结构：废弃 `GoodDirections/BadDirections` 方向映射模型，改为基于合表的精确天干喜忌 + 地支喜忌双层数据
- **重写** `CalcJinBuHuanDayun` 评级算法：前5年按天干查「调候用神喜忌」，后5年按地支查「金不换地支喜忌」
- **录入** 120 条合表数据（10天干 × 12月支），每条包含：天干喜忌、地支喜忌、特注
- **保留** 前端"前五年 / 后五年"双评级展示逻辑（已完成，仅需确认兼容性）

## Capabilities

### New Capabilities
- `dayun-heibiao-engine`: 基于梁湘润合表的大运评级引擎，实现精确到天干/地支的喜忌查表和前后五年独立评级

### Modified Capabilities
- `bazi-advanced-data`: 金不换数据结构从方向映射模型变更为天干喜忌+地支喜忌双层模型

## Impact

- **后端核心文件**：`backend/pkg/bazi/jin_bu_huan_dict.go` — 数据结构重构 + 120条数据重录
- **后端算法**：`CalcJinBuHuanDayun` 函数完全重写
- **API 响应格式**：`jin_bu_huan` 字段的 `qian_level/qian_desc/hou_level/hou_desc` 保持不变，但评级来源和描述文本会变化
- **前端**：无需改动（已适配前后五年双评级展示）
- **旧字段废弃**：`GoodDirections`、`BadDirections`、`Verse`、`SpecificZhi` 全部移除
