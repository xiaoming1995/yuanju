# Tasks

- [x] T1 后端 `CalculateInput` 增加 `DisplayName string`，handler 调用现有 `normalizeChartDisplayName` 后写入 `BaziChart.DisplayName`
- [x] T2 后端 handler 单测：提交合法/空/超长 display_name 三种情况，验证落库结果与错误返回
- [x] T3 前端 `lib/api.ts` `CalculateInput` 增加 `display_name?: string`
- [x] T4 前端 `BirthProfileForm`（或直接 `HomePage`）新增 displayName 输入框（可选，maxLength=20，placeholder "例如：我 / 小王"）
- [x] T5 前端 `HomePage` state + submit 把 displayName 拼进 `CalculateInput`
- [x] T6 前端 `ResultPage.tsx` 删除 `.chart-archive-tools` section (lines 731-784)，并移除关联的：
  - `chartDisplayNameDraft / setChartDisplayNameDraft`
  - `chartDisplayNameSaving / setChartDisplayNameSaving`
  - `chartDisplayNameError / setChartDisplayNameError`
  - `handleSaveChartDisplayName`
  - `launchCompatibilityFromResult`（结果页唯一调用者已删除）
- [x] T7 前端 `ResultPage.css` 删除 `.chart-archive-tools / .chart-archive-kicker / .chart-archive-editor / .chart-archive-actions / .chart-archive-error` 五条规则及对应媒体查询分支
- [x] T8 手动验证：登录态起盘填写称呼后，HistoryPage 列表显示该 display_name；不填称呼时 ResultPage 不再出现命名编辑器
- [x] T9 openspec change folder 完成（proposal / tasks / specs）
