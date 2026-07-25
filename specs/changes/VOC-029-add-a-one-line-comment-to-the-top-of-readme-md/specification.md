# VOC-029 — Add a One-Line Comment to the Top of README.md: Specification

## Objective and requirement source

Add a single one-line comment at the very top of `README.md`, noting the date
the line was added, as a trivial diagnostic test of the planner→package
drafting pipeline. Requirement source: the supplied free-text request, which
explicitly states this is "not a real change" and directs that this package
must not be adopted. There is no product, engineering, or governance document
authorizing a substantive change here, because none is being made.

## Scope and non-goals

In scope: one new line at the top of `README.md`, above the existing
`# Vocanova` heading, recording the date the line was added (e.g. as an HTML
comment so it renders invisibly, consistent with `README.md` being a Markdown
document rendered by GitHub).

Out of scope: any edit to `README.md`'s existing headings, prose, or links; any
edit to any other file; any test, build, or governance script change; adoption
or implementation authorization of this package; any production or user-facing
effect.

## Risk and protected areas

Proposed **R0** — not a determination. `README.md` is a documentation-only
file with no runtime, build, or deployment role in this repository; the path
classifier floors `*.md` paths at `R0`. No protected area
(`docs/governance/protected-areas.md`) is touched.

## Decisions, contradictions, security, and privacy

`VOC-029-D00` — **Open, by design.** The exact comment wording and date format
are left as an implementer-facing detail for whoever eventually implements this
package (if it is ever adopted, which the request explicitly says it should
not be). No security, privacy, secret, or personal-data concern applies: the
change adds no code, no credential, and no user data of any kind.

## Data, migrations, analytics, and accessibility

None. This change touches no database, schema, migration, analytics event, or
user-facing accessibility surface — `README.md` is a repository-root
documentation file consumed by developers and GitHub's file browser, not by
the application.
