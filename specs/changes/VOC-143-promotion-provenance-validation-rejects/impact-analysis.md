# VOC-143 — Impact Analysis

## Security and privacy

This package repairs VOC-112 capture provenance so a legitimate current
`AGENTS.md` documentation update cannot fail promotion-path required
`validate` (`squash-safe-push`) or `ci / ci` (promotion `pr-validation`) while
the historical fixture remains unmodified. It does not introduce new secret
values, OAuth/session material, production-data access, or user-facing data
flows. It does not rotate `KARSIFT_BOT_APP_ID` / `KARSIFT_BOT_PRIVATE_KEY`.
It does not change the VOC-140 two-token contract.

Security controls that must remain:

- `squash-safe-push` binds fixture `agents_sha256` to an immutable ancestor
  of the validation tip. A hash that is not present on that ancestry fails
  closed.
- Promotion `pr-validation` keeps navigator hashes HEAD-bound and still
  rejects a non-ancestor promotion base.
- Ordinary `pr-validation` stays merge-base anchored. `local` and
  `pr-ancestry` retain working-tree `AGENTS.md` equality.
- Promotion check identity is unchanged: `validate` stays `squash-safe-push`
  for same-repository `main` ← `develop`; `ci / ci` stays `--promotion-pr` /
  `pr-validation`.
- The production merge guard still requires an effective active
  repository-owned ruleset, a pull-request rule, strict non-empty required
  checks, and `bypass_actors: []`.
- Exact-head binding remains fail-closed through the App-authored
  independent-review comment/check and merge-gate mismatch rejection.
- `config/roles.yml` is unchanged; no OpenAI execution route is added.
- Raw errors remain sanitized. No credential values are printed in logs,
  tests, or evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect. Provenance assertions are CI
orchestration only.

## Risks, dependencies, and evidence

- `VOC-143-R00`: **High release-safety risk** if `squash-safe-push` still
  requires working-tree `AGENTS.md` equality, repeating promotion `validate`
  failure on PR #1119. Mitigation: `VOC-143-D01`, `VOC-143-AC-00`,
  `VOC-143-TEST-00`.
- `VOC-143-R01`: **High attestation-safety risk** if working-tree equality is
  skipped without an ancestor bind, accepting a tampered fixture hash.
  Mitigation: `VOC-143-D02`, `VOC-143-TEST-01`.
- `VOC-143-R02`: **High release-safety risk** if promotion `pr-validation`
  remains HEAD-bound for `agents_sha256`, so `ci / ci` on #1119 still fails.
  Mitigation: `VOC-143-D03`, `VOC-143-AC-02`, `VOC-143-TEST-05`.
- `VOC-143-R03`: **High correctness risk** if `local`, `pr-ancestry`, or
  ordinary merge-base `pr-validation` are weakened. Mitigation:
  `VOC-143-D04`, `VOC-143-TEST-02`, `VOC-143-TEST-03`, `VOC-143-TEST-04`.
- `VOC-143-R04`: **High correctness risk** if navigator HEAD-binding is
  dropped so merge-base-only hashes pass promotion `pr-validation`.
  Mitigation: `VOC-143-D03`, `VOC-143-TEST-06`.
- `VOC-143-R05`: **High process risk** if promotion check identity is
  switched (`ci.yml` to `--squash-safe-push`, or `validate` away from
  `squash-safe-push`) or required checks are bypassed. Mitigation:
  `VOC-143-D05`, `VOC-143-TEST-07`.
- `VOC-143-R06`: **High coverage risk** if tests only comment the equality
  skip without exercising ancestor bind and fail-closed tamper. Mitigation:
  `VOC-143-D07`, `VOC-143-TEST-00`, `VOC-143-TEST-01`.
- `VOC-143-R07`: **High review-safety risk** if committed evidence is
  required to contain the SHA of the same commit that contains it.
  Mitigation: `VOC-143-D12`, `VOC-143-TEST-08`.
- `VOC-143-R08`: **High delivery risk** if this package snapshots the
  current develop/main gap (`karsift-ai-infra#15`), manually merges #1119,
  creates a duplicate promotion PR, or adds a VOC-097 evidence-carrier task.
  Mitigation: `VOC-143-D09`, `VOC-143-TEST-09`.
- `VOC-143-R09`: **Medium documentation risk** if current docs still claim
  `squash-safe-push` requires working-tree `AGENTS.md` equality. Mitigation:
  `VOC-143-D08`, `VOC-143-TEST-07`.
- `VOC-143-R10`: **Low application-runtime release risk** because ordinary
  sync is tree-equivalent; rollback is test/doc reversion to the last
  reviewed `develop` merge.
- `VOC-143-R11`: **Medium classification risk** if `repository-governance.yml`
  is edited unnecessarily and silently raises the path floor to R4.
  Mitigation: `VOC-143-D06`; prefer not to edit that workflow.
- Protected surfaces: `scripts/foundation/voc112-navigation-benchmark.test.mjs`,
  this package directory, and named current-state docs.
- `VOC-143-DEP-00` through `VOC-143-DEP-09`: see `change.yaml`.
- `VOC-143-EV-00`: T00 evidence — implementation PR base, provenance-assertion
  repair, negative-case results, validation after commit, exact-head binding
  contract, and release handoff.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.
This is governance CI provenance repair, not a product route or critical API
endpoint.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required
for engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R3** because it changes required-check provenance used
before production promotion (CI/CD / governance enforcement). The path
classifier may report R1 for `scripts/foundation/*.mjs`. The path classifier
and independent verifier remain authoritative. This is a draft proposal, not
a determination.
