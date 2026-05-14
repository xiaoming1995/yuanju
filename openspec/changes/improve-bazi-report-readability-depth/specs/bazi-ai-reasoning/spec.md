## MODIFIED Requirements

### Requirement: Analytical LLM Prompts
The backend system SHALL supply detailed astrological data (including all pillars and generated hidden stems) to the LLM and utilize a Chain of Thought prompt to derive Yongshen and Jishen. The prompt SHALL also require user-facing report content to be detailed, structured, and understandable to non-professional readers.

#### Scenario: Intelligent Interpretation Generation
- **WHEN** the backend asks the LLM to generate the report
- **THEN** it first instructs the AI to evaluate the five elements, the month command (月令), and the hidden stems, explicitly outputting an analysis of the Day Master's strength
- **THEN** the AI determines the Yongshen based on this analysis

#### Scenario: Detailed chapter generation
- **WHEN** the backend asks the LLM to generate structured report chapters
- **THEN** each concise chapter SHALL provide a clear conclusion suitable for quick reading
- **THEN** each detailed chapter SHALL include a stable long-form interpretation covering conclusion, astrological basis, real-life manifestation, and practical advice

#### Scenario: Plain-language terminology
- **WHEN** the generated report uses specialized terms such as 印星, 官杀, 食伤, 财星, 用神, 忌神, 调候, or 格局
- **THEN** the report SHALL explain the practical meaning in plain Chinese near the term
- **THEN** the report SHALL avoid presenting dense terminology without user-facing interpretation

#### Scenario: Analysis overview depth
- **WHEN** the backend asks the LLM to generate the overall chart analysis
- **THEN** the generated analysis SHALL provide enough detail to explain the main judgment, its basis, and its real-life meaning
- **THEN** the prompt SHALL avoid overly short limits that force the overview into a terse summary
