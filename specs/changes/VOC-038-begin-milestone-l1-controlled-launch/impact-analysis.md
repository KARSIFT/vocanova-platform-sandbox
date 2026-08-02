# VOC-038 — Impact Analysis

## Production impact

None from this package itself (docs-only proposal). Once adopted and tasks dispatched, the
highest-impact tasks are:

- `T01` (cohort/allowlist implementation) — touches `apps/api` auth/signup path; must not
  regress the existing kill-switch behavior verified in `VOC-037-EV-03`.
- `T04` (enable AI for the cohort) — first real (non-test-row) production AI usage; real cost
  and real user-facing safety exposure, exactly the class of risk DOC-12's L1 trigger list
  exists for.

## Security/privacy impact

The allowlist mechanism (`T00`/`T01`) is itself a security control — it must fail closed (an
account not on the list must be rejected, not admitted-by-default on any error path), consistent
with the existing kill-switch design philosophy documented in `VOC-037-EV-03`.

## Rollback impact

`T06` extends the existing rollback rehearsal technique to cover real cohort/AI state; no new
rollback mechanism is introduced.

## Dependencies on other in-flight work

Blocked on `VOC-037-AC-05` (R2 go/no-go) per `change.yaml`. No other in-flight package exists
as of this drafting pass (issue #265 is the only open issue; no open PRs).
