# VOC-085 — Impact Analysis

## Security and privacy

- **Synthetic-only verification:** gates use the reserved synthetic smoke
  identity and the workflow-minted session. No magic-link sends, OAuth
  completion, real-user mutation, or state-changing learning actions.
- **No privileged real-user reads:** verification must not inspect or alter
  real user rows via shortcuts outside the normal authenticated synthetic
  path.
- **Secrets:** reuse existing production deploy secret plumbing. Do not add
  new production secrets for this package. Never log `DATABASE_URL`
  credentials, mint tokens, or session opaque values beyond existing redaction
  conventions.
- **Content writes:** limited to repository-owned canonical P1 seed records
  via the existing idempotent upsert tool. Do not introduce destructive
  truncate/reseed paths.
- **Isolation:** preserve staging/production secret files, directories, deploy
  users, databases, and Docker networks under the single shared-edge nginx
  architecture.

## Data and migrations

- No schema migration.
- Additive/upsert content seed via existing `apps/api/cmd/seed` /
  `voc026-p1.json` against the production-private database during deploy,
  after migrations and the synthetic-user seed, before application
  convergence.
- Synthetic-user SQL already refreshes `onboarding_status='completed'`
  idempotently; this package relies on that rather than manual DB edits.
- Rollback of workflow/check revisions is a repository revert + redeploy.
  Canonical seed rows are preserved rather than destructively removed on
  rollback (content remains available; gates revert to prior strictness if a
  check revision is reverted).

## Analytics and accessibility

- **Analytics:** None expected — evidence-backed non-applicability for product
  analytics instrumentation.
- **Accessibility:** No UI redesign. Route sweep is functional reachability
  under authenticated synthetic session; do not regress existing a11y while
  adding checks. Preserve any `.karsift/lessons.md` layout workarounds if
  pages are touched (they should not be for this package's primary path).

## Risks, dependencies, and evidence

- `VOC-085-R00`: **Empty Discover remains deploy-green** if smoke still
  accepts HTTP 200 with `[]`. Mitigation: AC-03/TEST-04 fail-closed empty
  body.
- `VOC-085-R01`: **Seed failure after partial host mutation** if seed runs
  after `up -d` or with `continue-on-error`. Mitigation: AC-00/AC-01; run
  before convergence; fail closed.
- `VOC-085-R02`: **Duplicate or destructive content** from a non-idempotent
  or redesigned seed. Mitigation: reuse existing upsert tool; AC-02/TEST-03;
  no dataset replacement.
- `VOC-085-R03`: **Real-user impact** from verification shortcuts or learning
  mutations. Mitigation: AC-05/AC-06; synthetic session only; read-only
  checks.
- `VOC-085-R04`: **False green route sweep** that skips dynamic discover
  routes or drops auth cookie. Mitigation: AC-05/TEST-06/TEST-07.
- `VOC-085-R05`: **Isolation/topology regression** while editing production
  deploy. Mitigation: AC-08/TEST-10; VOC-067 invariants.
- `VOC-085-R06`: **Over-claiming monitoring coverage** by inventing naive
  unauthenticated page monitors. Mitigation: out of scope; later
  monitoring-inventory package owns stable IDs.
- `VOC-085-DEP-00`: Root cause resolved at drafting (deferred production seed
  + status-only smoke).
- `VOC-085-DEP-01`: Reuse existing seed tool (resolved).
- `VOC-085-DEP-02`: Route-sweep harness shape (open implementer choice;
  coverage fixed).
- `VOC-085-EV-00`: T00 seed bundling/order/fail-closed evidence
  (`t00-evidence.md`).
- `VOC-085-EV-01`: T01 content-aware smoke evidence (`t01-evidence.md`).
- `VOC-085-EV-02`: T02 route sweep + live Cloudflare/isolation evidence
  (`t02-evidence.md`).

## Monitoring impact

Update existing production deploy/smoke synthetic coverage in this package.
The later monitoring-inventory package will adopt it under a stable
repository synthetic-check ID and governance mapping; do not create naive
unauthenticated page monitors here.
