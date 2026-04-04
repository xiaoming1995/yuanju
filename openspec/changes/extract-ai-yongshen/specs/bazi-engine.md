## ADDED Requirements

### Requirement: AI Report JSON Format Extraction
The Bazi Engine's AI Report module must require the LLM to return a JSON object encompassing the analysis report, favoring god (Yongshen), and unfavorable god (Jishen). The system must securely parse this JSON and persist the structured data into the database.

#### Scenario: Generate AI Report successfully parses JSON
- **WHEN** the LLM successfully returns a JSON block with `yongshen`, `jishen`, and `report` fields
- **THEN** the system updates the `bazi_charts` table with the extracted `yongshen` and `jishen` values
- **AND** the system saves the `report` text to the `ai_reports` table
- **AND** the API response contains both the report and the updated chart properties

#### Scenario: Generate AI Report receives malformed JSON
- **WHEN** the LLM returns unstructured text without valid JSON or with malformed JSON
- **THEN** the system must not crash
- **AND** the system must fallback to treating the entire response as the raw report content
- **AND** the `yongshen` and `jishen` fields remain blank for that chart
