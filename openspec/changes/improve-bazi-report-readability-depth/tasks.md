## 1. Backend Prompt Depth

- [x] 1.1 Update `buildBaziPrompt()` to require concise chapters around 80-120 Chinese characters and detailed chapters around 220-350 Chinese characters.
- [x] 1.2 Update detailed chapter instructions to require conclusion, astrological basis, real-life manifestation, and practical advice without forcing mechanical visible subheadings.
- [x] 1.3 Add plain-language terminology rules for common terms such as 印星, 官杀, 食伤, 财星, 用神, 忌神, 调候, and 格局.
- [x] 1.4 Relax the `analysis.logic` length target from 300-500 characters to a deeper but bounded range around 500-800 characters.

## 2. Frontend Reading Experience

- [x] 2.1 Change the structured report default mode from concise to detailed.
- [x] 2.2 Adjust report mode labels so users understand the difference between quick conclusions and full interpretation.
- [x] 2.3 Improve report body typography in `ResultPage.css`, including readable body size, line height, and spacing for long-form Chinese content.
- [x] 2.4 Verify legacy reports without `content_structured` still render through the fallback path without a mode switcher.

## 3. Tests and Verification

- [x] 3.1 Add or update backend tests that assert the Prompt contains the new chapter length, structure, and plain-language terminology constraints.
- [x] 3.2 Run backend tests covering report generation or Prompt construction.
- [x] 3.3 Run frontend build or relevant tests to verify the ResultPage changes compile.
- [x] 3.4 Manually generate or inspect a sample report for `1996-02-08 20:00` and confirm default detailed display, richer chapter content, and plain-language explanations.
