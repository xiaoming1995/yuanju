## 1. 原生数据结构扩展 (Model Extension)

- [x] 1.1 在 `backend/pkg/bazi/engine.go` 内找到 `LiuNianItem`，为其新增 `IsTransition` (bool), `TransMonth` (int), `TransDay` (int), 和 `PrevDayun` (string) 字段及对应的 `json` tag。

## 2. 核心排盘逻辑升级 (Engine Logic Update)

- [x] 2.1 在 `backend/pkg/bazi/engine.go` 的 `Calculate` 方法中，准备提取或缓存上一个大运干支 `prevDayunGanzhi` 的变量。
- [x] 2.2 在遍历 `yun.GetDaYun()` 以及其内嵌的 `dy.GetLiuNian()` 时，增加首行断层逻辑：当流年索引为 `0` 时，将该流年的 `IsTransition` 设为 `true`。
- [x] 2.3 获取 `startSolar.GetMonth()` 和 `startSolar.GetDay()`，赋值给首个流年记录。
- [x] 2.4 在流年索引为 `1~9` 时，将 `IsTransition` 设为 `false`。

## 3. 测试与验证 (Testing & Validation)

- [x] 3.1 运行或扩展 `backend/pkg/bazi/engine_test.go`。
- [x] 3.2 增加断言验证某个实际命盘返回的 `DayunItem` 列表中，大运首年是否成功挂载 `is_transition` 为 `true`，且 `trans_month` / `trans_day` 与 `start_yun_solar` 的预期一致。
