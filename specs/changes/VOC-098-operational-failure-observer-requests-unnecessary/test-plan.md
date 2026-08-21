# VOC-098 — Test Plan

## VOC-098-TEST-00 — App mint requests issues-write only

- Covers: `VOC-098-AC-00`
- Preconditions: T00 workflow diff present
- Procedure: Static-read
  `.github/workflows/operational-failure-monitoring.yml` mint step; assert
  `permission-issues: write` present and `permission-actions` absent.
- Expected result: Match; foundation test fails if Actions permission returns.
- Evidence: `VOC-098-EV-00`

## VOC-098-TEST-01 — App mint uses create-github-app-token@v3 with bot secrets

- Covers: `VOC-098-AC-00`
- Preconditions: T00 workflow present
- Procedure: Assert `actions/create-github-app-token@v3`,
  `KARSIFT_BOT_APP_ID`, `KARSIFT_BOT_PRIVATE_KEY`, and fail-closed credential
  gate remain.
- Expected result: App identity path unchanged except permission list.
- Evidence: `VOC-098-EV-00`

## VOC-098-TEST-02 — Issue writer receives App token only

- Covers: `VOC-098-AC-01`
- Preconditions: T00 workflow present
- Procedure: Assert the open/dedupe step env `GH_TOKEN` is
  `steps.app-token.outputs.token`; assert no `github.token` /
  `secrets.GITHUB_TOKEN` wired into `open-failure-issue.sh`.
- Expected result: Issue create/dedupe remains App-only.
- Evidence: `VOC-098-EV-00`

## VOC-098-TEST-03 — Classifier uses non-App job token for Actions jobs API

- Covers: `VOC-098-AC-01`
- Preconditions: T00 workflow present; classifier still called before open
- Procedure: Assert classifier step `GH_TOKEN` is `github.token` (or documented
  equivalent non-App job token); assert workflow/job `permissions` include
  `actions: read`; assert classifier still invokes
  `/actions/runs/${FAILURE_RUN_ID}/jobs` and does not read logs.
- Expected result: VOC-094 classification remains available without App Actions.
- Evidence: `VOC-098-EV-00`

## VOC-098-TEST-04 — Sanitization, marker, concurrency, and fail-closed regressions

- Covers: `VOC-098-AC-02`
- Preconditions: T00 complete
- Procedure: Run `node --test scripts/foundation/voc088-failure-to-issue.test.mjs`
  and `node --test scripts/foundation/voc094-deploy-concurrency.test.mjs`
  (updated assertions). Confirm concurrency group, marker format, unlabeled
  issues, and fail-closed classifier fixtures still pass.
- Expected result: All regression tests green; no permission-actions assertions
  remain for the App mint.
- Evidence: `VOC-098-EV-00`

## VOC-098-TEST-05 — Live observer success on controlled watched non-success

- Covers: `VOC-098-AC-03`
- Preconditions: T00 live on observer execution branch; operator-owned live
  evidence contract for T01
- Procedure: Controlled cancel (or failure) of a watched workflow per
  `docs/operations/staging-controlled-signup.md`; wait for
  `operational-failure-monitoring` conclusion `success`.
- Expected result: Observer run succeeds; mint no longer fails on Actions
  permission.
- Evidence: `VOC-098-EV-01` (metadata only)

## VOC-098-TEST-06 — Single sanitized App issue create-or-dedupe; no duplicate

- Covers: `VOC-098-AC-04`
- Preconditions: TEST-05 observer success for a known fingerprint
- Procedure: Inspect the open issue owning the marker; confirm App authorship,
  unlabeled, allowlisted body fields only; repeat same fingerprint; confirm no
  second open issue for that marker.
- Expected result: Exactly one open owner; duplicate path is no-op.
- Evidence: `VOC-098-EV-01` (metadata only)

## VOC-098-TEST-07 — Operator doc token-role consistency (when doc touched)

- Covers: `VOC-098-AC-05`
- Preconditions: T00 may or may not edit `docs/operations/staging-controlled-signup.md`
- Procedure: If the doc is in the task diff, assert it describes App = issues and
  job token = classifier Actions metadata (or equivalent accurate wording). If
  untouched, assert it does not claim App Actions permission.
- Expected result: No false doc claim about App Actions scope.
- Evidence: `VOC-098-EV-00`

Include positive, negative, authorization, and failure coverage as above. Tests
must not use secrets or production data beyond public Actions metadata and issue
fields already governed as sanitized.
