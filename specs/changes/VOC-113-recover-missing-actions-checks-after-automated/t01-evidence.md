# VOC-113-T01 — Promotion recovery evidence

Date: `2026-08-25`

This record contains allowlisted GitHub metadata only. It contains no workflow
logs, credentials, token values, or application data. Auto-advance created draft
carrier PR `#955`; no implementer run was started for this operator-owned task.

## Exact identities

- Evidence carrier: PR `#955`, branch
  `agent/voc-113-voc-113-t01`.
- Promotion PR: `#947`, `develop` → `main`.
- Promotion head: `88a94d49ef0d37ef030464ba20b5ed055d21161f`.
- Promotion merge result on `main`:
  `fea24f18c434f9e4add33d86f784dfa3ca381789`.
- Carrier head verified by the operator workflow:
  `231c0c82d1be925bc9845172715ea98c538a892e`.
- Qualified result head created by the repository-controlled reconciler:
  `571f6fee4feea849c20a4e1ab9641524e0cf89b9`.

## Genuine exact-head checks and merge outcome

All required checks completed successfully on promotion head
`88a94d49ef0d37ef030464ba20b5ed055d21161f`:

- `governance-policy`: run `32735430356`, job `97456970671`, `success`;
- `validate`: newest run `32736408968`, job `97460183750`, `success`;
- `ci / ci`: run `32735432092`, job `97457005161`, `success`.

Release run `32735432092` selected the authoritative exact-head evidence. Its
`release / converge` job `97457936739` completed with `success`, and PR `#947`
merged once as `fea24f18c434f9e4add33d86f784dfa3ca381789`. No manual merge or
fabricated check was used.

## Contract-bound verification

Operator workflow-dispatch run `32787318872` executed on exact carrier head
`231c0c82d1be925bc9845172715ea98c538a892e`:

- workflow: `pipeline.yml` / `pipeline`;
- event: `workflow_dispatch`;
- `verify-promotion-check-recovery / verify`, job `97621787019`: `success`;
- started `2026-08-24T23:00:12Z`, completed `2026-08-24T23:00:26Z`;
- duration: `14` seconds.

Repository-controlled observe run `32787642471` qualified that allowlisted
metadata under `.karsift/live-evidence/VOC-113-T01.result.json`, published the
trusted App attestation, and advanced the carrier without force-pushing.

Two earlier operator dispatch attempts, runs `32786665074` and `32786804489`,
failed before job creation with `startup_failure` because the old carrier branch
lacked current job-scoped permissions required by the reusable release workflow.
The carrier was merged forward to current `main`; the PR-visible task diff
remained evidence-only, and the exact-head verifier then succeeded. Those failed
runs remain visible in GitHub history.

## Validation posture

- `VOC-113-TEST-08`: passed by the exact-head checks, single release-converge
  merge, contract-bound verifier success, and trusted qualified result above.
- No status fabrication, ruleset weakening, log replay, secret exposure, or
  implementer-owned Actions dispatch occurred.
- The qualified result head requires a fresh exact-SHA independent review before
  this draft carrier may be marked ready and merged.
