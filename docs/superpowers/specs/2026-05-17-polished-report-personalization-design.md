# 命理解读「润色版」个性化设计

**日期**：2026-05-17
**范围**：在原版 AI 命理解读基础上，新增「润色版」 tab —— 用户输入当前情况后，AI 基于原版报告 + 用户描述，逐章重写一份贴近用户处境的版本。
**目的**：缓解原版报告「太泛、太冷冰冰」的体感，给用户一份"师傅听完你的话后写的信"。

## 背景

当前 AI 命理解读输出是基于八字算法 + 通用 prompt 生成的"专业批断风格"，用户反馈三个症状：
- 太冷冰冰、不像人在说话
- 重复、套路、结构化过重
- 太泛、谁看都适用

原计划用 prompt 改造（去 AI 味方案 B）单点修复，但用户真正想要的是**新功能**：让用户输入当下场景，生成针对该场景的润色版。原版不动，润色版是额外产物。

## 设计决策

### 1 · UI 流程

ResultPage 的「AI 命理解读」区顶部加 tab 切换：

```
◉ 原版    ○ 润色版
```

- **原版 tab**：保持当前显示逻辑不变（章节渲染 + 行动建议 / 词典 / 工具栏 ...）
- **润色版 tab**：
  - 未生成过：显示输入引导（自由文本框 300 字 + 「生成润色版」按钮）
  - 已生成过：显示「你的情况描述: ..."[修改][重新润色]」+ 润色版 5 章内容

### 2 · 触发前提

润色版 **依赖原版报告存在**。如果原版还没生成，点润色版 tab 提示「请先生成原版命理解读，再尝试润色」。

### 3 · 输入

- 单个 textarea
- 占位提示：「例：今年在考虑跳槽 / 跟对象有点摩擦 / ...」
- 长度限制：**20-300 字**（client + server 校验，下限避免一句话场景信息量太少）
- 必填，不可空
- 字符过滤：保留中文、英文、数字、标点；不限制内容主题（让用户自由）

### 4 · 后端数据流

```
POST /api/bazi/polished-report/:chart_id
  body: { user_situation: string }

Service:
  1. 校验 chart 存在、属于当前用户、原版 ai_report 存在
  2. 校验 user_situation: 非空、≤300 字
  3. 构建 polish prompt（原版章节 + 用户输入 + 命理数据 + 指令）
  4. 调用 LLM（与原版同 provider/model）
  5. 解析 markdown → 拆 5 章
  6. UPSERT 到 ai_polished_reports（chart_id 唯一）
  7. 返回 { content, content_structured, user_situation }

GET /api/bazi/polished-report/:chart_id
  - 读 ai_polished_reports，无则返回 404 / null
```

### 5 · 数据库

新增表 `ai_polished_reports`：

```sql
CREATE TABLE ai_polished_reports (
  id                  SERIAL PRIMARY KEY,
  chart_id            VARCHAR(64) NOT NULL UNIQUE REFERENCES bazi_charts(id),
  user_situation      TEXT NOT NULL,
  content             TEXT NOT NULL,
  content_structured  JSONB,
  model               VARCHAR(128),
  provider_id         INT,
  prompt_tokens       INT DEFAULT 0,
  completion_tokens   INT DEFAULT 0,
  total_tokens        INT DEFAULT 0,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ai_polished_reports_chart ON ai_polished_reports(chart_id);
```

DDL 加到 `backend/pkg/database/database.go`。

UNIQUE (chart_id) 约束 → 一个命盘只有一份最新润色版，重生即 UPSERT。

### 6 · Prompt 大致结构

```
你将基于以下命理报告，结合用户的当下情况，逐章重写一份贴近用户处境的版本。

[原版报告 markdown 全文]

[命理数据]
  - 用神 / 忌神
  - 命格 / 调候 / 大运 / 神煞
  - 当前年份、当前大运

[用户当下情况]
"{user_situation}"

[改写要求]
- 保持 5 章结构（性格特质/感情运势/事业财运/健康提示/大运走势）
- 每章 200-300 字
- 每章首段引出与用户情况相关的事，再带回命理依据
- 第二人称「你」叙述、师傅口吻、避免报告体
- 不能改变命局结论（用神/格局/十神等）
- 已禁词同原版：百分比 / 加权 / 由此可见 / ...
- 输出 markdown ## 章节标题

输出格式（一字不差）：
## 【性格特质】
...
## 【感情运势】
...
## 【事业财运】
...
## 【健康提示】
...
## 【大运走势】
...
```

### 7 · 前端实现

**ResultPage.tsx 改动**（新增 ~120 行）：

```tsx
// 顶部加 state
const [reportTab, setReportTab] = useState<'original' | 'polished'>('original')
const [polishedReport, setPolishedReport] = useState<PolishedReport | null>(null)
const [userSituation, setUserSituation] = useState('')
const [polishing, setPolishing] = useState(false)

// 进入 chart 时同时拉原版 + 润色版
useEffect(() => {
  baziAPI.getPolishedReport(chartId).then(setPolishedReport).catch(() => {})
}, [chartId])

// Tab 渲染
<div className="report-tab-row">
  <button className={tab === 'original' ? 'active' : ''} onClick={() => setReportTab('original')}>原版</button>
  <button className={tab === 'polished' ? 'active' : ''} onClick={() => setReportTab('polished')}>润色版</button>
</div>

// 润色版 panel
{reportTab === 'polished' && (
  <PolishedPanel
    polishedReport={polishedReport}
    userSituation={userSituation}
    onChange={setUserSituation}
    onSubmit={generatePolished}
    loading={polishing}
  />
)}
```

新组件 `PolishedPanel.tsx` 约 80-100 行。

### 8 · 后端实现

**新文件**：
- `backend/internal/handler/polished_report_handler.go` (~120 行)
- `backend/internal/service/polished_report_service.go` (~150 行)
- `backend/internal/repository/polished_report_repository.go` (~80 行)
- `backend/internal/model/polished_report.go` (~30 行)

**修改**：
- `backend/cmd/api/main.go` — 注册 2 个新路由
- `backend/pkg/database/database.go` — 加表 DDL

## 不在范围

- 多份润色版历史 / 不同场景 切换 —— UNIQUE chart_id 约束，只存最新
- 流式生成（润色版一次性返回，不走 SSE）
- 润色版 PDF 导出 —— 先做屏幕展示，PDF 后续考虑
- 词典 / 工具栏 在润色版里展示（沿用原版）
- 提示词模板的 admin UI 配置（润色 prompt 写死在 service 里）
- 移动端独立交互
- 输入内容审核 / 敏感词过滤（信任用户）
- 多 chart 间润色版分享 / 跨 chart 引用

## 风险与缓解

| 风险 | 缓解 |
|---|---|
| LLM 输出格式不符（章节标题错） | 复用现有 `ParseMarkdownToStructured` 解析，校验 5 章齐全；缺章 fallback 走原文展示 |
| 用户输入太短 / 无具体场景 | 长度校验 ≥ 20 字、prompt 里加「如果情况描述过简，按一般人当下处境写」 |
| 输入含敏感个人信息 | 后端记日志时脱敏；DB 加 user_situation 索引时不索引内容 |
| 润色版与原版命理结论不一致 | prompt 显式约束「不可改变用神 / 格局 / 十神等命局结论」 |
| Token 超限 | 监控 input + output token，估算 8-13K，主流模型支持 |
| 单 chart 一份润色版限制太严 | 接受先做最简化，后续按反馈决定是否多版本 |

## 验收

**自动化**：
- backend `go test ./internal/service/...` 通过（含新 polished 测试）
- frontend `tsc -b --force` 0 errors
- 现有 55 tests 不退化

**人工**：
- 任选一份命盘，原版已生成
- 切到润色版 tab → 输入「最近考虑跳槽，想转行做创意类工作」→ 生成
- 验证润色版 5 章都围绕「跳槽 / 创意类工作」展开，引述命理依据
- 改输入再生成 → 内容变化合理
- 检查 DB ai_polished_reports 表数据正确

## 工作量

```
后端
─────────────────────────────────────────────────────────────────
新 handler / service / repository / model            ~380 行
DDL migration                                        ~15 行
prompt builder                                       ~80 行

前端
─────────────────────────────────────────────────────────────────
ResultPage tab + state                               ~60 行
PolishedPanel 组件                                   ~80 行
API client (baziAPI.getPolishedReport / .generate)   ~20 行

总: 8-15 个 task、单 PR、估 4-8 小时实施
```

## 后续可能扩展（非本次）

- 多场景润色（事业 / 感情 / 健康 三份独立润色）
- 润色版 PDF 导出（与原版共用样式）
- 用户对润色质量的反馈 / 评分
- 流式生成（章节级 SSE 一章一章出）
- 多模型对比（同一情况下用不同模型生成对照）
