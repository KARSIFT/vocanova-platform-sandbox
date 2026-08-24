# VOC-114-T00 — Evidence

Task: `VOC-114-T00` — Restore recovery metadata-read token contract, localize
errors, add tests.

Evidence date: 2026-08-24

Authority issue: `KARSIFT/vocanova-platform-sandbox#958`

## Issue #956 observations (metadata-only)

| Item | Value |
|------|-------|
| Merged integration SHA | `b97e9575fd30671c336a2e92ca00db6e29b86416` (PR #954) |
| Integration recovery failure | Pipeline run `32696249484`, job `97339047384` — `integration_push` mode, immediate `github_metadata_read_failed` before bounded wait or dispatch |
| Release recovery failure | Pipeline run `32696549963`, job `97339655608` — `reconcile-release` for release issue #946, same immediate metadata failure in `promotion_pr` mode |
| Blocked downstream | VOC-113-T01; promotion PR #947 lacks genuine exact-head required checks |
| Prior mutation posture | Merge, task-completion publication, and issue closure succeeded before the recovery metadata-read phase failed |

## Verified cause boundary (`VOC-114-D00`, `VOC-114-D01`, `VOC-114-D05`)

The two live runs prove that the App installation token used by recovery lacked
effective access at the first metadata phase; they do not identify the rejected
endpoint because pre-T00 code collapsed every `gh` failure to
`github_metadata_read_failed`. Code review also shows that all three mint paths
feeding `actions-check-recovery-runner.py` omitted explicit `permission-*`
inputs. Omission alone is **not** a verified cause: the token action documents
that an omitted permission list inherits the installation's permissions. The
available metadata therefore supports an effective App-permission failure, but
does not distinguish a missing App/installation grant from an endpoint-specific
denial. T01 owns that live installation confirmation.

The complete REST permission contract is:

- `/commits/{sha}/check-runs`: **Checks read**
- `/commits/{sha}/status`: **Commit statuses read** (distinct from Checks)
- `/actions/runs?head_sha=…`: **Actions read**, satisfied by the existing
  **Actions write** grant also required for allowlisted workflow dispatch
- `/commits/{sha}`: **Contents read**
- `/pulls/{number}`: **Pull requests read**

T00 makes that contract explicit at every mint site. This is fail-closed in two
places: `actions/create-github-app-token` refuses a requested permission that the
installation has not been granted, and the runner refuses any endpoint failure
before dispatch planning. Workflow-job `GITHUB_TOKEN` permissions are irrelevant
to these calls because the runner authenticates with the minted App token.

**Installation contract for the T01 operator:** confirm the KARSIFT App and its
caller-repository installation grant Checks read, Commit statuses read, Actions
read/write, Contents read/write, and Pull requests read/write at the levels
requested by the applicable carrier. Existing Issues write remains required by
the merge/release mutation path. Do not weaken the requested contract if minting
fails; update the App/installation grant and retry through T01.

Primary references: GitHub's REST documentation assigns Checks read to the
[check-runs endpoint](https://docs.github.com/en/rest/checks/runs), Commit
statuses read to the [combined-status endpoint](https://docs.github.com/en/rest/commits/statuses?apiVersion=2022-11-28),
and the token action documents inherited permissions plus fail-closed explicit
requests in its [permission inputs](https://github.com/actions/create-github-app-token/blob/main/README.md?plain=1).

## Remediation applied (T00)

The primary implementation merged in `KARSIFT/karsift-ai-infra#136` from
reviewed head `72b3742f41bed1e7306b9dccc20a700a2bc467ec` as immutable merge
`30cc0a6f443b95e45527b03094767b8357b0a2dc`. Live T01 execution then exposed
adjacent causal defects in the same recovery mechanism. Those corrections
merged as source PRs #137 through #142, culminating in immutable merge
`bdc6736568827103b48255521f4bc83d5103bd3b`. The caller fixture is synchronized
to that final source merge and `PINNED_SHA.txt` advances to the same SHA.

| Target | Change |
|--------|--------|
| `config/actions-check-recovery-runner.py` | Localized metadata-read failures to `check_runs_read_failed`, `workflow_runs_read_failed`, and `commit_metadata_read_failed`; extracted `run_metadata_phase()` so read failures abort before dispatch planning |
| `.github/workflows/merge-gate.yml` | App mint explicitly preserves Contents/Issues/Pull requests/Actions write and adds Checks read plus Commit statuses read |
| `.github/workflows/release.yml` | Same complete read/write contract on the converge recovery token |
| `.github/workflows/recover-actions-checks.yml` | App mint requests Actions write exactly once, Checks read, Commit statuses read, Contents read, and Pull requests read; hosted verification uses valid repository context |
| Source PRs #137–#142 | Removed invalid `gh api --repo` use in recovery/verification, bound promotion suppression, completion, and final verification to required contexts, replaced incompatible paginated `gh api --slurp --jq` usage with standalone `jq`, and added a bounded caller `recover-integration-push` dispatch whose target is the internally resolved current `develop` head rather than a free-form SHA |
| `.github/workflows/repository-governance.yml` | Keeps actual pull requests in `pr-validation` (and fixture-changing originals in strict `pr-ancestry`), while authenticated promotion recovery of the immutable `develop` tip uses the existing `squash-safe-push` contract because the promotion aggregates already-squashed task commits |
| `tests/test_voc114_actions_check_recovery.py` and caller tests | Deterministic coverage includes both positive modes, complete and omitted mint contracts, duplicate input rejection, endpoint classes, no planning/dispatch after read failure, required-context filtering, hosted verifier CLI contracts, exact target resolution, and fixture pin assertions |
| `README.md` (shared + fixture) | Documents recovery metadata read contract and sanitized endpoint classes |
| `docs/operations/11-devops-and-ci-cd.md` | Documents caller-facing recovery App read contract, endpoint classes, and bounded operator integration recovery |

Mutation posture is unchanged except for the declared read scopes required by
the metadata phase. Recovery dispatch still uses `permission-actions: write`
only on allowlisted workflows; no broadened mutation grants were introduced.

## Validation results

Commands run at implementation time (2026-08-24):

```text
# karsift-ai-infra PR #136, head 72b3742f41bed1e7306b9dccc20a700a2bc467ec
PYTHONPATH=config python3 -m unittest discover -s tests -p 'test_*.py' -v
  → Ran 254 tests — OK
git diff --check
  → no conflicts
self-ci: actionlint, shellcheck, yaml-parse, policy-tests
  → all passed before merge 30cc0a6f443b95e45527b03094767b8357b0a2dc

# karsift-ai-infra final corrective merge bdc6736568827103b48255521f4bc83d5103bd3b
PYTHONPATH=config python3 -m unittest discover -s tests -p 'test_*.py' -v
  → Ran 260 tests — OK
actionlint; shellcheck; YAML parse; policy tests; git diff --check
  → all passed across source PRs #137–#142

# caller fixture mirror (authoritative for this PR)
PYTHONPATH=config python3 -m unittest discover -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*voc114*'
  → Ran 12 tests — OK
node --test scripts/foundation/voc114-actions-check-recovery.test.mjs
  → 3 tests — OK
node --test scripts/foundation/voc113-actions-check-recovery.test.mjs
  → existing VOC-113 caller tests — OK
node --test scripts/foundation/voc097-fixture-matrix.test.mjs \
  scripts/foundation/voc104-ready-for-review-reuse.test.mjs \
  scripts/foundation/voc108-authoritative-lifecycle.test.mjs \
  scripts/foundation/voc113-actions-check-recovery.test.mjs \
  scripts/foundation/voc114-actions-check-recovery.test.mjs
  → 25 tests — OK
PYTHONPATH=tooling/governance/fixtures/karsift-ai-infra/config \
  python3 -m unittest discover \
  -s tooling/governance/fixtures/karsift-ai-infra/tests -p 'test_*.py'
  → Ran 176 tests — OK
PYTHONPATH=tooling/governance/tests python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'
  → 160 tests — OK
bash scripts/governance/validate-governance.sh
  → passed
bash scripts/governance/classify-change-risk.sh
  → path floor R4 (package + infra fixture paths)
git diff --check
  → no conflicts
```

## Shared-infra adoption state

The source dependency is satisfied by merged PRs
`https://github.com/KARSIFT/karsift-ai-infra/pull/136` through
`https://github.com/KARSIFT/karsift-ai-infra/pull/142`. The authoritative final
source revision and caller fixture pin are both
`bdc6736568827103b48255521f4bc83d5103bd3b`. T01 still waits for this caller
task PR to merge so its evidence-only carrier can execute the trusted pinned
template revision.

## T01 dependency

Live integration_push and `reconcile-release` recovery proof for promotion PR
#947 remains operator-owned under `VOC-114-T01` after the shared infra changes
land on the branch the caller pipeline executes from.
