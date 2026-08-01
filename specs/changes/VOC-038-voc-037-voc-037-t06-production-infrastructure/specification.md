# VOC-038 — VOC-037-T06: Production Infrastructure Provisioning: Specification

## Objective and requirement source

Execute `VOC-037-T06` (dispatched via issue
[#269](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/269)): build the
actual production target that `VOC-037-D00` (production hosting/deploy-target
decision, accepted 2026-08-01, Option A-modified) and `VOC-037-D01` (production
secrets management decision, accepted 2026-08-01, corrected mechanism 4A) already
decided, but which no task in `VOC-037`'s original `T00`-`T05` roster actually
built — `T00` and `T01` were scoped, by their own explicit non-goals, as
decision-record tasks only ("No infrastructure provisioning in this task", "does
not create real credentials, write secret values, provision infrastructure, or
deploy production"). `VOC-037-T03` (kill-switch/rollback verification) and
`VOC-037-T04` (monitoring/alerting readiness) both list `VOC-037-T06` as a
dependency in `VOC-037`'s own `tasks.md`, because they need a real production
target to verify against, not just a decision record.

This is a follow-up task added after `VOC-037`'s adoption (founder-gate delegate,
2026-08-01), per issue #269's framing: "it does not make new decisions" — it
executes `VOC-037-D00`/`VOC-037-D01`'s already-accepted decisions.

## Scope and non-goals

In scope:

1. A `/opt/vocanova/production/` directory tree on the shared host, fully separate
   from staging's existing `/opt/vocanova/infra/` tree — per `VOC-037-D01` section
   4A's corrected mechanism.
2. A dedicated, least-privilege production deploy OS user, distinct from the OS
   user staging's deploy path uses.
3. A production Compose file (e.g. `infra/docker-compose.production.yml`) defining
   its own `vocanova-production` Compose project (`-p`/`COMPOSE_PROJECT_NAME`),
   reusing the existing four-service shape (`postgres`, `api`, `web`, `nginx`) with
   explicit per-service resource limits (`mem_limit`/`cpus` or equivalent) to
   mitigate the shared-host CPU/RAM contention `VOC-037-D00` explicitly
   acknowledged.
4. A `production` GitHub Actions environment with founder-controlled required
   reviewers, scoping production secrets at environment level (never
   repository-global), per `VOC-037-D01`'s INV-2.
5. A new `.github/workflows/deploy-production.yml` workflow that declares
   `environment: production`, writes secret files only under
   `/opt/vocanova/production/secrets/`, and operates only on the
   `vocanova-production` Compose project — never touching staging's directory
   tree, deploy user, Compose project, or the existing `STAGING_SSH_*` secrets.
6. Executing `VOC-037-D01`'s `INS-9` through `INS-11` negative-access rehearsal
   (staging's deploy path/user cannot read `/opt/vocanova/production/secrets/`)
   and recording `VOC-037-EV-01` alongside this package's own `VOC-038-EV-00`.
7. Documenting the production secret-path, Compose-project, and resource-limit
   conventions in `infra/README.md`, and adding a production-target amendment
   note to `docs/operations/11-devops-and-ci-cd.md`.
8. Variable-name-only updates to `apps/api/.env.example`/`apps/web/.env.example`,
   if and only if a new variable name is needed to distinguish production from
   staging configuration — no real secret value is ever added to either file.

Non-goals (explicitly excluded from this package):

- Choosing or purchasing a final production domain/DNS name. `VOC-037-D00`'s
  hosting decision record states final names are "founder-confirmed during
  implementation"; this package uses an explicit placeholder hostname (see open
  question 1 below) rather than guessing a real one.
- Any change to `VOC-037-D00` or `VOC-037-D01` themselves. Both are already
  founder-accepted; this package executes them and does not reopen either
  decision.
- `VOC-037-T03` (kill-switch/rollback verification) and `VOC-037-T04`
  (monitoring/alerting readiness) themselves — this package only unblocks them by
  making the production target they need to verify against exist.
- `VOC-037-T05` (R2 release PR and founder go/no-go record).
- Publishing, drafting, or referencing any legal/privacy content (`VOC-037-T02`'s
  scope).
- Provisioning real production third-party credentials (AI-provider keys,
  email-provider keys, Google OAuth client secret, Sentry DSN) — this package
  creates the storage/injection mechanism only; the follow-up of actually
  obtaining and entering those real values is a separate, non-repository action
  outside this package's authority per AGENTS.md's "Agents do not receive
  production secrets and do not deploy directly to production."
- Migrating production off the shared host to a dedicated host. `VOC-037-D00`
  requires the layout to keep that migration a "clean cut" later, not perform it
  now.

## Risk and protected areas

Builder assessment: this package's known target areas —
`.github/workflows/deploy-production.yml` (new), `infra/` (compose and secrets-path
documentation), and a GitHub Actions `production` environment — match
`docs/governance/protected-areas.md`'s "Production infrastructure"
(`/infrastructure/`, `/infra/`, environment configuration) and "Deployment and
rollback" (`/.github/workflows/`, deploy scripts) rows, both of which "create an R3
floor" per that document's own opening line. `docs/governance/change-risk-
classification.md`'s R3 row explicitly lists "production infrastructure ... CI/CD"
as R3.

This package does not, on its own, reach R4 under that same document's
classification-test question 1 ("Does this decide an R4 matter, commit the company
publicly or financially, change user rights/trust, or authorize the initial/major
launch?") — because the R4-grade decisions this task depends on (production
architecture shape, secrets mechanism) were already made and accepted as R4/R3
decisions respectively by `VOC-037-D00`/`VOC-037-D01`. This package's own tasks
execute those decisions rather than deciding a new R4 matter. If, during
implementation, a task's actual diff or a discovered fact (e.g. the shared-host
resource-limit values chosen, or the production hostname once confirmed) turns out
to have a materially different production/user-trust effect than anticipated here,
the implementer must re-run `scripts/governance/classify-change-risk.sh` against
the real diff and raise the class if the floor is higher — this specification's R3
proposal does not bind that later, path-based result.

Protected effects flagged for reviewer attention:

- The `production` GitHub Actions environment's required-reviewers list is itself
  a security control; misconfiguring it (e.g. no required reviewer, or a reviewer
  who is not the founder) would silently weaken `VOC-037-D01`'s INV-2. Independent
  verification must check the environment's protection-rule configuration, not
  just that the environment exists.
- Because production and staging share one physical host (`VOC-037-D00`'s
  acknowledged condition), any implementation step that touches staging's existing
  `/opt/vocanova/infra/` tree, `docker-compose.yml`, or deploy workflow to add
  production is out of scope and a specification violation — the whole point of
  `VOC-037-D01`'s corrected 4A mechanism is that production gets its own tree, not
  a shared one.

## Decisions, contradictions, security, and privacy

No new decision (`VOC-038-D00` or similarly numbered) is defined by this package.
This package's task, `VOC-038-T00` in `tasks.md`, executes `VOC-037-D00` and
`VOC-037-D01` exactly as accepted; see those decision records for the authorization
this package relies on.

Contradictions found: none. `VOC-037-D00` and `VOC-037-D01` are internally
consistent with each other (the "Relationship to VOC-037-D00" section of
`t01-production-secrets-decision-record.md` already reconciles a drafting-order
inconsistency from `VOC-037`'s own history), and issue #269's framing of this task
is consistent with both.

Security and privacy:

- Authorization: the `production` GitHub Actions environment's required-reviewer
  gate is the sole control-plane authorization boundary for production deploys.
  This package must not weaken it (e.g. by adding a bypass workflow trigger, or by
  scoping the environment to a branch pattern broader than intended).
- Secrets: this package creates the mechanism (directory tree, permissions,
  Compose project isolation, environment scoping) but writes zero real secret
  values anywhere in the repository, consistent with `VOC-037-D01`'s INV-3. Any
  `.env.example` change is variable-name documentation only.
- Abuse impact: an incorrectly-scoped `deploy-production.yml` trigger (e.g. a
  `pull_request`- or `pull_request_target`-triggered job with `environment:
  production` declared) would let an untrusted PR reach the production
  environment's secrets. `VOC-037-D01`'s `INS-3` finding already confirms the
  existing `deploy-staging.yml` avoids this pattern; the new workflow must be
  built the same way (`push`/`workflow_dispatch` only) and independent
  verification must re-check this explicitly for the new file.

## Data, migrations, analytics, and accessibility

- Data/migrations: None. This package provisions infrastructure and does not
  create, migrate, or touch application data or database schema. The production
  Postgres container is a fresh instance per `VOC-037-D00`'s isolation requirement
  ("separate database (own Postgres instance/container, not a shared instance or
  schema)"); no data migrates into it as part of this package.
- Analytics: None. This package does not add, change, or remove any analytics
  instrumentation.
- Accessibility: Not applicable. This package has no user-facing UI surface; it is
  infrastructure/CI/CD tooling only.

## Open questions for human review

1. **Production hostname/DNS placeholder.** `VOC-037-D00`'s decision record states
   final production domain/hostname is "founder-confirmed during implementation."
   This package's tasks use an explicit placeholder (`production.vocanova.internal`
   suggested, pending founder confirmation of the real public hostname) rather than
   guessing a real domain name, and flag the real value as a required founder
   confirmation before `deploy-production.yml`'s health checks or any DNS/Cloudflare
   configuration are finalized. This is not something this planning pass can
   resolve; it is recorded here for the human reviewer and for `VOC-038-T00`'s
   implementer.
2. **Shared-host resource-limit values.** `VOC-037-D00` requires explicit
   per-service resource limits to mitigate CPU/RAM contention between staging and
   production on the shared 2-vCPU/4GB host, but does not itself specify the exact
   numeric limits (e.g. how much memory each of `postgres`/`api`/`web`/`nginx`
   should be capped at, in either tier). This package's `tasks.md` proposes that
   the implementer size these based on the host's actual current staging usage
   (inspected at implementation time) plus a safety margin, and record the chosen
   values and their rationale in `infra/README.md` — but the exact numbers are an
   implementer/reviewer-time decision, not one this drafting pass can responsibly
   invent without inspecting live host metrics it does not have access to.
