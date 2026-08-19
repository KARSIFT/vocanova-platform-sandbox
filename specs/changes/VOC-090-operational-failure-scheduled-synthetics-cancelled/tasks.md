# VOC-090 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-090-T00 — Fix staging core-journey CI budget in scheduled-synthetics

- Requirement source: issue #759; `VOC-090-D00`–`D04`
- Acceptance criteria: `VOC-090-AC-00`, `VOC-090-AC-01`, `VOC-090-AC-02`,
  `VOC-090-AC-03`, `VOC-090-AC-04`
- Tests: `VOC-090-TEST-00`–`VOC-090-TEST-05`, regression on `VOC-086-TEST-11`,
  `VOC-086-TEST-12`
- Evidence: `VOC-090-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Read run 32271016931 job logs for `synthetic.staging.authenticated-core-journey`
   at implementation time. Record in `t00-evidence.md` which phase consumed the
   budget (dependency install, Playwright install, SSH seed, journey execution).
   Do not copy secrets, session values, OAuth state, or personal data into evidence.
2. Add pnpm and Playwright browser caching to the
   `staging-authenticated-core-journey` job in `.github/workflows/scheduled-synthetics.yml`.
   Prefer the same cache keys / store paths `deploy-staging.yml` would use if a
   shared composite is introduced; document the choice in the task PR.
3. Set `timeout-minutes` on the staging core-journey job to a value that covers:
   - corrected setup with warm cache (measured or conservatively estimated),
   - SSH seed (2m command timeout),
   - session mint,
   - `playwright.staging.config.ts` journey timeout (240s),
   - review-card loop headroom (`MAX_REVIEW_CARDS = 8`).
   Minimum: match `deploy-staging.yml`'s proven 40-minute deploy-job budget for
   the same journey unless evidence shows a lower safe value with caching.
4. Update `infra/monitoring/synthetics.yaml`
   `synthetic.staging.authenticated-core-journey.timeout_seconds` to equal the
   job wall clock in seconds. Add or extend a comment in the registry or workflow
   header documenting that this field tracks the GitHub Actions job budget.
5. Optionally extend `infra/monitoring/scheduled-synthetics.mjs` with a validator
   that compares job `timeout-minutes` to registry `timeout_seconds` for the staging
   core-journey check_ref.
6. Add deterministic tests (`scripts/foundation/voc090-scheduled-synthetics-budget.test.mjs`
   or extend `voc086-scheduled-synthetics.test.mjs`) covering caching presence,
   timeout/registry alignment, and unchanged SSH-seed → mint → Playwright ordering.
7. Run `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs` and
   the new VOC-090 tests; run governance validation if the task diff touches paths
   that require it.

### Explicitly out of scope for this task

- Live `workflow_dispatch` proof (T01).
- Application or Playwright test logic changes unless `VOC-090-D03` is proven and
  explicitly escalated in the task PR (default: defer to a follow-up package).
- Changes to `operational-failure-monitoring.yml`, `deploy-staging.yml` (unless
  a shared composite is explicitly chosen and justified), or production synthetic
  harness scripts.

## VOC-090-T01 — Record live verification that the hourly suite completes green

- Requirement source: issue #759; `VOC-090-D00`
- Acceptance criteria: `VOC-090-AC-05`, `VOC-090-AC-06`
- Tests: `VOC-090-TEST-06`
- Evidence: `VOC-090-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-090-T00`

### Required work

1. After T00 merges to `develop`, dispatch `scheduled-synthetics.yml` via
   `workflow_dispatch` with empty `synthetic_id` (full suite) or with
   `synthetic.staging.authenticated-core-journey` if staging-only proof is
   sufficient for the first verification pass. Record which mode was used.
2. Confirm conclusion `success` and that job
   `synthetic.staging.authenticated-core-journey` completes within the new job
   timeout. Record run URL, total duration, and job duration — no secrets or
   personal data.
3. Confirm no new open issue exists with marker
   `<!-- operational-failure:scheduled-synthetics:cancelled -->` unless issue
   #759 remains the sole open fingerprint owner.
4. Note whether a subsequent schedule-triggered hourly run also succeeded, or
   defer that confirmation to package closure if timing does not allow waiting
   one hour within the task window.

### Explicitly out of scope for this task

- Code changes (T00 owns all workflow/registry edits).
- Closing issue #759 manually outside the normal package roster closure path.

## Task ordering notes

- T00 blocks T01: live proof requires the caching and timeout fix on `develop`.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
