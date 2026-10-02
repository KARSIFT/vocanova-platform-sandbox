# Documentation

Start with the [current product state](product/current-state.md), the
[product bible](product/00-product-bible.md), and the
[development guide](development.md). The product state distinguishes implemented
features, dated live observations, and remaining readiness work.

## Source authority

- Owner instructions and [AGENTS.md](../AGENTS.md) define current scope, safety,
  and delivery. Ordinary work uses a PR against `main`; the old plan/adopt/roster
  and `develop` promotion process is retired.
- Product and design documents describe intended learner behavior. Later
  owner-authorized deliveries can extend the original MVP.
- Code, migrations, tests, and [workflow definitions](../.github/workflows/)
  establish implementation. Capability switches and fixture tests do not prove
  successful live use by a learner.
- Dated verification records establish only the revision, environment, and
  checks they describe. Web/API `/version` responses identify deployed releases.
  Merging deploys staging; production deployment is manual.
- [Decisions](decisions/README.md), [change packages](../specs/README.md), and
  [migration records](archive/README-migration-notes.md) preserve rationale and
  history. Old approval/adoption labels do not activate a retired process.

Some older category indexes and runbooks refer to removed `docs/governance/`
files. Follow AGENTS.md for current delivery and actual workflow triggers for
deployment; historical comments can describe former arrangements.

## Working guides

The [release-readiness plan](product/release-readiness.md) orders the remaining
delivery work and defines the evidence needed before a public launch.

| Need                                     | Read                                                                                                                                                                    |
| ---------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Product behavior and remaining evidence  | [Current state](product/current-state.md), [Product bible](product/00-product-bible.md), [MVP PRD](product/01-mvp-prd.md)                                               |
| Local setup and validation               | [Development](development.md)                                                                                                                                           |
| Current visual direction                 | [Learning workspace](design/learning-workspace.md)                                                                                                                      |
| Learning continuity and sentence history | [Mature learning and account experience](product/mature-learning-and-account-experience.md)                                                                             |
| Passwords, profile, and appearance       | [Delivery scope](product/password-profile-and-theme.md), [email activation](development/account-and-release-operations.md)                                              |
| Cohort operations                        | [Staging controlled signup](operations/staging-controlled-signup.md), [production controlled signup](operations/production-controlled-signup.md)                        |
| Infrastructure and monitoring            | [Infrastructure guide](../infra/README.md), [monitoring runbook](operations/monitoring.md)                                                                              |
| API and persistence                      | [Backend design](engineering/06-backend-design.md), [API contract](engineering/07-api-contract-and-dto-design.md), [database design](engineering/05-database-design.md) |
| AI feedback                              | [AI requirements](engineering/09-ai-features.md), [implementation](../apps/api/business/aifeedback/)                                                                    |
| Agent context                            | [AGENTS.md](../AGENTS.md), [agent skills](development/agent-skills.md)                                                                                                  |

## Document corpus

The use column explains how to read these reconciled documents today.
Frontmatter retains the original adoption record.

| ID     | Document                                                                                        | Current use                                                              |
| ------ | ----------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| DOC-00 | [Product Bible](product/00-product-bible.md)                                                    | Product baseline                                                         |
| DOC-01 | [MVP PRD](product/01-mvp-prd.md)                                                                | Original MVP requirements; read with later deliveries                    |
| DOC-02 | [Market Research](research/02-market-research.md)                                               | Dated research                                                           |
| DOC-03 | [UI/UX Design](design/03-ui-ux-design.md)                                                       | Design baseline; read with learning workspace guidance                   |
| DOC-04 | [Technical Architecture](engineering/04-technical-architecture.md)                              | Architecture baseline; verify against source                             |
| DOC-05 | [Database Design](engineering/05-database-design.md)                                            | Domain design; schemas/migrations show current persistence               |
| DOC-06 | [Backend Design](engineering/06-backend-design.md)                                              | Service design baseline                                                  |
| DOC-07 | [API Contract and DTO Design](engineering/07-api-contract-and-dto-design.md)                    | Contract design; compare with current API/client                         |
| DOC-08 | [Web Application Design](design/08-web-app-design.md)                                           | Screen and interaction baseline                                          |
| DOC-09 | [AI Features](engineering/09-ai-features.md)                                                    | Feedback and evaluation requirements                                     |
| DOC-10 | [Development Workflow](operations/10-development-workflow.md)                                   | Historical delivery rules superseded by AGENTS.md                        |
| DOC-11 | [DevOps and CI/CD Plan](operations/11-devops-and-ci-cd.md)                                      | Infrastructure history; current workflows supersede branch/release rules |
| DOC-12 | [MVP Implementation Plan](product/12-mvp-implementation-plan.md)                                | Historical implementation sequence                                       |
| DOC-13 | [F1 Foundation Execution Package](operations/13-f1-repository-foundation-execution-package.md)  | Historical completed foundation work                                     |
| DOC-14 | Historical KARSIFT AI Development Automation Architecture                                       | Not adopted; context in DOC-19                                           |
| DOC-15 | [AI-Native Operating Model](operations/15-ai-native-product-and-engineering-operating-model.md) | Historical operating model                                               |
| DOC-16 | Autonomous Development Operating Model                                                          | Former governance document; removed from this checkout                   |
| DOC-17 | [Autonomous Development Architecture](architecture/17-autonomous-development-architecture.md)   | Historical unbuilt Control Plane proposal                                |
| DOC-18 | [Autonomous Development Roadmap](planning/18-autonomous-development-implementation-roadmap.md)  | Historical Control Plane roadmap, not current product work               |
| DOC-19 | [Governance Reconciliation Notes](operations/19-governance-reconciliation-notes.md)             | Historical reconciliation                                                |

## Migration evidence

- [Migration manifest](archive/migration-manifest.yaml): source hashes, coverage,
  and original disposition.
- [Document graph](archive/document-graph.yaml): derived relationships, not authority.
- [Migration notes](archive/README-migration-notes.md) and
  [adoption notes](archive/README-adoption-notes.md): reconciliation decisions.

DOC-17/18 described an unbuilt standalone Control Plane. A later reusable GitHub
Actions pipeline shipped earlier work and was itself retired. Today's workflow
is the small set of repository GitHub Actions described in AGENTS.md. Preserve
that history without treating either earlier system as a prerequisite for
delivering the learning product.
