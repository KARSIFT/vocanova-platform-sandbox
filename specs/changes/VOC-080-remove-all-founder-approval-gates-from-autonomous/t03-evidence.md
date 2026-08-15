# VOC-080-EV-03 — T03 release, remediation, and deploy-path gates

Evidence for `VOC-080-AC-02`, `VOC-080-AC-06`, and `VOC-080-AC-07`.
The integrated live promotion/deploy rehearsal remains `VOC-080-T06`.

## Shared workflow delivery

`KARSIFT/karsift-ai-infra` PR
[#39](https://github.com/KARSIFT/karsift-ai-infra/pull/39), merge commit
`489dd82b5403a36082e70c95185463f445d02c13`, landed the release behavior
used by callers pinned to `@main`:

- completed rosters automatically open or reuse a `develop -> main`
  promotion PR;
- no `issue_comment`, founder identity, or literal `approved` signal is a
  release authority path;
- `action=reconcile-release` plus `release_issue_number` provides an
  idempotent retry after an interrupted promotion;
- retry validates the `karsift:release` audit issue, package roster, and
  closure of every roster task before creating a promotion PR;
- both normal and retry paths wait for all promotion checks and merge only
  the exact checked head SHA;
- `remediate.yml` retains its existing fail-closed independent-review and
  bounded-retry behavior; it has no founder override.

Infra PR #39 passed actionlint, shellcheck, YAML parsing, and eight policy
tests, including `tests/test_release_policy.py` assertions that the founder
comment job is absent and retry is dispatch-driven.

## Repository-controlled production environment

Live GitHub API inspection on 2026-08-15 returned one `production`
environment with only a deployment branch-policy protection rule and
`reviewers: null`. Therefore no repository-controlled founder or other human
environment reviewer remains on the production deploy path. The custom
deployment-branch policy remains a non-human safety control.

## Boundaries

- Caller `.github/workflows/pipeline.yml` wiring and canonical documentation
  are completed in T04 after the reusable inputs stabilize.
- T06 must exercise the live completed-roster promotion, retry behavior, and
  push-to-main deployment path; this evidence does not claim that rehearsal.
- Failed CI, review, promotion, or deployment checks remain blocking and
  cannot be overridden by a founder comment.
