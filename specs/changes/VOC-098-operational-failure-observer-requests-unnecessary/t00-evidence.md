# VOC-098-T00 evidence — least-privilege App token and classifier token split

Recorded at implementation time. No secrets, tokens, or live run metadata.

## Diagnosis (before → after)

| Item | Before (VOC-094-T00) | After (VOC-098-T00) |
|------|----------------------|---------------------|
| App mint permissions | `permission-issues: write` + `permission-actions: read` | `permission-issues: write` only |
| Workflow `permissions` floor | `contents: read` | `contents: read`, `actions: read` |
| Classifier `GH_TOKEN` | `steps.app-token.outputs.token` (App) | `github.token` (job token) |
| Issue writer `GH_TOKEN` | `steps.app-token.outputs.token` (App) | unchanged — App only |

Root cause (issue #840): the App installation does not grant Actions, so
`actions/create-github-app-token` failed when `permission-actions: read` was
requested — including for `scheduled-synthetics` failures that never invoke the
classifier.

## Files changed

- `.github/workflows/operational-failure-monitoring.yml` — drop App Actions
  permission; add workflow `actions: read`; wire classifier to `github.token`
- `scripts/foundation/voc088-failure-to-issue.test.mjs` — assert no App Actions
  permission; assert `actions: read` workflow floor
- `scripts/foundation/voc094-deploy-concurrency.test.mjs` — assert dual-token
  split (classifier = job token, issue writer = App token)
- `docs/operations/staging-controlled-signup.md` — document dual-token contract

`infra/scripts/classify-deploy-concurrency-cancel.sh` unchanged (env wiring only
in the workflow).

## Deterministic validation (VOC-098-EV-00)

Recorded at implementation time (`2026-08-21`):

```text
$ node --test scripts/foundation/voc088-failure-to-issue.test.mjs
✔ VOC-088-TEST-08 through TEST-11 (4/4 pass)

$ node --test scripts/foundation/voc094-deploy-concurrency.test.mjs
✔ VOC-094-TEST-00 through TEST-05 (7/7 pass)

$ bash scripts/governance/validate-governance.sh
Repository foundation validation passed.
Monitoring impact validation passed.
Governance structure validation passed.

$ bash scripts/governance/classify-change-risk.sh
Detected path-based risk floor: R3
  - .github/workflows/operational-failure-monitoring.yml

$ git diff --check
(no whitespace errors)
```

All suites exited 0.

## Acceptance criteria mapping

| AC | Status | Notes |
|----|--------|-------|
| VOC-098-AC-00 | satisfied (deterministic) | App mint requests issues-write only; tests assert no `permission-actions` |
| VOC-098-AC-01 | satisfied (deterministic) | Classifier on `github.token`; `open-failure-issue.sh` on App token only |
| VOC-098-AC-02 | satisfied (regression) | VOC-088/VOC-094 foundation suites pass unchanged behavior |
| VOC-098-AC-05 | satisfied | `staging-controlled-signup.md` documents dual-token contract |

Live proof (VOC-098-AC-03, AC-04) is operator-owned in T01 per
`.karsift/live-evidence/VOC-098-T01.yaml`.
