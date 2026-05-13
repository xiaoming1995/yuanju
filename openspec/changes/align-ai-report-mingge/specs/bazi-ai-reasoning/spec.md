## MODIFIED Requirements

### Requirement: Analytical LLM Prompts

The backend system SHALL supply detailed astrological data to the LLM and use the system-computed `ming_ge` as the single authoritative source for the report's main格局 conclusion.

#### Scenario: AI report generation uses system MingGe as authoritative source

- **WHEN** the backend asks the LLM to generate the natal report and the chart result contains `ming_ge`
- **THEN** the Prompt MUST inject `ming_ge` and any available explanatory text
- **AND** the Prompt MUST explicitly instruct the LLM to treat that value as the only valid main格局 label
- **AND** the LLM MUST NOT rename the chart into a different main格局

#### Scenario: AI may explain the main格局 without re-judging its name

- **WHEN** the LLM interprets the chart's 格局 in `analysis.logic`
- **THEN** it MAY explain whether the system main格局 is 成格、破格、有救、偏弱 or mixed with兼象
- **BUT** it MUST present any other structural tendency only as a secondary tendency, not as a replacement for the system main格局

### Requirement: Main格局-first analysis ordering

The `analysis.logic` section SHALL describe the report in a fixed order anchored on the system-computed `ming_ge`.

#### Scenario: analysis.logic starts from the system main格局

- **WHEN** the LLM writes `analysis.logic`
- **THEN** it MUST first state the system main格局
- **AND THEN** state whether that格局 is 成格、破格、有救 or 偏弱 in the natal chart
- **AND THEN** explain the relationship between that格局, 调候, and the final Yongshen/Jishen conclusion

#### Scenario: 调候 and 格局 direction conflict

- **WHEN** the system main格局's usual tendency differs from the 调候 direction
- **THEN** the report MUST explicitly explain the priority relationship
- **AND** it MUST preserve the system main格局 label while allowing 调候 to override the final喜忌 conclusion

### Requirement: Interpretation-oriented 格局 knowledge base

The格局 knowledge base used by the LLM SHALL guide explanation of the system main格局, not re-judgment of the格局 name.

#### Scenario: kb_gejv explains rather than renames

- **WHEN** the LLM consumes the `kb_gejv` knowledge block
- **THEN** that knowledge MUST define how to interpret 成格、破格、有救、偏弱、兼象
- **AND** it MUST define how 格局 and 调候 interact
- **AND** it MUST NOT instruct the LLM to overwrite the injected system main格局 with a different primary格局 name
