## ADDED Requirements

### Requirement: ValidateYearNarrative SHALL return a structured result preserving valid sentences

The system SHALL replace the legacy `(bool, string)` validator contract with a structured `NarrativeValidationResult` exposing the cleaned narrative text, cleared-sentence list, soft-warned-term list, and hard-keywords-hit list.

#### Scenario: Clean narrative passes unchanged
- **WHEN** a year narrative contains only命理 terms attested in its evidence
- **THEN** `CleanedNarrative` SHALL equal the input verbatim
- **AND** `ClearedSentences`, `SoftWarnedTerms`, and `HardKeywordsHit` SHALL all be empty

#### Scenario: Empty input returns empty result
- **WHEN** the year narrative is the empty string
- **THEN** `CleanedNarrative` SHALL be the empty string
- **AND** all three diagnostic slices SHALL be empty

### Requirement: Hard-keyword class violations SHALL clear only the offending sentence

The system SHALL drop only the sentence in which a hard-keyword-class violation appears, preserving all other sentences of the narrative.

#### Scenario: One hard-relation violation in a four-sentence narrative
- **WHEN** the narrative contains four sentences and one sentence references "伏吟"
- **AND** the year evidence has no "伏吟" anchor
- **THEN** the cleared sentence SHALL be removed from `CleanedNarrative`
- **AND** the other three sentences SHALL appear in `CleanedNarrative` in their original order
- **AND** `ClearedSentences` SHALL contain the dropped sentence
- **AND** `HardKeywordsHit` SHALL contain "伏吟"

#### Scenario: Position-anchor violation clears the sentence
- **WHEN** the narrative contains a sentence claiming "用神位受冲"
- **AND** the year evidence has no "用神位" reference
- **THEN** that sentence SHALL be removed from `CleanedNarrative`
- **AND** `HardKeywordsHit` SHALL contain "用神位"

#### Scenario: Event-marker violation clears the sentence
- **WHEN** the narrative contains a sentence stating "本年双重命中"
- **AND** the year evidence has no "双重命中" anchor
- **THEN** that sentence SHALL be removed from `CleanedNarrative`
- **AND** `HardKeywordsHit` SHALL contain "双重命中"

#### Scenario: Multiple sentences with the same hard-keyword violation
- **WHEN** two sentences in the same narrative both reference "伏吟"
- **AND** the year evidence has no "伏吟" anchor
- **THEN** both sentences SHALL be removed from `CleanedNarrative`
- **AND** `HardKeywordsHit` SHALL contain "伏吟" once (de-duplicated)

#### Scenario: All sentences hit hard-keyword violations
- **WHEN** every sentence in the narrative references unattested hard-keyword-class terms
- **THEN** `CleanedNarrative` SHALL be the empty string
- **AND** `ClearedSentences` SHALL contain all original sentences

### Requirement: 神煞 class violations SHALL preserve the sentence with an inline marker

The system SHALL keep sentences that reference unattested 神煞 names, appending an inline marker `(注：未在本年算法 evidence 中识别)` immediately after the offending term.

#### Scenario: Single 神煞 violation gains an inline marker
- **WHEN** a narrative sentence reads "驿马动象明显"
- **AND** the year evidence has no "驿马" anchor
- **THEN** `CleanedNarrative` SHALL contain that sentence with the marker appended after "驿马"
- **AND** the surrounding sentences SHALL be unchanged
- **AND** `SoftWarnedTerms` SHALL contain "驿马"
- **AND** `ClearedSentences` SHALL NOT contain that sentence

#### Scenario: Multiple distinct 神煞 in one sentence
- **WHEN** a narrative sentence references both "桃花" and "天乙"
- **AND** neither term is in the year evidence
- **THEN** the sentence SHALL be preserved with markers after each unattested term
- **AND** `SoftWarnedTerms` SHALL contain both "桃花" and "天乙"

### Requirement: Mixed-class violations SHALL apply independent treatment per sentence

The system SHALL process each sentence against the bifurcated keyword sets independently, allowing a hard-keyword drop on one sentence and a 神煞 soft-warn on another in the same narrative.

#### Scenario: Hard violation in sentence A, 神煞 violation in sentence B
- **WHEN** sentence A references unattested "伏吟" and sentence B references unattested "驿马"
- **THEN** sentence A SHALL be removed from `CleanedNarrative`
- **AND** sentence B SHALL remain with a "(注：未在本年算法 evidence 中识别)" marker after "驿马"
- **AND** `HardKeywordsHit` SHALL contain "伏吟"
- **AND** `SoftWarnedTerms` SHALL contain "驿马"

### Requirement: Sentence segmentation SHALL respect Chinese full-width punctuation

The system SHALL segment narratives on Chinese sentence terminators `。！？；…` as well as ASCII counterparts `!?.;` while preserving terminators with their parent sentence.

#### Scenario: Mixed terminators are recognized
- **WHEN** a narrative contains sentences ending in "。", "！", "？", and a trailing fragment without a terminator
- **THEN** `splitChineseSentences` SHALL return four segments, each ending with its original terminator (the trailing fragment has none)

#### Scenario: Single-sentence narrative
- **WHEN** a narrative has no internal terminator
- **THEN** segmentation SHALL return a slice of length 1 containing the original text

### Requirement: Cache writes SHALL stamp the new algorithm version

The system SHALL write `algorithm_version = 'v3-narrative-graceful'` to newly inserted or upserted rows in `ai_dayun_summaries` and `ai_reports` once Phase 2 ships.

#### Scenario: New dayun summary row carries v3 version
- **WHEN** `UpsertDayunSummary` writes a row after Phase 2 ships
- **THEN** the row's `algorithm_version` SHALL equal `"v3-narrative-graceful"`

#### Scenario: Pre-Phase-2 rows remain readable at their original version
- **WHEN** the system reads a row with `algorithm_version = 'v2-yongshen-shishen'` or `NULL`
- **THEN** the row SHALL be served as-is without retroactive validation rewrite
