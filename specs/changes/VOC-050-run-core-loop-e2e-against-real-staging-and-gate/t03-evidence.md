# VOC-050-EV-03 — T03 fail-closed signal evidence

Evidence for `VOC-050-T03` (`VOC-050-AC-04`, test `VOC-050-TEST-04`).

`VOC-050-T03` has two deliverables:

1. Prove with a live-observed run (not YAML inspection alone) that a failure in
   the staging core-loop gate produces a **failed** `deploy-staging.yml` job
   conclusion.
2. Record the still-open cross-repo dependency: whether
   `karsift-ai-infra`'s `release.yml` actually consumes this failure signal for
   `develop` -> `main` auto-promotion.

This file is the durable in-repo record because the implementer in this
workflow does not author the generated PR body directly.

## 1. Live-observed failing run on `deploy-staging.yml`

Observed run:

- Workflow: `deploy-staging`
- Run number: `#144`
- Run id: `31285044504`
- Commit: `a029e5fc50945b50009b692b5280bdb17d4ef3f2`
- URL: <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31285044504>
- Job URL: <https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/31285044504/job/93172237205>

Step-level conclusion evidence from the GitHub Actions REST API:

```
$ curl -fsS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31285044504/jobs" \
  | jq -r '.jobs[] | [.name, .status, (.conclusion // "")] | @tsv'
deploy to staging	completed	failure

$ curl -fsS "https://api.github.com/repos/KARSIFT/vocanova-platform-sandbox/actions/runs/31285044504/jobs" \
  | jq -r '.jobs[] | .steps[] | [.number, .name, .status, (.conclusion // "")] | @tsv'
...
16	Poll api-staging.vocanova.site/healthz	completed	success
17	Poll staging.vocanova.site/	completed	success
18	Mint synthetic smoke-test session for the staging core-loop	completed	failure
19	Set up Node for the staging core-loop check	completed	skipped
20	Install pnpm for the staging core-loop check	completed	skipped
21	Install workspace dependencies for the staging core-loop check	completed	skipped
22	Install Playwright Chromium for the staging core-loop check	completed	skipped
23	Run the staging core-loop journey	completed	skipped
...
```

Interpretation:

- The workflow reached the post-deploy core-loop gate region (`steps 18-23`)
  after successful deploy and readiness polling (`steps 15-17`).
- Failure in that region produced an overall job conclusion of `failure`.
- There is no `continue-on-error` masking path for this failure in the observed
  execution.

This satisfies `VOC-050-AC-04`'s repository-local requirement that a failed
staging core-loop gate records as a blocking failure signal on this repository's
side.

## 2. Cross-repo dependency explicitly recorded (still open here)

`VOC-050-T03` cannot modify or prove behavior inside
`karsift-ai-infra`'s `release.yml`. This task proves only that
`vocanova-platform-sandbox` now emits an accurate failure signal in
`deploy-staging.yml` when the staging core-loop gate fails.

Still-required human follow-up (outside this repository's scope):

- Confirm (or change, via a companion package in `karsift-ai-infra`) that the
  release workflow's auto-promotion decision truly gates on this workflow
  conclusion for the relevant `develop` commit.

This is exactly `specification.md` open question 1 and `change.yaml`
`VOC-050-DEP-01`; it remains open after this task.

## 3. Limits and reproducibility notes

- The observed failure happened at the session-mint step in the core-loop gate
  region, so the Playwright journey step itself was skipped in that run. For
  `AC-04`, that is acceptable: the criterion is fail-closed workflow conclusion
  when the gate fails, not the specific failing sub-step.
- Job-log download via unauthenticated API returned `403` in this environment,
  so this evidence uses public run metadata plus step conclusions from the jobs
  endpoint instead of raw per-line logs.
