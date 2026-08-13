# VOC-077 — Acceptance Criteria

## VOC-077-AC-00 — VOC-072-T00 evidence matches confirmed provisioned secret

- Requirement source: issue #578; `VOC-077-D00`
- Tasks: `VOC-077-T00`
- Tests: `VOC-077-TEST-00`
- Evidence: `VOC-077-EV-00`
- Result: pending

Observable outcome:

1. `t00-token-provisioning-evidence.md` has `gate_status` set to a confirmed /
   resolved value (not `pending_operator_execution`).
2. The evidence narrative states that
   `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` is present in the GitHub
   **production** environment, with redacted audit fields filled (operator,
   date, secret name, redacted `gh secret list` excerpt showing name +
   updated-at — **never** the token string).
3. AC-00 / TEST-00 language in that evidence file no longer claims "NOT
   satisfied" solely because of a stale pending template.
4. Live re-check at implementation time
   (`gh secret list --env production --repo KARSIFT/vocanova-platform-sandbox`)
   still lists the secret; evidence cites that check without inventing values.

## VOC-077-AC-01 — VOC-072 `change.yaml` DEP text no longer contradicts presence

- Requirement source: issue #578 Medium finding on `f49ffc50`; `VOC-077-D00`
- Tasks: `VOC-077-T00`
- Tests: `VOC-077-TEST-01`
- Evidence: `VOC-077-EV-00`
- Result: pending

Observable outcome: VOC-072's `change.yaml` dependency entries for
`VOC-072-DEP-00` / `VOC-072-DEP-01` no longer claim that
production-environment presence is still outstanding, and no longer leave
DEP-01 as unresolved-drafting-only when the dedicated secret name and
presence are confirmed. Status text matches the evidence file (dedicated
secret chosen; name
`PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`; provisioned).

## VOC-077-AC-02 — VOC-072-AC-00 Result aligned; no secret leakage; no out-of-scope edits

- Requirement source: issue #578; `specification.md` non-goals
- Tasks: `VOC-077-T00`
- Tests: `VOC-077-TEST-02`
- Evidence: `VOC-077-EV-00`
- Result: pending

Observable outcome:

1. `VOC-072-AC-00` Result in VOC-072 `acceptance-criteria.md` is updated to
   satisfied (or equivalent) when the evidence supports it — not left
   `pending` while evidence claims resolved.
2. Diff contains **no** secret token strings, Cloudflare API responses with
   credentials, or `.env` / secret files.
3. Diff does **not** edit `.github/workflows/`, `infra/scripts/`,
   `infra/README.md`, or implement VOC-072-T01/T02 behavior.
4. Independent review of the exact revision reports `VERDICT: PASS` or
   `PASS WITH NON-BLOCKING FINDINGS` (no open Critical/High; no unwaived
   Medium that re-asserts the pending contradiction).
