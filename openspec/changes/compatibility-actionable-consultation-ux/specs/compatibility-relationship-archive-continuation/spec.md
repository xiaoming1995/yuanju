## ADDED Requirements

### Requirement: Compatibility history supports continuation
The compatibility history page SHALL help users recognize and continue from prior personality-fit readings instead of only listing saved records.

#### Scenario: User views compatibility history
- **WHEN** the user opens compatibility history with saved readings
- **THEN** each reading card shows the relationship stage, primary question, personality match type, overall decision state, and a continuation action

#### Scenario: Reading has no context
- **WHEN** a saved compatibility reading lacks relationship stage or primary question context
- **THEN** the history card remains visible with a general relationship fallback label

### Requirement: History cards expose next best action
The compatibility history page SHALL choose a clear next action for each reading based on available personality-fit, reading, and report state.

#### Scenario: Reading has no deep report
- **WHEN** a history item indicates no deep compatibility report is available, or the detail flow can generate one
- **THEN** the card offers a continuation path that leads the user toward generating or viewing the deep reading without hiding the saved result

#### Scenario: Reading has enough result data
- **WHEN** a history item already has a decision result
- **THEN** the card offers a view/continue action and highlights the personality match type, latest question, or validation focus

### Requirement: Archive remains scan-friendly on mobile
The compatibility history page SHALL keep relationship archive cards readable and tappable on mobile.

#### Scenario: User views history on mobile
- **WHEN** compatibility history is rendered on a narrow viewport
- **THEN** context labels, score snippets, summary tags, and continuation actions stack without overlapping or truncating essential labels
