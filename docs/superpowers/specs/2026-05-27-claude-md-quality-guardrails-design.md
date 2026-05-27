# CLAUDE.md 代码质量护栏补充 — 设计稿

**日期**：2026-05-27
**作者**：基于代码扫描证据的提议
**状态**：草案

---

## 1. 背景

当前 `CLAUDE.md`（64 行）是**纯通用 LLM 行为准则**，明确声明"请根据需要将其与特定项目的指令相结合"。项目特定规则分布在 `ENGINEERING.md`（53 行，毛泽东思想风格的工程准则）和 `AGENTS.md`（219 行，技术栈/目录/约定）。

用户提出疑问：现有规则是否覆盖代码复用、抽象、设计模式等代码质量维度？

## 2. 代码扫描结论

### 2.1 现有规则覆盖盘点

| 维度 | 已写在哪里 | 内容 |
|---|---|---|
| 反过度抽象 | CLAUDE.md §2 | 切勿为单次使用代码引入抽象层；切勿擅自增加灵活性/可配置性 |
| 反过度设计 | ENGINEERING.md §二 | 一个计数器不需要工厂模式 + 观察者模式 |
| 复用优先 | ENGINEERING.md §三 | 发现已有组件/工具函数必须复用 |
| 文件 500 行上限 | ENGINEERING.md §二 | 单文件超过 500 行未拆分是对代码库的犯罪 |
| 分层禁忌 | AGENTS.md §开发规范 | 严禁在 handler 中直接写 SQL；环境变量统一通过 Config |
| 编码前扫描 | ENGINEERING.md §一 | 必须先 grep / find 检索关键目录 |

**规则其实相当完整。** 但代码扫描显示这些规则正在被广泛违反。

### 2.2 真实违反证据（截至 2026-05-27）

#### A. 重复实现（违反"复用优先"）

- `frontend/src/lib/wuxingColorSystem.ts` 已存在五行色彩系统，但：
  - `frontend/src/pages/ResultPage.tsx:61` 重复定义 `GAN_WUXING` 天干→五行 Map
  - `frontend/src/components/DayunTimeline.tsx:55` 又复制一份
  - `frontend/src/components/LiuYueDrawer.tsx:28` 第三份
- `frontend/src/pages/ResultPage.tsx:30-44` 把 44 行的 `SHENSHA_POLARITY` 神煞极性表硬编码在页面里，未抽到 lib

#### B. 分层违规（违反 AGENTS.md "严禁 handler 写 SQL"）

- `backend/internal/handler/admin_handler.go:241-250` `AdminGetStats` 里直接 `database.DB.QueryRow()` 6 次做统计
- `backend/internal/handler/admin_handler.go:263-291` `AdminGetAIStats` 里 `rows.Query/Scan`
- `backend/internal/handler/admin_handler.go:315-340` `AdminGetUsers` 完整 SQL + 行迭代

#### C. 欠抽象（违反"三次以上即抽离"——但该规则尚未明文）

- `ShouldBindJSON` 错误返回三行模板在 handler 里出现 31 次
- 前端 `useState + useEffect + setLoading + setError` 数据加载模板在多个页面重复，未抽成 hook
- 后端 stats 统计的 `database.DB.QueryRow + Scan` 模式连续出现 3 次（admin_handler.go:241-250）

#### D. 文件体积超标（违反 ENGINEERING.md 500 行上限）

| 文件 | 行数 |
|---|---|
| `backend/internal/service/report_service.go` | 1493 |
| `frontend/src/pages/ResultPage.tsx` | 1332 |
| `frontend/src/pages/CompatibilityResultPage.tsx` | 981 |
| `frontend/src/components/PrintLayout.tsx` | 735 |
| `frontend/src/pages/admin/TokenUsagePage.tsx` | 702 |
| `frontend/src/lib/api.ts` | 694 |
| `backend/internal/service/report_service_test.go` | 664 |
| `backend/internal/service/compatibility_service.go` | 604 |
| `backend/internal/service/ai_client.go` | 547 |
| `backend/internal/handler/admin_handler.go` | 514 |

#### E. 命名/约定不一致（轻微）

- Admin 接口 handler 命名混乱：`admin_handler.go` 用 `AdminGetStats` 前缀，`admin_prompt.go` 却用 `GetPrompts / UpdatePrompt` 无前缀
- 前端 API 调用混合："对象方法"（`baziAPI.fetchLiuYue`）与"裸函数"（`fetchShenshaAnnotations`）共存

### 2.3 过度抽象（C）几乎不存在

仅发现 `backend/internal/service/cleanup_service.go:19` 的 `Clock` interface，是为单元测试设计的合理最小抽象。**说明 CLAUDE.md §2 反过度抽象的导向是有效的，导向另一边的规则才是缺口。**

## 3. 真正的缺口

不是"缺规则"，而是 CLAUDE.md（通用准则部分）存在三个**通用层面的语义空白**：

### 3.1 缺：复用 / 抽象的可量化阈值

§2 说"切勿为单次使用引入抽象"，但**没有说什么时候应该抽**。结果是 LLM 倾向于解读成"能不抽就不抽"，导致欠抽象。

### 3.2 缺：§2 与"复用优先"的张力消解

§2（CLAUDE.md）和"复用优先"（ENGINEERING.md）单独看都对，**合起来读 LLM 会困惑**：
- 看到一段类似已有函数的逻辑，要复用还是新写？
- 看到 2 处复制，要不要抽？看到 3 处呢？
- 这个张力需要一个明确的判定线。

### 3.3 缺：§3 与"500 行上限"在遗留代码上的冲突

§3（精准修改）说"切勿重构本身并未损坏的代码"。但当目标文件已经 1493 行，**新增一行也是在恶化体积问题**。LLM 没有明确指引：到底是先拆再加，还是直接加？

## 4. 设计方案

### 4.1 选择的方案

**方案 A：在 CLAUDE.md 末尾（§4 之后、总结段之前）新增 §5「代码质量护栏」**，只补上述 3 个通用缺口。

理由：
- 保持 CLAUDE.md 的通用性（不掺项目细节）
- 不重复 ENGINEERING.md / AGENTS.md 已有内容
- 修补的是 LLM 可被推理出的歧义点，而非项目知识

### 4.2 拒绝的备选

- **方案 B（原地强化既有条款）**：会让 §2 内部出现"不要抽象 / 必须抽象"的字面矛盾，需要重新组织语境，改动面大
- **方案 C（拆成多文件）**：与项目已有 ENGINEERING.md / AGENTS.md 的分工冲突
- **直接修改 ENGINEERING.md**：项目特定规则其实已经全在，缺的是通用层面的判定规则——属于 CLAUDE.md 职责

### 4.3 §5 提议内容（最终落地文本）

```markdown
## 5. 代码质量护栏（与项目无关的通用补充）

§2 反对"无中生有的抽象"，本节补充"何时不抽是错的"——两者依靠**阈值**和**层次**区分，并不矛盾。

### 5.1 复用与抽象的阈值

- **复用优先**：动手写任何工具函数、常量表、组件、SQL 语句之前，先用 `grep` 在相关目录搜一遍是否已经存在。已存在就复用；找不到再写。
- **第 3 次复制即抽离**：第 1、2 次复制保持原样（避免过早抽象），第 3 次出现时停下来抽出公共函数 / 组件 / hook，并把前两处一起迁移过去。
- **抽象的最低标准**：抽出来的单元必须有清晰的单一职责、有意义的命名、稳定的接口。如果只是为了"少写几行"而抽出意义不明的 helper，不如不抽。

### 5.2 分层与边界

- 项目本身的分层约束（哪一层可以做什么）记录在项目级文档中（如 `ENGINEERING.md` / `AGENTS.md` / 项目根 `CLAUDE.md`）。**编码前必读这些文件**，未确认分层规则之前不要新增代码。
- 破坏分层（例：把本应在数据访问层的逻辑塞进控制器层、把本应在配置层的环境变量读取散落到业务代码里）一律视为 Bug，即使"只用一次"也不允许。

### 5.3 在遗留代码中工作

§3 的"精准修改"针对**结构尚可**的代码。当待修改文件本身已经违反项目硬约束（如文件行数上限、单一职责、分层规则），处理顺序如下：

1. **先指出**：在动手前明确告知用户"此文件已超过 X 行 / 已混合多种职责"，征求处理方式
2. **优先拆分**：用户同意后，先做最小拆分把新增功能放进合理位置（新文件或合适的既有层），不要"再加一点点就好"
3. **不允许借口**：现有违反规则的代码不构成新增违反规则代码的理由

### 5.4 命名与返回值一致性

修改文件时观察周围 5-10 个同类函数 / 接口的命名、返回值约定、错误处理风格，**对齐它们**而不是发明新写法。如果发现项目内部不一致（同一类操作存在多套风格），向用户指出但不擅自统一。
```

## 5. 不在本次范围内（后续跟进项）

以下属于代码层面的具体债务，**不通过修改 CLAUDE.md 解决**，需要后续单独立项：

| 债务 | 建议处理方式 |
|---|---|
| `admin_handler.go` 三处 SQL 违规 | 抽到 `admin_repository.go` 的 `CountByDay / ListUsersPaged` 方法 |
| `GAN_WUXING` 三处重复 | 抽到 `frontend/src/lib/wuxingColorSystem.ts` 并迁移 3 处引用 |
| `SHENSHA_POLARITY` 44 行硬编码 | 抽到 `frontend/src/lib/shenshaPolarity.ts` |
| 10 个超过 500 行的文件 | 按 5.3 流程，**新增功能时**逐步拆分；不做一次性大重构 |
| Handler 命名前缀不一致 | 遇到时顺手对齐为 `Admin*` 前缀 |

这些项可在后续会话中由用户按优先级择机推进，不构成本次 CLAUDE.md 修改的前置条件。

## 6. 验证标准

CLAUDE.md 修改后，**新会话中** Claude 应当能够：
- 看到一段似曾相识的代码，会先 grep 而不是直接写
- 看到 3 处重复代码，会主动提议抽离
- 改动 1493 行的文件时，会先提示用户"建议先拆分"而不是默默加行
- 修改 handler 时不会再写新的 SQL 调用

（无法用单测验证，需通过实际使用观察）
