## ADDED Requirements

### Requirement: Compatibility report is generated from structured pair analysis
The system SHALL generate the compatibility report from the saved compatibility reading and its structured evidences, rather than from raw birth inputs alone.

#### Scenario: Report generation uses existing reading
- **WHEN** a user requests the compatibility report for an existing reading
- **THEN** the system loads the reading's dimension scores and evidences as the primary AI context
- **AND** the system does not require the client to resubmit both birth profiles

### Requirement: Compatibility report uses a dedicated prompt module and structured output format
The system SHALL use a dedicated prompt module for compatibility readings and SHALL return a structured report shape that covers overall judgment, per-dimension interpretation, risks, and advice.

#### Scenario: Prompt module is isolated from natal and liunian prompts
- **WHEN** an administrator manages prompt templates
- **THEN** the compatibility report template is available as an independent module
- **AND** editing it does not change the single-chart natal or liunian prompt behavior

#### Scenario: Generated report includes required sections
- **WHEN** a compatibility report is generated successfully
- **THEN** the response contains an overall summary section
- **AND** the response contains content for the four core dimensions plus explicit risks and advice

### Requirement: Compatibility reports are cached per reading
The system SHALL save generated compatibility reports under the compatibility reading resource so repeated viewing can reuse the latest generated result instead of creating duplicate reports by default.

#### Scenario: Reopening a reading returns the latest saved report
- **WHEN** a user opens a compatibility reading detail after a report was already generated
- **THEN** the detail response includes the latest saved compatibility report for that reading

#### Scenario: Report storage is isolated from single-chart report cache
- **WHEN** a compatibility report is created
- **THEN** it is stored under a compatibility-specific report resource
- **AND** it does not appear in the existing single-chart `ai_reports` history
