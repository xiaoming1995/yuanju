## Why

To match professional Bazi software (like Wenzhen Bazi), our calculation engine currently lacks three critical indicators: Xingyun (Day Master's 12 life stages relative to each pillar's branch), Xun Kong (Void branches based on the Day or Year pillar), and Shensha (Gods and Demons). Adding these data layer points is a prerequisite for building a high-density, professional "data grid" on the frontend.

## What Changes

1. Extend the `BaziResult` backend struct to include `xing_yun` (星运), `xun_kong` (空亡), and `shen_sha` (神煞) for the Four Pillars.
2. Utilize `lunar-go`'s robust built-in astronomical functionality (`GetXunKong()`, `GetYearShenSha()`, `GetDayShenSha()`, and `LunarUtil.GetZhiXing(dayGan, branch)`) to accurately populate these fields during the `Calculate()` function execution.
3. Ensure these arrays/strings are safely exposed via the `/api/bazi/calculate` JSON output without breaking existing UI.

## Capabilities

### New Capabilities
- `bazi-advanced-data`: Accurate calculation and delivery of Xingyun, Xun Kong, and Shensha indicators for the Four Pillars.

### Modified Capabilities

## Impact

- **Affected Code**: `backend/pkg/bazi/engine.go`, `backend/internal/model/model.go`
- **APIs**: Restful payload for `POST /api/bazi/calculate` will seamlessly expand.
- **Frontend**: Currently no UI breaking changes, as adding new data simply prepares the payload for the future UI Grid refactor.
