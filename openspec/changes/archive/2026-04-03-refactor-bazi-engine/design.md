## Context

The system currently relies on `pkg/bazi/engine.go` which uses hardcoded dates for Jieqi (e.g., Feb 4th for Lichun) and simplified math for Dayun (hardcoded start age of 3). It does not support true solar time (真太阳时). Yongshen is calculated using naïve mathematical averages. To achieve professional quality and generate accurate AI reports, we need precision astronomical methods. Moving to the mature `lunar-go` open-source library will eliminate the maintenance burden of custom astronomical programming.

## Goals / Non-Goals

**Goals:**
- Replace the bespoke Bazi math in `pkg/bazi` with `github.com/6tail/lunar-go`.
- Calculate precise Dayun transition dates rather than fixed-year approximations.
- Expose Hidden Stems (地支藏干) to the frontend.
- Offload the Yongshen logic from Go to the LLM.
- Allow optional longitude input from the frontend to correct to True Solar Time.

**Non-Goals:**
- Implementing an astronomical planetary engine from scratch.
- Full Geocoding integration (backend resolution of strings to coordinates); the frontend will provide the estimated longitude based on basic cities.

## Decisions

1. **Adopt `github.com/6tail/lunar-go`** 
   - *Rationale*: It is the industry standard Go library for traditional Chinese calendrical calculations. It natively supports EightCharacters (Bazi), accurate JieQi bounds down to the second, Dayun, hidden stems, and True Solar Time. It eliminates hundreds of lines of brittle, hardcoded logic.
   - *Alternative Considered*: Porting `swisseph` directly. Rejected due to extreme complexity and lack of direct Chinese calendrical mapping.
2. **Shift Yongshen Reasoning to the LLM**
   - *Rationale*: Assessing strength/weakness and Yongshen requires evaluating the Day Master against the Month Branch, Hidden Stems, and various interactions (clashes, combinations). It is nearly impossible to rule-set cleanly without a massive expert system. LLMs with proper Chain-of-Thought prompts handle this pattern recognition much better.
3. **True Solar Time via Frontend Longitude Map**
   - *Rationale*: Users do not know their longitude. The frontend will present a City/Province selector (or allow direct input) that maps to an approximate longitude, passing `longitude: number` to the backend's `/calculate` endpoint.

## Risks / Trade-offs

- **[Risk] Changes to historical charts** → Switching algorithms means users re-calculating previous charts may get slightly different Dayun timings or even a different month/day pillar if they were born right on a Jieqi boundary.
  - *Mitigation*: This is a fix, not a regression. The previous charts were objectively incorrect. We will accept the behavioral change.
- **[Risk] LLM hallucinating Yongshen** → LLM might output unpredictable Yongshens if the prompt is poor.
  - *Mitigation*: Use strict Chain-of-Thought prompting ("First examine the month branch... then count the supporting stems... then conclude Yongshen/Jishen").
