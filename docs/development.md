# Local development

## Prerequisites

- Node.js `24.18.0` (LTS), declared in `.nvmrc` and `package.json`.
- pnpm `11.14.0`, declared by the root `packageManager` field.
- Go `1.26.5`, declared by `apps/api/go.mod` (`go1.26.0` language level and
  `go1.26.5` toolchain).

From a clean checkout, enable Corepack and install the exact frozen dependency graph:

```bash
corepack enable
corepack prepare pnpm@11.14.0 --activate
pnpm install --frozen-lockfile
```

Go downloads the declared toolchain when needed. This requires ordinary access to the
official Go toolchain distribution and no repository secret.

## Root commands

| Command             | Purpose                                                                         |
| ------------------- | ------------------------------------------------------------------------------- |
| `pnpm dev`          | Run the Next.js web development server.                                         |
| `pnpm validate`     | Run workspace, format, lint/vet, type, test, and build validation.              |
| `pnpm lint`         | Run Next.js-aware web lint, package ESLint, and `go vet` for the API.           |
| `pnpm typecheck`    | Generate Next.js route types and type-check the web and shared packages.        |
| `pnpm test`         | Run workspace foundation tests and API tests.                                   |
| `pnpm build`        | Build the Next.js web app, TypeScript packages, and Go API skeleton.            |
| `pnpm format:check` | Check Prettier and `gofmt` formatting without writing.                          |
| `pnpm format`       | Apply Prettier and `gofmt` formatting.                                          |
| `pnpm audit`        | Fail when the pnpm production dependency graph has a high or critical advisory. |

The audit policy permits moderate and low advisories to be reported without failing;
all reported advisories remain visible and must be recorded in the pull request.

## Project-specific commands

Use `pnpm --filter @vocanova/web dev`, `build`, `start`, `lint`, or `typecheck` for
the Next.js application. `start` serves a prior production build. The root page is a
technical framework-validation placeholder and contains no product UI.

Run API commands from `apps/api`:

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./...
```

`ent/` and `migrations/` are non-executable structural foundations only.

## Troubleshooting

- An engine or package-manager mismatch means the exact declared Node/pnpm versions
  are not active; switch versions and repeat the frozen install.
- A frozen-install failure means `package.json` and `pnpm-lock.yaml` disagree. Do not
  bypass it with a non-frozen CI install; reconcile dependencies in an authorized
  change.
- `pnpm validate` stops at the first failing child command and preserves its output.
- Go may download `go1.26.5` on first use. A network failure is not a passing API
  check; restore official toolchain access and rerun it.
- An agent/CI sandbox may have an unreachable internal `GOPROXY` (e.g. a private-network
  mirror) and/or `GOSUMDB=off` set by default. Neither is a repository requirement -
  `go env GOPROXY` / `go env GOSUMDB` show the active values. If the toolchain download
  or a module fetch fails, retry with the public defaults rather than assuming the
  build is broken:
  ```bash
  GOPROXY=https://proxy.golang.org,direct GOSUMDB=sum.golang.org go build ./...
  ```
  `GOSUMDB=off` specifically blocks the Go _toolchain_ download itself (distinct from
  disabling module checksum verification, which it also does) - re-enabling it is
  required to fetch a missing `go1.26.5`, e.g. via `go install golang.org/dl/go1.26.5@latest`
  then `go1.26.5 download`.
- No deployment, migration, integration, accessibility, staging, or production check
  exists in this foundation.
