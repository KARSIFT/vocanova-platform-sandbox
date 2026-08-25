# VOC-120 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #987 is one
cache-hygiene outcome. Coordinated caller and infrastructure pull requests remain
one task; repository count is not a split reason.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for the portable ignore
rules and infra tests. The implementer opens the infra PR for that behavior;
this package is the authorizing change package for the required outcome. Do not
treat the untracked local `karsift-ai-infra/` checkout (if present) as this
repo's tracked tree. Caller untrack, ignore verification, validation, and
evidence land in this repository under the same task. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a closing
keyword.

## VOC-120-T00 — Untrack caller bytecode, add portable infra ignores, and add hygiene validation

- Requirement source: issue #987; `VOC-120-D00` through `VOC-120-D06`
- Acceptance criteria: `VOC-120-AC-00` through `VOC-120-AC-05`
- Tests: `VOC-120-TEST-00` through `VOC-120-TEST-07`
- Evidence: `VOC-120-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #987 discovery in `t00-evidence.md` (caller refs, tracked
   blob path, introducing commit, infra SHA, missing infra `.gitignore`).
2. In `KARSIFT/karsift-ai-infra`, add a repository-root `.gitignore` that covers
   `__pycache__/`, `*.pyc`, and related bytecode variants (`*.py[cod]` is the
   preferred portable form). Add deterministic tests that fail closed when those
   patterns are missing.
3. Untrack
   `infra/scripts/__pycache__/cloudflare_origin_port_remap.cpython-312.pyc` in
   the caller without changing `infra/scripts/cloudflare_origin_port_remap.py`.
   If `git ls-files` finds additional tracked `__pycache__/`, `*.pyc`, `*.pyo`,
   or `*.pyd` paths, untrack those in this same task.
4. Keep and verify the caller `.gitignore` entries `__pycache__/` and
   `*.py[cod]`.
5. Add caller deterministic validation that:
   - fails if any tracked Python bytecode/cache artifact exists;
   - fails if required caller ignore patterns are missing;
   - inspects tracked paths (`git ls-files`), not untracked working-tree cache;
   - is invoked from `bash scripts/governance/validate-governance.sh`;
   - has unit coverage under `tooling/governance/tests/`, including a synthetic
     tracked `.pyc` / `__pycache__` negative case.
6. Land the infra change through one reviewed infra PR. Pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA only
   if the fixture actually consumes the change. Do not copy `.gitignore` into the
   policy fixture merely to force a pin.
7. Run applicable validation and record results in `t00-evidence.md`:
   - `git ls-files` showing no tracked bytecode/cache artifacts on the caller;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - `git diff --check`;
   - exact reviewed infra SHA and pin applicability.
8. Preserve independent exact-SHA review, risk classification, protected checks,
   and retry limits.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential, provider, or
  model-routing changes.
- Git history rewrite to purge the historical blob.
- Expanding infra `.gitignore` into unrelated Python packaging ignores.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting caller untrack, infra ignore rules, tests, or evidence into
  separate tasks.

## Task ordering notes

- This package intentionally has one task because no concrete split boundary is
  required: coordinated source and caller PRs remain one outcome.
- Infra should merge first when the caller fixture/pin consumes that change;
  otherwise the two reviewed PRs may complete under the same task without a pin
  bump.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
