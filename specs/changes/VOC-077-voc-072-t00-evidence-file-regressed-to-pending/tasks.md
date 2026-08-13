# VOC-077 — Tasks

## VOC-077-T00 — Correct VOC-072-T00 evidence and DEP-01 contradiction

- Requirement source: issue #578; `VOC-077-D00`
- Acceptance criteria: `VOC-077-AC-00`, `VOC-077-AC-01`, `VOC-077-AC-02`
- Tests: `VOC-077-TEST-00`, `VOC-077-TEST-01`, `VOC-077-TEST-02`
- Evidence: `VOC-077-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

Fresh task / new PR — **do not** redispatch exhausted issue #543 and **do
not** force-merge PR #558 while it still has a FAIL verdict.

### Required work

1. **Re-verify presence (redacted only).** From an environment with
   production-environment secret list access:

   ```bash
   gh secret list --env production --repo KARSIFT/vocanova-platform-sandbox \
     | grep -E 'PRODUCTION_CLOUDFLARE'
   ```

   Confirm `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN` is listed. If it
   is **missing**, stop and escalate — do not invent a "resolved" evidence
   file; that would recreate the contradiction this package exists to fix.
   (Re-provisioning is **out of scope** here; that would return to VOC-072-T00
   human-ops authority.)

2. **Restore
   `specs/changes/VOC-072-same-request-as-github-issue-535-voc-067-t05/t00-token-provisioning-evidence.md`.**
   Prefer building on the substance of reviewed commit `f49ffc50` (confirmed
   provisioned state), **not** the regressed pending content of `4b021050` /
   current pending template. Required properties:
   - Frontmatter `gate_status`: confirmed/resolved (not
     `pending_operator_execution`).
   - Narrative: AC-00/TEST-00 satisfied / gate closed for presence.
   - §5 (or equivalent) operator confirmation filled with redacted facts
     (secret name, provisioning timestamp from `gh secret list` updated-at
     if available, operator handle if known from prior revision — never the
     token value).
   - Preserve DEP-00 dedicated-secret and DEP-01 name decisions already
     recorded in §1.
   - Keep runbook sections as historical operator record if useful; they
     must not override the resolved gate status.

3. **Fix VOC-072 `change.yaml` DEP-00/DEP-01 status text.** Remove any claim
   that production-environment presence is still outstanding. Record that
   DEP-00 chose dedicated secret and DEP-01 name is
   `PRODUCTION_CLOUDFLARE_ZONE_ORIGIN_RULES_TOKEN`, with presence confirmed
   (cite evidence file / `gh` check date). Do **not** flip VOC-072 adoption /
   authorization / `automatic_merge_allowed` fields as part of this task
   (VOC-075 may address `automatic_merge_allowed` separately).

4. **Align `VOC-072-AC-00` Result** in VOC-072 `acceptance-criteria.md` with
   the evidence (satisfied when evidence is resolved).

5. **Record `t00-evidence.md` in this VOC-077 package** summarizing: commands
   run, redacted list excerpt, files changed, SHA of the correction PR tip,
   and that no secret values appear in the diff.

6. **Independent review** must PASS on the exact tip before merge-gate can
   merge. Do not treat a founder `approved` comment alone as sufficient if
   review is FAIL.

### Explicitly out of scope for this task

- VOC-072-T01 workflow wiring and VOC-072-T02 `--verify-only`.
- Editing deploy workflows or cutover scripts.
- Investigating/fixing implementer regeneration (`VOC-077-DEP-01`) unless
  adoption expanded that dependency into this task.
- Closing or commenting on PR #558 / issue #543 beyond what adoption
  recorded for `VOC-077-DEP-00` (if adoption asks the implementer to close
  #558 after merge, do that as a follow-through note in evidence — not a
  substitute for the file corrections above).

## Task ordering notes

- Single task; one PR is expected.
- No task may be dispatched before this package is adopted.
- Closing this task with PASS evidence is what unblocks VOC-072-T01/T02 from
  an evidence-gate perspective; those tasks remain separate VOC-072 work.

Tasks preserve scope, separation of duties, and rollback safety.
