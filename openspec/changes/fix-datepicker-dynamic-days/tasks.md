## 1. 核心实现

- [x] 1.1 在 `frontend/src/pages/HomePage.tsx` 中新增 `getDaysInMonth(year, month)` 辅助函数，使用 `new Date(year, month, 0).getDate()` 计算实际天数
- [x] 1.2 将第 93 行固定 `days` 数组改为动态：`const days = Array.from({ length: getDaysInMonth(form.year, form.month) }, (_, i) => i + 1)`

## 2. 联动重置逻辑

- [x] 2.1 将年份 `<select>` 的 `onChange` 改为独立处理函数：计算新年月的最大天数，若当前 `form.day` 超出则重置为 1，再更新 `form.year`
- [x] 2.2 将月份 `<select>` 的 `onChange` 改为独立处理函数：计算新年月的最大天数，若当前 `form.day` 超出则重置为 1，再更新 `form.month`


## 3. 验证

- [x] 3.1 选择闰年（如 2024），月份选 2 月，确认日下拉最大选项为 29
- [x] 3.2 选择平年（如 2023），月份选 2 月，确认日下拉最大选项为 28
- [x] 3.3 选择 31 日后切换到 4 月（30天），确认日自动重置为 1
- [x] 3.4 选择 15 日后切换月份，确认日不被重置，保持 15
