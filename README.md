# Vocanova

Vocanova is an AI-powered platform for practical English learning. It is maintained
by KARSIFT as a pnpm and Go monorepo. The repository is public; see
`docs/governance/repository-settings.md` for the security posture that implies
(secret scanning/push protection enabled, no assumption of a private audience).

Its canonical roots are `apps/web`, `apps/api`, and the shared packages under `packages/`.
Real, shipped product surfaces exist today - `apps/web` has working Home, Progress, and
Journey/Discover (including situation drill-down) screens (VOC-018 through VOC-022), built
against the wired design-token system - not skeletons awaiting later approved changes. See
the [local development guide](docs/development.md) for exact tools and commands.

## Documentation

- [Documentation index](docs/README.md)
- [Product documentation](docs/product/)
- [Architecture documentation](docs/architecture/)
- [Planning documentation](docs/planning/)
- [Architecture Decision Records](docs/decisions/README.md)
- [Executable change packages](specs/README.md)
- [Autonomous development governance](docs/governance/README.md)
- [Workflow templates](docs/templates/README.md)
- [GitHub repository configuration](.github/README.md)
- [Contribution guidelines](CONTRIBUTING.md)
- [Security policy](SECURITY.md)

The repository uses three distinct knowledge systems: `docs/` for approved living
documentation, `docs/decisions/` for material decision rationale, and `specs/` for
bounded executable change packages. Documents 00–13 were migrated and adopted as canonical
(VOC-007/VOC-008); DOC-14 was deliberately reconciled but not adopted (see
[docs/README.md](docs/README.md) for the full index and each document's actual status -
that index, not this paragraph, is the source of truth for migration state going forward).

