## Why

The current homegrown Bazi calculation algorithm relies on imprecise hardcoded dates for solar terms (Jieqi) and simplified math for Dayun (e.g., starting exactly at age 3). Additionally, the Yongshen (用神) calculation is overly simplistic and ignores actual traditional factors like month pillar strength and hidden stems. To ensure professional-grade accuracy and prevent incorrect AI readings, we must migrate to a proven astronomical calculation library and delegate nuanced pattern-matching logic like Yongshen to the LLM.

## What Changes

- Replace the manual Bazi mathematical calculations in `pkg/bazi/engine.go` with the `github.com/6tail/lunar-go` library.
- **BREAKING**: Change the data structure of Bazi output to include Hidden Stems (地支藏干) and precise Dayun timestamps.
- Remove the algorithmic Yongshen/Jishen code; instead, alter the LLM prompt to compel the AI to deduce Yongshen through a Chain-of-Thought (CoT) process before generating the final report.
- Enhance the frontend input form to accept birthplace to compute true solar time (真太阳时) correction via longitude.
- Display Hidden Stems (地支藏干) and precise Dayun start times in the professional view on the frontend.

## Capabilities

### New Capabilities

- `bazi-precision-engine`: The core calculation capability integrating `lunar-go` for solar term boundaries, true solar time, hidden stems, and exact Dayun steps.
- `bazi-ai-reasoning`: Modifies the AI prompt pipeline to use Chain of Thought, forcing the model to evaluate the month pillar weighting and deduce Yongshen independently.

### Modified Capabilities


## Impact

- `pkg/bazi/engine.go`: Will be entirely replaced.
- Bazi API Handlers: Will need to accept optional longitude parameters.
- Frontend `HomePage.tsx`: Added form fields for location.
- Frontend `ResultPage.tsx`: Minor UI components added for displaying NaYin and Hidden Stems.
- Database: Schema may need minor adjustments if we wish to cache the calculated true solar time or location.
