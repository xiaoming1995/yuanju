## 验证记录（2026-11-04）

- `node --test miniapp/tools/*.test.js`：111 项通过。
- `openspec validate simplify-miniapp-date-selection --strict`：通过。
- 微信开发者工具 `compile_wxml` 与 `compile_wxss`：成功；`simulator_open_page` 成功打开 `pages/moving-date/index`。
- 模拟器 534 × 1156 截图检查了默认页和展开首个日期后的页面：事项、范围摘要、三组数量及简洁日期条目顺序清楚；标签换行、日期详情和底部安全区未见截断。截图分别保存在本机 `/tmp/yuanju-date-selection.png` 和 `/tmp/yuanju-date-selection-detail.png`。
- 未在真机或更窄的模拟器尺寸上核对。
