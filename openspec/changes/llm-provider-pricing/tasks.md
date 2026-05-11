## 1. 后端：数据库迁移

- [x] 1.1 在 `pkg/database/database.go` 的 `Migrate()` 中添加：
  ```sql
  ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS input_price_cny DECIMAL(10,4) DEFAULT 1.0;
  ALTER TABLE llm_providers ADD COLUMN IF NOT EXISTS output_price_cny DECIMAL(10,4) DEFAULT 2.0;
  ```

## 2. 后端：Model struct 更新

- [x] 2.1 在 `internal/model/admin.go` 的 `LLMProvider` struct 中新增：
  ```go
  InputPriceCny  float64 `json:"input_price_cny"`
  OutputPriceCny float64 `json:"output_price_cny"`
  ```

## 3. 后端：Repository 更新

- [x] 3.1 在 `internal/repository/admin_repository.go` 的 `GetLLMProviders` 查询中，SELECT 增加 `input_price_cny, output_price_cny` 并在 Scan 中读取
- [x] 3.2 在 `CreateLLMProvider` 的 INSERT 语句中增加这两个字段及对应参数
- [x] 3.3 在 `UpdateLLMProvider` 的 UPDATE 语句中增加这两个字段及对应参数
- [x] 3.4 新增 `GetPriceByModel(model string) (inputPrice, outputPrice float64, err error)` 函数，查询 `SELECT input_price_cny, output_price_cny FROM llm_providers WHERE model = $1 LIMIT 1`

## 4. 后端：定价服务更新

- [x] 4.1 修改 `internal/service/llm_pricing.go` 的 `GetModelPrice(modelName string)`：改为调用 `repository.GetPriceByModel(modelName)`，若 `err != nil` 或查询结果为空则 fallback 到原有常量（1.0 / 2.0），删除原有的 `matchModelTier` 和 `algo_config` 查询逻辑

## 5. 后端：Handler 更新

- [x] 5.1 在 `internal/handler/admin_handler.go` 的 Provider 创建/更新请求体 struct 中新增 `InputPriceCny float64 \`json:"input_price_cny"\`` 和 `OutputPriceCny float64 \`json:"output_price_cny"\``，并传递给对应 repository 函数

## 6. 前端：类型和 API Client 更新

- [x] 6.1 在 `frontend/src/lib/adminApi.ts` 的 `Provider` interface（或相关类型）中新增 `input_price_cny: number` 和 `output_price_cny: number`

## 7. 前端：AdminLLMPage 更新

- [x] 7.1 在 Provider 列表表格的表头新增「输入价（¥/百万）」和「输出价（¥/百万）」两列
- [x] 7.2 在 Provider 列表的每行 `<td>` 中渲染 `p.input_price_cny` 和 `p.output_price_cny`（保留两位小数）
- [x] 7.3 在 `initialForm` 中新增 `input_price_cny: 1.0` 和 `output_price_cny: 2.0`
- [x] 7.4 在添加/编辑 Modal 中新增两个 `<input type="number" step="0.01">` 字段，分别绑定 `form.input_price_cny` 和 `form.output_price_cny`
- [x] 7.5 `handleSave` 中将这两个字段传入 `adminLLMAPI.create` / `adminLLMAPI.update` 的请求体

## 8. 验证

- [x] 8.1 运行 `go build ./...` 确认后端无编译错误
- [ ] 8.2 重启后端服务，确认 `llm_providers` 表已新增两列且存量行有默认值 1.0 / 2.0
- [ ] 8.3 在 AdminLLMPage 添加一个新 Provider（如 Kimi），填入输入价 `0.12`、输出价 `0.12`，保存后列表中正确显示
- [ ] 8.4 编辑该 Provider，将输出价改为 `0.15`，保存后列表更新
- [ ] 8.5 通过 token 用量统计页确认该模型的 `estimated_cost_cny` 按新价格计算
