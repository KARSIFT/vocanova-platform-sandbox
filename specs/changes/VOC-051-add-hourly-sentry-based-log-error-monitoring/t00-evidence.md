# VOC-051-EV-04 — T00 Sentry organization/plan precondition check

Evidence for `VOC-051-T00` and `VOC-051-TEST-09`.

## Gate status: BLOCKED — `VOC-051-T01` and `VOC-051-T02` must not start

**`VOC-051-DEP-00` is NOT resolved.** This record is the "documented blocking
constraint" branch of `VOC-051-TEST-09`'s expected result, not the "written
confirmation" branch.

Closing this task's GitHub task issue does **not** clear the precondition that
`implementation-plan.md` places on `VOC-051-T01`/`VOC-051-T02` ("both depend on
the confirmed Sentry project/DSN layout"). The gate lives in this package's own
state, not in the task issue: `tasks.md` records `VOC-051-T00` as `blocked`, and
records `VOC-051-T01`/`VOC-051-T02` as blocked on it. Whoever dispatches T01 or
T02 must first see this file replaced by a completed layout record (§3 below,
with its unknown cells filled by a human) and the corresponding `tasks.md` and
`change.yaml` states flipped.

The blocking constraint is the one named in §4: three org-specific facts that
require signed-in access to the founder's Sentry organization, which this
workspace does not have (no Sentry API token, no Sentry session, no GitHub
Actions secret access — verified: no `SENTRY_*` value is present in this
environment, and `AGENTS.md`'s "Safety" section says agents do not receive
production secrets). Nothing in this repository can substitute for that access,
and this task will not infer the answers from source code.

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
its justification in that task's pull request. It is recorded here rather than
acted on because §4's confirmation has not happened.

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

Neither is chosen here. Under the Developer plan's shared 5,000 errors/month
account-wide cap, Layout B's isolation benefit is partly illusory (the cap is
per-account, not per-project), which is a point for Layout A; but that trade-off
is a human decision about the founder's real org, not this task's to make.

## 3. Layout record — to be completed by the human with Sentry access

Everything below except the four bracketed unknowns is fixed by this repository's
existing code and workflows and does not need confirming. Fill the brackets, or
replace this table wholesale if Layout B is chosen, and this section becomes the
written confirmation `VOC-051-TEST-09` asks for.

| App | Tier | Sentry project slug | GitHub Actions secret holding the DSN | Runtime env var | `SENTRY_ENVIRONMENT` value | Status |
| --- | --- | --- | --- | --- | --- | --- |
| `apps/api` | production | `[unknown — confirm]` | `PRODUCTION_SENTRY_DSN` (exists) | `SENTRY_DSN` | `production` | wired today |
| `apps/api` | staging | `[unknown — confirm]` | `STAGING_SENTRY_DSN` (does not exist yet) | `SENTRY_DSN` | `staging` | not wired; `VOC-051-T01` adds |
| `apps/web` | production | `[unknown — confirm]` | `PRODUCTION_WEB_SENTRY_DSN` (proposed) | `NEXT_PUBLIC_SENTRY_DSN` (proposed) | `production` | not wired; `VOC-051-T01` adds |
| `apps/web` | staging | `[unknown — confirm]` | `STAGING_WEB_SENTRY_DSN` (proposed) | `NEXT_PUBLIC_SENTRY_DSN` (proposed) | `staging` | not wired; `VOC-051-T01` adds |

Secret and variable names in the "proposed" rows follow the existing
`PRODUCTION_SENTRY_DSN` / `PRODUCTION_WEB_HOST` naming convention in
`.github/workflows/deploy-production.yml`; `VOC-051-T01` owns confirming them,
including the `NEXT_PUBLIC_` prefix question that task already raises.

The DSN **values** are deliberately absent and must never be written into this
file or any other repository file — they belong in GitHub Actions secrets only.
Recording the project slug and the mapping is what T00 needs; the secret material
is not.

Also to be recorded here at the same time: the Sentry **organization slug**
(`[unknown — confirm]`), needed by `VOC-051-T02`'s API calls.

## 4. The three org-specific facts still required

A human with signed-in access to the founder's Sentry organization must confirm
and record, in §3 above:

1. **Plan tier and headroom.** Which plan the organization is on, and — if
   Developer — whether the 5,000 errors/month cap and 1 user seat are acceptable
   once `apps/web`'s browser errors start flowing on top of `apps/api`'s. Browser
   error volume is typically higher and noisier than server-side volume, and
   under `VOC-051-R01` every distinct issue can become a GitHub issue and thence
   a `plan-from-issue` run, so quota exhaustion here degrades the monitoring this
   package exists to add.
2. **Token creation.** That an internal integration with exactly `project:read` +
   `event:read` (§1b) can be created in this organization, and its actual granted
   scope list once created — `VOC-051-TEST-02` checks the real configured scope,
   not the intent.
3. **Layout choice and project inventory.** Layout A or Layout B (§2), the
   organization slug, the existing project slug(s), and the resulting per-tier
   mapping filled into §3.

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
