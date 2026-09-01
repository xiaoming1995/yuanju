## ADDED Requirements

### Requirement: AI reports explain backend wealth profile
The backend AI report prompt SHALL include backend-computed wealth-profile context when available and SHALL instruct the LLM to explain it without recalculating or overriding it.

#### Scenario: AI prompt includes wealth profile
- **WHEN** the backend builds an AI report prompt for a Bazi result containing `wealth_profile`
- **THEN** the prompt includes the wealth grade, label, summary, tags, risk flags, current hint when present, and bounded evidence
- **AND** the prompt states that the LLM must not invent another wealth grade or contradict the backend-computed profile

#### Scenario: AI explains wealth conservatively
- **WHEN** the LLM generates a report using wealth-profile context
- **THEN** the generated copy frames the result as wealth structure, money/resource visibility, carrying capacity, flow, retention, and risk-control guidance
- **AND** it does not claim guaranteed wealth, fixed social class, exact income, investment timing, or investment advice

#### Scenario: Wealth profile is missing
- **WHEN** the backend builds an AI report prompt for a result without `wealth_profile`
- **THEN** the prompt omits the wealth-profile block
- **AND** the LLM is not asked to invent a deterministic wealth grade
