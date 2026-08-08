# VOC-051-EV-04 — T00 Sentry organization/plan precondition check

Evidence for `VOC-051-T00` and `VOC-051-TEST-09`.

## Gate status: RESOLVED — `VOC-051-T01` and `VOC-051-T02` may proceed

**`VOC-051-DEP-00` is resolved as of 2026-08-08.** The founder confirmed
signed-in access to the organization's real Sentry account, chose Layout B
(§2), created all four projects, created a scoped internal-integration token,
and provided all DSN values directly (through the founder-gate overseer
session, not committed to this repository). This record is now the "written
confirmation" branch of `VOC-051-TEST-09`'s expected result.

§3's layout table below is filled. The following GitHub Actions secrets exist
on `KARSIFT/vocanova-platform-sandbox` as of this resolution (values are not
recorded in this file or anywhere else in the repository, per this file's own
"DSN values are deliberately absent" rule below):

- `PRODUCTION_SENTRY_DSN` — updated in place. **Important operational note:**
  the founder deleted the pre-existing Sentry project this secret originally
  pointed to (while cleaning up before creating the four new projects), which
  silently disabled `apps/api` production error reporting for a window of
  time. The secret has been corrected to point at the new `prod-api` project,
  but the fix only takes effect on `apps/api`'s **next production deploy** —
  it does not retroactively apply to the already-running container.
- `STAGING_SENTRY_DSN` — new.
- `PRODUCTION_WEB_SENTRY_DSN` — new.
- `STAGING_WEB_SENTRY_DSN` — new.
- `SENTRY_API_TOKEN` — new; internal-integration token scoped to `project:read`
  + `event:read` per §1b, for `VOC-051-T02`'s hourly workflow.

## 1. What this task did establish

The following are vendor-published and repository-verified facts. They narrow
`specification.md`'s open question 3 substantially, but none of them is a
statement about the founder's specific Sentry organization, so none of them
discharges the gate.

### 1a. Sentry does not cap project count on any plan

Sentry's published pricing (`https://docs.sentry.io/pricing/`) defines one free
plan (Developer) and three paid plans (Team, Business, Enterprise). None
documents a project-count limit. The only "5 projects" figure in Sentry's docs is
an internal divisor inside the spike-protection threshold formula
(`https://docs.sentry.io/pricing/quotas/spike-protection/`), not an account cap.

Consequence: the drafting-time worry that "some free/starter tiers restrict
project count" (`specification.md` open question 3) does not appear to be real
for Sentry. Adding an `apps/web` project should not be plan-gated.

The Developer plan's real binding limits are **5,000 errors/month** and **1 user
seat**. Both matter to this package and are recorded as findings in §5.

### 1b. An org-scoped read-only token is available on every plan, but the
plain "organization auth token" is probably the wrong token type

Sentry offers three token types
(`https://docs.sentry.io/account/auth-tokens/`,
`https://docs.sentry.io/api/permissions/`):

| Token type | Bound to | Scopes |
| --- | --- | --- |
| Organization auth token (`sntrys_` prefix) | organization | **fixed, not selectable** — a limited CI/`sentry-cli` set |
| Internal-integration token | organization | **selectable and editable** after creation |
| Personal token | a user account | selectable at creation, not editable; capped by that user's own permissions |

The workflow `VOC-051-T02` will build needs to list issues and read events, which
requires `event:read` (Sentry's Issues & Events `GET` scope) and `project:read`
(project `GET`). Because the plain organization auth token's permission set is
fixed rather than selectable, it cannot be assumed to carry `event:read`.

The token type that satisfies **both** "org-scoped, not tied to a person" (the
issue's requirement, and the right call given the Developer plan's 1-seat limit —
a personal token would die with that seat) **and** "read-only, least privilege"
is therefore an **internal-integration token scoped to exactly `project:read` +
`event:read`**, with no `*:write` or `*:admin` scope.

Creating an internal integration requires an org-level role of **Manager or
Admin** (`https://docs.sentry.io/api/guides/create-auth-token/`). The founder, as
org owner, qualifies.

This is a substantive narrowing of `specification.md`'s open question 2 for
`VOC-051-T02`'s implementer, who still owns recording the final chosen scope and
its justification in that task's pull request. §4's confirmation has since
happened (see §3/§4 below); this subsection is left as-is because it
correctly predicted the actual choice made (internal-integration token,
`project:read` + `event:read`).

### 1c. In Sentry, an "environment" is a tag inside a project, not a project

DOC-11 §1 requires "separate Sentry environments per environment tier." In
Sentry's data model that separation is carried by the `environment` tag on each
event (set by the SDK), and the DSN identifies the **project**, not the tier. One
project can therefore serve both staging and production, distinguished by tag.

`apps/api`'s existing wiring already works exactly this way, and that is the
repo-consistent default:

- `apps/api/app/api/production.go` lines 177–179 read `SENTRY_DSN`,
  `SENTRY_ENVIRONMENT` (defaulting to `ENVIRONMENT`, itself defaulting to
  `staging`), and `SENTRY_RELEASE`.
- `.github/workflows/deploy-production.yml` line 188 injects the
  `PRODUCTION_SENTRY_DSN` secret and writes it to the host as `SENTRY_DSN`
  (line 211), then sets `SENTRY_ENVIRONMENT=production` as a literal (line 322)
  and `SENTRY_RELEASE=sha-<short_sha>` (line 323).
- `.github/workflows/deploy-staging.yml` references no `SENTRY_*` value at all,
  so staging currently runs with `SENTRY_DSN` unset and Sentry disabled
  (`apps/api/cmd/api/main.go` line 57 prints `sentry disabled (SENTRY_DSN
  unset)`).
- `apps/web` has no Sentry reference of any kind.

So the layout choice §3 puts to the human is a real, consequential one, and it is
narrower than "map out whatever exists": it is A-or-B.

## 2. The two candidate layouts

**Layout A — two projects, environment carried by tag (repo-consistent).**
One Sentry project per application (`apps/api`, `apps/web`); each has exactly one
DSN, used by both tiers; staging and production are separated by the
`SENTRY_ENVIRONMENT` / `environment` tag. This is what `apps/api` already
implements. Two DSN values total. Quota is pooled per project across tiers, so a
staging error storm consumes production's quota.

**Layout B — four projects, one per application per tier.** Four DSN values.
Stronger isolation (a staging storm cannot exhaust the production project's
quota, and per-project alert rules differ cleanly), at the cost of diverging from
`apps/api`'s existing single-DSN-plus-tag wiring, which `VOC-051-T01` is
explicitly instructed to mirror rather than reinvent.

Neither was chosen by this task itself — that decision belonged to the human
with access to the founder's real org. Under the Developer plan's shared
5,000 errors/month account-wide cap, Layout B's isolation benefit is partly
illusory (the cap is per-account, not per-project), which was a point for
Layout A; the founder chose Layout B anyway (see §3/§4), knowingly accepting
that trade-off.

## 3. Layout record — confirmed by the founder 2026-08-08

**Layout B chosen** (§2): four projects, one per application per tier — not
the `apps/api`-precedent Layout A. The founder's stated reasoning: stronger
isolation per environment, accepted knowingly despite §2's noted caveat that
the Developer plan's quota is pooled account-wide regardless of project count.
`VOC-051-T01` must follow this table, not silently default to Layout A/apps/api's
existing single-DSN-plus-tag pattern.

**Sentry organization slug:** `vocanova` (organization ID `4511838056480768`).

| App | Tier | Sentry project slug | GitHub Actions secret holding the DSN | Runtime env var | `SENTRY_ENVIRONMENT` value | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `apps/api` | production | `prod-api` | `PRODUCTION_SENTRY_DSN` (updated in place — see gate-status note above re: the deleted predecessor project) | `SENTRY_DSN` | `production` | wired today; secret corrected, takes effect next deploy |
| `apps/api` | staging | `stage-api` | `STAGING_SENTRY_DSN` (new) | `SENTRY_DSN` | `staging` | not wired; `VOC-051-T01` adds |
| `apps/web` | production | `prod-web` | `PRODUCTION_WEB_SENTRY_DSN` (new) | `NEXT_PUBLIC_SENTRY_DSN` (proposed — `VOC-051-T01` confirms) | `production` | not wired; `VOC-051-T01` adds |
| `apps/web` | staging | `stage-web` | `STAGING_WEB_SENTRY_DSN` (new) | `NEXT_PUBLIC_SENTRY_DSN` (proposed — `VOC-051-T01` confirms) | `staging` | not wired; `VOC-051-T01` adds |

All four secrets (plus `SENTRY_API_TOKEN` for `VOC-051-T02`) are already set on
the repository as of this resolution. The DSN **values** are deliberately
absent from this file and must never be written into this file or any other
repository file — they were provided directly to the founder-gate overseer
session and written straight to GitHub Actions secrets, never committed.

## 4. The three org-specific facts (now confirmed)

All three, confirmed 2026-08-08 and recorded in §3:

1. **Plan tier and headroom.** Developer (free) plan confirmed. The founder was
   told explicitly that the 5,000 errors/month cap is pooled account-wide
   across all four projects, not per-project, and accepted that constraint
   knowingly when choosing Layout B. `VOC-051-T02`'s implementer should treat
   quota exhaustion as a real, live operating condition (see §5's existing
   quota-interaction finding), not a theoretical edge case.
2. **Token creation.** Confirmed possible. An internal integration named
   `vocanova-monitoring-agent` was created, scoped to `Issue & Event: Read` +
   `Project: Read` only (no write/admin on any resource), and its token is
   stored as the `SENTRY_API_TOKEN` secret. `VOC-051-TEST-02` should still
   verify the token's actual granted scope against the live API when T02 is
   implemented, per this section's original caution — this record reflects
   what was configured, not an independent re-check.
3. **Layout choice and project inventory.** Layout B, organization slug
   `vocanova`, four project slugs (`prod-api`, `prod-web`, `stage-api`,
   `stage-web`), full per-tier mapping in §3.

## 5. Findings this task raises for the reviewing human

- **The Developer plan's 1 user seat makes a personal token a bad choice** for
  the monitoring workflow's credential even though personal tokens have the most
  flexible scope selection: the token would be bound to the founder's single
  seat. §1b's internal-integration token avoids this. Non-blocking for T00, but
  `VOC-051-T02` should not silently pick a personal token.
- **`apps/api` staging has no Sentry wiring at all today.** `specification.md`
  scope item 1 and `VOC-051-T01` are written around `apps/web`, but the hourly
  workflow (`VOC-051-AC-02`) is required to query "both the staging and
  production Sentry environments." If staging never emits events, that criterion
  is vacuously satisfiable. Adding `STAGING_SENTRY_DSN` to
  `.github/workflows/deploy-staging.yml` for `apps/api` is inside `VOC-051-T01`'s
  stated scope ("Add or update `.github/workflows/deploy-staging.yml`"), but the
  package text frames that task as `apps/web`-only. Recorded as a scope
  clarification for the reviewing human rather than resolved here.
- **Quota interaction with the duplicate-check guard.** Under Layout A on the
  Developer plan, a single bad deploy could consume the whole monthly error quota
  and cause Sentry to drop later events, which the hourly workflow cannot
  distinguish from "no new problems." `VOC-051-R02` covers API *errors* but not
  silent quota-drop. Worth a look when `VOC-051-T02` is implemented.

## 6. Method and limitations

- Repository facts in §1c were read directly from the tree at this branch's tip:
  `apps/api/app/api/production.go`, `apps/api/cmd/api/main.go`,
  `apps/api/.env.example`, `.github/workflows/deploy-production.yml`,
  `.github/workflows/deploy-staging.yml`, and a repository-wide search for
  `SENTRY` under `apps/` and `.github/workflows/`.
- Vendor facts in §1a and §1b were read from Sentry's own published documentation
  (pricing, quotas/spike-protection, account/auth-tokens, api/permissions,
  api/guides/create-auth-token). They are current as of this task's run and are
  not org-specific.
- No Sentry API call was made and no Sentry UI was accessed: no credential for
  the founder's organization exists in this environment.
- No code, workflow, secret, or DOC-11 change is made by this task, per
  `tasks.md`'s "This task makes no code change."
