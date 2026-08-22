# VOC-107-T00 — Evidence

## evidence_id

VOC-107-EV-00

## gate_status

pending

## drafting-time diagnosis

Confirmed `karsift-ai-infra/.github/workflows/implement.yml` creates
`base_sha..HEAD` bundles and that attempt 2+ sets `base_sha` to `HEAD` after
checkout/rebase of the existing agent branch. The isolated publish job fetches
only `integration_branch` before `git bundle verify`. Issue #891 records run
`32539352323` where a valid remediation commit was rejected before publication
because that thin-bundle prerequisite was absent from the clean bare repository.

## commands

_To be filled by the implementer after adoption._

## results

_To be filled by the implementer after adoption._

## privacy

Evidence must remain allowlisted metadata only (commands, pass/fail, SHAs, run
IDs). No logs, artifacts, secrets, or user identifiers.
