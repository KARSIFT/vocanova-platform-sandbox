# VOC-079 — Impact Analysis

## Security and privacy

- **TLS / secrets:** Production and staging nginx conf and certificate trees
  remain tier-isolated. Shared edge continues read-only mounts from each
  tier's own paths. No secret values may appear in evidence files or commits.
- **Cloudflare token:** Uses existing
  `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` (zone-scoped Origin Rules).
  Do not broaden Workers-AI `PRODUCTION_CLOUDFLARE_API_TOKEN`. Redact logs.
- **Isolation:** Must not regress VOC-037-D01 / VOC-067-AC-03 write
  isolation or copy secrets across `/opt/vocanova/infra` and
  `/opt/vocanova/production`.
- **Outage class:** Premature bridge removal while a remap to `:8443` still
  existed would recreate issue #485. Mitigated by mandatory repository
  verify-only (`VOC-079-AC-00` / `VOC-079-DEP-02`) before T02.

No new personal-data processing, auth model change, or cookie-scope change.

## Data and migrations

None. No database schema, seed, or data backfill.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** None — no UI change.

## Risks, dependencies, and evidence

- `VOC-079-R00`: **Issue #485 recurrence** if bridge is retired while Cloudflare
  still remaps to origin `:8443`. Mitigation: T00 gate; bridge-retention test
  remains binding until `absent`; independent review FAIL without transcript.
- `VOC-079-R01`: **Shared-edge single fault domain** — bad reload or mis-mounted
  production conf can affect both tiers. Mitigation: keep fail-closed
  `nginx -t`; routine deploys do not recreate shared-edge; monitoring window
  in T03.
- `VOC-079-R02`: **OAuth / CORS / middleware breakage** after stripping
  `:8443` from `API_BASE_URL` and redirect allowlists. Mitigation: T01 before
  T02; smoke and readiness on canonical `:443`; rollback via prior revision.
- `VOC-079-R03`: **Orphan-removal blast radius** if `--remove-orphans` (or
  equivalent) is applied to the wrong compose project and removes shared-edge
  or staging services. Mitigation: `VOC-079-DEP-01` requires production
  project scope only; TEST-04 inspects the workflow.
- `VOC-079-R04`: **Rollback path change** — after bridge retirement, Cloudflare
  `--restore` alone is insufficient unless the prior compose revision (with
  bridge) is redeployed. Mitigation: AC-07 documents both mechanisms; T03
  names owner.
- `VOC-079-R05`: **Stale foundation tests** still requiring `:8443` could block
  correct PRs or, if naively deleted without replacement, lose coverage.
  Mitigation: invert tests in T01; AC-06 single-edge invariants in T02.
- `VOC-079-DEP-00`: Predecessor VOC-067-T04 / VOC-072-T02 disposition.
- `VOC-079-DEP-01`: Orphan-removal mechanism scope.
- `VOC-079-DEP-02`: Repository verify-only mandatory vs dashboard-only.
- `VOC-079-EV-00`: verify-only run URL + EV-05 `absent` update.
- `VOC-079-EV-01`: URL normalization diff + inverted foundation tests.
- `VOC-079-EV-02`: compose/workflow/orphan/invariant evidence.
- `VOC-079-EV-03`: live deploy, four-hostname checks, ports/container
  absence, rollback + monitoring owner.
