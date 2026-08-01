# VOC-037-EV-03 — Kill-switch and rollback evidence (T03)

## Standing of `VOC-037-AC-03` at this revision

**`VOC-037-AC-03` is NOT satisfied at this revision.** Its observable
outcome requires each of the four switches to be "toggled against the
production target" and a "rollback-by-redeploy rehearsal against the
production target" to complete. Neither has happened yet: this task's
implementation has no production host access, no `PRODUCTION_*` secret,
and no authority to dispatch `deploy-production`. What is delivered here
is everything needed for that live rehearsal to be run once and
believed, plus fixes for two defects that made `AC-03` unreachable
through the sanctioned deploy path at all.

This is deliberately not recorded as "partially satisfied". `AC-03`'s
criterion is behavioral, and nothing below observes production
behavior. `VOC-037-T05` must read this file's status directly rather
than inferring it from a green pipeline.

| AC-03 clause | Status at this revision |
| --- | --- |
| Each of the four switches, toggled against the production target, observably changes behavior as documented | **Not verified.** The rehearsal that performs the toggles and asserts each documented effect exists and is proven to fail when a switch is ignored (see "Harness result" below), but it has not run against production. |
| Toggling one switch does not affect unrelated features | **Partially covered, not verified live.** Each probe targets exactly one path, and every switch is restored before the next is toggled, so a cross-effect surfaces as a failed probe. Only a live run can observe it. |
| A rollback-by-redeploy rehearsal completes without data loss beyond an intentionally reverted change | **Not verified.** Before this task there was no rollback path to rehearse at all (defect 1 below). The path now exists and the rehearsal asserts artifact identity, health, and row-count invariance across it. |

## Defects found in the existing production launch controls

Both were found by reading `T06`'s deploy path against what `AC-03`
requires, and both are fixed here. Neither is a new feature: without
them, DOC-11 §3's controls exist in the application but are not
operable through the deploy path that owns production.

1. **There was no rollback path.** `deploy-production.yml` always built
   the checked-out revision and deployed `sha-<current-sha>`. DOC-11 §2
   requires production to "deploy by digest, never by rebuilding from
   source", and §3's rollback is "redeploy previous known-good artifact
   by digest" — neither was possible, because no run could name an
   existing artifact. Re-running the workflow at an older commit is not
   the same thing: it rebuilds, producing a different artifact than the
   one that was known good. The workflow now takes a `deploy_mode` of
   `build-and-deploy` (unchanged behavior, the default) or
   `redeploy-existing-image` with an `existing_image_tag`, validated
   against `^sha-[0-9a-f]{7}$` so only an immutable per-commit tag can
   be named. In rollback mode both build steps are skipped, both
   manifests are resolved through `docker buildx imagetools inspect`
   before anything on the host is touched, and the migration apply is
   skipped because DOC-11 §3 forbids reversing a production migration
   automatically.

2. **The four kill switches were hardcoded in the workflow.**
   `AI_FEATURES_ENABLED=true` and the three `false` literals were
   written into `api.env` on every deploy, so the only way to toggle a
   switch was to hand-edit `api.env` on the host — an unaudited change
   that the next deploy silently reverted. A control that cannot be
   operated through the reviewed path is not a launch control. All four
   are now `workflow_dispatch` inputs, defaulting to exactly the values
   the workflow previously hardcoded, so this changes no deployed
   behavior by itself while making the switches operable and recorded
   per run, inside the founder-reviewed `production` environment gate.

3. **A bare recreate would have silently moved production onto the
   mutable `prod` tag.** `docker-compose.production.yml` falls back to
   `${PRODUCTION_IMAGE_TAG:-prod}`, so any `docker compose up` that did
   not export the variable — which is exactly what a naive on-host
   toggle does — would replace the running immutable `sha-` artifact
   with whatever `prod` currently points at, destroying the artifact
   identity the rollback model depends on. The rehearsal script always
   passes the intended tag explicitly and carries that reason in a
   comment at the call site.

## Repository deliverables

| Deliverable | Verification performed |
| --- | --- |
| `infra/scripts/rehearse-production-killswitch-rollback.sh` | The on-host rehearsal: `INS-12`–`INS-16`. Executed end to end against the deterministic harness below; not yet against production. |
| `infra/scripts/rehearse-production-killswitch-rollback.selftest.sh` | Executed here; ten cases, `SELFTEST PASS` (output below). |
| `.github/workflows/deploy-production.yml` | `deploy_mode`/`existing_image_tag`/four switch inputs added; YAML parses; both build steps gated on `build-and-deploy`; migration apply gated out of rollback mode; the rehearsal script added to the deploy bundle. |
| `scripts/foundation/production-launch-controls.test.mjs` | Eight static guards, run by `pnpm run test` on every pull request. Mutation-checked: re-hardcoding one switch value fails the suite. |
| `infra/README.md` | New "Launch kill switches and rollback (DOC-11 §3)" section: each switch's api.env key, default, and observable disabled effect; why toggling goes through a workflow run rather than a host edit; the rollback invocation. |

## What the rehearsal checks

Each switch is toggled off, then on, in `api.env`; the api container is
recreated (not restarted — Compose bakes `env_file` in at creation, as
`T06` found live); and two independent observations are required per
state:

- the api's own startup line (`apps/api/cmd/api/main.go` logs
  `ai=/magic=/oauth=/signups=`), which proves the process loaded the
  state the deploy wrote rather than that the two merely agree on disk;
- an HTTP probe of the exact path DOC-11 §3's switch gates.

| Check | Switch off | Switch on |
| --- | --- | --- |
| `INS-12` | `POST /api/v1/auth/magic-links` → 503 | not 503 |
| `INS-13` | `POST /api/v1/auth/oauth/google/start` → 503 | not 503 (404 "not configured" while no production Google client exists) |
| `INS-14` | consuming a valid link for an unknown address → 503, and no `users` row created | → 200, and exactly one `users` row created |
| `INS-15` | `POST /api/v1/sentence-feedback` → 200 carrying `AI_FEEDBACK_GENERATION_DISABLED` | → 404 (the request passed the generation gate and reached the deliberately non-existent target) |
| `INS-16` | rollback to a named prior tag: the api container runs that tag's image id, `/healthz` returns `status=ok`, and every invariant table's row count is unchanged; then roll forward to the original artifact and re-assert both | — |

`INS-15`'s probe deliberately uses a random, non-existent `attemptId`.
The generation gate is checked before the target lookup
(`apps/api/business/aifeedback/service.go`), so the enabled and disabled
outcomes are distinguishable without creating a saved word, a review, or
any AI-provider call — which also means the check costs nothing and
writes nothing.

The three probes that need server state (a session for `INS-15`, a link
row for `INS-14`) write it directly, under one reserved e-mail namespace
(`voc037-t03-<run>-*@rehearsal.invalid`), mirroring `VOC-034`'s
disposable-identity precedent. Tokens are generated exactly the way
`auth.generateToken` does — padded URL-safe base64 of 32 random bytes,
SHA-256 of the same bytes persisted — and the raw bytes never leave a
temporary file. Every disposable row is deleted on exit, including on
failure, and both the deletion and the pre-rehearsal row counts are
re-asserted afterward.

Restoration is a `trap`, not a final step: the original switch values
(restored by absence where a key was originally absent) and the
originally deployed image tag are put back even if a check fails, and
the restoration itself is then verified against the api's startup line.
An unevaluated check is a failure everywhere in this script — a probe
that could not complete reports `000` and fails rather than reading as
"not disabled".

## Harness result

`VOC-037-TEST-03` cannot run before a live rehearsal, and the one thing
nobody can check while writing a rehearsal script is whether it would
actually fail when the thing it rehearses is broken. `T06`'s review
found precisely that defect in the secrets-boundary rehearsal: a script
that printed observations without asserting them and passed
unconditionally.

`rehearse-production-killswitch-rollback.selftest.sh` closes that gap
without a production host. It builds a fake tier from `docker` and
`curl` stubs that model the real contracts (recreate-not-restart, the
startup log line, the 503/200/404 response mapping, image identity per
tag), runs the real rehearsal against it once expecting a pass, and once
per deliberately broken control expecting a failure. It needs no root,
no docker, and no network.

```
=== case: every kill switch honored and rollback changes the artifact (expect PASS) ===
--- as expected: the rehearsal passed
=== case: the passing run left no trace behind ===
--- as expected: switch values, image tag, and row counts all restored
=== case: magic kill switch has no effect on behavior (expect FAIL) ===
  FAIL: magic-link request with EMAIL_MAGIC_LINK_ENABLED=false returned HTTP 200, expected 503
--- as expected: the rehearsal failed
=== case: oauth kill switch has no effect on behavior (expect FAIL) ===
  FAIL: oauth start with GOOGLE_OAUTH_ENABLED=false returned HTTP 404, expected 503
--- as expected: the rehearsal failed
=== case: signups kill switch has no effect on behavior (expect FAIL) ===
  FAIL: first-time sign-up with NEW_USER_SIGNUP_ENABLED=false returned HTTP 200, expected 503
  FAIL: a user row was created while NEW_USER_SIGNUP_ENABLED=false
--- as expected: the rehearsal failed
=== case: ai kill switch has no effect on behavior (expect FAIL) ===
  FAIL: sentence feedback with AI_FEATURES_ENABLED=false returned HTTP 404, expected 200
--- as expected: the rehearsal failed
=== case: rollback leaves the pre-rollback artifact running (expect FAIL) ===
  FAIL: the api container still runs the pre-rollback image after redeploying sha-2222222
--- as expected: the rehearsal failed
=== case: rollback loses a row (expect FAIL) ===
  FAIL: row counts changed across the rollback; the artifact rollback lost or wrote data
  FAIL: row counts differ from the pre-rehearsal baseline
--- as expected: the rehearsal failed
=== case: disposable rehearsal rows are never cleaned up (expect FAIL) ===
  FAIL: disposable rehearsal rows survived cleanup
  FAIL: row counts differ from the pre-rehearsal baseline
--- as expected: the rehearsal failed
=== case: the requested rollback tag was never published (expect FAIL) ===
  FAIL: could not resolve the image id of ghcr.io/karsift/vocanova-api:sha-9999999
--- as expected: the rehearsal failed
=== case: the api container never becomes healthy after a toggle (expect FAIL) ===
FATAL: api container did not become healthy after EMAIL_MAGIC_LINK_ENABLED=false
--- as expected: the rehearsal failed
SELFTEST PASS: the rehearsal script accepts a correctly behaving tier and rejects every broken control above
```

The harness models behavior, not identity: any non-empty session cookie
with a matching CSRF header authenticates against the stub, because the
fake has no token verification. Token and hash handling is covered by
the API's own Go tests; what the harness proves is that the rehearsal's
assertions, ordering, and restoration are correct.

## Deterministic checks run

```bash
bash infra/scripts/rehearse-production-killswitch-rollback.selftest.sh
node --test scripts/foundation/production-launch-controls.test.mjs
bash -n infra/scripts/rehearse-production-killswitch-rollback.sh
bash -n infra/scripts/rehearse-production-killswitch-rollback.selftest.sh
python3 -c "import yaml; yaml.safe_load(open('.github/workflows/deploy-production.yml'))"
bash scripts/governance/validate-governance.sh
```

Observed: `SELFTEST PASS`; 8/8 node tests pass (and 7/8 with one switch
value re-hardcoded, confirming the guard bites); both scripts parse; the
workflow parses and still declares `environment: production`; governance
validation passes.

`classify-change-risk.sh --files-from` over this task's changed files
reports a path floor of **R3**, established by
`.github/workflows/deploy-production.yml`, both `infra/scripts/`
rehearsal scripts, and `infra/README.md` — matching the risk
`change.yaml` proposed for `T03`. The independent verifier owns any
semantic escalation.

Not run here, and not claimed as passing: `pnpm run test` in full
(`pnpm install` was not available in this environment — the new node
test was executed directly with `node --test`, which is what
`pnpm run test` invokes for `scripts/foundation/*.test.mjs`),
`prettier --check`, and `go test ./...` (this task changes no Go
source). CI runs all three on the pull request.

## Live rehearsal required to close AC-03

Owner: founder, or the founder-gate delegate with production access.
Ordered, because step 2 needs the artifact step 1 publishes.

1. Confirm at least two published production artifacts exist. If only
   one has ever been deployed, run `deploy-production` once in
   `build-and-deploy` mode (default inputs) so a prior
   `sha-<short-sha>` exists to roll back to, and note the tag currently
   deployed and the tag to roll back to.
2. On the shared host, as the `vocanova-production` deploy user:

   ```bash
   bash /opt/vocanova/production/infra/scripts/rehearse-production-killswitch-rollback.sh \
     api-production.vocanova.site sha-<previous-short-sha>
   ```

   Expect `PASS: all four kill switches and the rollback path behaved as
   documented`. Record the full output, the run's start and end
   timestamps, and both image tags. The script restores the tier's
   starting state itself; confirm from its `[restore]` section that it
   did.
3. Separately exercise the sanctioned rollback path end to end, which
   the on-host script deliberately does not do for itself (a rehearsal
   should not be able to dispatch a production deploy): dispatch
   `deploy-production` with `deploy_mode: redeploy-existing-image` and
   `existing_image_tag: sha-<previous-short-sha>`, approve the
   `production` environment gate, and confirm the run skips both build
   steps, logs `rollback mode: skipping migration apply`, and passes
   both health polls. Then dispatch it again in `build-and-deploy` mode
   to roll forward.
4. Optionally confirm a switch toggle through the workflow, which is the
   operational path this task makes possible: dispatch
   `deploy-production` with one switch input flipped, confirm the api's
   startup line reports the new state, then flip it back. This is the
   same behavior `INS-12`–`INS-15` assert; it verifies the input
   plumbing rather than the switch.
5. Append the recorded output and timestamps to this file, mark
   `VOC-037-AC-03` satisfied in `acceptance-criteria.md`, and note in
   `tasks.md` that `T03` is complete.

## Notes

- No production behavior changes by merging this task: every new switch
  input defaults to the value the workflow previously hardcoded, and
  `deploy_mode` defaults to the existing build-and-deploy behavior.
- This task neither closes R2 nor authorizes a release. `VOC-037-T05`
  remains the founder go/no-go gate, and autonomous production release
  remains disabled (`AGENTS.md`).
- **Follow-up, not fixed here (out of `T03`'s scope):**
  `deploy-production.yml` still builds its images from source in
  `build-and-deploy` mode, whereas DOC-11 §2 specifies "build once →
  test in staging → promote exactly to production" — promotion of the
  staging-tested digest, not a fresh production build. The rollback mode
  added here is the first half of that model (deploying a named,
  already-published artifact); making the forward path a promotion of
  the staging artifact is a separate change with its own risk
  classification.
- **Follow-up, not fixed here:** `EMAIL_MAGIC_LINK_ENABLED` and
  `GOOGLE_OAUTH_ENABLED` can now be turned on per deploy, but turning
  them on is only useful once production-tier e-mail and Google OAuth
  credentials exist (`VOC-032-DEP-07`'s production equivalent, still
  open). The rehearsal's "on" assertions account for this: they require
  a non-503 response rather than a successful sign-in.
