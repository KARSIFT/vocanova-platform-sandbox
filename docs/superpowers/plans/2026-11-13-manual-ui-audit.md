# Manual UI audit implementation plan

> **For agentic workers:** Use repository browser-testing, systematic debugging and verification guidance. Root coordinates builds, servers and integration; bounded independent reviewers inspect source and screenshots.

**Goal:** Inspect every page family and repair observed alignment, hierarchy and style inconsistencies without changing learning or authentication semantics.

**Architecture:** Reuse existing Surface, AuthShell, headword and action roles. Keep shared control defaults in the components layer so authored pressed/selected utilities apply. Preserve canonical prompts, request identities, server-confirmed progress and credential/session policy.

**Tech stack:** Existing Next.js/React, generated Tailwind tokens and pinned Playwright. Local web3100/API8180; no new dependency.

**Spec:** Owner's page-by-page manual test request, `docs/design/learning-workspace.md`, and findings in `docs/product/manual-ui-ux-audit-2026-11-13.md`.

## Tasks

- [x] Inventory all43 source page patterns; distinguish aliases/redirects and create synthetic dynamic sessions through UI actions. Save and inspect before screenshots, including actual questions, feedback and completion.
- [x] Correct Sentence History return/header spacing, Plan wrapping, Home empty typography and Account export Surface. Verify geometry, enlarged text and saved-word typography.
- [x] Correct action hierarchy and reuse existing button roles in Stories, Plan, Practice, Lists and account forms. Verify selected audio/writing states and disabled controls; preserve every mutation and destination.
- [x] Separate typed recall's focused heading from its intact server prompt and restore actual teaching-word typography. Verify no answer leakage, unchanged grading and keyboard progression.
- [x] Reuse AuthShell for magic/email outcomes and fallback, preserving token/CSRF handling, refs, retry and announcements. Replace generated status-only errors with the screen fallback while retaining authored API guidance and401 routing.
- [x] Inspect repaired screenshots at360/430/1280px across light/dark; run focused cross-engine regressions, existing feature/layout suites, unit tests, types, lint and build. Request independent source and visual review; qualify harness corrections and provider/device limits.

## Review focus

- A CSS reuse change must not hide selected states beneath shared defaults.
- A condensed typed prompt must neither expose answers nor change the server text.
- Auth wrappers must not lose outcome focus or alter token consumption/returnTo.
- Enlarged text and wrapped controls must retain distinct44px targets and reachable actions.
- An empty/failure/redirect screenshot must not be mislabeled as a populated/successful page.

## Integration

Use one consolidated draft/hold PR. Inspect required CI/reviews and use the normal merge queue. Strict staging and separate production promotion require exact release identity and their actual acceptance results. Release receipts belong to the PR and private handoff; this implementation plan is not rollout evidence.
