## 1. Backend Data Structures & Logic

- [x] 1.1 Create `jin_bu_huan_dict.go` in `pkg/bazi/` with initial structs (`JinBuHuanRule`, `JinBuHuanResult`)
- [x] 1.2 Populate `jin_bu_huan_dict.go` with base rules for Day Masters (e.g., 甲木, 乙木)
- [x] 1.3 Add a helper function `GetZhiDirection(zhi string) string` mapping 12 branches to 4 directions
- [x] 1.4 Implement `calcJinBuHuanDayun(dayGan, monthZhi, dayunZhi string)` matching logic

## 2. Backend Engine Integration

- [x] 2.1 Update `DayunItem` struct to include `JinBuHuan *JinBuHuanResult` with JSON tags
- [x] 2.2 Modify `calcDayun` in `engine.go` to invoke Jin Bu Huan calculation and attach to each pillar
- [x] 2.3 Run backend tests to ensure new JSON structure outputs correctly for `/api/bazi/calculate`

## 3. Frontend Timeline Integration

- [x] 3.1 Update `BaziResult` Typescript interface in frontend to include `jin_bu_huan` on `DayunItem`
- [x] 3.2 Update `DayunTimeline.tsx` to conditionally render `JinBuHuan` badges below Shishen data
- [x] 3.3 Add minimal CSS styling for "大吉", "吉", "平", "凶" badge variations
