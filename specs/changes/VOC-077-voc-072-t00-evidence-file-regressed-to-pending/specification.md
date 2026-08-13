# VOC-077 — VOC-072-T00 Evidence File Regressed to Pending: Specification

## Objective and requirement source

Restore VOC-072-T00's repository evidence and package dependency text so they
accurately record that the production GitHub environment secret
`PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` is already provisioned —
unblocking VOC-072-T01/T02 without re-running exhausted task #543 or
force-merging the FAILed PR #558.

Requirement source:
[GitHub issue #578](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/578).

## Confirmed findings (from issue #578)

- Production secret exists:
  `gh secret list --repo KARSIFT/vocanova-platform-sandbox --env production`
  lists `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`, added
  `2026-08-13T21:24:26Z`.
- PR #558 commit `f49ffc50bdb3f9183660ba3715cb9436718e6e81` (reviewed
  2026-08-13T21:29:46Z) correctly documented provisioning and was down to one
  **Medium** finding: `change.yaml` DEP-01 still contained a sentence claiming
  "production-environment presence is still outstanding," contradicting the
  evidence file's own confirmed state.
- Commit `4b021050d75b93b9ac54444419dda9e0c598d3cd` (reviewed
  2026-08-13T22:06:26Z, PR #558 HEAD) **regressed**
  `t00-token-provisioning-evidence.md` to
  `gate_status: pending_operator_execution` with AC-00/TEST-00 marked "NOT
  satisfied." Independent review correctly returned High / FAIL for
  contradicting reality.
- Task issue #543 shows both attempts (2/2) exhausted as of
  `2026-08-13T22:07:47Z`; automation stopped retrying.
- Founder `approved` on PR #558 (`2026-08-13T22:10:35Z`) correctly did not
  force-merge — merge-gate requires positive `VERDICT: PASS` / `PASS WITH
  NON-BLOCKING FINDINGS` proof and does not override a confirmed FAIL. That
  gate behavior is working as designed, not a bug to fix here.
- Drafting-time read of the current
  `t00-token-provisioning-evidence.md` on this branch still shows
  `gate_status: pending_operator_execution` and "AC-00 is NOT satisfied" —
  the regressed content is what a new task must correct.

## Scope and non-goals

In scope:

- `VOC-077-T00`: In a new task PR (not a redispatch of #543):
  1. Update
     `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md`
     so frontmatter `gate_status` is `resolved` (or equivalent confirmed
     status), AC-00/TEST-00 are marked satisfied, operator §5 fields are
     filled with redacted facts (never the token string), and the narrative
     matches the confirmed provisioned state (substance of `f49ffc50`,
     re-verified against live `gh secret list` at implementation time).
  2. Fix VOC-072 `change.yaml` DEP-00/DEP-01 `status` text so it no longer
     claims production-environment presence is still outstanding or leaves
     DEP-01 as unresolved-drafting when the secret name and presence are
     confirmed.
  3. Align `VOC-072-AC-00` Result in `acceptance-criteria.md` with the
     evidence if it still says `pending`.
  4. Obtain independent review PASS (or PASS WITH NON-BLOCKING FINDINGS) on
     that exact revision so merge-gate can proceed.

Non-goals / explicitly excluded:

- Not creating, rotating, deleting, or printing the secret value.
- Not implementing VOC-072-T01 (`deploy-production.yml` / script wiring) or
  VOC-072-T02 (live `--verify-only`).
- Not changing Cloudflare token scopes or Workers AI sync credentials.
- Not force-merging PR #558 while it still carries a FAIL verdict.
- Not redispatches of exhausted issue #543.
- Not fixing karsift-ai-infra implementer regeneration behavior unless
  adoption expands `VOC-077-DEP-01` (default: separate follow-up).
- Not adopting, authorizing, implementing, or merging this package from
  within the draft itself.

## Risk and protected areas

Builder assessment: expected edits are confined to
`specs/changes/VOC-072-…` evidence / metadata files. Path classifier floor
measured at drafting time: **R1**.

This package proposes **R1** — documentation/evidence correction matching
already-confirmed production presence. No governance authority change, no
workflow edit, no secret mutation. EHR is not triggered. Under active A-003,
no standing technical-steward or founder approval is required merely for
this class; independent verification still applies.

If implementation expands into workflows, infra scripts, or secret
settings APIs, the verifier must raise the class.

## Decisions, contradictions, security, and privacy

`VOC-077-D00` (recorded here for traceability; formal decision numbering
applies after adoption): When production presence is independently
confirmable via redacted `gh secret list` (name + updated-at only) and a
prior reviewed revision already documented that fact, the package record
must not claim `pending_operator_execution` / "NOT satisfied." Correcting
the record is required; inventing a new secret or pasting secret values
into git is forbidden.

No contradiction with VOC-072's dedicated-secret decision (`VOC-072-DEP-00`
dedicated; `VOC-072-DEP-01` name
`PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`) — this package preserves
those decisions and only updates status text / evidence to match reality.

Open questions for the reviewing human:

1. **`VOC-077-DEP-00` — PR #558 / issue #543 disposition.** Recommended
   default: after VOC-077-T00 merges with PASS, close PR #558 without merge
   and treat VOC-072-T00's acceptance obligation as satisfied by
   `VOC-077-EV-00` (do not redispatch #543). Confirm at adoption.
2. **`VOC-077-DEP-01` — Implementer regression root cause.** Recommended
   default: out of this package; open a separate unlabeled issue if pipeline
   hardening is wanted (e.g. implementer overwrote confirmed evidence from a
   stale template). Confirm at adoption.

Security / privacy: no secret values in git, PR bodies, or evidence files.
Re-verification uses name-only `gh secret list` output. No personal data.

## Data, migrations, analytics, and accessibility

- **Data / migrations:** None.
- **Analytics:** None.
- **Accessibility:** None. Explicit non-applicability: evidence /
  package-metadata correction only.
