## Why

目前 `AlgoConfigPage` 的"吉凶判定参数"表格将 `llm_price_*` 六个价格 key 和算法参数混排，且直接展示原始 key 名（如 `llm_price_flash_input`），管理员无法直观辨认各档位的输入/输出价对，价格管理体验差。

## What Changes

- 在 `AlgoConfigPage.tsx` 的通用参数表格中过滤掉 `llm_price_*` 前缀的 key，避免混排。
- 新增"LLM 模型定价"独立区块：以三行结构化表格展示 flash / pro / default 三个档位，每行显示档位名称、适配模型关键词说明、输入价、输出价，支持行内编辑（同时编辑输入/输出两个值）。
- 纯前端改动，不涉及后端接口变更（复用现有 `PUT /api/admin/algo-config/:key` 端点）。

## Capabilities

### Modified Capabilities
- `algo-config-ui`（已有）：算法参数管理页面，新增 LLM 定价专属区块，与调候规则区块并列。

## Impact

仅影响 `frontend/src/pages/admin/AlgoConfigPage.tsx`，对其他页面、后端逻辑、用户侧功能无任何影响。
