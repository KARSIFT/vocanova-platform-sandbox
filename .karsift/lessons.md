<!-- karsift lessons: keep this file under ~10KB. When it grows past that,
     move the oldest/least-relevant entries into .karsift/lessons-archive.md
     (create if needed) rather than deleting them - unbounded growth dilutes
     signal for anyone (human or agent) actually reading it. Archive by hand;
     no automated pruning exists. -->

# Lessons learned - vocanova-platform-sandbox

Append-only reference. No longer auto-injected into any pipeline prompt (the
custom plan/implement/review pipeline that did this was retired) - read it
manually, or point an agent at it, when it's relevant to what you're doing.
Seeded and maintained manually by the founder-gate operator - no automated
write-path exists yet (deliberately: unsupervised auto-written lessons were
ruled out to avoid bad lessons being written silently).

## 2026-07-29: Tailwind `max-w-{size}` token collision - ROOT CAUSE STILL UNFIXED

`packages/design-tokens/src/spacing.ts` generates
`apps/web/src/app/tokens.generated.css`'s `@theme static { --spacing-md: 16px;
--spacing-xl: 32px; ... }` block. Tailwind v4 resolves named `max-w-{size}`
container utilities (sm/md/lg/xl/2xl/...) against this same named scale
before the intended `--container-{size}` default (28rem/36rem/etc.) when
both happen to share a key name - this repo doesn't define its own
`--container-*` namespace, so every one of them collides. Confirmed by
inspecting the compiled CSS: `.max-w-md{max-width:var(--spacing-md)}` (16px,
not 448px), `.max-w-xl{max-width:var(--spacing-xl)}` (32px, not 576px).

**Do not use `max-w-md`, `max-w-xl`, `max-w-lg`, `max-w-sm`, `max-w-2xl`, etc.
anywhere in `apps/web` until this is actually fixed.** The current workaround
at every call site that has hit this so far is an explicit arbitrary value
with an inline comment, e.g.:

    <div className="w-full max-w-[28rem] ...">
    {/* max-w-[28rem] (not max-w-md): this repo's tokens.generated.css only
        defines a --spacing-* scale, which shadows Tailwind's intended
        28rem max-w-md container size - see .karsift/lessons.md */}

Known locations already patched this way (as of 2026-07-29):
`apps/web/src/app/auth/magic/_components/magic-link-page-content.tsx`,
`apps/web/src/app/auth/magic/page.tsx`,
`apps/web/src/app/onboarding/page.tsx` (max-w-xl -> max-w-[36rem]),
`apps/web/src/app/signin/page.tsx`,
`apps/web/src/app/(app)/reviews/_components/review-session.tsx`.
If you touch any of these files, keep the arbitrary-value pattern - don't
"simplify" it back to `max-w-md`/`max-w-xl`, it will silently shrink the
container to 16px/32px and was the confirmed root cause of VOC-031-T08's
core-loop test reporting a heading as CSS-hidden (word-wrapped to 0px
measured width under Playwright's visibility check).

If you're implementing a task that touches design tokens or Tailwind config
and can permanently fix this (e.g. by adding an explicit `--container-*`
namespace to `tokens.generated.css`'s generator so `max-w-*` no longer
aliases onto `--spacing-*`), that is the correct real fix - do it if in
scope, and delete this lesson entry once every workaround site above is
reverted to the plain `max-w-{size}` utility and confirmed correct.

## 2026-07-29: Lighthouse `LIGHTHOUSE_CHROME_PATH` needs shell expansion, not `env:` expansion

`.github/workflows/lighthouse.yml`'s "Run Lighthouse suite" step resolves the
Playwright-installed Chromium binary path with a glob
(`~/.cache/ms-playwright/chromium-*/chrome-linux/chrome`) because the exact
revision directory name changes with Playwright version bumps. GitHub
Actions does NOT shell-expand `env:` block values (only `run:` script bodies
are shell-expanded) - so the `ls -d ... | head -1` command substitution that
resolves the real path MUST happen inside the `run:` step body itself
(`export LIGHTHOUSE_CHROME_PATH="$(ls -d ...)"`), not as a literal string in
the step's `env:` block. Putting it in `env:` silently passes the unexpanded
glob pattern as a literal path, and Lighthouse's `chrome-launcher` then fails
to find the binary with `ECONNREFUSED`. Keep this in mind if this workflow's
Chrome-path resolution is ever refactored - the same rule applies to any
similar path resolution added elsewhere in CI.

## 2026-07-29: Lighthouse 12's `emulatedFormFactor` was renamed to `formFactor`

`apps/web/tests/lighthouse/runner.mjs` currently uses `formFactor` correctly
(both per-layout, e.g. `formFactor: "mobile"`, and in the resulting Lighthouse
config object). Older Lighthouse docs/examples (pre-v10ish) use
`emulatedFormFactor` instead - if you're implementing or planning against
web-fetched or memorized Lighthouse config examples, `emulatedFormFactor` is
STALE for the Lighthouse version this repo runs (pinned to v12+ via the
lockfile) and is silently ignored rather than erroring - Lighthouse just
falls back to its own default form factor ("mobile"), defeating the point of
the mobile/desktop layout matrix this test suite deliberately covers. Use
`formFactor` only.

## 2026-07-29: Atlas v1.x's `-- atlas:txmode transaction` directive is INVALID

`apps/api/migrations/*.sql` (VOC-025 through VOC-031, 13 files in total)
each start with the header comment `-- atlas:txmode transaction`. Atlas
v1.x (the only version this repo will plausibly be on, given the
2026-07-29 era) accepts exactly three `atlas:txmode` values: `none`,
`file`, and `all` (and `all` is rejected inside a per-file directive -
it is global-only). The literal value `transaction` is not in that
set; Atlas errors with `unknown txmode "transaction" found in file
directive "<name>.sql"` and `atlas migrate apply` aborts before
running any SQL. Confirmed by downloading Atlas v1.2.0-canary and
applying the existing migration set against a fresh disposable
Postgres 16 - every migration is rejected at the directive-parsing
step, so the entire apply fails. **This means the existing
`apps/api/migrations/*.sql` files cannot be applied by Atlas as
written.** The T06 deliverable (apps/api/atlas.hcl + the
apps/api/scripts/migrate.sh wrapper + apps/api/migrations/atlas.sum)
adds the tooling but does not silently fix the directive - that fix
touches the protected `apps/api/migrations` area, which the VOC-032
package scopes T06 as "read-only, tooling only". The T06 PR's PR
description records the directive bug as a VOC-032-T06 follow-up
needing a separate (small) package or an explicit, narrow exception
in T06's own scope; either way, the apply command in the wrapper
will fail against the current migration set with the above error,
and `atlas migrate apply` is not silently passing this.

If you are touching Atlas tooling or the migration files for any
reason, the fix for the directive is one of:
  1. Change `-- atlas:txmode transaction` to `-- atlas:txmode file`
     in every *.sql file (file = default, so omitting the directive
     entirely also works; deleting the line is the minimum-diff
     change). `file` means "wrap this file in its own transaction",
     which is exactly what the original `transaction` comment was
     trying to express. The T06 wrapper's comment block anticipates
     this fix and explicitly does not override the per-file
     txmode from the CLI.
  2. If a global "all migrations in one transaction" semantic is
     actually wanted, set `txmode = "all"` in the `dev` env block of
     apps/api/atlas.hcl and remove the per-file directives (Atlas
     rejects `atlas:txmode all` inside a file). Today, every file
     has its own BEGIN/COMMIT boundary implicit in the
     `file`-mode default, so option (1) is the safe minimum.

If you're reviewing T06 and see this lesson, the existing migration
files are not Atlas-applicable and that is a pre-existing
incompatibility the package documented as a follow-up rather than
silently editing in scope. Don't assume the migration apply works
in staging until both this directive is fixed and the separate
duplicate-index bug in
`apps/api/migrations/20260725130002_voc030_p4_gamification_tables.sql`
(line 33's `user_id uuid NOT NULL UNIQUE` plus line 47-48's
`CREATE UNIQUE INDEX streak_states_user_id_key` collide - Postgres
auto-creates a unique index named `<table>_<col>_key` for the
inline UNIQUE constraint, and the explicit CREATE UNIQUE INDEX then
errors with `relation "<table>_<col>_key" already exists (42P07)`)
is also resolved. The T06 PR's notes section points at both.

## 2026-07-29: Atlas's default forward-apply file glob is `*.sql`; the `.example` suffix is the load-bearing protection for recovery down-files

`apps/api/migrations/README.md` already states the recovery
`.down.sql.example` files are "deliberately not executable by Atlas
and exist only for disposable recovery rehearsal". The mechanism is
Atlas's default versioned-mode file glob, which is `*.sql` (and
NOT `*.sql` OR `*.down.sql` OR `*.down.sql.example`). Confirmed by
renaming one `.down.sql.example` to `.down.sql` in a temp
directory: Atlas immediately picks it up as a forward migration and
tries to apply it (in this case, `DROP TABLE magic_links;` against
an empty database fails the test, but the point is that the file
was applied at all). The `.example` suffix is what keeps recovery
down-files outside the glob, and
`apps/api/migrations/atlas_tooling_test.go::TestMigrationsDirectoryHasNoForwardDiscoveredDownFiles`
guards the naming convention (fails fast if any file ending in
`.down.sql` - without the `.example` suffix - is ever committed).
If you are adding a recovery down-file, the convention is
`<version-prefix>_<table>.down.sql.example` (note: `.example` is
part of the extension, so the file appears as a single
dot-separated name to most tooling and is sorted distinctly from
the `.sql` and `.down.sql` files in directory listings). Don't
"tidy up" the suffix - it's load-bearing.
