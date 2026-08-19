# VOC-088-T02 — Failure-to-issue evidence

## Implemented observer

- `.github/workflows/operational-failure-monitoring.yml` is a standalone,
  serialized `workflow_run` observer for `scheduled-synthetics`,
  `deploy-staging`, and `deploy-production`.
- It handles only `failure`, `cancelled`, and `timed_out` terminal conclusions.
- It rejects runs whose head repository is not this repository and does not
  observe itself, so neither observer failures nor fixtures recurse.
- It fails closed when the automation App credentials are unavailable and mints
  an issues-write installation token. There is no `GITHUB_TOKEN` fallback.

## Sanitization and deduplication

`infra/scripts/open-failure-issue.sh` accepts only the three exact workflow names,
the three expected conclusions, this repository's `owner/name` form, and a
canonical numeric GitHub Actions run URL. It reads no job logs or step output.
The created issue has no labels and contains a fixed summary plus only the
allowlisted workflow, conclusion, and run URL.

The stable marker is workflow name plus failure category. The workflow's constant
concurrency group serializes the open-issue lookup and create operation. If an
open issue already owns the marker, the helper exits successfully without a
second POST.

## Deterministic verification

`scripts/foundation/voc088-failure-to-issue.test.mjs` uses a local mock GitHub CLI
and proves:

- first failure creates exactly one issue;
- a repeat open fingerprint creates no duplicate;
- no label, credential, log, or unbounded metadata reaches the issue;
- success is rejected;
- cancellation and timeout are in the observer's terminal-failure set;
- only a GitHub App installation token is wired to issue creation;
- all three intended operational workflows are observed externally.

The existing `.github/workflows/error-monitoring.yml` remains unchanged and keeps
its Sentry-only responsibility.

Validation commands for the task PR:

```text
node --test scripts/foundation/voc088-failure-to-issue.test.mjs
pnpm validate
git diff --check
```

Live App-created issue, downstream planning, and repeated-event proof are deferred
to VOC-088-T03 after this observer is present on the default branch. The bounded
proof must not include logs, credentials, personal data, or a recursive observer
failure.
