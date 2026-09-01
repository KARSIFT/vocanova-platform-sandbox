# VOC-145-T00 — Evidence

Task: `VOC-145-T00` — Governed reconciliation of the unauthorized
role-binding and VOC-117 expectation drift.

Do not record secrets, credentials, session values, OAuth material, personal
data, complete CI logs, raw provider responses, or App token values.

This file is the T00 evidence contract. Implementation fills the recorded
fields below after the repair is tracked and committed. Do not require this
file to contain the SHA of the same commit that contains it.

## Implementation PR base

Recorded before the first in-scope edit:

`10df745c42ed283405d6bdf5b01180afdfed7d26`

Issue-creation infrastructure pin was
`8993e867640dfb604dec0466c4e0787e68d8e258`.
Plan/adoption/roster commits after that SHA are governance-only and do not
count as protected-file drift.

## Infrastructure merge

Coordinated infrastructure carrier consumed by this caller pin:

`ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f`

Publication: `KARSIFT/karsift-ai-infra#175` (`agent/voc-145-voc-145-t00` →
`main`), head `ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f`, parent
`d8720829b176cf1287e633f9382989fc8f258105`, message
`VOC-145: VOC-145-T00 coordinated source carrier (attempt 1)`.

Independent exact-revision review and merge of PR #175 to live `main` are
required before caller `@main` reusable workflows and this pin are jointly
authoritative. Attempt 1 incorrectly recorded non-existent SHA
`a724f05bc5eae29076953295dc03f68367a92185`; remediation binds the verified
carrier head above. Must not pin `d8720829b176cf1287e633f9382989fc8f258105`.

Last governed VOC-142 infrastructure base (historical audit):
`8993e867640dfb604dec0466c4e0787e68d8e258`.

Unauthorized infra head at issue creation (audit only, not a pin target):
`d8720829b176cf1287e633f9382989fc8f258105`.

## Authorized path

Record Path A (default restore) or Path B (only if adoption recorded
`VOC-145-DEP-07`):

`Path A` — adoption did not record `VOC-145-DEP-07` Path B.

## Issue #1124 incident record

| Item | Value |
|------|-------|
| Last governed VOC-142 infrastructure base | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| Unauthorized infra head | `d8720829b176cf1287e633f9382989fc8f258105` |
| Infra self-CI run claimed green | `33443684483` |
| Caller pin at issue creation | `8993e867640dfb604dec0466c4e0787e68d8e258` |
| `reviewer` drift | `effort=high,fast=false` → `effort=xhigh,fast=false` |
| `reviewer_fast_retry` drift | `effort=high,fast=false` → `effort=xhigh,fast=true` |
| `plan_reviewer` drift | `effort=high,fast=false` → `effort=xhigh,fast=false` |
| VOC-117 tests | rewritten on infra `main` to bless the new values |
| Unreconciled | `README.md`, `CHANGELOG.md`, caller fixtures/pin, VOC-142 DEP-06/AC-09 current-state evidence |
| Explicitly excluded | issue #1120 VOC-112 provenance EHR escalation |

A green test changed in the same direct sequence does not establish
governed authority.

## Binding contract

| Role | Path A (authorized current) |
|------|-----------------------------|
| `implementer` | `cursor/composer-2.5` |
| `implementer_escalation` | `cursor/composer-2.5` |
| `planner` | `cursor/grok-4.6[effort=high,fast=false]` |
| `reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |
| `reviewer_fast_retry` | `cursor/grok-4.6[effort=high,fast=false]` |
| `plan_reviewer` | `cursor/grok-4.6[effort=high,fast=false]` |

Historical VOC-117 expected bindings (restored as current-state on Path A):
the Path A column above.

## Exhaustive source-search disposition

| Path / pattern | Disposition |
|----------------|-------------|
| `karsift-ai-infra/config/roles.yml` | **update** — restore Path A in nested checkout for infra carrier PR #175 |
| `karsift-ai-infra/tests/test_voc117_role_bindings.py` | **update** — restore exact `VOC117_BINDINGS` and `prepare_cursor_model` assertions |
| `karsift-ai-infra/README.md` | **update** — document Path A current lineup; note `d8720829…` drift is not current |
| `karsift-ai-infra/CHANGELOG.md` | **update** — record VOC-145 governed reconciliation |
| `tooling/governance/fixtures/karsift-ai-infra/config/roles.yml` | **mirror** — byte-identical to carrier head (unchanged bytes from prior pin) |
| `tooling/governance/fixtures/karsift-ai-infra/tests/test_voc117_role_bindings.py` | **mirror** — byte-identical to carrier head |
| `tooling/governance/fixtures/karsift-ai-infra/CHANGELOG.md` | **update** — mirror carrier VOC-145 entry |
| `tooling/governance/fixtures/karsift-ai-infra/README.md` | **update** — fixture provenance + pin advance to `ad2b277…`; notes PR #175 merge prerequisite |
| `tooling/governance/fixtures/karsift-ai-infra/PINNED_SHA.txt` | **update** — `ad2b27784e6fc33b3ac7e9dab48245dd6d08ac7f` |
| `tooling/governance/tests/test_voc145_caller_replacement.py` | **add** — binding, pin, mirror-hash, fail-closed, and safety regressions |
| `tooling/governance/tests/test_voc117_role_bindings.py` | **unchanged** — already asserted Path A fixture bindings |
| `tooling/governance/tests/test_voc136_caller_replacement.py` | **update** — live `CURRENT_PIN` literal and mirrored `CHANGELOG.md` hash |
| `tooling/governance/tests/test_voc137_pr_sha_scan.py` | **update** — live `CURRENT_PIN` literal and mirrored `CHANGELOG.md` hash |
| `tooling/governance/tests/test_voc121`–`126`, `129`, `138`–`142` tests | **update** — live pin literals (17-path lock) |
| `scripts/foundation/voc097` / `voc104` / `voc108` tests | **update** — live pin literals |
| `effort=xhigh`, `fast=true` in current tracked source | **absent** after restore (historical only in evidence/spec) |
| `specs/changes/VOC-117-…/`, `VOC-142-…/`, `VOC-141-…/`, `VOC-140-…/` | **historical** — not modified |
| VOC-112 fixtures, `#1120` resume helpers | **excluded** — not touched |

## Validation commands (after the repair is tracked and committed)

Record exact commands and results. Required:

```bash
bash scripts/governance/validate-governance.sh --base 10df745c42ed283405d6bdf5b01180afdfed7d26 --head <implementation-head>
bash scripts/governance/classify-change-risk.sh --base 10df745c42ed283405d6bdf5b01180afdfed7d26 --head <implementation-head>
python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
git diff --check
```

Working-tree verification before workflow commit (all passed 2026-09-01):

```text
PYTHONPATH=tooling/governance/tests python3 -m unittest tooling.governance.tests.test_voc145_caller_replacement
# Ran 11 tests in 0.022s — OK

PYTHONPATH=tooling/governance/tests python3 -m unittest tooling.governance.tests.test_voc117_role_bindings
# Ran 8 tests — OK

PYTHONPATH=tooling/governance/tests python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
# Ran 302 tests in 26.066s — OK

PYTHONPATH=tooling/governance/tests python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
# OK

git diff --check
# (no output — clean)
```

Post-commit governance validation with exact `<implementation-head>` is
recorded by the calling workflow after it stages and commits this repair.

Do not treat a missing suite or an untracked-only pass as acceptance.

## Exact-head binding contract

The App-authored independent-review comment/check on the implementation PR
must bind the live head exactly. Merge-gate must reject any mismatch. This
file must not be required to contain that live head SHA in the same commit.

Independent review must explicitly evaluate: authorized path; six exact
current bindings; historical-versus-current VOC-117 split;
README/CHANGELOG/fixture reconciliation; pin not equal to `d8720829…`;
unchanged retry/exact-SHA/fail-closed controls; exclusion of issue #1120;
and that coordinated infrastructure PR #175 is independently reviewed and
merged to `main` at `ad2b277…` before caller merge makes the pin live on
`@main`.

## Release handoff

After the exact reviewed caller merge, ordinary later promotion uses
existing release evaluation. Do not snapshot the develop/main gap. Do not
resume issue #1120. Do not create a duplicate promotion PR or release
audit. Root issue #1124 closes only after allowlisted metadata from a
successful implement/release path exists. Unauthorized head `d8720829…`,
governed base `8993e867…`, and run `33443684483` remain incident evidence
only.
