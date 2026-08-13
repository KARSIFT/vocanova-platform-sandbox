# VOC-076-T00 — Disabled-button root-cause evidence

**Task:** VOC-076-T00  
**Package:** VOC-076  
**Evidence ID:** VOC-076-EV-00  
**Investigation date:** 2026-08-13  
**Reviewed revision:** `c8f6a7f` (current branch tip; failing staging deploy was `e3f732c` — `review-session.tsx` and `reviewOneCard` logic are unchanged between those revisions)

## Confirmed root cause

**Sustained `isLoading === true` in `review-session.tsx` keeps multiple-choice
option buttons disabled while the MC fieldset remains visible.** The only
in-product rule that produces a visible MC group with `disabled` buttons and
`aria-pressed="false"` on an unselected option during learner input is:

```170:170:apps/web/src/app/(app)/reviews/_components/review-session.tsx
  const isLoading = isSubmitting || isRefetching;
```

```270:276:apps/web/src/app/(app)/reviews/_components/review-session.tsx
                const isDisabled = isLoading || phase === "feedback";
                return (
                  <button
                    key={option.meaningId}
                    type="button"
                    aria-pressed={isSelected}
                    disabled={isDisabled}
```

`isSubmitting` is set around `submitReview` (`review-session.tsx:125-167`);
`isRefetching` is set around the batch-end `listDueWords` refetch in `advance()`
(`review-session.tsx:80-103`). The API client uses bare `fetch` with **no
request timeout** (`packages/api-client/src/index.ts:797-818`), so either call
can hold `isLoading` true for the full Playwright test timeout when staging
network or API latency stalls.

**T01 must include a product-side fix** (not E2E timeout inflation alone).
E2E enabled-wait hardening is still recommended as defense-in-depth (see below).

## Evidence chain

### 1. Run 31748423831 — authoritative failure log (public Actions summary)

| Item | Value |
|------|--------|
| Run | [31748423831](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831) |
| Trigger | `develop` push after VOC-075 plan merge (`e3f732c`) |
| Job / step | `deploy to staging` → `Run the staging core-loop journey` |
| Failing line | `core-loop.staging.spec.ts:258` inside `reviewOneCard` |
| Timeout | 240000 ms (full test timeout) |

Playwright call log (from the run annotation):

```
waiting for getByRole('group', { name: /^Choose the meaning for / }).getByRole('button').first()
  locator resolved to <button disabled type="button" aria-pressed="false" ...>
attempting click action
  2 × waiting for element to be visible, enabled and stable
    - element is not enabled
  ... (retries for ~240s) ...
  - element was detached from the DOM, retrying
```

**Interpretation:**

- The MC fieldset **did render** (not VOC-074's vacuous empty-queue case).
- The locator's first option stayed **`disabled` with `aria-pressed="false"`**
  for the entire enabled-wait window — consistent with `isLoading === true`
  during the prompt phase, not with a completed feedback-phase selection
  (`aria-pressed="true"` on the chosen option).
- Late **detach** is consistent with `advance()` / refetch eventually replacing
  `dueWords`, index change, or caught-up remount after the prolonged stall.

### 2. Product disabled-state path — confirmed mechanism

Investigation order from `tasks.md` §1:

| State flag | Set where | Cleared where | Disables MC options when true |
|------------|-----------|---------------|-------------------------------|
| `isSubmitting` | `submitAttempt` L125 | `finally` L165-167 | Yes (`isLoading`) |
| `isRefetching` | `advance` batch-end L80 | `finally` L101-103 | Yes (`isLoading`) |
| `phase === "feedback"` | MC `onClick` L277-280 | `useEffect` on card change L67-72 | Yes (direct) |

On a **fresh** `/reviews` mount, both flags initialise `false`
(`review-session.tsx:43-44`) and `phase` is `"prompt"` (`review-session.tsx:47`),
so MC buttons should be enabled. A **240s** disabled window therefore implies a
**prior in-session async operation** (submit or batch refetch) whose promise
did not settle before the next MC interaction, **or** a sustained staging
`fetch` hang — not a one-frame render glitch alone.

`advance()` after the last card in a batch fires `listDueWords` without awaiting
it in the caller; `reviewOneCard` can return while `isRefetching` is still
`true` if the refetch is slow (`review-session.tsx:80-103` plus
`core-loop.staging.spec.ts:258-268` flow). The step-5 outer loop can then call
`reviewOneCard` again while `isLoading` is still elevated.

### 3. E2E readiness gap — confirmed, but not sufficient as sole root cause

```241:258:apps/web/tests/staging-e2e/core-loop.staging.spec.ts
  await expect(
    showAnswerButton.or(multipleChoiceGroup).or(caughtUpHeading).first(),
  ).toBeVisible();
  // ...
  await multipleChoiceGroup.getByRole("button").first().click();
```

`reviewOneCard` waits only for **visibility** of the MC group (or self-check /
caught-up) before clicking. That is a real readiness gap relative to
`VOC-076-AC-02`.

However, Playwright's `locator.click()` **already auto-retries until the
target is enabled** (documented in the run call log:
`waiting for element to be visible, enabled and stable` / `element is not
enabled`). The test burned the **full 240s** on that enabled wait. Therefore:

- **Ruled out as sole root cause:** "E2E race — click before enabled wait."
  A missing explicit `toBeEnabled` does not explain a multi-minute hang; the
  button was genuinely disabled in the product for the timeout duration.
- **Still required in T01:** explicit enabled / readiness assertion for clearer
  failure messages and to guard brief sub-second readiness windows (see §5).

### 4. Staging API latency / errors — inconclusive (log access unavailable)

Checked:

- Public Actions summary for run 31748423831 (above).
- API client: no `AbortSignal` / timeout on `listDueWords` or `submitReview`
  (`packages/api-client/src/index.ts:549-565`, `595-613`, `797-818`).

Not available in this environment:

- GitHub Actions job log stream (`gh run view` returned auth error — no
  `GH_TOKEN`).
- Staging API server logs for `listDueWords` / `submitReview` around the run
  timestamp (~2026-08-13 22:04 UTC).
- Live staging reproduction (no synthetic-session credentials in this runner).

**Status:** Staging API hang or extreme latency is **consistent** with
sustained `isLoading` and the 240s enabled-wait, but **not directly proven** from
server-side logs. T02 staging verification remains the authoritative proof that
T01 fixes the failure mode.

### 5. Detach / re-render — explained by card replacement after stall

Late detach in the call log aligns with code paths that replace the card tree:

- `advance()` refetch: `setDueWords` + `setCurrentIndex(0)` (`review-session.tsx:85-88`)
- In-batch advance: `setCurrentIndex` (`review-session.tsx:75-77`)
- Queue empty: `setCompleted(true)` (`review-session.tsx:89-90`)

Any of these remount the MC `<button>` nodes after a long `isLoading` window,
matching "disabled for ~240s, then detached."

**Secondary product defect (brief windows, not 240s alone):** phase reset runs in
`useEffect` **after paint** when `currentIndex` or `dueWords` changes
(`review-session.tsx:67-72`). For one render, a new card can show MC options
while `phase` is still `"feedback"` from the prior card, disabling all options
with `aria-pressed="false"`. This matches the failure DOM signature but normally
clears on the next commit; T01 should still harden this (e.g. `useLayoutEffect`
or deriving prompt readiness without a stale `phase`).

**Inconclusive / noted:** `buildMultipleChoiceOptions` calls `shuffleArray` on
every render (`review-session.tsx:412-435`), which can reorder options between
renders and may contribute to Playwright "detached" retries; it does not by
itself set `disabled` and is not the primary 240s cause.

### 6. VOC-074 distinction — ruled out

VOC-074's confirmed failure mode is `reviewedCards = 0` with an empty or
caught-up queue (no card interaction). Run 31748423831 resolved an MC fieldset
and a disabled answer button — a card was present and the journey reached the MC
click site. This is a **distinct** defect from VOC-074's vacuous-pass /
queue-exhaustion path.

## Candidate summary

| Candidate | Status | Notes |
|-----------|--------|-------|
| Stuck `isLoading` (`isSubmitting` / `isRefetching`) | **Confirmed mechanism** | Only product rule matching 240s `disabled` + `aria-pressed="false"` on prompt-ready MC |
| Hung `listDueWords` in `advance()` | **Consistent, log-inconclusive** | No client timeout; refetch can overlap next `reviewOneCard` |
| Hung `submitReview` in `submitAttempt` | **Consistent, log-inconclusive** | Same no-timeout `fetch`; would also block `reviewOneCard` at rating click |
| E2E visibility-only wait (no enabled assertion) | **Confirmed gap, ruled out as sole cause** | Playwright `click()` already waited for enabled 240s |
| `phase === "feedback"` stale across card change | **Confirmed brief defect** | `useEffect` reset lag; insufficient alone for 240s |
| Staging API latency / errors | **Inconclusive** | No log access from this environment |
| Detach / re-render | **Explained** | Card replacement after prolonged stall |
| VOC-074 empty-queue vacuous pass | **Ruled out** | MC group rendered on this run |

## Implications for T01

| Area | Direction |
|------|-----------|
| **Product (`review-session.tsx`)** | **Required.** Narrow when MC options are disabled during prompt phase — e.g. ensure `isRefetching` does not block interaction on a prompt-ready card after refetch data is applied; reset `phase` synchronously on card change; verify `finally` paths always clear flags. Do not remove intentional disabled states during active submit or post-selection feedback. Preserve CSRF, session handling, a11y, and `max-w-[28rem]` workaround. |
| **E2E (`core-loop.staging.spec.ts`)** | **Required (defense-in-depth).** Add explicit `toBeEnabled` (or equivalent) on the MC answer control before click; do not rely on timeout inflation alone. Preserve VOC-074 `reviewedCards >= 1` hardening. |
| **`deploy-staging.yml`** | **Out of scope** unless T01 investigation during implementation proves diagnostic seeding is necessary (`VOC-076-DEP-01`). |
| **API (`apps/api/`)** | **Not in default T01 scope.** No log proof of API defect; if T01 product + E2E changes do not close T02, open a separate package for API hang investigation. |

## Commands run during T00

```bash
# Package adoption gate
# Read specs/changes/VOC-076-.../change.yaml — status: adopted, implementation.authorized: true

# Source inspection (paths above)
# apps/web/src/app/(app)/reviews/_components/review-session.tsx
# apps/web/tests/staging-e2e/core-loop.staging.spec.ts
# packages/api-client/src/index.ts

# Failing revision diff check
git show e3f732c:apps/web/src/app/\(app\)/reviews/_components/review-session.tsx
# — identical disabled-state logic to current tip

# Attempted CI log fetch (blocked)
gh run view 31748423831 --log-failed
# → gh: set GH_TOKEN (401 in this runner)

# Public run summary (browser)
# https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31748423831
```

No product fix was written in this task (investigation only).
