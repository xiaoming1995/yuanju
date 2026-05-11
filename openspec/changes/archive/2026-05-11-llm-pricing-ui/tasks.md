## 1. 过滤通用表格中的价格 key

- [x] 1.1 在 `AlgoConfigPage.tsx` 中，将渲染"吉凶判定参数"表格时的 `params` 过滤为 `params.filter(p => !p.key.startsWith('llm_price_'))`，使价格 key 不再出现在通用表格里

## 2. 新增 LLM 模型定价区块

- [x] 2.1 在组件内定义 `PRICE_TIERS` 常量，描述三个档位的元数据：
  - `flash`：显示名"Flash"，匹配说明"v4-flash / chat"，inputKey `llm_price_flash_input`，outputKey `llm_price_flash_output`
  - `pro`：显示名"Pro"，匹配说明"v4-pro / reasoner"，inputKey `llm_price_pro_input`，outputKey `llm_price_pro_output`
  - `default`：显示名"默认"，匹配说明"其他模型 fallback"，inputKey `llm_price_default_input`，outputKey `llm_price_default_output`

- [x] 2.2 新增 `editingTier`（`string | null`）、`editInput`、`editOutput`（`string`）三个 state，用于控制价格行的编辑状态

- [x] 2.3 在调候规则区块上方新增"LLM 模型定价"`<section>`，渲染包含以下列的表格：档位、适配模型、输入价（¥/百万tokens）、输出价（¥/百万tokens）、操作

- [x] 2.4 每行非编辑态：从 `params` 中查找对应 inputKey / outputKey 的值显示，操作列显示"编辑"按钮

- [x] 2.5 点击"编辑"进入编辑态：输入价和输出价各自变为 `<input type="number">`，操作列变为"保存"/"取消"

- [x] 2.6 实现 `handleSaveTierPrice(tier)` 函数：依次调用 `adminAlgoConfigAPI.update(inputKey, { value: editInput })` 和 `adminAlgoConfigAPI.update(outputKey, { value: editOutput })`，保存后调 `fetchParams()` 刷新，并退出编辑态

## 3. 验证

- [ ] 3.1 启动前端 dev server，进入管理后台算法配置页，确认通用参数表格不再出现 `llm_price_*` 行
- [ ] 3.2 确认"LLM 模型定价"区块正确显示三个档位及当前价格值
- [ ] 3.3 修改 Flash 档输入价为 `0.30`，保存后确认表格值更新，并通过 `GET /api/admin/algo-config` 确认数据库已写入
- [ ] 3.4 点击取消，确认价格值恢复原样未被修改
