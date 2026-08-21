# VOC-093 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01**.

## VOC-093-T00 — Diagnose and fix production route-sweep failure

- Requirement source: issue #771; `VOC-093-D00`–`D04`
- Acceptance criteria: `VOC-093-AC-00`, `VOC-093-AC-01`, `VOC-093-AC-02`,
  `VOC-093-AC-03`
- Tests: `VOC-093-TEST-00`–`VOC-093-TEST-05`, regression on `VOC-085-TEST-06`,
  `VOC-085-TEST-07`, `VOC-086-TEST-11`, `VOC-086-TEST-12`
- Evidence: `VOC-093-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Read run 32288703894 job logs for
   `synthetic.production.authenticated-route-content-sweep` at implementation
   time. Record in `t00-evidence.md` which smoke check failed (route label or
   `FAIL:` line). Do not copy secrets, session values, OAuth state, or personal
   data into evidence.
2. Compare with sibling job `synthetic.production.journey-content` in the same
   run to confirm healthz, kill switches, and journey API checks passed.
3. Implement the smallest correct fix:
   - **Harness path:** update `infra/scripts/smoke-test-production.sh` and
     `smoke-test-production.selftest.sh` when the checker is wrong.
   - **Application path:** fix the failing production route in `apps/web/` (or
     shared packages) when evidence shows a real rendering or redirect defect.
   - **Workflow path:** only if evidence shows miswired env on the route-sweep
     job (unlikely given journey-content success); update
     `.github/workflows/scheduled-synthetics.yml` minimally.
4. Extend deterministic tests to lock the fix or a regression fixture matching
   the failing behavior from step 1.
5. Run:
   - `bash infra/scripts/smoke-test-production.selftest.sh`
   - `node --test scripts/foundation/voc085-production-route-sweep.test.mjs`
   - `node --test scripts/foundation/voc086-scheduled-synthetics.test.mjs`
   - governance validation if the task diff touches paths that require it.

### Explicitly out of scope for this task

- Live `workflow_dispatch` proof (T01).
- Weakening route-sweep coverage or accepting sign-in redirects on protected routes.
- Changes to `operational-failure-monitoring.yml`, staging synthetics, or
  VOC-090 staging core-journey caching unless explicitly proven necessary.

## VOC-093-T01 — Record live verification that production route sweep completes green

- Requirement source: issue #771; `VOC-093-D00`, `VOC-093-D04`
- Acceptance criteria: `VOC-093-AC-04`, `VOC-093-AC-05`
- Tests: `VOC-093-TEST-06`
- Evidence: `VOC-093-EV-01` (`t01-evidence.md`)
- Live-evidence contract: `.karsift/live-evidence/VOC-093-T01.yaml` (operator-owned;
  governed reconcile per `docs/operations/live-evidence.md`)
- Status: pending — depends on `VOC-093-T00`

### Required work

1. After T00 merges to `develop`, dispatch `scheduled-synthetics.yml` via
   `workflow_dispatch` with `synthetic_id`:
   `synthetic.production.authenticated-route-content-sweep` (preferred) or empty
   for the full suite. Record which mode was used.
2. Confirm conclusion `success` and job
   `synthetic.production.authenticated-route-content-sweep` success. Record run
   URL, total duration, and job duration — no secrets or personal data.
3. Confirm no new open issue exists with marker
   `<!-- operational-failure:scheduled-synthetics:failure -->` unless issue #771
   remains the sole open fingerprint owner.
4. Optionally note whether the next schedule-triggered hourly run also succeeded,
   or defer that confirmation to package closure if timing does not allow waiting
   one hour within the task window.

### Explicitly out of scope for this task

- Code changes (T00 owns all fixes).
- Closing issue #771 manually outside the normal package roster closure path.

## Task ordering notes

- T00 blocks T01: live proof requires the route-sweep fix on `develop`.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
