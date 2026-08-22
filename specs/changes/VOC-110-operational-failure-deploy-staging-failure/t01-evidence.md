# VOC-110-T01 evidence — live deploy-staging verification

Pending until `VOC-110-T00` merges to `develop` and operator-owned live evidence
reconciliation completes per `.karsift/live-evidence/VOC-110-T01.yaml`.

## gate_status

pending

## Required proof

1. A push-triggered `deploy-staging` run on `develop` whose HEAD SHA contains the
   T00 fix reaches conclusion `success`.
2. Job `deploy to staging` succeeds.
3. No new open issue with marker `<!-- operational-failure:deploy-staging:failure -->`
   beyond issue #911.

## Allowlisted fields to record

- Run URL and run ID
- Conclusion and total duration
- Head SHA and branch
- Reconcile timestamp

Do not record logs, secrets, session values, or personal data.
