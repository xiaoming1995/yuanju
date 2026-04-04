## Context

在目前的系统中，八字计算模块（`bazi_engine.go`）被设计为不做硬编码的喜忌推算，相关字段留空以兼容未来的扩展，真正的喜忌推理全部移交到了通过大语言模型进行解读的环节。但当前大语言模型的输出仅为包含文字段落的 Markdown 报告，导致这些分析出的属性未进行结构化截取，无法点亮前端的界面。

## Goals / Non-Goals

**Goals:**
- 从 AI 输出中精准且具有高容错率地解析出《喜用神》和《忌神》。
- 解析后将这些核心属性写入 `bazi_charts` 的数据库记录中。
- 保障接口格式稳定返回，即使在 AI 发生“幻觉”（不按规矩返回 JSON）时也能最大程度不引发服务错误。
- 前端页面在未获得结果前妥善呈现“计算中”样式。

**Non-Goals:**
- 不重新在本地计算服务实现这套推断流程。
- 不引入任何外部专用的 JSON Schema Parser 和 Validation（比如依赖第三方大型类库），纯使用原生正则及原生 `encoding/json`。

## Decisions

### 1. 修改系统 Prompt 强制 JSON 化
将 `report_service.go` 的 prompt 修改为让 AI 最终输出一段标准的 JSON。结构要求如下：
```json
{
  "yongshen": "木火",
  "jishen": "金水",
  "report": "【性格特质】...\n..."
}
```

### 2. 构建安全的 JSON 文本截取器
考虑到 LLM 容易“自作多情”，返回可能会包裹在 ` ```json ` 块中或在最外层乱加说话头（例如“好的！这是您的...”），必须使用通过字符串查找或简单的正则提取（先找到首个 `{` 并找到最后一个 `}`）将其有效截取成单纯的 JSON 字符串送去 `Unmarshal`。

### 3. 数据层更新与模型调整
由于生成报告期间 `chart` 早已创建（在 `bazi_handler` 内保存），`report_service` 获取并解封 JSON 后将通过 `chart_repository.go` 中新增或原有的 `UpdateChartFields` 方法（专门负责只更新特定属性如 yongshen 和 jishen 避免全量覆盖冲突）来进行变更持久化。

## Risks / Trade-offs

- **Risk: AI 始终无法按 JSON 输出（结构毁坏）**
  - **Trade-off/Mitigation**: 提供无损降级逻辑（Fallback）。当 `Unmarshal` 解析失败时，不再继续提取特征，只将捕获到的 AI 返回视作纯文本（赋值给 `report_content`），这样前端尽管依然不能呈现特定的属性徽章，依然能够呈现长报告。
