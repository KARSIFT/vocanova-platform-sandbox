# VOC-034 — Staging and Live-Verification Evidence

## Purpose

This document records the evidence required by `VOC-034-AC-10` (via
`VOC-034-TEST-10`) — the live-staging verification issue #216 requires to actually
unblock `KARSIFT/vocanova-platform-sandbox#185` (VOC-032-T09). It is drafted before
adoption/implementation, mirroring `VOC-032`'s and `VOC-031`'s `staging-evidence.md`
convention, and is updated once `VOC-034-T03` actually executes.

## Current status

**Passed on 2026-07-31.** `VOC-034-T00`–`T02` merged, merge commit
`f990e86efeef73730d747d53ea2d1ca7cd77bf84` was deployed by
`deploy-staging` run
[30618654496](https://github.com/KARSIFT/vocanova-platform-sandbox/actions/runs/30618654496),
and the live disposable-identity exercise below proved that an ordinary safe
sentence now reaches the real OpenCode moderation and feedback path. Both required
database rows were present before cleanup, and all disposable rows were deleted and
verified absent afterward. `VOC-034-DEP-01` is resolved and `VOC-034-AC-10` passes.

## Evidence summary

| Evidence | Requirement | Status |
| --- | --- | --- |
| `EV-05` | API restarts healthy, logs `ai=on`, after redeploy | **Passed** — API container started `2026-07-31T09:07:26Z`, reported healthy, and public `/healthz` returned `status=ok`, `database=ok` at `09:07:58Z` |
| `EV-05` | Real `POST /api/v1/sentence-feedback` for an ordinary safe sentence, via a disposable identity, does not return `SAFETY_MODERATION_UNAVAILABLE` | **Passed** — HTTP 200, no error code, real feedback status `correct` |
| `EV-05` | A `learner_sentences` row and an `ai_feedback_attempts` row are created for the attempt, then deleted along with the disposable identity, session, word fixture, saved word, and review used to reach the sentence-feedback step | **Passed** — both rows counted `1` before cleanup; all disposable row groups counted `0` afterward |

## Exercise procedure

See `tasks.md`'s `VOC-034-T03` for the full ordered procedure. Summary: confirm
healthy restart with `ai=on` → create disposable identity → save word → complete
review → submit ordinary safe sentence → confirm real (non-fail-closed) result and
persisted rows → delete every disposable artifact created during the exercise →
record timestamps/row-count evidence below → update issue #216 and
`KARSIFT/vocanova-platform-sandbox#185` noting the blocker is resolved.

## Execution record

### Deployment and service health

- Fixed revision: `f990e86efeef73730d747d53ea2d1ca7cd77bf84`.
- Deploy run: `30618654496`, manually dispatched at `2026-07-31T09:05:44Z`
  because the GitHub App merge did not emit the expected `push` workflow event;
  completed successfully.
- API image: `ghcr.io/karsift/vocanova-api:dev`, image digest
  `sha256:e4741f094541dbf8aecd43f77f3daa19d851670f8c176bdc7601271f1e86ae43`.
- API container start: `2026-07-31T09:07:26Z`; Docker health:
  `healthy`.
- Public health observation at `2026-07-31T09:07:58Z`:
  HTTP success with `status=ok`, `database=ok`.
- Public web root returned HTTP 200 through Cloudflare.
- Sanitized startup log: `env=staging, ai=on, magic=off, oauth=off,
  signups=off`.
- `vocanova-opencode.service` was active and listening only on the private
  Docker gateway `172.18.0.1:4096`; the API container successfully fetched the
  private OpenCode `/doc` endpoint.

### Disposable public-API core loop

- Run tag: `20260731T091043Z-7c5c2a5a`.
- Started: `2026-07-31T09:10:43Z`; completed before cleanup:
  `2026-07-31T09:10:57Z`.
- Identity: synthetic non-production user
  `3b766c48-76d4-4d7d-b63b-2b416a552a05`.
- Synthetic word/meaning fixture:
  `456f819f-40c1-45ab-9f8c-42354c4dc10f` /
  `10feebc0-85d3-43a1-9fe4-fbf1eea60db5`.
- Save-word request: HTTP 200; user-word ID
  `a960b70f-c458-48da-a6bd-96d5a6c69877`.
- Review submission: HTTP 200; review-attempt ID
  `e5322adb-fb84-4236-bb98-a1654574e5cb`.
- Sentence-feedback request used the ordinary synthetic sentence described by
  `VOC-034-T03`: HTTP 200 in `13,423 ms`, no `errorCode`, feedback status
  `correct`.
- Learner-sentence ID:
  `b0e58083-bdb2-402a-9c1c-e2df9060bf65`.
- AI-feedback-attempt ID:
  `4a5a01cd-6a7d-4662-9467-5811f272c155`.
- The feedback row matched `status=succeeded`, `provider=opencode`, and
  `model=opencode-go/hy3`.

Pre-cleanup row counts scoped to the disposable IDs:

| Row group | Count |
| --- | ---: |
| `users` | 1 |
| `user_words` | 1 |
| `review_attempts` | 1 |
| `learner_sentences` | 1 |
| successful Hy3 `ai_feedback_attempts` | 1 |

### Cleanup proof

Cleanup ran in one explicit transaction at `2026-07-31T09:10:57Z`. It deleted:

- 1 AI feedback attempt;
- 1 learner sentence;
- 1 review attempt;
- 3 idempotency records;
- 1 session;
- 1 user word;
- 1 disposable user;
- 1 word meaning; and
- 1 canonical word.

Post-cleanup counts were all zero for the disposable user, sessions, user words,
reviews, learner sentences, AI feedback attempts, word meaning, and canonical word.
No real learner account or content was used, and no credential or token was recorded.

### Finding carried to VOC-032-T10

The successful two-provider request took `13,423 ms`, above DOC-09's nominal
10-second combined backend target. This does not invalidate `VOC-034-AC-10`, whose
criterion is reachability, persistence, and cleanup, but it must be evaluated
honestly by VOC-032-T10's live threshold procedure rather than treated as a passing
latency result here.
