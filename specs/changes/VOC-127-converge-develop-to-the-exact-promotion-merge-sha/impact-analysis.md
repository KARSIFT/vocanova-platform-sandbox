# VOC-127 — Impact Analysis

## Security and privacy

This package repairs `release.yml` so a successful `--merge` promotion
advances the long-lived integration ref to the exact merge commit, and adds
an exceptional adopted path to reconcile authorized main-only hotfix/evidence
back onto `develop`. It does not introduce new secret values, OAuth/session
material, production-data access, or user-facing data flows. It does not
rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY` and does not change
the `karsift-ai-infra-bot` installation.

Security controls that must remain:

- Promotion still uses one exact-head `--merge --match-head-commit`; the merge
  commit is preserved as audit evidence.
- Develop-ref updates use compare-and-swap / non-force Git ref mutation
  against an expected current SHA. Unique develop commits fail closed and are
  never erased.
- Exceptional main-only reconciliation identity is a merged PR number plus
  adopted package/task bindings, not free-form SHAs on caller
  `workflow_dispatch`.
- The model-controlled implementer runner never receives the GitHub App
  token and still has no general `actions` permission.
- Release App token remains limited to contents/issues/pull-requests write;
  recovery metadata reads stay on the job `GITHUB_TOKEN`.
- No credential values are printed in logs, tests, or evidence.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-127-R00`: **High correctness risk** if `release.yml` still leaves
  `develop` at `CHECKED_HEAD_SHA` after `--merge`, repeating the #1033 / 0/46
  ancestor gap. Mitigation: `VOC-127-D01`, `VOC-127-AC-00`.
- `VOC-127-R01`: **High data-loss risk** if sync force-updates `develop` over
  unique commits. Mitigation: `VOC-127-D02`, `VOC-127-AC-01`.
- `VOC-127-R02`: **High release-loop risk** if the develop-ref push opens a
  new promotion PR or if `ahead_by == 0` with unequal SHAs is treated as
  done. Mitigation: `VOC-127-D03`, `VOC-127-D04`, `VOC-127-AC-02`.
- `VOC-127-R03`: **High authority risk** if exceptional main-only
  reconciliation accepts operator SHAs or runs without an adopted package,
  normalizing direct-to-main. Mitigation: `VOC-127-D06`, `VOC-127-AC-05`.
- `VOC-127-R04`: **Medium operational risk** if tree-equivalent develop sync
  schedules a full staging deploy. Mitigation: `VOC-127-D05`, `VOC-127-AC-04`.
- `VOC-127-R05`: **Medium coverage risk** if tests still assert restoration at
  `CHECKED_HEAD_SHA` (`test_release_policy.py` today) and miss the already-merged
  reconcile path. Mitigation: `VOC-127-D08`, `VOC-127-TEST-00`,
  `VOC-127-TEST-05`, `VOC-127-TEST-07`.
- `VOC-127-R06`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is workflow/config/test/doc reversion.
  Exceptional main-only recon may introduce already-authorized main tree
  onto staging's source branch, still gated by VOC-111 path selection.
- Protected surfaces: `KARSIFT/karsift-ai-infra` `release.yml` / helpers /
  tests / README, caller pipeline and `deploy-staging.yml`, caller
  `tooling/governance/` fixtures and tests, current-state release/branch docs,
  and this package directory.
- `VOC-127-DEP-00` through `VOC-127-DEP-07`: see `change.yaml`.
- `VOC-127-EV-00`: T00 evidence — bind-and-sync mechanism, race/unique-commit
  fail-closed proof, reconcile-release already-merged path, staging skip,
  exceptional main-only authority, validation commands, exact infra SHA, and
  pin applicability.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because it changes protected CI/CD orchestration and
`tooling/governance/` fixtures, but the path classifier and independent
verifier remain authoritative. This is a draft proposal, not a determination.
