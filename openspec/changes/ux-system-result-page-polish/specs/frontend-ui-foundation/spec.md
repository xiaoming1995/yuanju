## ADDED Requirements

### Requirement: Shared frontend design tokens

The frontend SHALL expose shared CSS variables for semantic colors, text hierarchy, spacing, radius, and status states so pages do not define primary visual rules independently.

#### Scenario: Page uses semantic surface tokens
- **WHEN** a new user-facing page or shared component needs a background, border, muted text, success, warning, danger, or primary accent color
- **THEN** it SHALL use a semantic CSS variable from the global token set instead of a hard-coded color literal.

#### Scenario: Component uses shared spacing and radius
- **WHEN** a new shared UI primitive renders card, button, input, dialog, or panel spacing
- **THEN** it SHALL use shared spacing and radius variables rather than page-local magic numbers as its primary layout source.

### Requirement: Shared UI primitives

The frontend SHALL provide lightweight shared UI primitives for page layout, sections, buttons, tabs, badges, empty states, confirmation, toast feedback, and form fields.

#### Scenario: Page shell composition
- **WHEN** a user-facing page needs standard width, page padding, and mobile safe-area spacing
- **THEN** it SHALL be able to compose the layout with `PageShell` without duplicating page container CSS.

#### Scenario: Section composition
- **WHEN** a page renders a titled content section
- **THEN** it SHALL be able to use `SectionPanel` for consistent title, description, spacing, and content framing.

#### Scenario: Button variants
- **WHEN** a page renders primary, secondary, ghost, or dangerous actions
- **THEN** it SHALL be able to use a shared `Button` component with those variants and consistent disabled/loading visual states.

#### Scenario: Status badges
- **WHEN** a page renders active, inactive, success, failed, pending, warning, or danger status
- **THEN** it SHALL be able to use `StatusBadge` with a consistent visual mapping and Chinese-friendly label support.

### Requirement: Unified feedback primitives

The frontend SHALL provide non-browser-native feedback primitives for confirmation and transient operation feedback.

#### Scenario: Confirmation dialog
- **WHEN** a user action requires confirmation, especially a dangerous action
- **THEN** the UI SHALL be able to show `ConfirmDialog` with title, description, cancel action, confirm action, danger styling, and pending state.

#### Scenario: Toast feedback
- **WHEN** an operation succeeds, fails, or needs non-blocking information feedback
- **THEN** the UI SHALL be able to show a toast notification without using browser-native `alert()`.

### Requirement: Shared form field structure

The frontend SHALL provide a reusable form field structure for labels, hints, errors, and controls.

#### Scenario: Field validation error
- **WHEN** a form control has a validation error
- **THEN** the UI SHALL render the error inline near the control using shared form field styling rather than a browser-native alert.

#### Scenario: Field helper text
- **WHEN** a form control needs explanatory text
- **THEN** the UI SHALL render the helper text with shared muted text styling and consistent spacing.
