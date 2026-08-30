## ADDED Requirements

### Requirement: Dayun road explains natal structure interaction
The Dayun road calculation SHALL evaluate each Dayun against the current natal assessment's climate urgency, Fuyi use-god support, and pattern-support or pattern-break ten gods, in addition to its existing road factors.

#### Scenario: A Dayun supports a natal pattern
- **WHEN** a Dayun stem or branch ten god matches a configured natal pattern support condition
- **THEN** its road evidence includes a positive "格局作用" record that identifies the support direction.

#### Scenario: A Dayun aggravates a natal pattern risk
- **WHEN** a Dayun stem or branch ten god matches a configured natal pattern break condition
- **THEN** its road evidence includes a negative "格局作用" record that identifies the risk direction.

#### Scenario: A legacy chart is displayed
- **WHEN** a legacy chart is opened on the result page
- **THEN** lazy assessment backfill produces the same structure-aware road evidence before the road data is returned.
