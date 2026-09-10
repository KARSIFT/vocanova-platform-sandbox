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

- [x] Landing, sign-in, onboarding, and app screens share a visual system.
- [x] A first-time learner can discover, save, review, and practise a word.
- [x] Home prioritizes the current mission and handles empty/completed states.
- [x] Journey and saved vocabulary support clear browsing and word detail.
- [x] Reviews provide focused recall, useful rating guidance, and completion.
- [x] Sentence practice preserves input, handles pending/errors, and shows feedback.
- [x] Progress and settings are understandable and usable on mobile and desktop.
- [x] Browser checks cover 360px, 430px, desktop, keyboard, and accessibility.
- [x] Relevant local unit, browser, and build checks pass; database checks run in CI.
- [x] Documentation reflects verified behavior and deployment limitations.
- [x] One PR contains the delivery, with readiness gated on final CI and review.

## Verification evidence (2026-09-10)

- Production web build, workspace validation, formatting, lint, and type checks pass.
- Foundation checks, 31 shared API-client tests, and 59 web helper/middleware tests pass.
- Local browser suite: 142 passed, 17 existing layout-specific skips. The complete
  learning loop now runs at all three layouts, not just desktop.
- Retry tests verify one pending request, stable request identity after an ambiguous
  transport failure, fresh identity after editing, the original checked sentence,
  and removal of empty validation alerts while a revised sentence is pending.
- API CI command (`go test -skip TestControlledSignupOAuth ./...`) and Go build pass.
- Production smoke self-tests: all 15 scenarios pass, including rejecting `/login`
  redirects instead of counting the login page as a rendered authenticated route.
- Manually inspected rendered Home, Journey, word detail, Progress, and sign-in
  screens on mobile, plus desktop Journey. Browser fixtures are synthetic, not
  evidence of real-provider AI quality or live Google authentication.
- All 12 Lighthouse audits pass the unchanged 85/95/90 budgets. Minimum scores:
  performance 100, accessibility 100, best practices 96. Reduced-motion transition
  behavior is also verified in Chromium.

## Deployment boundaries and remaining gates

- [PR #1460](https://github.com/KARSIFT/vocanova-platform-sandbox/pull/1460) is the
  single delivery PR and live source of final CI/review status. Required checks
  and automated review are inspected before making it ready for the merge queue.
- Local Docker is unavailable. Database-backed OAuth verification must pass in
  the dedicated CI job; the local API command intentionally excludes that harness.
- Merging deploys staging. Production requires a separate manual workflow dispatch
  on the reviewed revision; documentation now describes that distinction.
- Real AI generation fails closed when the selected provider lacks its required
  credentials. Test mocks cannot silently become production learner feedback.
- Live Google account login requires a human/provider session; the OAuth callback
  harness and readiness checks do not claim to replace that audit.
- Issue #1459 still tracks the production authenticated route sweep. Do not close
  it without a successful deployed check.

## Baseline findings

- Baseline web production build passes.
- Home put review navigation below up to ten sentence-practice forms.
- Landing and learning used unrelated visual styles.
- Canonical seed contains seven situations, 39 words, and 42 meanings.
- Local browser execution initially lacks system libraries; local Docker is an
  unavailable WSL integration. Database-backed verification requires a functioning
  disposable database environment or the dedicated CI checks.
- Issue #1459 tracks a failing production authenticated route sweep; the same
  workflow's staging learning journey and production content checks pass.

This checklist records work in progress, not a release claim.
