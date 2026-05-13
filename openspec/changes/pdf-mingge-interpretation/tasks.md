## 1. Print Layout Data Wiring

- [x] 1.1 Extend `PrintLayout` props to accept `ming_ge` and `ming_ge_desc`
- [x] 1.2 Update `ResultPage` to pass MingGe fields into the print layout component

## 2. PDF MingGe Section

- [x] 2.1 Add a `命格解读` block to `PrintLayout` before the existing `命局分析总览` block
- [x] 2.2 Render the MingGe name and MingGe description in a compact print-friendly format
- [x] 2.3 Derive an optional `本局落点` short line from `structured.analysis.logic` and omit it when unavailable
- [x] 2.4 Ensure the MingGe block is skipped cleanly when `ming_ge` is absent

## 3. Verification

- [x] 3.1 Verify PDF / print output renders the MingGe block in the intended order
- [x] 3.2 Verify reports without MingGe still export correctly with no empty placeholders
- [x] 3.3 Verify the `本局落点` line stays concise and does not duplicate the full analysis block
