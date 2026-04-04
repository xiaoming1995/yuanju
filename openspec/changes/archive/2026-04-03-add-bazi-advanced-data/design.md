## Context

The current `BaziResult` only includes `di_shi`, which corresponds to the "Self-sat life stage" (自坐) of a pillar's Heavenly Stem against its own Earthly Branch. Professional setups heavily rely on the Day Master's (日干) Twelve Life Stages (星运) across all four branches, the Void branches (空亡), and the Gods and Demons (神煞). The backend relies on `github.com/6tail/lunar-go` which already possesses internal abstractions for these calculations.

## Goals / Non-Goals

**Goals:**
- Extract Star Fortune (Xingyun) of the Day Master for Year, Month, Day, and Hour branches.
- Extract Void (Xun Kong) status associated with each branch.
- Extract Gods and Demons (Shensha) arrays for each pillar.
- Keep JSON output backward compatible for current UI layout, merely expanding available properties.

**Non-Goals:**
- Frontend Data Grid layout generation (this change is strictly for backend extraction and types modification; UI redesign will be a separate change).
- Detailed logic modification of Lunar-go.

## Decisions

1. **Naming Conventions**: 
    - 星运: `year_xing_yun`, `month_xing_yun`, `day_xing_yun`, `hour_xing_yun`
    - 空亡: `year_xun_kong`, `month_xun_kong`, `day_xun_kong`, `hour_xun_kong`
    - 神煞: `year_shen_sha`, `month_shen_sha`, `day_shen_sha`, `hour_shen_sha` (as arrays of strings `[]string`)
2. **Extraction Methodology**: 
    - Use `LunarUtil.GetZhiXing(dayGan, branch)` for 星运.
    - Use `bazi.GetYearXunKong()`, etc., for 空亡.
    - Use `bazi.GetYearShenSha()`, etc., iterating over the Maps returned to collect array of strings for 神煞.

## Risks / Trade-offs

- **Risk**: Lunar-go `Get...ShenSha()` returns a Map mapping rule to Shensha string. We need to collect the keys/values into a clean Go slice of strings to easily serialize to JSON.
- **Trade-off**: Slightly larger payload size for the `calculate` endpoint, but Bazi JSON remains well within minimal size thresholds.
