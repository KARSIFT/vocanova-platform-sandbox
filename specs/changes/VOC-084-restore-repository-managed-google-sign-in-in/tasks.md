# VOC-084 — Tasks

None of the tasks below is implementation-authorized by this package.
Adoption and each task's own implementation authorization are separate.
Order is mandatory: **T00 → T01 → T02**.

## VOC-084-T00 — Staging Google OAuth sync, canonical config, allowlist, workflow tests

- Requirement source: issue #691; `VOC-084-D00`–`D03`; `VOC-084-D06`;
  `VOC-084-DEP-00`; `VOC-084-DEP-02`
- Acceptance criteria: `VOC-084-AC-00`, `VOC-084-AC-01`, `VOC-084-AC-02`
  (config half), `VOC-084-AC-03`, `VOC-084-AC-05` (workflow/config tests),
  `VOC-084-AC-06` (isolation)
- Tests: `VOC-084-TEST-00`, `VOC-084-TEST-01`, `VOC-084-TEST-02`,
  `VOC-084-TEST-03`, `VOC-084-TEST-04`
- Evidence: `VOC-084-EV-00` (`t00-evidence.md` in this package directory)
- Status: pending

### Required work

1. In `.github/workflows/deploy-staging.yml`, add Google OAuth credential
   synchronization mirroring the safe production pattern:
   - both secrets unset → coherent skip / disabled posture
   - exactly one set → fail before application convergence
   - both present → write to staging secret file only, `chmod 600`, never
     log values
2. On each staging deploy convergence, write/overwrite:
   - `OAUTH_REDIRECT_URI=https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`
   - staging-appropriate `OAUTH_REDIRECT_ALLOWLIST` for the staging web host
   - `GOOGLE_OAUTH_ENABLED` from actual complete credential availability
   - `NEW_USER_SIGNUP_ENABLED=false`
   - `NEW_USER_SIGNUP_ALLOWLIST` from the repository-workflow control
     accepted at adoption (recommended default: workflow_dispatch input
     defaulting to empty)
3. Update any staging deploy comments/docs in-scope that would otherwise
   claim these keys remain exclusively founder-manual.
4. Add deterministic tests covering both-present, both-absent,
   partial-pair rejection, canonical callback, and allowlist control /
   default-empty behavior.
5. Do not change production OAuth sync behavior or shared-edge topology.
6. Record commands, results, and AC mapping in `t00-evidence.md`.

### Explicitly out of scope for this task

- UI capability gating (T01).
- Live OAuth-start check wiring and Google Console disposition (T02).
- Completing a real Google login.
- Manual SSH/host edits.

## VOC-084-T01 — Sign-in UI consumes OAuth capability signal

- Requirement source: issue #691; `VOC-084-D04`
- Acceptance criteria: `VOC-084-AC-04`, `VOC-084-AC-05` (UI tests)
- Tests: `VOC-084-TEST-05`
- Evidence: `VOC-084-EV-01` (`t01-evidence.md`)
- Status: pending — depends on `VOC-084-T00` merging to `develop` for
  end-to-end staging proof that enabled/disabled tracks deploy state; the
  UI code change itself may land against the existing `/healthz` signal

### Required work

1. Make `apps/web` sign-in consume a deploy-derived / auth-capability signal
   (prefer `GET /healthz` → `kill_switches.oauth_enabled`).
2. Do not render "Continue with Google" when capability reports disabled.
3. Avoid staging-only hardcoded lies and avoid enabling the button merely
   because a build succeeded.
4. Add deterministic UI tests for disabled-method rendering (and enabled
   rendering when capability is true, if the harness supports it without
   inventing live Google login).
5. Preserve accessibility of remaining methods and the
   `max-w-[28rem]` sign-in layout workaround.
6. Record evidence in `t01-evidence.md`.

### Explicitly out of scope for this task

- Deploy-staging secret sync (T00).
- Post-deploy OAuth-start check (T02).
- Auth redesign beyond capability-gated rendering.

## VOC-084-T02 — Post-deploy OAuth-start check + Google callback disposition

- Requirement source: issue #691; `VOC-084-D01`; `VOC-084-D05`;
  `VOC-084-DEP-01`
- Acceptance criteria: `VOC-084-AC-02` (live half), `VOC-084-AC-05`,
  `VOC-084-AC-06`, `VOC-084-AC-07`
- Tests: `VOC-084-TEST-06`, `VOC-084-TEST-07`, `VOC-084-TEST-08`
- Evidence: `VOC-084-EV-02` (`t02-evidence.md`)
- Status: pending — depends on `VOC-084-T00` (and preferably `VOC-084-T01`)
  merging to `develop`

### Required work

1. Add a post-deploy OAuth-start check to `deploy-staging.yml` that:
   - POSTs staging `/api/v1/auth/oauth/google/start` with a staging web
     `redirectUri` (issue evidence used
     `https://staging.vocanova.site/home`)
   - requires HTTP 200 when credentials are present and OAuth is enabled
   - requires an `accounts.google.com` authorization URL
   - requires `redirect_uri` exactly
     `https://api-staging.vocanova.site/api/v1/auth/oauth/google/callback`
   - does **not** follow Google or complete OAuth
2. When the pair is absent, assert coherent disabled behavior (not a fake
   200).
3. Confirm public health endpoints remain healthy after deploy; note
   isolation preserved.
4. From available access/evidence, verify Google client staging-callback
   authorization **or** record the exact external configuration requirement
   without claiming it complete.
5. Record run URLs and redacted evidence in `t02-evidence.md` (no secrets).

### Explicitly out of scope for this task

- Inventing Google Cloud mutations without access.
- Completing real OAuth login in CI.
- Production deploy behavior changes.

## Task ordering notes

- T00 blocks T01's staging end-to-end proof and T02's live OAuth-start gate.
- T01 may implement against `/healthz` independently but should not claim
  staging advertise-while-disabled closure until T00 has converged real
  capability.
- T02 is the package's live proof + external-dependency disposition task.
- No task may be dispatched before this package is adopted and
  implementation-authorized.
- Closing issue #691 is gated on AC results with evidence, not on task issue
  closure alone.

Tasks preserve scope, separation of duties, and rollback safety.
