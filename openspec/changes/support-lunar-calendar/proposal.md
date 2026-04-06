## Why

目前首页的起盘表单仅支持「公历（阳历）」日期输入。但在华人八字命理场景中，大量用户（特别是长辈或传统信仰者）只知道自己的「农历（阴历）」生日。不支持农历输入是一个巨大的痛点，导致用户不得不在外部工具中先手动转换成公历才能在平台排盘，体验极差。

## What Changes

- **新增历法选择**：在生辰日期选择器上方新增「公历 / 农历」切换选项。
- **农历由于闰月引发的 UI 联动**：若选择农历，在选择年份后，月份列表需动态判定当年是否存在闰月（如“闰四月”），且天数列表也要受到农历大小月（30天或29天）的动态控制。
- **API 请求参数扩展**：前端向 `POST /api/bazi/calculate` 发送的请求体中，新增 `calendar_type` (solar/lunar) 和 `is_leap_month` 标识。
- **后端算法适配**：后端在 `bazi.Calculate` 引擎中拦截，当用户输入为农历时，借助 `lunar-go` 库自动将带闰月的农历转换为公历，进而复用原有的天文历法排盘运算。

## Capabilities

### New Capabilities
- `lunar-input-support`: 前后端对农历日期的全面支持，包括农历闰月的识别、日期联动以及后端的历法转换支持。

### Modified Capabilities
<!-- 无 spec 层行为变化，核心算命逻辑原封不动 -->

## Impact

- `frontend/src/pages/HomePage.tsx`: 日期表单状态扩展（增加 `calendarType`, `isLeapMonth`），年月日的下拉菜单选项渲染逻辑大幅升级。
- `backend/internal/handler/bazi_handler.go`: 入参 `CalculateInput` 增加参数，校验规则随之变更。
- `backend/pkg/bazi/engine.go`: 核心引擎需追加公历转换前置预处理，其他排盘逻辑无痛向后集成。
