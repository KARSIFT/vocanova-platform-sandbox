# VOC-041 — deploy-production.yml Writes BASE_URL/OAUTH_REDIRECT_URI/OAUTH_REDIRECT_ALLOWLIST Without the Required :8443 Port, Breaking Real Google Sign-In via CORS: Specification

## Objective and requirement source

Restore real Google sign-in against production by making the three browser-facing,
CORS/redirect-relevant configuration values `deploy-production.yml` writes into
`/opt/vocanova/production/secrets/api.env` — `BASE_URL`, `OAUTH_REDIRECT_URI`, and
`OAUTH_REDIRECT_ALLOWLIST` — consistently include the `:8443` port that production
actually serves on, matching the same workflow's own already-correct health-check
and smoke-test URLs. Grounded in [issue #312](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/312)
in full, including its confirmed-by-live-reproduction root cause (nginx access logs,
API boot logs, and the live `/opt/vocanova/production/secrets/api.env` file itself,
all quoted in the issue) and its named suggested fix. Not yet approved by a founder
or technical steward — see `change.yaml`'s `requirement_approval_status`.

## Scope and non-goals

In scope:
- `.github/workflows/deploy-production.yml`'s "Write production application
  configuration" step (currently lines 256-322): append `:8443` to the host used in
  the `BASE_URL`, `OAUTH_REDIRECT_URI`, and `OAUTH_REDIRECT_ALLOWLIST` lines (lines
  299-301), so the written values become:
  - `BASE_URL=https://${PRODUCTION_API_HOST}:8443`
  - `OAUTH_REDIRECT_URI=https://${PRODUCTION_API_HOST}:8443/api/v1/auth/oauth/google/callback`
  - `OAUTH_REDIRECT_ALLOWLIST=https://${PRODUCTION_WEB_HOST}:8443/onboarding,https://${PRODUCTION_WEB_HOST}:8443/home`
- Correcting or removing the step's own comment (lines 289-298) that currently
  asserts the disproven "Cloudflare forwards plain :443 to origin :8443" claim, so
  a future reader does not re-introduce the same bug on the same mistaken belief.
  The replacement comment must record the issue's live disproof (nginx `444`
  catch-all on the unqualified host; the existing sibling "Poll production API
  health endpoint" step's own comment already recording an earlier instance of the
  same Cloudflare-fallback finding) as the reason the port is required, mirroring
  that sibling step's comment style.
- A regression test (or, if this repository has no existing test harness for
  workflow YAML content, a documented deterministic check runnable via `pnpm
  validate` or a narrower script) that fails against the pre-fix unqualified values
  and passes against the post-fix `:8443`-qualified ones, so this specific class of
  regression (a config-writing step's value silently drifting from the deploy
  target's actual serving port) is caught before the next real dispatch rather than
  only via a live production reproduction.
- Confirming (not necessarily changing) that `SESSION_COOKIE_DOMAIN` (line 312) is
  correctly unaffected, since it already derives from the parent domain rather than
  the full host and cookie domain-matching ignores port, per the issue's own
  statement and this package's own reading of the line.

Non-goals:
- `production_api_base_url`'s workflow-input default (line 20) — see "Open
  questions" below. Not fixed by this package; flagged instead.
- Any change to production's live, already-deployed `api.env` file. This package
  changes what a *future* dispatch of `deploy-production.yml` writes; it does not
  reach into the currently-running host. Whether a one-time manual correction of the
  live file is needed independent of the next scheduled deploy is a deployment/
  operational decision for the reviewing human (see `README.md`'s recommended next
  action 3), not something this package can or does perform.
- Any change to Cloudflare's own configuration. Out of this package's access and
  out of the issue's suggested fix, which explicitly frames the correct remediation
  as "until either the production/staging host-sharing setup goes away or a
  verified-working Cloudflare port-forwarding rule for the plain :443 host actually
  exists (it currently does not, per the direct reproduction)".
- Re-diagnosing or re-fixing VOC-039's already-merged Edge-runtime middleware fix.
  That fix is presumed correct and unrelated to this CORS/port defect.
- Any change to `apps/api/app/api/production.go`'s `corsAllowedOrigins` derivation
  logic itself. That logic is correct as written (it derives the allow-list from
  `OAUTH_REDIRECT_ALLOWLIST` verbatim); the defect is entirely in the value the
  deploy workflow writes into that input, not in how the API consumes it.

## Risk and protected areas

`.github/workflows/deploy-production.yml` is CI/CD deploy tooling that directly
writes the production authentication/CORS configuration boundary (OAuth redirect
URIs and the CORS allow-list consumed by `apps/api/app/api/production.go`). Per
`docs/governance/change-risk-classification.md`, "CI/CD, rollback, security,
governance enforcement, or agent authority" and "authentication/authorization" are
both named explicitly as R3-floor categories, regardless of diff size (in this
case, three string literals plus a comment). This package proposes `R3` (see
`change.yaml`); it does not touch schema migrations, billing, or a new secret, so no
higher class is proposed, but the reviewing human's own judgment governs this, not
this proposal. `scripts/governance/classify-change-risk.sh` has not been run against
a real, task-scoped file list at drafting time — consistent with how VOC-039/VOC-040
handled this field, that computation belongs to each task's own implementation PR.

## Decisions, contradictions, security, and privacy

No `VOC-041-D00`-style founder/product decision is defined by this draft — the fix
itself (matching the port the workflow's own health-check step already uses) is
fully specified by the issue and requires no product judgment call. If the
reviewing human disagrees, they should record why at adoption time rather than this
package inventing a decision.

**Contradiction recorded, not silently resolved:** the step's current comment
(lines 289-298) and the sibling health-check step's comment (lines 476-486) make
directly opposing empirical claims about the same Cloudflare behavior. The issue
resolves this by direct live reproduction (browser + nginx logs) in favor of the
health-check step's claim being correct and the config-writing step's claim being
wrong. This package adopts that resolution as its premise, per the issue's evidence,
but records that the underlying tension — this repository temporarily serves two
environments (staging, production) from one host on two different ports, purely as
an interim measure per VOC-037/T06's D00 supersession note — remains real and will
resurface if that host-sharing arrangement changes without every affected
"which port does the target actually respond on" call site (there are now at least
five across this one workflow) being re-audited together.

Security/privacy: this fix does not change *what* is authenticated, *who* can sign
in, or *what* the CORS allow-list logic itself does — it corrects the *value* fed
into an existing, already-deployed mechanism, from one that does not match any
real browser origin to one that does. No new attacker-controlled surface, secret,
or personal-data field is introduced. Widening the CORS allow-list to the port a
real browser actually uses is the intended, narrower fix, not an accidental
broadening: the allow-list still contains only production's own web host (with its
correct serving port), not a wildcard or any additional origin.

## Data, migrations, analytics, and accessibility

None. This package touches only `.github/workflows/deploy-production.yml` and its
test coverage; no schema, migration, analytics event, or accessibility surface is
affected.

## Open questions

1. **`production_api_base_url`'s workflow-input default (line 20) is very likely
   affected by the same bug class, and the issue's own claim about it does not
   match the file's current content.** The issue's description (Suggested fix
   paragraph) states the client-facing default "`https://api-production.vocanova.site:8443`
   \[is\] already fixed for the browser-bundle case earlier today" — but this
   package's drafting-time read of `.github/workflows/deploy-production.yml` line
   20 shows the default is still `https://api-production.vocanova.site`, with no
   port, and its own description text (line 17) still asserts the same disproven
   Cloudflare-forwarding claim this package's in-scope fix removes elsewhere in the
   same file. `NEXT_PUBLIC_API_BASE_URL` is baked into the web bundle at build time
   and used for real browser `fetch()` calls to the API — if this input's default is
   dispatched unqualified, browser calls to the API would hit the same nginx `444`
   catch-all the issue reproduces for the unqualified web host, independent of this
   package's `BASE_URL`/`OAUTH_REDIRECT_URI`/`OAUTH_REDIRECT_ALLOWLIST` fix. This
   package does not fix this input (the issue's suggested fix does not name it, and
   silently expanding scope past what was asked risks masking that the issue's own
   evidence claim is stale or wrong). Flagged for the reviewing human to decide:
   fix it here, in a fast follow-up package, or confirm via a fresh live check
   that some other verified mechanism actually does make the unqualified default
   work for this one input specifically before the next dispatch.
2. **Whether the currently-live production `api.env` needs a one-time manual
   correction.** This fix only changes what a future `deploy-production.yml`
   dispatch writes. The issue's own evidence shows the *currently deployed* file
   already has the unqualified (broken) values. Whether production sign-in stays
   broken until the next full dispatch, or whether an operator manually corrects
   the three lines on the live host sooner, is an operational decision outside
   this package's scope (this package has no production access) — flagged in
   `README.md`'s recommended next action 3.
