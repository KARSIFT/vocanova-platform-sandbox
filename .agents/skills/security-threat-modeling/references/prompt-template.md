# Threat model output contract

Produce Markdown with these sections:

## 1. Scope and assumptions

- In-scope paths and components
- Deployment model (staging/production) with doc references
- Explicit assumptions and open questions

## 2. System overview

- Components and data flows (evidence-backed)
- Runtime vs CI/dev separation

## 3. Trust boundaries

| Boundary | Protocol | Auth / validation | Notes |
|----------|----------|-------------------|-------|

## 4. Assets

| Asset | Sensitivity | Location / evidence |

## 5. Entry points

| Entry | Exposure | AuthN / AuthZ |

## 6. Threats and abuse paths

| ID | Threat | Asset | Likelihood | Impact | Priority | Existing controls | Recommendations |

## 7. Mitigation summary

- Quick wins vs structural changes
- Links to files or packages for implementation follow-up

Keep the report concise and reviewable. Mark conditional recommendations when assumptions are unresolved.
