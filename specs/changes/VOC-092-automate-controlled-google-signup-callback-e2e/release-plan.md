# VOC-092 — Release Plan

## Release and deployment authorization

This package does not authorize production deployment by itself. Task merges to
`develop` follow the repository-controlled promotion path when the package roster
closes. Under **active A-004**, no founder `approved` comment is required on
engineering-workflow gates.

## Preconditions, monitoring, and outcome

| Item | Requirement |
|------|-------------|
| Exact revision | Each task PR head SHA independently verified |
| Repository validation | `pnpm validate` and governance checks per touched paths |
| Harness CI | Green run showing allowlisted-success and unlisted-503 cases |
| Staging synthetic | `synthetic.staging.oauth-expected-state` remains green post-deploy |
| Monitoring inventory | Unchanged (`monitoring_impact.state: none`) |
| Secrets / PII | None in logs, artifacts, issues, PRs, or evidence |
| Outcome owner | unassigned (set at adoption) |

## Rollback

- **Trigger:** Harness introduces CI instability, leaks OAuth material, or
  accidentally touches production auth wiring.
- **Mechanism:** Revert VOC-092 task commits on `develop`; normal promotion
  carries revert to `main` when release gates pass.
- **Validation:** CI green without harness job; staging synthetic still passes;
  no test-auth routes in production wiring.
- **Accountable owner:** unassigned (set at adoption).
- **Last-known-good:** `develop` commit immediately before T00 merge.

## Independent verification, human approvals, and closure

- **Authority model:** A-004 active. Strengthened R3 evidence obligations apply;
  no founder merge gate.
- **Verifier obligations:** Confirm harness uses disposable DB and localhost fake
  provider only; confirm VOC-088 staging synthetic unchanged; confirm VOC-088
  evidence remediation; bind report to exact reviewed SHA.
- **Human audit (ongoing):** Periodic interactive Google sign-in per updated
  operations doc — not replaced by this package.
- **EHR:** not triggered.
- **Closure evidence:** `t03-evidence.md` with CI run URL, staging synthetic run
  URL, and validation commands inspected.

Do not conflate repository merge, release, activation, or closure. Historically
under A-003, R4 merge required founder approval. **Under active A-004,**
engineering-workflow gates require no founder `approved` comment.
