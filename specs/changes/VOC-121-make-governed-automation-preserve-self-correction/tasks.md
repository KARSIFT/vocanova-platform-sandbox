# VOC-121 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
This package intentionally defaults to **one task** because issue #994 is one
autonomous-delivery reliability outcome. Coordinated caller and infrastructure
pull requests remain one task; repository count, workflow count, and the three
related findings are not split reasons.

Cross-repo note: T00 changes `KARSIFT/karsift-ai-infra` for implementer
staging/self-correction/publication and required-check recovery. The
implementer opens the infra PR for that behavior; this package is the
authorizing change package for the required outcome. Do not treat the
untracked local `karsift-ai-infra/` checkout (if present) as this repo's
tracked tree. Caller fixture/pin, tests, and evidence land in this repository
under the same task. Infra PRs must say
`Relates to KARSIFT/vocanova-platform-sandbox#<task>` and MUST NOT use a
closing keyword.

## VOC-121-T00 — Preserve self-correction helpers, finish coordinated carriers without silent loss, and recover required checks from GitHub satisfaction state

- Requirement source: issue #994; `VOC-121-D00` through `VOC-121-D10`
- Acceptance criteria: `VOC-121-AC-00` through `VOC-121-AC-07`
- Tests: `VOC-121-TEST-00` through `VOC-121-TEST-08`
- Evidence: `VOC-121-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. Record the issue #994 live failures in `t00-evidence.md` (VOC-120-T00 run
   `32899479806` / job `97970681892`, self-correction missing
   `prepare_cursor_model.py`, promotion PR #993, release run `32902418610`).
2. In `KARSIFT/karsift-ai-infra/.github/workflows/implement.yml`, copy every
   helper self-correction later invokes to an immutable location before
   `rm -rf karsift-ai-infra`. At minimum preserve `config/run-app-checks.sh`
   and `config/prepare_cursor_model.py`. Point self-correction at those copies.
   Continue to delete the nested checkout before caller `git add -A`.
3. Stop silent discard of authorized nested source edits. Preferred: isolated
   source commit, bundle, clean-runner publisher, source PR, independent
   review, merge, exact SHA capture, and caller pin when consumed. Fallback if
   general multi-carrier cannot be made safe without weakening isolation,
   least privilege, exact-SHA review, or protected checks: fail closed with
   precise recovery instructions. Record the chosen path. Never publish a
   gitlink. Never return an App token to the implementer runner.
4. Keep caller PRs on local `Closes #N`. Cross-repository PR bodies use
   `Relates to KARSIFT/vocanova-platform-sandbox#<task>` with no closing
   keyword.
5. Change required-check recovery so a cancelled or failed required check-run
   on the unchanged exact head is rerun or redispatched even when another run
   or same-named status succeeded. Do not treat status attestation as
   overriding a check-run GitHub still selects.
6. Add deterministic tests that reproduce:
   - caller-only staging after nested edits (silent discard versus isolated
     publish or fail-loud);
   - cross-repository publisher races and stale heads;
   - deletion-before-self-correction helper loss;
   - cancelled exact-head required check-run versus successful alternate run
     or same-named status.
7. Update current-state comments/docs (`implement.yml` header comments,
   recovery/release comments, `karsift-ai-infra/README.md`) so they no longer
   describe authorized nested edits as safely disposable or claim that
   same-SHA status attestation satisfies a cancelled required check-run if
   GitHub does not.
8. Land the infra change through one reviewed infra PR. Pin
   `tooling/governance/fixtures/karsift-ai-infra/` to that exact merge SHA when
   the fixture consumes the change. Update caller governance tests in the same
   task.
9. Run applicable validation and record results in `t00-evidence.md`:
   - `python3 -m unittest discover -s tests -p 'test_*.py'` in the primary
     `KARSIFT/karsift-ai-infra` checkout;
   - `bash scripts/governance/validate-governance.sh`;
   - `bash scripts/governance/classify-change-risk.sh`;
   - `python3 -m unittest discover -s tooling/governance/tests -p 'test_*.py'`;
   - `git diff --check`;
   - exact reviewed infra SHA and pin applicability;
   - any narrower targeted commands added by the implementation.
10. Preserve independent exact-SHA review for each carrier, risk
    classification, protected checks, App-token isolation, and two-attempt
    limits. Do not introduce OpenAI/Codex or `OPENAI_API_KEY`.

### Explicitly out of scope for this task

- Application runtime, deployment topology, credential-value, provider, or
  monitor-inventory changes.
- Introducing an OpenAI/Codex route or weakening Cursor fail-closed auth.
- Weakening exact-SHA review, risk floors, protected checks, or retry caps.
- Splitting self-correction, multi-carrier publication, and check recovery
  into separate tasks.
- Treating this task's possible bootstrap recovery of its own infra PR as a
  new package (`VOC-121-D10`).

## Task ordering notes

- This package intentionally has one task because no concrete split boundary
  is required: the three live failures are one reliability outcome, and
  coordinated source and caller PRs remain one task.
- Infra should merge first when the caller fixture/pin consumes that change;
  otherwise the two reviewed PRs may complete under the same task without a
  pin bump.
- No task may be dispatched before this package is adopted and
  implementation-authorized.

Tasks preserve scope, separation of duties, and rollback safety.
