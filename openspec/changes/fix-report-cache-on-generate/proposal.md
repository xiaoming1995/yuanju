# 修复：生成解读接口不应读缓存

## 问题

用户点击「生成解读」时，后端生成接口（POST）内部优先检查缓存——只要库里有记录就直接返回，LLM 从未被调用。切换模型后仍返回旧内容，用户无法获得新模型的解读。

## 根因

缓存的正确用途是**历史记录查看**（GET），而非拦截**主动生成**（POST）。目前两者混用了同一条缓存读取逻辑。

## 目标

**生成接口（POST）永远调 LLM，生成完覆盖存库。**
**历史查看接口（GET /history/:id）保持不变，仍从库中读取。**

## 改动范围

**后端（`backend/internal/service/report_service.go`）：**

| 函数 | 操作 |
|------|------|
| `GenerateAIReport` | 删除开头的缓存检查（3 行） |
| `GenerateAIReportStream` | 删除开头的缓存检查（4 行） |
| `GenerateLiunianReport` | 删除开头的缓存检查（3 行） |
| `GeneratePastEventsStream` | 删除开头的缓存检查（3 行） |
| `GenerateDayunSummariesStream` | 删除每段大运循环内的缓存命中分支 |

历史查看、管理后台等读库路径**不做任何修改**。
