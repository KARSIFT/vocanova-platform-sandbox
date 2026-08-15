# Protected Areas

Protected areas create an R3 floor unless a more consequential R4 rule applies.
Protection covers both current paths and equivalent future paths.

| Area | Current or anticipated paths/change types | Required specialist concern |
|---|---|---|
| Authentication and authorization | `/apps/**/auth/`, `/packages/auth/`, middleware, sessions, roles, access policy | Security and authorization review |
| Database schemas and migrations | `/packages/database/`, `/database/`, `/db/`, `/migrations/`, schema files | Data integrity, forward/backward compatibility, recovery |
| Billing and payments | `/packages/billing/`, `/payments/`, checkout, invoices, webhooks | Financial correctness, authorization, idempotency |
| Secrets and credentials | secret policy, environment bindings, token issuance, `.env` handling | Exposure prevention, rotation, least privilege |
| Production infrastructure | `/infrastructure/`, `/infra/`, Cloudflare/Wrangler config, environment configuration | Deployment permissions, isolation, rollback |
| AI provider configuration | `/packages/ai/`, provider clients, prompts used as controls, model routing, quotas | Privacy, cost, safety, evaluation |
| Sensitive-data handling | personal data models, exports, deletion, retention, logs, analytics identifiers | Privacy, minimization, retention, access |
| Audio and voice storage | recording upload, object storage, transcription inputs, retention/deletion | Consent, privacy, encryption, lifecycle |
| Backup and recovery | backup schedules, restore scripts/runbooks, recovery objectives | Restore testing and evidence |
| Deployment and rollback | `/.github/workflows/`, deploy scripts, release/rollback policy | Supply chain, permissions, environment gates |
| Repository governance | `/.github/CODEOWNERS`, branch/ruleset policy, `/CONTRIBUTING.md`, `/SECURITY.md`, `/docs/governance/`, agent instructions | Independent policy and authority review |

## Semantic protection

Protection applies by effect even when a file is outside a listed path. For example,
a generic utility that changes authorization, a product document that changes data
retention, or a package script that deploys production is protected.

The following are R4 by decision type even when no protected path is changed:

- pricing, monetization, contractual, legal, or material privacy decisions;
- product strategy, material user-facing scope, and public commitments;
- consequential user-trust decisions;
- initial public launch and major launch decisions; and
- difficult-to-reverse company or product actions.

## Ownership

`CODEOWNERS` routes review; it is not approval evidence and does not express every
required approval combination. Historically under A-003, routine R3 ownership routing
did not imply standing technical-steward or founder approval. **Under active A-004,**
R4 path floors and strengthened evidence remain; founder `approved`
comments are not engineering-workflow merge gates. Claude Code may independently
verify protected changes but is never a human approval authority.

EHR may obtain qualified human expertise for an exceptional triggered matter. It
must not become permanent ownership or a routine replacement approval layer.

## Bootstrap boundary

The initial DOC-16/A-002 adoption is protected R4 governance. Its narrowly scoped
bootstrap rule permits founder approval, independent Claude Code verification, and
repository validation to adopt the framework without claiming nonexistent steward
approval. It authorizes no production or R3 protected technical work. Immediately
after merge, ordinary R3 steward requirements apply and R3 production remains blocked
until a qualified human steward and enforcement are configured.
