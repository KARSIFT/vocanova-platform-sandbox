# Vocanova

Vocanova helps A2–B1 English learners discover vocabulary in real situations,
save useful words, review them, and practise original sentences with AI feedback.
The responsive web product is a Next.js and Go monorepo maintained by KARSIFT.

The application includes daily missions, Journey discovery and word detail,
saved vocabulary, spaced reviews, sentence practice and history, progress,
onboarding, and account settings. Home, Journey, and Progress are the three
primary destinations. Sign-in methods depend on environment configuration;
implemented password flows do not mean email is enabled live.

Start with the [current product state](docs/product/current-state.md) for the
implemented surfaces, dated deployment observation, and remaining readiness work.
The separate ChatGPT Site prototype is design input, not the Next.js/Go app.

## Work locally

Use the exact tools and prerequisites in the [development guide](docs/development.md):

```bash
corepack enable
pnpm install --frozen-lockfile
pnpm dev
```

`pnpm validate` runs the repository validation sequence. Browser checks and
database integration tests have additional prerequisites documented in that guide.

## Repository map

| Path                      | Purpose                                                       |
| ------------------------- | ------------------------------------------------------------- |
| `apps/web/`               | Next.js public, authentication, learning, and account screens |
| `apps/api/`               | Go services, PostgreSQL schemas, and migrations               |
| `packages/api-client/`    | Shared typed API client                                       |
| `packages/design-tokens/` | Shared visual tokens                                          |
| `infra/`                  | Deployment, shared edge, monitoring, and operational scripts  |
| `docs/`                   | Product, design, engineering, and operational guidance        |
| `specs/changes/`          | Historical implementation packages and evidence               |
| `.agents/`                | Canonical repository skills and agents                        |

## Delivery and documentation

[AGENTS.md](AGENTS.md) defines the current workflow: PRs target `main`, required
CI and the merge queue gate merging, and a draft or `hold` label keeps a PR from
auto-merging. Automated review is advisory. Staging deploys from pushes to
`main`; production requires a separate manual dispatch of
[`deploy-production.yml`](.github/workflows/deploy-production.yml).

- [Documentation index and source authority](docs/README.md)
- [Release-readiness plan](docs/product/release-readiness.md)
- [Product bible](docs/product/00-product-bible.md) and [MVP PRD](docs/product/01-mvp-prd.md)
- [Learning workspace design](docs/design/learning-workspace.md)
- [Account email activation and release identity](docs/development/account-and-release-operations.md)
- [Repository agent skills](docs/development/agent-skills.md)
- [Architecture decisions](docs/decisions/README.md)
- [Security policy](SECURITY.md)

Older documents preserve delivery history. Their adoption metadata does not
reinstate the retired plan/adopt/roster pipeline or former `develop` branch.
Use current instructions, executable workflows, and source code when historical
documents disagree with today's process.
