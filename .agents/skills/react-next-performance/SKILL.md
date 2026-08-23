---
name: react-next-performance
description: Use when optimizing React 19 or Next.js 16 web performance in apps/web—rendering, bundles, data fetching, and Core Web Vitals without changing product behavior outside an approved scope.
---

# React and Next.js performance (vocanova web)

Repository-native guidance for `apps/web` (Next.js `16.3.x`, React `19.2.x`). Authoritative docs:

- [React Profiler](https://react.dev/reference/react/Profiler)
- [Next.js production checklist](https://nextjs.org/docs/app/guides/production-checklist)
- [Next.js Server and Client Components](https://nextjs.org/docs/app/getting-started/server-and-client-components)

## Governance precedence

When this skill conflicts with `AGENTS.md`, `CLAUDE.md`, approved change packages, tests, or source code, the repository sources win.

## Rejected upstream source

`vercel-labs/agent-skills` at commit `dd089a8c752c966dee8bf0f27cb625ba193ffd9e` was reviewed and **not** copied — no compatible license was detected for the React skill. This skill is authored from current official React/Next documentation and this repository's patterns.

## When to use

- Slow pages or layouts under `apps/web/src/app/`
- Large client bundles or unnecessary client components
- Layout shift, LCP, or hydration issues (see `apps/web/tests/lighthouse/`)
- Server Component vs Client Component boundary questions

## Workflow

1. **Measure** — use Lighthouse (`pnpm --filter @vocanova/web test:lighthouse`) or Next build output before optimizing.
2. **Identify layer** — server render, client hydration, data fetch, CSS, or asset delivery.
3. **Apply smallest fix** — one change per hypothesis; re-run targeted tests.
4. **Verify** — `pnpm --filter @vocanova/web build`, relevant Playwright or Lighthouse suites, then `pnpm validate` when in doubt (see `docs/development.md`).

## Practices for this codebase

### Server Components by default

- Keep `apps/web/src/app/**/page.tsx` as Server Components unless interactivity requires `"use client"`.
- Push client boundaries to leaf components; avoid client-wrapping entire routes.

### Data fetching

- Fetch on the server in Server Components or route handlers; avoid client-side waterfalls for initial paint.
- Colocate fetches with the component that consumes the data.

### Bundles and imports

- Import only needed symbols from shared packages (`@vocanova/design-tokens`, `@vocanova/api-client`).
- Audit new client dependencies for bundle impact before adding.

### Images and fonts

- Use Next.js `next/image` and documented font loading patterns from Next.js optimizing guides.
- Prefer static generation or caching where product requirements allow.

### Tailwind and design tokens

- Generated tokens live in `apps/web/src/app/tokens.generated.css`.
- **Do not use `max-w-md`, `max-w-xl`, or similar named `max-w-*` utilities** — they collide with spacing tokens (see `.karsift/lessons.md`). Use explicit arbitrary widths such as `max-w-[28rem]` with a short comment when needed.

### Lists and state

- Memoize only when profiling shows wasted renders; prefer stable keys and lean props.
- Avoid lifting state above the smallest subtree that needs it.

## Validation

| Check          | Command                                       |
| -------------- | --------------------------------------------- |
| Typecheck      | `pnpm --filter @vocanova/web typecheck`       |
| Build          | `pnpm --filter @vocanova/web build`           |
| E2E / a11y     | `pnpm --filter @vocanova/web test:e2e`        |
| Lighthouse     | `pnpm --filter @vocanova/web test:lighthouse` |
| Full workspace | `pnpm validate`                               |

## Safety

Performance work does not bypass governance — behavior changes need an adopted change package. Do not read `.env*` files. Do not paste raw CI logs into chat or issues. Never log secrets, tokens, or personal data while profiling.
