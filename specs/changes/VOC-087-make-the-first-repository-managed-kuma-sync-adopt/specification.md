# VOC-087 — Make the first repository-managed Kuma sync adopt live monitors safely: Specification

## Objective and requirement source

Close the first-apply safety gap reported in
[GitHub issue #728](https://github.com/KARSIFT/vocanova-platform-sandbox/issues/728):
the merged VOC-086 inventory and synchronizer cannot safely adopt the live
production monitors, can strip notification bindings on update, records
stale T01/T02 evidence, does not execute documented shell harnesses in the
normal test entry point, and can strand credential rotation after a
successful host reset if proof transfer fails.

**This draft package still does not adopt or authorize itself**; adoption
remains a separate A-004 plan-review / adopt path.

Primary context (issue #728 + drafting-time repo read):

| Item | Value |
|------|--------|
| Live API monitor (2026-08-19) | name `VocaNova Production API`; URL `https://api-production.vocanova.site/healthz` |
| Live web monitor (2026-08-19) | name `VocaNova Production Web`; URL `https://production.vocanova.site` (no trailing slash) |
| Inventory adoption names | `Production API /healthz`; `Production Web` |
| Inventory production web URL | `https://production.vocanova.site/` (trailing slash) |
| Adoption matcher | exact `name` and `url` string equality in `plan.mjs` `adoptionMatches` |
| Notification payload | `inventoryEntryToDesiredMonitor` sets `notificationIDList: {}` |
| VOC-086 T01 evidence | records inventory values `Production Web` / `https://production.vocanova.site/` and `Production API /healthz` / `https://api-production.vocanova.site/healthz` as live adoption keys |
| VOC-086 T02 evidence | claims password is stored from the runner-local file *before* host→runner fetch; merged workflow stores only after proof fetch, then always scrubs |
| Shell harnesses | `infra/scripts/kuma-rotate-credentials.selftest.sh` and `infra/scripts/sync-kuma-inventory.selftest.sh` exist; `pnpm test` does not execute them |
| First live apply | must not run until this package is merged and deployed |

**Objective:** after this package's implementation, planning against the
exact two live monitor records produces adoption/update operations rather
than collisions or duplicate creates; a second sync is a no-op; remote
notification bindings survive adoption and later managed updates unless
inventory explicitly takes ownership; stale VOC-086 T01/T02 evidence
describes the merged implementation and real live preconditions; both shell
harnesses and Node tests run in CI; and credential rotation remains
recoverable after reset-success/proof-transfer-failure without exposing
credentials.

## Confirmed findings (issue #728 + drafting-time re-read)

- `infra/monitoring/monitors.yaml` production entries use
  `name: Production Web` / `Production API /healthz` and adoption
  `match_name`/`match_url` of those same incorrect live identities. Production
  web `url` and `match_url` include a trailing slash the live monitor does
  not have.
- `plan.mjs` `adoptionMatches` compares `monitor.name` to `adoption.match_name`
  and `monitor.url` to `adoption.match_url` with exact `String(...) ===`
  equality. Unmanaged URL collision uses the same exact URL equality against
  `entry.url`, not the live URL and not a shared normalizer.
- Consequence for the API monitor: name mismatch prevents adoption; URL
  matches exactly, so unmanaged URL collision fails closed.
- Consequence for the web monitor: name mismatch prevents adoption; URL
  also mismatches on trailing slash, so exact URL collision misses and the
  planner can emit a **create** that duplicates the live site.
- `monitor-payload.mjs` always sets `notificationIDList: {}` on the desired
  monitor. `EDITABLE_MONITOR_FIELDS` includes `notificationIDList`, so
  `buildEditPayload` / `snapshotMonitorForRollback` will send that empty
  object on adopt/update.
- `scripts/foundation/voc086-kuma-sync.test.mjs` VOC-086-TEST-04 seeds
  adoption fixtures from `productionWeb.name` / `productionWeb.url` (the
  inventory values), not the 2026-08-19 live names/URLs, so current tests
  cannot catch this defect.
- `pnpm test` runs `node --test scripts/foundation/*.test.mjs`. The
  workflow test only `existsSync`s the two selftest scripts.
- Merged `.github/workflows/sync-monitoring.yml` fetches reset proof with
  `always()`, stores `KUMA_PASSWORD` only when this-attempt proof is
  present, then always scrubs host and runner password files. If host reset
  succeeded and scp of proof fails, the new password can be lost.

## Scope and non-goals

In scope:

1. Align production inventory `name`, `url`, `adoption.match_name`, and
   `adoption.match_url` for
   `kuma.availability.production.web` and
   `kuma.availability.production.api-healthz` to the verified live identity.
2. Shared URL comparison used by both adoption matching and unmanaged-URL
   collision, covering at least trailing-slash equivalence for HTTP(S) URLs,
   without fuzzy name matching or host/path weakening.
3. Preserve remote `notificationIDList` on adopt and later managed updates
   unless inventory explicitly takes ownership. Canonical `monitors.yaml`
   remains preserve-by-default.
4. Deterministic tests using the exact live monitor names/URLs; update
   payload tests proving notification bindings survive; second sync no-op;
   collision safety retained for distinct URLs and name mismatches.
5. Correct VOC-086 `t01-evidence.md` and `t02-evidence.md` so they describe
   the merged implementation and real live preconditions, without secrets.
6. Execute `kuma-rotate-credentials.selftest.sh` and
   `sync-kuma-inventory.selftest.sh` from the normal deterministic test
   entry point (`pnpm test` / `node --test scripts/foundation/*.test.mjs`
   or an equivalent CI path that already runs on every PR).
7. Address or fail safely on reset-success/proof-transfer-failure so a
   credential rotation remains recoverable without exposing credentials and
   without blindly resetting a second time.
8. Keep the operation idempotent and fail closed on ambiguous adoption or
   partial application. No SQLite. Preserve unrelated manual monitors.

Non-goals / explicitly excluded:

- Dispatching or claiming the first live `sync-monitoring` inventory apply.
- Direct SQLite read/write.
- Cloudflare configuration changes.
- Changing Sentry/`error-monitoring.yml`.
- Application schema migrations or real-user mutation.
- Redesigning scheduled synthetics.
- Managing a full notification-provider catalog in inventory unless the
  optional explicit-ownership field from `VOC-087-DEP-03` is used.
- Self-adoption / self-authorization of this package.

## Risk and protected areas

- **Builder/issue proposal:** none stated in issue #728.
- **Measured path floor:** **R3** for `infra/monitoring/*`, `infra/scripts/*`,
  and `.github/workflows/sync-monitoring.yml`. Not R4 unless a task actually
  touches `scripts/governance/*` or another R4 path.
- **Draft package proposal:** **R3** (highest of stated builder class and
  measured floor). This is a draft proposal for the reviewing human at
  adoption time, never a determination. Semantic first-live-apply and
  credential-recovery consequences may still be raised to R4 by the
  independent verifier or reviewing human.
- Protected areas: `.github/workflows/sync-monitoring.yml`, live Kuma
  inventory mutation (deferred), monitoring credentials, VOC-081 topology
  invariants.
- EHR: not triggered.
- Active authority model: **A-004**. No founder `approved` comment is a
  merge/adopt/release gate.

## Decisions

`VOC-087-D00`: The 2026-08-19 verified live production monitor identity is
authoritative for first adoption:

- `kuma.availability.production.api-healthz`: name and `match_name`
  `VocaNova Production API`; URL and `match_url`
  `https://api-production.vocanova.site/healthz`
- `kuma.availability.production.web`: name and `match_name`
  `VocaNova Production Web`; URL and `match_url`
  `https://production.vocanova.site`

Aligning managed `name`/`url` to the live identity avoids a first-apply
rename or URL rewrite of the live monitors. Staging and monitor-host
entries are not live-adopted and need not be renamed.

`VOC-087-D01`: Adoption still requires exact name equality plus URL equality
after a single shared normalizer used for both adoption matching and
unmanaged-URL collision. Required URL equivalence: HTTP(S) trailing-slash
differences on an otherwise identical URL. Do not add substring, host-rewrite,
or scheme-dropping matching. If implementer evidence shows additional Kuma
storage normalization, record it in T00 evidence and apply it only in that
same shared function.

`VOC-087-D02`: `notificationIDList` is not owned by the canonical inventory
by default. On adopt/update, copy the remote monitor's existing notification
bindings into the edit payload. Do not send `notificationIDList: {}` as a
default desired value. `monitorsMatch` must not treat preserved notification
bindings as drift. If inventory explicitly takes ownership (shape per
`VOC-087-DEP-03`), the inventory value is sent instead.

`VOC-087-D03`: This package must not perform or claim the first live
`sync-monitoring` inventory apply. VOC-086-T05 / operator dispatch remains
blocked until this package is merged and deployed.

`VOC-087-D04`: After a successful host password reset for this attempt, the
generated password must remain recoverable until it is stored in GitHub
`monitoring` `KUMA_PASSWORD`. Proof-of-reset remains required when available
so an unused generated password is not stored over a secret when reset did
not happen. If reset succeeded and proof/metadata transfer fails, the
workflow must fail closed, must not log credentials, must not invoke
`reset-password.js` again, and must not scrub the last remaining copy of the
new password before a recoverable store path exists.

`VOC-087-D05`: `infra/scripts/kuma-rotate-credentials.selftest.sh` and
`infra/scripts/sync-kuma-inventory.selftest.sh` must actually execute in the
normal deterministic CI/test entry point, not merely exist.

`VOC-087-D06`: VOC-086 `t01-evidence.md` and `t02-evidence.md` must be
corrected in the tasks that change the corresponding behavior so they
describe the merged implementation and the real live preconditions, without
credentials or notification configuration.

`VOC-087-D07`: No SQLite. Fail closed on ambiguous adoption, duplicate
candidates, or partial apply. Preserve unrelated manually owned monitors.
Never log Kuma credentials or notification configuration.

`VOC-087-D08`: This package's `monitoring_impact.state` is `update` because
it corrects adoption identity and update payload behavior for existing
VOC-086 availability monitors and does not add new stable IDs.

## Contradictions / open questions

1. **Rotation recovery shape** (`VOC-087-DEP-02`): implementer may retain the
   runner-local password until secret store succeeds, store from the
   runner-local file after rotate-step success even if proof fetch fails,
   and/or add a store-only retry that never resets again. AC-05 is the
   requirement; do not guess a second live reset.
2. **Notification ownership schema** (`VOC-087-DEP-03`): issue #728 requires
   preservation unless inventory explicitly takes ownership. Canonical
   `monitors.yaml` must stay preserve-by-default. Exact reserved field name
   is an implementer choice if tests cover both default preserve and
   explicit override.
3. **Risk class:** measured path floor is R3. Semantic live-apply and
   credential recovery may be raised to R4 by the reviewing human or
   verifier. Do not silently touch `scripts/governance/*` to force an R4
   path floor.
4. **Additional URL normalization:** only trailing-slash equivalence is
   required from verified evidence. Further Kuma storage rewrites are an
   evidence-backed addition in T00, not a guess at drafting time.

## Security and privacy

- Credentials only in GitHub secrets / workflow environment; never committed
  and never printed in logs, issues, or evidence.
- Notification configuration is sensitive operational data; tests may use
  synthetic numeric IDs, never live provider secrets or destination details.
- No SQLite access from deploy/sync tooling.
- Preserve Kuma authentication as the admin boundary.
- Rotation recovery must not create a new log/artifact that contains the
  generated password.

## Data, migrations, analytics, and accessibility

- **Data/migrations:** None for application databases. Kuma monitor
  definitions change only via supported Socket.IO after this package is
  merged and deployed and a later authorized sync runs. Credential rotation
  recovery does not change application data.
- **Analytics:** None expected — evidence-backed non-applicability.
- **Accessibility:** No product UI redesign. Operator evidence/docs only.
