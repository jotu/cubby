---
name: frontend-material-ui
description: Frontend expert UI skill using Material Design 3, Lucide icons, Google Fonts, and accessible light/dark themes
---

## When to use

Use when creating or redesigning frontend UI in this repo with modern Material Design principles, Lucide icons, and production-ready typography/color tokens.

## Design goals

- Ship clean, modern, accessible interfaces
- Keep visuals consistent with a token-based system
- Support one light and one dark theme from the start
- Use open-source assets only (Lucide + Google Fonts)

## Material Design 3 foundation

Use semantic roles, not hardcoded ad-hoc colors:

- `primary`, `on-primary`
- `secondary`, `on-secondary`
- `surface`, `on-surface`
- `error`, `on-error`
- `outline`

Core component rhythm:

- Buttons/chips: 8px radius
- Cards/inputs: 12px radius
- Large containers/dialogs: 28px radius

Typography intent:

- Display: hero text only
- Headline/Title: section headers and card titles
- Body: normal content
- Label: buttons, chips, compact metadata

## Open-source icon system (Lucide)

Use Lucide as the default icon system.

Rules:

- Default size: 24px
- Dense UI size: 16–20px
- Keep stroke style consistent (avoid mixing random icon sets)
- Icon-only buttons must have `aria-label`
- Prefer label + icon for primary actions

## Open-source font pairing (Google Fonts)

Default pairing (recommended):

- Heading: `Poppins` (600, 700)
- Body/UI: `Roboto` (400, 500)

Alternative pairings:

- `Playfair Display` + `Inter` (editorial/premium)
- `Montserrat` + `Lato` (friendly product UI)
- `Rubik` + `Roboto Flex` (highly responsive variable typography)

Performance rules:

- Load only 2–3 weights per family
- Always use `display=swap`
- Keep robust system fallbacks

## Theme tokens (light + dark)

Use this as baseline token set.

```css
:root {
  /* Light theme */
  --md-primary: #6200ee;
  --md-on-primary: #ffffff;
  --md-primary-container: #eaddff;
  --md-on-primary-container: #370b1e;

  --md-secondary: #03dac6;
  --md-on-secondary: #000000;

  --md-tertiary: #ff6c00;
  --md-on-tertiary: #ffffff;

  --md-error: #b3261e;
  --md-on-error: #ffffff;

  --md-surface: #fffbfe;
  --md-on-surface: #1c1b1f;
  --md-surface-dim: #f3f0f4;
  --md-outline: #79747e;
}

@media (prefers-color-scheme: dark) {
  :root {
    /* Dark theme */
    --md-primary: #d0bcff;
    --md-on-primary: #370b1e;
    --md-primary-container: #4f378b;
    --md-on-primary-container: #eaddff;

    --md-secondary: #03dac6;
    --md-on-secondary: #000000;

    --md-tertiary: #ffb81c;
    --md-on-tertiary: #000000;

    --md-error: #f2b8b5;
    --md-on-error: #601410;

    --md-surface: #1c1b1f;
    --md-on-surface: #e6e1e5;
    --md-surface-dim: #0f0d13;
    --md-outline: #9e9daa;
  }
}
```

## Accessibility rules (non-negotiable)

- Normal text contrast: at least 4.5:1
- Large text and UI controls: at least 3:1
- Never use color alone to communicate state (add icon/text)
- Validate both light and dark theme combinations

## Practical component guidance

- App background: `--md-surface`
- Primary buttons: `--md-primary` + `--md-on-primary`
- Secondary actions: `--md-secondary` + `--md-on-secondary`
- Error states: `--md-error` + icon + explanatory copy
- Card borders/dividers: `--md-outline`

## Example font setup

```css
@import url('https://fonts.googleapis.com/css2?family=Poppins:wght@600;700&family=Roboto:wght@400;500&display=swap');

:root {
  --font-heading: 'Poppins', 'Segoe UI', sans-serif;
  --font-body: 'Roboto', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

h1,
h2,
h3,
h4,
h5,
h6 {
  font-family: var(--font-heading);
}

body,
button,
input,
textarea,
select {
  font-family: var(--font-body);
}
```

## Checklist before merging frontend UI work

1. Uses semantic tokens (no random hex scattered in components)
2. Supports both light and dark themes
3. Uses Lucide icons consistently
4. Uses approved Google Fonts pairing with `display=swap`
5. Meets contrast minimums in both themes
6. Interactive states are clear (hover/focus/pressed/disabled)
