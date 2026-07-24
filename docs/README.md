# Documentation

`docs/` contains approved and proposed living documentation. Only documents whose frontmatter
status is `approved` are authoritative implementation inputs. Executable work authority lives in
adopted packages under [`specs/`](../specs/README.md); decision rationale lives in
[`docs/decisions/`](decisions/README.md).

## Categories

- [Product](product/README.md)
- [Research](research/README.md)
- [Design](design/README.md)
- [Engineering](engineering/README.md)
- [Architecture](architecture/README.md)
- [Planning](planning/README.md)
- [Operations](operations/README.md)
- [Governance](governance/README.md)
- [Decisions](decisions/README.md)
- [Templates](templates/README.md)

## Canonical document index

| ID | Title | Status | Owner | Canonical path | Related artifacts |
|---|---|---|---|---|---|
| DOC-00 | [VocaNova Product Bible](product/00-product-bible.md) | approved | founder | `docs/product/00-product-bible.md` | DOC-01, DOC-02, DOC-03, DOC-05, DOC-09, DOC-12 |
| DOC-01 | [VocaNova MVP PRD](product/01-mvp-prd.md) | approved | founder | `docs/product/01-mvp-prd.md` | DOC-00, DOC-03, DOC-08, DOC-09, DOC-12 |
| DOC-02 | [VocaNova Market Research](research/02-market-research.md) | approved | founder | `docs/research/02-market-research.md` | DOC-00, DOC-01 |
| DOC-03 | [VocaNova UI/UX Design](design/03-ui-ux-design.md) | approved | founder | `docs/design/03-ui-ux-design.md` | DOC-00, DOC-01, DOC-08, DOC-09 |
| DOC-04 | [VocaNova Technical Architecture](engineering/04-technical-architecture.md) | approved | founder | `docs/engineering/04-technical-architecture.md` | DOC-05, DOC-06, DOC-07, DOC-08, DOC-09, DOC-10, DOC-11, DOC-17 |
| DOC-05 | [VocaNova Database Design](engineering/05-database-design.md) | approved | founder | `docs/engineering/05-database-design.md` | DOC-04, DOC-06, DOC-07, DOC-09 |
| DOC-06 | [VocaNova Backend Design](engineering/06-backend-design.md) | approved | founder | `docs/engineering/06-backend-design.md` | DOC-04, DOC-05, DOC-07, DOC-09, DOC-10 |
| DOC-07 | [VocaNova API Contract and DTO Design](engineering/07-api-contract-and-dto-design.md) | approved | founder | `docs/engineering/07-api-contract-and-dto-design.md` | DOC-04, DOC-05, DOC-06, DOC-08, DOC-09 |
| DOC-08 | [VocaNova Web Application Design](design/08-web-app-design.md) | approved | founder | `docs/design/08-web-app-design.md` | DOC-03, DOC-04, DOC-07, DOC-09 |
| DOC-09 | [VocaNova AI Features](engineering/09-ai-features.md) | approved | founder | `docs/engineering/09-ai-features.md` | DOC-00, DOC-01, DOC-04, DOC-05, DOC-06, DOC-07 |
| DOC-10 | [VocaNova Development Workflow](operations/10-development-workflow.md) | approved | founder | `docs/operations/10-development-workflow.md` | DOC-11, DOC-15, DOC-16, DOC-19 |
| DOC-11 | [VocaNova DevOps and CI/CD Plan](operations/11-devops-and-ci-cd.md) | approved | founder | `docs/operations/11-devops-and-ci-cd.md` | DOC-10, DOC-16, DOC-19 |
| DOC-12 | [VocaNova MVP Implementation Plan](product/12-mvp-implementation-plan.md) | approved | founder | `docs/product/12-mvp-implementation-plan.md` | DOC-00, DOC-01, DOC-03, DOC-04, DOC-10, DOC-11, DOC-13, DOC-18 |
| DOC-13 | [VocaNova F1 Repository Foundation Execution Package](operations/13-f1-repository-foundation-execution-package.md) | historical (F1 complete) | founder | `docs/operations/13-f1-repository-foundation-execution-package.md` | DOC-10, DOC-12, DOC-15, DOC-16 |
| DOC-14 | Historical KARSIFT AI Development Automation Architecture | not adopted | founder | Preserved as research; see DOC-19 | DOC-19 |
| DOC-15 | [AI-Native Product and Engineering Operating Model](operations/15-ai-native-product-and-engineering-operating-model.md) | approved | founder | `docs/operations/15-ai-native-product-and-engineering-operating-model.md` | DOC-16, DOC-19 |
| DOC-16 | [Autonomous Development Operating Model](governance/16-autonomous-development-operating-model.md) | approved | founder | `docs/governance/16-autonomous-development-operating-model.md` | DOC-15, DOC-17, DOC-19 |
| DOC-17 | [Autonomous Development Architecture](architecture/17-autonomous-development-architecture.md) | approved | founder | `docs/architecture/17-autonomous-development-architecture.md` | DOC-16, DOC-18, DOC-19 |
| DOC-18 | [Autonomous Development Implementation Roadmap](planning/18-autonomous-development-implementation-roadmap.md) | approved | founder | `docs/planning/18-autonomous-development-implementation-roadmap.md` | DOC-17, DOC-19 |
| DOC-19 | [Governance Reconciliation Notes](operations/19-governance-reconciliation-notes.md) | proposed | founder | `docs/operations/19-governance-reconciliation-notes.md` | DOC-10, DOC-11, DOC-15, DOC-16, DOC-17, DOC-18, A-002, A-003 |

## Migration and relationships

- [Migration manifest](archive/migration-manifest.yaml) records source hashes, coverage, status, and disposition.
- [Document graph](archive/document-graph.yaml) is a derived impact aid and does not override authority.
- [Migration notes](archive/README-migration-notes.md) preserve the reconciliation evidence trail.
- [Adoption notes](archive/README-adoption-notes.md) record VOC-008 semantic corrections.

DOC-17 and DOC-18 are adopted together per VOC-004 (canonical adoption), but describe a system
that was never built and is not the project's actual direction (noted 2026-07-24; both remain
`approved`/adopted as historical planning documents, not deleted, but should not be read as
describing current or planned engineering work). They specify a standalone Control Plane service
(a durable PostgreSQL work queue, an AI Budget Governor, an Execution Lease Manager, an MCP
founder interface, etc.) and an 18-phase roadmap to build it. The system that actually shipped
VOC-010 through VOC-022 is architecturally unrelated: a set of reusable GitHub Actions workflows
(`KARSIFT/karsift-ai-infra`) wired into this repo's own `.github/workflows/pipeline.yml` - no
Postgres queue, no Budget Governor, no MCP interface, no Change Contract Registry. This was a
deliberate decision (see `karsift-ai-infra`'s own README and commit history), not an oversight.
Their adoption does not implement the Control Plane, activate RL1/RL2, or enable production
deployment or autonomous production release. It also does not, on its own, cover automatic merge
into `develop` - that narrower capability (A-003 §10) is separately implemented and live; see
`docs/governance/a003-transition-state.yaml`.
