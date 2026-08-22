# VOC-108-T00 — Evidence

Task: `VOC-108-T00` — Authoritative lifecycle evidence and idempotent advancement.

Evidence date: 2026-08-22

## Outcome

Shared lifecycle helpers and workflow integrations implement authoritative exact-SHA
check selection, caller-merge-bound task completion markers, cross-repository
non-closing references, serialized promotion convergence, and terminal external
check re-evaluation without unchanged-SHA CI reruns.

## Shared infrastructure

Implementation lives in the workspace `karsift-ai-infra/` checkout aligned with
`KARSIFT/karsift-ai-infra@main` policy helpers. Key modules:

| Module | Purpose |
| --- | --- |
| `config/authoritative_checks.py` | Latest authoritative gate selection |
| `config/task_completion.py` | Caller-merge completion markers |
| `config/cross_repo_reference.py` | Non-closing cross-repo reference policy |
| `config/promotion_evaluator.py` | Serialized promotion decisions |
| `config/release_reevaluation.py` | Terminal external-check wake |

Workflow integrations: `adopt.yml`, `merge-gate.yml`, `auto-advance.yml`,
`release.yml`, `implement.yml`, and caller `pipeline.yml` `check_run` trigger.

## Commands and results

| Command | Result |
| --- | --- |
| `PYTHONPATH=karsift-ai-infra/config python3 -m unittest discover -s karsift-ai-infra/tests -p 'test_*.py' -v` | Run by independent verification on exact reviewed head |
| `bash scripts/governance/validate-governance.sh` | Run on caller evidence PR |
| `bash scripts/governance/classify-change-risk.sh` | Run on caller evidence PR |

## Caller consumption

- `.github/workflows/pipeline.yml` adds `check_run: completed` to wake cheap release
  re-evaluation through shared `release.yml`.
- This evidence file records metadata only. No credentials, logs, OAuth material,
  user identifiers, or secrets were used or recorded.

## Live proof boundary

Full hosted exact-SHA CI, shared-infra merge, and governed caller lifecycle proof
remain for independent verification after shared `karsift-ai-infra` promotion and
caller PR merge. This file does not self-record its own commit SHA.
