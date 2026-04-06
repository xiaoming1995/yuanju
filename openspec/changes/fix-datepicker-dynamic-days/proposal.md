## Why

起盘表单的「日」下拉列表永远固定显示 1~31，不随年份和月份动态变化，导致用户可以选择如「2月29日（平年）」「2月30日」「4月31日」等根本不存在的日期并提交，对八字算法产生错误输入。

## What Changes

- **新增 `getDaysInMonth` 辅助函数**：基于 `new Date(year, month, 0).getDate()` 动态计算指定年月的实际天数（自动覆盖公历闰年规则）
- **「日」下拉改为动态生成**：依据当前选中的年份和月份实时计算有效天数，生成对应选项（而非固定31项）
- **年份/月份变更时联动重置「日」**：当切换年份或月份后，若当前已选的 `day` 超出新月份的合法范围，自动重置为 1，防止表单携带非法日期提交

## Capabilities

### New Capabilities
- `datepicker-dynamic-days`：起盘表单日期选择器的动态天数校验能力，确保日下拉选项始终与所选年月的实际天数一致，并在切换时自动纠正越界日期

### Modified Capabilities
<!-- 无 spec 层行为变化，纯 UI 修复 -->

## Impact

- `frontend/src/pages/HomePage.tsx`：新增 `getDaysInMonth` 函数，修改 `days` 数组生成逻辑，修改年份/月份 `onChange` 处理逻辑，无破坏性变更
- 后端无改动
