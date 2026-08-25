# VOC-120 — Impact Analysis

## Security and privacy

This package removes a generated interpreter artifact, adds portable ignore
rules, and adds source-tree hygiene validation. No new secrets, OAuth/session
material, production data access, or user-facing data flows are introduced.

The tracked `.pyc` is not a credential store, but it is a committed binary on
protected branches. Untracking it reduces accidental binary drift. Validation
must inspect tracked path names only and must not copy bytecode, secrets, or CI
logs into evidence.

## Data and migrations

None. No database, schema, seed, or migration changes are proposed. Git history
is not rewritten; the historical blob remains in prior commits.

## Analytics and accessibility

None for product/runtime behavior. There is no analytics instrumentation or
user-interface accessibility effect.

## Risks, dependencies, and evidence

- `VOC-120-R00`: **Medium hygiene risk** if ignore rules are added but the
  already-tracked blob is left in the index. Mitigation: explicit untrack plus
  `git ls-files` fail-closed validation (`VOC-120-D00`, `VOC-120-D04`).
- `VOC-120-R01`: **Medium operational risk** if infra continues to rely on
  private exclude-file entries. Mitigation: repository-root `.gitignore` and
  infra tests (`VOC-120-D02`).
- `VOC-120-R02`: **Medium governance risk** if caller validation is added only as
  an unhooked unittest and `validate-governance.sh` still passes with tracked
  bytecode. Mitigation: invoke the caller check from
  `scripts/governance/validate-governance.sh` and from `tooling/governance/tests/`.
- `VOC-120-R03`: **Low false-positive risk** if validation treats untracked,
  gitignored cache directories as failures. Mitigation: inspect tracked files
  (`git ls-files`), not the working tree.
- `VOC-120-R04`: **Low release risk** because no runtime deployment change is
  intended; rollback is revert of ignore/validation commits and does not require
  re-adding bytecode.
- Protected surfaces: `scripts/governance/`, `tooling/governance/`, caller
  `infra/scripts/`, `KARSIFT/karsift-ai-infra` `.gitignore` and tests, and this
  package directory.
- `VOC-120-DEP-00`: issue #987 provides the change objective and reproduction.
- `VOC-120-DEP-01`: A-004 is active; founder comment gates stay removed, while
  exact-SHA independent verification and fail-closed enforcement remain mandatory.
- `VOC-120-DEP-02`: caller ignore rules exist but do not untrack the committed blob.
- `VOC-120-DEP-03`: shared infra has no repository `.gitignore` at the cited SHA.
- `VOC-120-DEP-04`: coordinated caller and infra PRs remain one task; pin only if
  the caller fixture consumes the infra change.
- `VOC-120-EV-00`: T00 evidence — untracked paths, ignore-rule text, validation
  commands/results, infra SHA, and pin applicability.

## Monitoring

`monitoring_impact.state: none` — no monitor or synthetic inventory changes.

## Authority and EHR

Active authority model: **A-004**. No founder `approved` comment is required for
engineering-workflow adopt/merge/release gates. EHR is not triggered.

This draft proposes **R4** because durable caller validation belongs under
`scripts/governance/` and/or `tooling/governance/`, but the path classifier and
independent verifier remain authoritative.
