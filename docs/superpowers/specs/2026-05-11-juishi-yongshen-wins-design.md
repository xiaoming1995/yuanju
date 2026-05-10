# 三合/三会局势力判断——用神赢信号设计

## 背景

过往事件推算模块（`event_signals.go`）目前的 Layer 0 信号（`collectYingqiSignals`）只检测流年/大运地支与原局干支位的**单对单**刑冲克穿交互。当流年地支与大运地支、原局地支**共同凑成三合/三会局**，且该局五行属用神并克制忌神时，当前算法无法识别"用神势力大，赢"这一吉象；反向（忌神局成）亦无法标注极凶信号。

## 目标

新增 `collectJuShiSignals` 函数：
- 检测流年地支（结合大运+原局地支）是否补全三合/三会局
- 判断局五行方向：用神局克忌 → 吉；忌神局 → 极凶（标星）
- 产生的信号归入 Layer 0，参与已有 Layer 0 vs Layer 4 压制规则

## 非目标

- 不处理半合（两支凑成）的势力判断，需三支全齐
- 不处理天干五合化局（仅处理地支三合/三会）
- 不修改现有 `isSanheTriggered`/`isSanhuiTriggered` 在 `GetYearEventSignals` 中的十神婚恋逻辑

## 函数签名

```go
func collectJuShiSignals(natal *BaziResult, lnZhi, dyZhi string) []EventSignal
```

### 集成位置

在 `GetYearEventSignals` 开头，与 `collectYingqiSignals` 合并为 Layer 0：

```go
layer0Sigs := collectYingqiSignals(natal, lnGan, lnZhi, dyGan, dyZhi)
layer0Sigs = append(layer0Sigs, collectJuShiSignals(natal, lnZhi, dyZhi)...)

layer0HasXiong := false
for _, s := range layer0Sigs {
    if s.Polarity == PolarityXiong {
        layer0HasXiong = true
        break
    }
}
signals = append(signals, layer0Sigs...)
layer0End := len(signals)
```

## 局成立条件

**existingZhi** = 原局四柱地支 + 大运地支（不含流年地支）：

```go
existingZhi := []string{natal.YearZhi, natal.MonthZhi, natal.DayZhi, natal.HourZhi}
if dyZhi != "" {
    existingZhi = append(existingZhi, dyZhi)
}
```

**三合局**（四组，需三支全齐）：

| 局名 | 三支 | 五行 |
|------|------|------|
| 申子辰 | 申、子、辰 | 水 |
| 寅午戌 | 寅、午、戌 | 火 |
| 亥卯未 | 亥、卯、未 | 木 |
| 巳酉丑 | 巳、酉、丑 | 金 |

**三会局**（四组，需三支全齐）：

| 局名 | 三支 | 五行 |
|------|------|------|
| 寅卯辰 | 寅、卯、辰 | 木 |
| 巳午未 | 巳、午、未 | 火 |
| 申酉戌 | 申、酉、戌 | 金 |
| 亥子丑 | 亥、子、丑 | 水 |

判断逻辑（三合/三会相同结构）：

```
for 每个局组:
    if lnZhi ∈ 该组:
        other2 = 该组其余两支
        if other2[0] ∈ existingZhi AND other2[1] ∈ existingZhi:
            局成立，localWx = 该局五行
```

`matchCount >= 2`（三支全齐），比现有 `isSanheTriggered` 的 `matchCount >= 1`（半合）更严格。

## 势力判断与信号输出

```
localWx = 局五行（pinyin）

Case 1：用神赢（吉）
  条件：yongSet[localWx]  AND  wxKe[localWx] ∈ jiSet
  Type     = "综合变动"
  Polarity = 吉
  Source   = SourceZhuwei
  Evidence = "流年{lnZhi}补全{局名}{局类型}{用神CN}局，
              {用神CN}势力大增，克制忌神{忌神CN}，用神赢，应期吉"

Case 2：忌神局（极凶）
  条件：jiSet[localWx]
  Type     = "局势_重"   ← 新 Type 常量
  Polarity = 凶
  Source   = SourceZhuwei
  Evidence = "★流年{lnZhi}补全{局名}{局类型}{忌神CN}局，
              忌神势力极强，用神承压，应期极凶"

其余（局成但五行无克忌关系）：静默跳过，不产生信号
```

## 新增常量

```go
const TypeJuShiZhong = "局势_重"  // 三合/三会忌神局极凶标星信号
```

## 测试场景

| 测试 | 场景 | 期望 |
|------|------|------|
| `TestJuShi_YongWins` | 用神=火；原局寅+戌；流年午 → 三合火局；忌神=金（火克金） | 吉，Evidence 含"用神赢" |
| `TestJuShi_JiXiong` | 忌神=水；原局申+辰；流年子 → 三合水局 | 凶，Type=局势_重，Evidence 含"★" |
| `TestJuShi_HalfHe_NoSignal` | 原局只有寅；流年=午（缺戌）→ 半合 | 无局势力信号 |
| `TestJuShi_NoKe_NoSignal` | 用神=火；三合火局成；忌神=土（无克） | 无信号，静默 |

所有测试放入 `backend/pkg/bazi/event_signals_test.go`。

## 影响范围

- `backend/pkg/bazi/event_signals.go`：新增 `TypeJuShiZhong` 常量、`collectJuShiSignals` 函数；修改 `GetYearEventSignals` Layer 0 收集块
- `backend/pkg/bazi/event_signals_test.go`：新增 4 个测试
