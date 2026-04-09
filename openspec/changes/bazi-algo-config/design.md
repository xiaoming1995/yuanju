## Context

`pkg/bazi/` 是纯算法层，无 DB 依赖，所有判断参数硬编码在 Go 源文件中：

- `dayun_jixiong.go`：极寒极热阈值（`Huo==0 && Shui+Jin>=4`）、身强判定线（`40%`）
- `tiaohou_dict.go`：120 条调候用神规则（Go `map[string]map[string]string`）

管理员后台已有 `ai_prompts` 表（module+content 文本），但仅用于拼装 LLM System Prompt，算法层无法读取。

## Goals / Non-Goals

**Goals:**
- 将算法数值参数（阈值、权重）迁移至 DB，管理员可在不发版的情况下调整
- 将 120 条调候用神规则迁移至 DB，支持按日干+月支维度的 CRUD
- `pkg/bazi/` 保持纯函数、零 DB 依赖——所有 DB 读取在服务层完成并注入缓存
- DB 无配置时自动 fallback 至现有硬编码默认值，保证零回归

**Non-Goals:**
- 金不换大运字典（120 条，复杂度高，留后续迭代）
- 用神推断权重精细化（留后续）
- 算法规则的版本历史与回滚（留后续）

## Decisions

### 决策 1：缓存层由 `internal/service` 持有，而非 `pkg/bazi`

**选择**：在 `internal/service/algo_config_service.go` 维护一个全局内存缓存（`sync.RWMutex` 保护），服务启动时加载，算法层通过函数注入（`bazi.SetAlgoConfig(cfg)`）接收配置，而非自己读 DB。

**理由**：`pkg/bazi` 是独立算法库，保持无副作用便于单元测试；DB 变更逻辑集中在 `internal/`，符合项目现有分层约定。

**备选方案**：`pkg/bazi` 直接读 DB → 破坏层级，且每次计算都有 IO 开销，否决。

---

### 决策 2：两张独立表，而非通用 key-value

**选择**：
- `algo_config`：`key VARCHAR PRIMARY KEY, value TEXT, description TEXT`（存数值参数）
- `algo_tiaohou`：`day_gan CHAR(1), month_zhi CHAR(1), xi_elements TEXT, text TEXT, PRIMARY KEY(day_gan, month_zhi)`

**理由**：`algo_tiaohou` 有固定结构（日干+月支复合主键），通用 key-value 无法高效做 CRUD，且前端表格渲染需要结构化字段。`algo_config` 参数少，key-value 足够。

---

### 决策 3：DB 规则优先，硬编码为 fallback

**选择**：缓存加载时合并：先取硬编码默认值，再用 DB 覆盖。DB 中不存在某条规则，则使用硬编码值。

**理由**：保证任何时候（DB 清空、新部署）算法行为与现在一致，零回归风险。管理员只需维护「与默认值不同」的条目。

---

### 决策 4：缓存热更新 API

**选择**：新增 `POST /api/admin/algo-config/reload` 端点，管理员修改配置后手动或自动触发内存缓存重载，无需重启服务。

**理由**：算法计算高频，缓存不能每次请求都查 DB；热更新在修改规则后立即生效，体验好于需重启。

## Risks / Trade-offs

- **[风险] 缓存与 DB 短暂不一致** → 管理员修改后点击「重载」即可，文档说明即可
- **[风险] 调候规则格式录入错误（天干写错字）** → 后端做校验（天干只能是 10 个字符之一），前端提供下拉选择
- **[风险] 120 条数据初始化工作量** → 提供一次性 seed 脚本，从现有硬编码 Go map 自动生成 SQL INSERT

## Migration Plan

1. 新增两张表（DDL in `pkg/database/database.go`）
2. 服务启动时，若 `algo_tiaohou` 表为空，自动从硬编码字典 seed 全量数据（保持历史行为）
3. 算法层通过 `SetAlgoConfig` 注入配置，fallback 到硬编码不变
4. 前端后台新增页面，管理员可逐条查看与修改

**回滚**：删除两张表、移除注入调用，恢复硬编码路径，零风险。
