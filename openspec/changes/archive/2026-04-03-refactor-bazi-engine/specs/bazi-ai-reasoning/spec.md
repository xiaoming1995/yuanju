## ADDED Requirements

### Requirement: Analytical LLM Prompts
The backend system SHALL supply detailed astrological data (including all pillars and generated hidden stems) to the LLM and utilize a Chain of Thought prompt to derive Yongshen and Jishen.

#### Scenario: Intelligent Interpretation Generation
- **WHEN** the backend asks the LLM to generate the report
- **THEN** it first instructs the AI to evaluate the five elements, the month command (月令), and the hidden stems, explicitly outputting an analysis of the Day Master's strength
- **THEN** the AI determines the Yongshen based on this analysis

### Requirement: Removal of Code-based Yongshen Algorithm
The system MUST NOT attempt to judge Yongshen or Day Master strength purely through code-based statistical averaging.

#### Scenario: Code simplification
- **WHEN** the bazi chart is generated
- **THEN** it does not run any hardcoded logic to assign Yongshen and Jishen prior to hitting the LLM (unless it's just a placeholder string)
