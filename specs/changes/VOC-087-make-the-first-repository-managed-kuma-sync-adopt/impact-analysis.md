# VOC-087 — Impact Analysis

## Security and privacy

- **Monitoring credentials:** this package changes credential-*rotation
  recovery*, not the existence of `KUMA_USERNAME` / `KUMA_PASSWORD`. The
  reset-success/proof-transfer-failure path must not log, commit, or issue-
  comment the generated password, and must not scrub the last remaining copy
  before a recoverable store exists.
- **Notification bindings:** live monitors may already have notification
  destinations. Sending `notificationIDList: {}` is a silent alert-disable.
  Default preserve; tests use synthetic numeric IDs only.
- **No SQLite path:** synchronizer and deploy/sync tooling must not read or
  write Kuma SQLite. Supported Socket.IO and the official password-reset
  tool remain the only mutation paths.
- **Manual monitor preservation:** ownership marker plus explicit adoption
  remain required. URL normalization must not adopt or overwrite an
  unrelated monitor.
- **Isolation:** preserve staging/production secret files, directories,
  deploy users, databases, Docker networks, loopback-only Kuma `3001`,
  single shared-edge nginx, and absence of 8081/8443.
- **Live apply:** this package must not itself mutate live Kuma inventory.
  First apply is deferred until after merge and deploy.

## Data and migrations

- No application schema migration.
- Kuma monitor definitions change via Socket.IO adopt/update only after a
  later authorized sync. This package changes repository inventory,
  matcher, payload, workflow recovery, tests, and evidence.
- Rollback: revert the responsible task commit(s); do not claim a
  destructive wipe of Kuma data; preserve manually owned monitors; rotate
  credentials again only if compromise is suspected (explicit input), never
  as a silent recovery from proof-transfer failure.

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability for
  product analytics instrumentation.
- **Accessibility:** No product UI redesign. Operator documentation and
  evidence only.

## Risks, dependencies, and evidence

- `VOC-087-R00`: **First live apply collides or duplicate-creates** because
  names/URLs still do not match live identity, or because trailing-slash
  mismatch bypasses both adoption and collision. Mitigation:
  AC-00/AC-01/TEST-00–TEST-03; align inventory to `VOC-087-D00`; shared URL
  normalizer.
- `VOC-087-R01`: **Notification bindings stripped** on adopt/update by empty
  `notificationIDList`. Mitigation: AC-02/TEST-06–TEST-08.
- `VOC-087-R02`: **URL normalization weakens collision safety** (substring or
  host-fuzzy matching). Mitigation: AC-01/TEST-02; exact names; shared
  normalizer limited to documented equivalences.
- `VOC-087-R03`: **Credential lockout** if host reset succeeds, proof
  transfer fails, and scrub deletes the only password copy. Mitigation:
  AC-05/TEST-10; no second reset; no credential logs.
- `VOC-087-R04`: **False CI confidence** because shell harnesses are only
  documented/`existsSync`d. Mitigation: AC-04/TEST-09 execute them.
- `VOC-087-R05`: **Stale VOC-086 evidence** continues to publish wrong live
  names or obsolete rotation ordering. Mitigation: AC-06/TEST-13.
- `VOC-087-R06`: **Live apply before this package deploys.** Mitigation:
  AC-08; no live dispatch in these tasks.
- `VOC-087-R07`: **Risk under-declaration** if semantic live-apply and
  credential recovery warrant R4 while the path floor is R3. Mitigation:
  draft proposes R3 from measured floor and flags verifier/human escalation
  in specification.md; this is a proposal, not a determination.

Dependencies: see `change.yaml` `VOC-087-DEP-00`–`DEP-03`, VOC-086, VOC-081.

Evidence IDs:

- `VOC-087-EV-00` — T00 live adoption identity (`t00-evidence.md`)
- `VOC-087-EV-01` — T01 notification preservation (`t01-evidence.md`)
- `VOC-087-EV-02` — T02 rotation recovery + CI harnesses (`t02-evidence.md`)

## Monitoring impact (this package)

`update`: corrects adoption identity for the two production availability
monitors and preserves notification bindings on managed updates of every
availability monitor. Declared in `change.yaml` under `monitoring_impact`
with those existing VOC-086 stable IDs. No new IDs.
