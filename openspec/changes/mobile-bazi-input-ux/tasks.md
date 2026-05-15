## 1. Birth Profile Component

- [x] 1.1 Update `BirthProfileForm` props so pages can opt into confirmation summary and advanced calibration without duplicating birth-field logic.
- [x] 1.2 Refactor the basic field layout to prioritize gender, calendar type, birth date, and birth time with mobile-first spacing and touch targets.
- [x] 1.3 Add a reusable birth profile summary renderer that includes gender, calendar type, date, time, leap month state, and calibration mode.
- [x] 1.4 Replace the current Zi hour checkbox behavior with explicit 23:00-23:59 and 00:00-00:59 mutually exclusive choices when Zi hour is selected.
- [x] 1.5 Preserve existing solar/lunar day normalization behavior when year, month, calendar type, or leap month changes.

## 2. Home Page Integration

- [x] 2.1 Move birth location and true solar time controls into a collapsed advanced calibration area on the Bazi homepage form.
- [x] 2.2 Keep the existing `CalculateInput` mapping unchanged, including `longitude`, `calendar_type`, `is_leap_month`, and early Zi hour fields.
- [x] 2.3 Place the confirmation summary close to the submit action so users can verify selections before starting calculation.
- [x] 2.4 Adjust mobile CSS so the form remains readable on narrow screens and the submit action is easy to reach after the last required field.

## 3. Compatibility Page Integration

- [x] 3.1 Update both compatibility participant forms to use the improved basic birth input flow.
- [x] 3.2 Confirm the two-form compatibility layout remains readable on mobile and desktop after the component changes.
- [x] 3.3 Ensure compatibility request payloads still use the existing `CompatibilityProfileInput` shape without adding backend requirements.

## 4. Verification

- [x] 4.1 Add or update frontend tests for summary text, lunar leap month display, date normalization, and Zi hour disambiguation state.
- [x] 4.2 Run the frontend test suite or the relevant component/page tests.
- [x] 4.3 Run the frontend build or typecheck command used by the project.
- [x] 4.4 Verify the homepage form manually in desktop and mobile viewports, including default Beijing-time flow, advanced calibration, lunar date, and early Zi hour submission mapping.
- [x] 4.5 Verify compatibility input manually in desktop and mobile viewports.

## 5. Mobile Compactness Follow-up

- [x] 5.1 Compact gender and calendar controls into a mobile-friendly primary selector area.
- [x] 5.2 Change mobile date layout to keep year full-width while placing month and day side by side.
- [x] 5.3 Add age helper text to year options so long mobile year lists are easier to scan.
- [x] 5.4 Verify homepage and compatibility forms still render correctly with the restored mobile bottom navigation.

## 6. Mobile Hero and Compatibility Steps

- [x] 6.1 Reduce the mobile homepage hero height so the birth form appears sooner.
- [x] 6.2 Add mobile tabs for compatibility participant input.
- [x] 6.3 Keep desktop compatibility input as two visible panels.
- [x] 6.4 Verify mobile homepage and compatibility tab behavior in browser.
