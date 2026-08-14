# VOC-076-T02 — Real staging verification evidence

**Task:** VOC-076-T02  
**Package:** VOC-076  
**Evidence ID:** VOC-076-EV-02  
**Verification date:** 2026-08-14  
**T01 revision verified:** `d305632b0e2040b7219f04685f66979b91caf353` (PR #594)

## Result: VOC-076-AC-03 not met

The first `deploy-staging.yml` run after VOC-076-T01 merged to `develop`
**failed** at step 5. Multiple-choice cards were exercised, but the journey did
**not** complete past the MC answer interaction. Issue #575's underlying
symptom — a visible MC meaning-option button that stays `disabled` — persists
on real staging under the T01 revision.

| Criterion | Status |
|-----------|--------|
| VOC-076-AC-03 — step 5 completes past MC click on real staging | **Fail** |
| VOC-076-AC-04 — package boundaries respected | **Pass** (no workflow expansion; VOC-074/VOC-050 intact) |
| MC coverage (AC-03 rule) | **Met** — MC fieldset rendered; failure at MC enabled wait |

## Authoritative verification run

| Item | Value |
|------|--------|
| Workflow | `deploy-staging.yml` run **#227** |
| Run URL | https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520 |
| Trigger | `push` to `develop` after VOC-076-T01 merge |
| Head SHA | `d305632b0e2040b7219f04685f66979b91caf353` |
| Conclusion | **failure** |
| Total duration | 11m 9s |
| Core-loop step duration | ~31s (`10:29:06Z` → `10:29:37Z`) |

Failing annotation (public Actions summary):

```
[staging-desktop-1280] › core-loop.staging.spec.ts:274:3 › … › 5. work the review queue

Error: expect(locator).toBeEnabled() failed

Locator: getByRole('group', { name: /^Choose the meaning for / }).getByRole('button').first()
Expected: enabled
Timeout: 20000ms
Error: element(s) not found

Call log:
  - Expect "toBeEnabled" with timeout 20000ms
  - waiting for getByRole('group', { name: /^Choose the meaning for / }).getByRole('button').first()
  - 6 × locator resolved to <button disabled … aria-pressed="false" …>
    - unexpected value "disabled"

  259 | await expect(firstMcOption).toBeEnabled();
      |                               ^
  260 | await firstMcOption.click();

at reviewOneCard (core-loop.staging.spec.ts:259:31)
```

**Interpretation:**

- The MC fieldset **rendered** (`Choose the meaning for …` group present) — not
  VOC-074's vacuous empty-queue case.
- The first meaning option stayed **`disabled` with `aria-pressed="false"`** for
  the full 20s `toBeEnabled` window — the same DOM signature as run 31748423831
  (issue #575).
- T01's explicit `toBeEnabled` wait **did** prevent the prior 240s click burn
  (`VOC-076-AC-02` defense-in-depth works as designed), but step 5 still does
  not pass because the product control never becomes interactable.
- No `reviewedCards` log line was emitted; the failure occurs before the step-5
  `reviewedCards >= 1` assertion (VOC-074 hardening never reached).

## Baseline comparison (same failure mode, different timeout surface)

| Run | SHA / package state | Step 5 failure site | Timeout | MC exercised |
|-----|---------------------|---------------------|---------|--------------|
| [31748423831](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831) #217 | `e3f732c` (pre-VOC-076; issue #575) | `multipleChoiceGroup…click()` L258 | **240000ms** | Yes |
| [31791692511](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791692511) #226 | `9b3d96f` (T00 merged, pre-T01 fix) | `multipleChoiceGroup…click()` L258 | **240000ms** | Yes |
| [31791701520](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520) #227 | `d305632` (**T01 merged**) | `expect(firstMcOption).toBeEnabled()` L259 | **20000ms** | Yes |

The T01 product changes (decouple MC disabled state from `isRefetching`; synchronous
`useLayoutEffect` phase reset; memoized MC options) and E2E enabled wait **did not**
restore interactable MC options on staging. The failure mode shifted from a 240s
`click()` hang to a faster, explicit readiness failure — not to a green step 5.

## MC coverage

Per `VOC-076-AC-03`, closure requires MC coverage when MC was the failure mode.
Run #227 satisfies the coverage rule: Playwright resolved the MC fieldset and a
disabled MC option button six times. No self-check-only bypass applies.

No subsequent `deploy-staging.yml` run on `develop` after `d305632` was found at
verification time (API query 2026-08-14; latest fifteen runs).

## Package boundaries (VOC-076-AC-04)

| Boundary | Status |
|----------|--------|
| `deploy-staging.yml` edited | **No** (`VOC-076-DEP-01` default scope preserved) |
| VOC-074 `reviewedCards >= 1` hardening | **Intact** — not reached because step 5 failed earlier |
| VOC-050 staging gate structure | **Intact** — core-loop still runs post-deploy; failure is fail-closed |
| VOC-074 cited as closure for #575 | **No** |

## Residual cause analysis (for follow-up T01 iteration)

T00 named sustained `isLoading` (`isSubmitting || isRefetching`) as the
mechanism. T01 removed `isRefetching` from the MC disabled rule
(`isMultipleChoiceOptionDisabled` — only `isSubmitting || phase === "feedback"`).
Run #227 shows MC options can still be disabled on staging with
`aria-pressed="false"`, which under the T01 rules implies either:

1. **`isSubmitting` stuck or slow** — e.g. hung `submitReview` on a prior card
   in the same session (step 5's loop can call `reviewOneCard` more than once), or
2. **`phase === "feedback"`** without a recorded selection — e.g. stale feedback
   phase not cleared before paint despite `useLayoutEffect`, or
3. **Staging API hang** — T00 left log-inconclusive; no client timeout on
   `listDueWords` / `submitReview` (`packages/api-client/src/index.ts`).

T01's `t01-evidence.md` residual-risk note anticipated this outcome: a hung
`listDueWords` leaving the prior card in feedback with MC disabled for the full
timeout. Run #227's 20s failure is consistent with that class of defect even
after the refetch/MC decoupling.

**Recommended next step (out of T02 scope):** a further **VOC-076-T01** iteration
(or a narrowly scoped follow-on package if API timeout is required) — not E2E
timeout inflation alone. T02 records verification; it does not implement fixes.

## Commands inspected during T02

```bash
# Package adoption gate
# specs/changes/VOC-076-.../change.yaml — status: adopted, implementation.authorized: true

# Confirm T01 on develop tip
git log develop --oneline -3
# d305632 VOC-076: VOC-076-T01 (#594)

# Public GitHub API — deploy-staging runs (no GH_TOKEN in this runner)
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/workflows/deploy-staging.yml/runs?per_page=15"
curl -sS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520/jobs"

# Public Actions summary annotations (browser fetch)
# https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31791701520

bash scripts/governance/validate-governance.sh
bash scripts/governance/classify-change-risk.sh
git diff --check
```

No staging secrets, production credentials, or server logs were accessed in this
environment. Log-level confirmation of `submitReview` / `listDueWords` timing
around run #227 remains unavailable here (same limitation as T00).

## Acceptance mapping

| Test | Result |
|------|--------|
| VOC-076-TEST-04 | **Fail** — step 5 did not complete; prior disabled-MC failure mode recurred |
| VOC-076-AC-03 | **Fail** |
| VOC-076-AC-04 | **Pass** |
