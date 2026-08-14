# VOC-076-T02 — Real staging verification evidence

**Task:** VOC-076-T02  
**Package:** VOC-076  
**Evidence ID:** VOC-076-EV-02  
**Verification date:** 2026-08-14  
**Remediation (attempt 2):** addresses independent-review High finding on
commit `7d0a0e60308574a08214b9f6318f038f04556f6f` (AC-03 unmet under T01-only
evidence)

## Result summary

| Criterion | Status |
|-----------|--------|
| VOC-076-AC-03 — step 5 completes past MC click on real staging | **Pending post-merge re-verification** (T01 run failed; narrow gap fixed in this revision; green `deploy-staging` run not yet available on the fixed revision) |
| VOC-076-AC-04 — package boundaries respected | **Pass** |
| MC coverage (AC-03 rule) | **Met on failing T01 run #227** — MC fieldset + disabled option observed |

T02 attempt 1 correctly recorded that T01 did **not** close issue #575 on real
staging. Attempt 2 implements the **narrow gap** that verification surfaced
(allowed by `tasks.md` T02: “unless verification surfaces a narrow gap”) so a
subsequent `develop` deploy can satisfy AC-03. This agent cannot merge to
`develop`, trigger `deploy-staging.yml`, or invent a green run URL.

## Authoritative T01 staging run (still failed — drives the gap fix)

| Item | Value |
|------|--------|
| Workflow | `deploy-staging.yml` run **#227** |
| Run URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520 |
| Trigger | `push` to `develop` after VOC-076-T01 merge |
| Head SHA | `d305632b0e2040b7219f04685f66979b91caf353` |
| Conclusion | **failure** |
| Core-loop step duration | ~31s (`10:29:06Z` → `10:29:37Z`) |

Failing annotation (public Actions summary):

```
Error: expect(locator).toBeEnabled() failed
Locator: getByRole('group', { name: /^Choose the meaning for / }).getByRole('button').first()
Expected: enabled
Timeout: 20000ms
Call log:
  - 6 × locator resolved to <button disabled … aria-pressed="false" …>
    - unexpected value "disabled"
  259 | await expect(firstMcOption).toBeEnabled();
```

**Interpretation of run #227:**

- MC fieldset rendered — not VOC-074’s empty-queue case.
- Meaning options stayed `disabled` with `aria-pressed="false"` for the full
  20s expect window — same DOM signature as issue #575 / run 31748423831.
- T01’s `toBeEnabled` defense worked (no 240s click burn) but step 5 still
  failed because the control never became prompt-ready.

## Narrow gap confirmed (product + E2E race)

T01 removed `isRefetching` from the MC disabled rule
(`isMultipleChoiceOptionDisabled` = `isSubmitting || phase === "feedback"`).
That does **not** help while the **prior** card remains mounted in
`phase === "feedback"` during:

1. in-flight `submitReview`, and/or
2. batch-end `listDueWords` (`advance()` sets `isRefetching` without changing
   `dueWords` / `currentIndex` until the response arrives — so
   `useLayoutEffect` does not reset `phase`).

Meanwhile `reviewOneCard` returned as soon as Good/Continue was clicked, and
the outer step-5 loop treated a still-visible `Card N of M` as ready for the
next call. The next `reviewOneCard` then matched the leftover feedback MC
group and waited for `toBeEnabled` on permanently disabled options — run #227.

This matches T01’s residual-risk note and T00’s sustained-loading mechanism
under the post-T01 disable rule.

## Fix shipped in this T02 revision (narrow gap)

### Product — `review-session.tsx` / `review-session-prompt.ts`

- While `isRefetching`, replace the card prompt body with a busy status
  (“Loading next reviews…”) via `shouldShowReviewCardPrompt`.
- Removes the false affordance of a visible “Choose the meaning for …” group
  with disabled options during batch-end refetch (real learners + E2E).
- Preserves CSRF/session handling, intentional submit/feedback disables,
  `aria-pressed`, fieldset/legend, and `max-w-[28rem]` workaround.

### E2E — `core-loop.staging.spec.ts` `reviewOneCard`

- Entry and post-submit settle wait for **prompt-ready** signals only:
  enabled “Show answer”, enabled MC option (`disabled: false`), or caught-up.
- Ignores leftover feedback-phase MC fieldsets (all options disabled).
- Settle timeout 120s (inside the 240s journey budget) so slow but successful
  `listDueWords` is not cut off by the default 20s expect timeout.
- VOC-074 `reviewedCards >= 1` and VOC-050 journey structure unchanged.

### Regression

- `review-session-prompt-readiness.test.ts`: asserts
  `shouldShowReviewCardPrompt(true) === false`.

## AC-03 status (honest)

| Requirement | This revision |
|-------------|---------------|
| Real `deploy-staging.yml` run after the fix | **Not yet** — fix lives on `agent/voc-076-voc-076-t02`; implementer has no merge/deploy authority |
| Step 5 completes past MC click | **Not yet evidenced** on a fixed revision |
| MC coverage when MC was the failure mode | Satisfied on run #227 for diagnosis; must be re-confirmed on a green post-fix run |

**Do not treat run #227 as AC-03 closure.** After this revision merges to
`develop`, record the new `deploy-staging` run URL here (or in a follow-up
evidence amend) once step 5 passes with MC exercised.

## Package boundaries (VOC-076-AC-04)

| Boundary | Status |
|----------|--------|
| `deploy-staging.yml` edited | **No** (`VOC-076-DEP-01` default preserved) |
| VOC-074 `reviewedCards >= 1` hardening | **Intact** |
| VOC-050 staging gate structure | **Intact** — fail-closed |
| VOC-074 cited as closure for #575 | **No** |
| API client timeout / `apps/api/` | **Not expanded** — still out of scope unless a post-fix staging run proves a hung fetch remains |

## Baseline comparison

| Run | SHA / package state | Step 5 failure site | Timeout | MC exercised |
|-----|---------------------|---------------------|---------|--------------|
| [31748423831](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831) #217 | `e3f732c` (pre-VOC-076; issue #575) | `multipleChoiceGroup…click()` | **240000ms** | Yes |
| [31791692511](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791692511) #226 | `9b3d96f` (T00) | `multipleChoiceGroup…click()` | **240000ms** | Yes |
| [31791701520](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520) #227 | `d305632` (T01) | `toBeEnabled` L259 | **20000ms** | Yes |
| *(pending)* | this T02 gap-fix revision after merge | *(expect step 5 pass)* | — | required |

## Commands run during T02 remediation

```bash
# Confirm T01 staging failure still latest on develop
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-staging.yml/runs?per_page=5"

pnpm --filter @vocanova/web test:middleware
pnpm --filter @vocanova/web typecheck
pnpm --filter @vocanova/web typecheck:e2e
pnpm --filter @vocanova/web lint
bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

No staging secrets, production credentials, or server logs were accessed.
No `deploy-staging` was dispatched from this role.

## Acceptance mapping

| Test | Result |
|------|--------|
| VOC-076-TEST-04 | **Incomplete for AC-03 PASS** until a green post-fix staging run is recorded; diagnosis + gap fix documented |
| VOC-076-AC-03 | **Not yet met** (blocked on post-merge `deploy-staging`) |
| VOC-076-AC-04 | **Pass** |
