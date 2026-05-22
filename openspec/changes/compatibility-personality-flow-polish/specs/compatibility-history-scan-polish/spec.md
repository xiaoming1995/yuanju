## ADDED Requirements

### Requirement: History cards prioritize personality recognition
The compatibility history page SHALL make personality match type the primary recognition signal on each card.

#### Scenario: User views saved compatibility readings
- **WHEN** compatibility history contains saved readings
- **THEN** each card prominently shows participant names, personality match type, relationship stage, selected question, and one continuation action
- **AND** raw score dimensions are visually secondary to the personality match type

### Requirement: History cards remain compact and tappable
The compatibility history page SHALL keep cards compact enough for repeated scanning while preserving tappable affordances.

#### Scenario: User scans history on mobile
- **WHEN** the page is rendered on a narrow viewport
- **THEN** names, match type, context labels, and continuation action stack clearly without truncating essential labels
- **AND** score details do not force the card to become overly tall

### Requirement: History continuation copy matches reading state
The compatibility history page SHALL use continuation labels that point users back into the personality-fit flow.

#### Scenario: User chooses a saved reading
- **WHEN** the card has enough result data to open the detail page
- **THEN** the continuation label uses personality-oriented language such as viewing personality fit or continuing validation
- **AND** the label does not imply persisted journal progress that does not exist
