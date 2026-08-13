# VOC-077 — Test Plan

## VOC-077-TEST-00 — Evidence file reflects provisioned secret

- Covers: `VOC-077-AC-00`
- Preconditions: `VOC-077-T00` diff available; implementer has (or cites)
  redacted `gh secret list --env production` output showing
  `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`.
- Procedure:
  1. Read
     `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md`.
  2. Assert frontmatter `gate_status` is not `pending_operator_execution`.
  3. Assert the file does not claim AC-00/TEST-00 are "NOT satisfied" due to
     missing provisioning.
  4. Assert redacted operator confirmation / audit fields name the secret
     and cite presence without including a token string.
  5. Cross-check: `gh secret list` (or the redacted excerpt in evidence)
     still lists the secret name.
- Expected result: Evidence matches confirmed production presence.
- Evidence: `VOC-077-EV-00`

## VOC-077-TEST-01 — DEP-00/DEP-01 text consistent with evidence

- Covers: `VOC-077-AC-01`
- Preconditions: VOC-072 `change.yaml` edited in the same PR tip.
- Procedure:
  1. Read `VOC-072-DEP-00` and `VOC-072-DEP-01` status fields.
  2. Grep the dependency block for phrases that claim presence is still
     outstanding / unresolved-drafting in a way that contradicts the
     evidence file.
  3. Confirm dedicated-secret choice and secret name
     `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` remain recorded.
  4. Confirm VOC-072 `approval_status` /
     `implementation_authorized` / other adoption fields were not altered
     by this task.
- Expected result: No contradictory "still outstanding" claim; DEP status
  matches evidence; adoption fields untouched.
- Evidence: `VOC-077-EV-00`

## VOC-077-TEST-02 — AC-00 Result, no secrets in diff, no out-of-scope files

- Covers: `VOC-077-AC-02`
- Preconditions: Full PR diff available.
- Procedure:
  1. Confirm `VOC-072-AC-00` Result is satisfied (or equivalent) when
     evidence is resolved.
  2. Scan the diff for high-entropy secret-like strings, Cloudflare token
     patterns, or pasted full secret values — expect none.
  3. Assert file list is limited to VOC-072 package evidence/metadata
     (and this package's own `t00-evidence.md` if added). No
     `.github/workflows/`, `infra/scripts/`, or cutover-behavior edits.
  4. Confirm independent-review comment on the tip is PASS or PASS WITH
     NON-BLOCKING FINDINGS.
  5. Run:

     ```bash
     bash scripts/governance/validate-governance.sh
     bash scripts/governance/classify-change-risk.sh
     git diff --check
     ```

- Expected result: Validation passes; declared risk meets or exceeds
  detected floor; scope and secrecy constraints hold; review PASS.
- Evidence: `VOC-077-EV-00`

## Rollback coverage

Rolling back means reverting the VOC-072 evidence / DEP / AC-00 Result
commits (and this package's `t00-evidence.md` if present). Validation:
re-run governance checks on the reverted tree. No data or secret rollback
is required — this package does not mutate secrets.

## Constraints

No test in this plan reads or prints secret values. No live
`--verify-only` / `--apply` cutover run is required for VOC-077 closure
(that remains VOC-072-T02). Name-only `gh secret list` is the allowed
production-environment check.
