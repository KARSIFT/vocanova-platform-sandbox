# Vocanova product maturity delivery

This delivery is one pull request against `main`. It turns the documented A2–B1
learning loop into a coherent, usable product. The product bible and MVP PRD
remain the scope baseline; implementation history is not an acceptance criterion.

## Product decisions

- Use one calm visual language across landing, authentication, onboarding, and
  learning: consistent color tokens, typography, surfaces, navigation, and actions.
- Put the daily mission and its next action first on Home. Keep saved vocabulary
  compact and sentence practice intentional, rather than repeating expanded forms.
- Keep Home, Journey, and Progress as the three primary destinations. Saved
  vocabulary belongs to Journey; reviews are focused sessions.
- Show backend-confirmed progress, completion, saved state, and feedback. Explain
  first-use, caught-up, loading, and failure states with a useful next action.
- Preserve passwordless authentication, learner ownership, CSRF protection,
  idempotent mutations, and existing learning history.

## Acceptance checklist

- [ ] Landing, sign-in, onboarding, and app screens share a visual system.
- [ ] A first-time learner can discover, save, review, and practise a word.
- [ ] Home prioritizes the current mission and handles empty/completed states.
- [ ] Journey and saved vocabulary support clear browsing and word detail.
- [ ] Reviews provide focused recall, useful rating guidance, and completion.
- [ ] Sentence practice preserves input, handles pending/errors, and shows feedback.
- [ ] Progress and settings are understandable and usable on mobile and desktop.
- [ ] Browser checks cover 360px, 430px, desktop, keyboard, and accessibility.
- [ ] Relevant unit, integration, build, and CI checks pass.
- [ ] Documentation reflects verified behavior and deployment limitations.
- [ ] One PR is reviewed and made ready only after the complete change is verified.

## Initial evidence

- Baseline web production build passes.
- Home currently puts review navigation below up to ten sentence-practice forms.
- Landing and learning use unrelated visual styles.
- Canonical seed contains seven situations, 39 words, and 42 meanings.
- Local browser execution initially lacks system libraries; local Docker is an
  unavailable WSL integration. Database-backed verification requires a functioning
  disposable database environment or the dedicated CI checks.
- Issue #1459 tracks a failing production authenticated route sweep; the same
  workflow's staging learning journey and production content checks pass.

This checklist records work in progress, not a release claim.
