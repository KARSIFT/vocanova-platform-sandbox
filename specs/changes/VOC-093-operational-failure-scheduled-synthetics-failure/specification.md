# VOC-093 — Fix scheduled-synthetics production route-sweep failure: Specification

## Objective and requirement source

Remediate the operational failure recorded in
[GitHub issue #771](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/771):
the hourly `scheduled-synthetics` workflow failed because
`synthetic.production.authenticated-route-content-sweep` exited with code 1.

**This draft package does not adopt or authorize itself**; adoption remains a
separate A-004 plan-review / adopt path.

Primary evidence (issue #771 + public run page and jobs API for run 32288703894):

| Item | Value |
|------|-------|
| Workflow | `scheduled-synthetics` |
| Run | [#27](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/32288703894), schedule-triggered 2026-08-19 18:40 UTC |
| Head branch / SHA | `main` @ `7851d570f1dc245a3b310b9705b54b10585c9ea0` |
| Conclusion | `failure` (workflow wall clock ~9m 10s) |
| Failing job | `synthetic.production.authenticated-route-content-sweep` |
| Failing step | Run production authenticated route/content sweep synthetic (~5s, exit 1) |
| Sibling success | `synthetic.production.journey-content` succeeded in the same run |
| Issue origin | VOC-088-T02 `operational-failure-monitoring.yml` sanitized issue body |

Drafting-time repo read of the failing job's path:

1. Checkout
2. Mint production session via `mint-synthetic-session.sh`
3. Invoke `run-scheduled-synthetic.sh production-authenticated-route-content-sweep`
   with `SMOKE_TEST_PROFILE=route-sweep` and kill-switch env matching
   `production-journey-content`
4. Harness runs healthz + journey-content sections, then section **# 6**
   authenticated route sweep (VOC-085-T02): ten fixed web routes and two
   API-derived discover routes, all non-mutating GETs

Because `journey-content` passed in parallel, the session cookie, kill switches,
and journey-situations API content were healthy when the route sweep ran. The
remediation target is therefore the route-sweep section itself or a production
web route it exercises.

## Scope and non-goals

In scope:

1. Read run 32288703894 job logs for
   `synthetic.production.authenticated-route-content-sweep` at implementation
   time. Record which `PASS`/`FAIL` line(s) caused exit 1 in `t00-evidence.md`
   without copying secrets, session values, OAuth state, or personal data.
2. Fix the identified root cause in the smallest correct surface:
   - **Production web regression** — fix the route or server behavior in
     `apps/web/` (or related shared packages) when evidence shows a real user-facing
     failure.
   - **Harness false negative** — fix `smoke-test-production.sh` (and selftests)
     when evidence shows the checker mis-handles an acceptable redirect or response.
   - **Production data edge case** — only when evidence shows missing slugs or
     empty content that journey-content should also have caught; if so, document
     and escalate rather than silently skipping coverage.
3. Preserve the non-mutating route-sweep contract from VOC-085-T02 and VOC-086-T03:
   GET-only, fail closed on sign-in redirects for protected routes, dynamic
   discover routes derived from live API slugs.
4. Extend deterministic tests (`voc085-production-route-sweep.test.mjs`,
   `smoke-test-production.selftest.sh`, and/or `voc086-scheduled-synthetics.test.mjs`)
   to cover the corrected behavior or the regression fixture.
5. Live verification: post-merge `workflow_dispatch` of
   `scheduled-synthetics.yml` targeting
   `synthetic.production.authenticated-route-content-sweep` (or the full suite)
   completes success; record scrubbed run URL in `t01-evidence.md`.

Non-goals / explicitly excluded:

- Weakening route-sweep coverage (skipping routes, removing dynamic discover
  paths, or accepting sign-in redirects on protected routes).
- Changing production signup policy, OAuth availability expectations, or Kuma
  inventory.
- Modifying `operational-failure-monitoring.yml` (the observer worked as designed).
- Re-opening VOC-090 staging core-journey timeout work unless T00 evidence proves
  a unrelated shared regression (default: out of scope).
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Draft package proposal:** **R3** (`infra/scripts/`, `infra/monitoring/`,
  possible `apps/web/` route fix).
- **Measured path floor at drafting:** **R3**. Not R4 unless a task touches
  `scripts/governance/*`.
- Protected areas: `infra/scripts/smoke-test-production.sh`,
  `.github/workflows/scheduled-synthetics.yml`, `infra/monitoring/synthetics.yaml`,
  production synthetic session mint secrets (read-only use only).
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-093-D00`: Run 32288703894 failed because job
`synthetic.production.authenticated-route-content-sweep` returned exit code 1 from
`smoke-test-production.sh` profile `route-sweep`. This is a functional smoke failure,
not a timeout or cancellation.

`VOC-093-D01`: Sibling job `synthetic.production.journey-content` succeeded in the
same run, bounding the failure to route-sweep web GET checks (section **# 6**) or
harness logic executed only in that section — not session mint, healthz, or
journey-situations API availability.

`VOC-093-D02`: The fix must preserve the VOC-085 non-mutating route-sweep contract:
ten fixed routes, two API-derived discover routes, fail closed on missing session
cookie and sign-in redirects for protected routes.

`VOC-093-D03`: If job logs show a real production route returns non-2xx, redirects
to sign-in with a valid minted session, or loops redirects, the primary fix is the
production route behavior (or seed/content if legitimately empty). Harness changes
that weaken assertions are not an acceptable default.

`VOC-093-D04`: If job logs show journey-content checks would fail on re-run (transient
production blip), T01 live verification still required; do not close the package on
theory alone.

## Open questions for the reviewing human

1. Accept proposed **R3**, or raise in writing if the adopting human treats
   production route-sweep remediation as R4 operational risk.
2. After T00 log review, confirm whether the fix belongs in `apps/web/` (product
   regression) or harness-only — the task PR must record the chosen branch.
3. If the failing route is `/discover/{slug}` dynamic content, confirm whether
   content seeding (VOC-085) needs a follow-up rather than expanding this package.

## Data, migrations, analytics, and accessibility

- No application schema migration unless T00 proves an unrelated data defect
  requiring a separate governed package.
- No intentional database mutation from the synthetic (read-only GETs).
- Possible product UI fix if a route fails to render — scoped by T00 evidence.
- No analytics change — evidence-backed non-applicability unless a route fix
  incidentally touches analytics (record in task PR if so).
- Accessibility impact follows any route fix; no standalone a11y scope unless a
  route fix requires it (record in task PR).
