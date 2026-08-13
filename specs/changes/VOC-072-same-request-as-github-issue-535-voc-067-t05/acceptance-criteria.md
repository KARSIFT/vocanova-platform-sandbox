# VOC-072 — Acceptance Criteria

## VOC-072-AC-00 — Zone-capable Cloudflare credential exists in production GitHub environment

- Requirement source: issue #535; `VOC-072-D00`; `VOC-072-DEP-00` / `DEP-01`
- Tasks: `VOC-072-T00`
- Tests: `VOC-072-TEST-00`
- Evidence: `VOC-072-EV-00`
- Result: satisfied

Observable outcome: the GitHub **production** environment holds a secret (name
recorded at adoption) whose Cloudflare API token is scoped at minimum to **Zone
Read** and **Origin Rules Edit** for `vocanova.site`. Evidence documents token
label, scope summary, and secret name — **never** the secret value. If
`VOC-072-DEP-00` chooses reuse, evidence explicitly records that
`PRODUCTION_CLOUDFLARE_API_TOKEN` was replaced/broadened and that Workers AI
sync was re-verified.

**Standing (2026-08-13):** `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` is
present in the production environment per `t00-token-provisioning-evidence.md`
(gate_status `resolved`; restored by VOC-077-T00 after implementer regression).

## VOC-072-AC-01 — Cutover workflow and docs reference the zone-capable credential

- Requirement source: issue #535; `specification.md` `VOC-072-T01`
- Tasks: `VOC-072-T01`
- Tests: `VOC-072-TEST-01`, `VOC-072-TEST-02`
- Evidence: `VOC-072-EV-01`
- Result: pending

Observable outcome: `deploy-production.yml` job `voc067-cloudflare-cutover`
and the post-smoke `--apply` step bind the adoption-chosen secret/env var for
Cloudflare cutover modes. Workers AI sync continues to use its existing secret
binding unless DEP-00 explicitly authorizes a shared token. `infra/README.md`
and `cloudflare-remove-production-origin-port-remap.sh` header comments describe
the correct env var precedence for operators. No secret values appear in git.

## VOC-072-AC-02 — Production `--verify-only` resolves zone and reports remap status

- Requirement source: issue #535; VOC-067-TEST-06 credential clause;
  `VOC-072-T02`
- Tasks: `VOC-072-T02`
- Tests: `VOC-072-TEST-03`
- Evidence: `VOC-072-EV-02`
- Result: pending

Observable outcome: a `workflow_dispatch` of `deploy-production.yml` with
`voc067_cloudflare_origin_cutover=verify-only` completes successfully (exit 0)
using the wired secret. Job log (redacted) shows zone resolution succeeded and
prints either `OK: no origin rules remap production hosts to port 8443` or an
explicit FOUND message — not `ERROR: zone not found` from an empty zone list.
This satisfies the **credential unblock** for VOC-067-TEST-06; it does not by
itself close VOC-067-AC-06 if remap is still present (apply remains separate per
open question 3).

Acceptance criteria must be observable, stable, security-aware, and bidirectionally
traceable to requirements, tasks, tests, and evidence.
