# Privacy Policy (Draft for Founder Review)

**Status:** Draft, not yet approved for publication  
**Prepared under:** `VOC-037-T02`  
**Last updated:** 2026-08-01

## 1. Scope

This Privacy Policy describes how VocaNova ("we", "our", "us") handles personal
data when you use the VocaNova web application and related services.

This document is a draft prepared for founder review and may be revised before
publication.

## 2. Data We Collect

Based on the current implemented product design (`DOC-05`, `DOC-06`, `DOC-07`,
`DOC-09`), VocaNova may process the following categories of data:

- Account and identity data:
  - Email address (when provided)
  - Authentication provider identifiers (Google OAuth subject and email-magic-link
    identity records)
  - Basic profile fields such as display name and avatar URL (if provided by
    identity provider)
- Learning data:
  - Saved words and meanings
  - Review attempts, review ratings, and review scheduling state
  - Daily mission progress and activity summaries
  - Confidence points, streak state, and grace-day ledger records
- Learner-created content:
  - Sentences you submit for learning practice
  - AI feedback results associated with those sentences
  - Feedback-quality reports you submit about AI results
- Settings data:
  - Timezone and selected app-learning settings
  - Notification and marketing preference flags
- Technical and operational data:
  - Request metadata and service logs needed for reliability and security
  - Security/audit records related to critical actions
  - Monitoring and uptime/error telemetry

## 3. How We Use Data

We process personal data to:

- Provide account access and authentication
- Deliver core learning features (saved words, review workflows, mission progress)
- Generate and return AI sentence feedback
- Detect abuse, maintain security, and operate the service
- Debug incidents and improve reliability
- Measure product health using privacy-conscious operational metrics
- Comply with legal obligations

## 4. AI Processing and Data Minimization

For sentence-feedback features, provider requests are intended to include only the
minimum data needed for the task, such as:

- CEFR level
- Target word or phrase and related learning metadata
- The learner sentence submitted for feedback

Per current engineering policy (`DOC-09`), provider requests should not
intentionally include unrelated account history or unnecessary identifiers.

## 5. Cookies and Session Data

VocaNova uses server-managed authenticated sessions and security mechanisms,
including HttpOnly cookies and CSRF protections, to keep accounts secure and
maintain signed-in state.

## 6. Data Sharing

We may share data with service providers that help us operate the platform, such
as:

- Hosting and infrastructure providers
- Authentication providers
- AI service providers for sentence feedback
- Error and uptime monitoring providers

We do not disclose data beyond what is needed for these services to perform their
roles.

## 7. Retention

Retention is governed by product and operational policies and may change as legal
requirements are finalized. Current design references include:

- Learning and account-related records retained while an account is active
- Structured operational records retained for bounded periods
- Account deletion workflows that deactivate access immediately and then perform a
  staged, verified purge/anonymization flow (`DOC-05`, `DOC-06`, `DOC-09`)

Final retention periods and legal bases must be founder-reviewed before
publication.

## 8. Security Controls

Current design and operations documentation includes controls such as:

- Environment separation between preview, staging, and production
- No production secrets reachable from lower environments
- HTTPS/TLS, secure cookies, and CSRF protections
- Access controls, structured logging, and monitoring

No security measure is guaranteed to be perfect, but we apply layered safeguards
appropriate to the service.

## 9. Your Rights and Choices

Depending on applicable law and your location, you may have rights to:

- Access your personal data
- Correct inaccurate data
- Request deletion of your account and associated personal data
- Object to or restrict certain processing

Support/contact procedures for these requests must be finalized before this policy
is published.

## 10. Children's Privacy

VocaNova is intended for learners within the approved product scope. Founder/legal
review must confirm final age-related terms and jurisdiction requirements before
publication.

## 11. International Processing

Service providers may process data in different regions based on deployment and
vendor configuration. Final region disclosures and transfer language require
founder/legal confirmation at publication time.

## 12. Changes to This Policy

We may update this policy as the product evolves. Material updates should include
an updated effective date and, where required, additional notice.

## 13. Contact

Contact details are `TBD` pending founder decision on support/legal inbox and
publication process.

---

## Founder Review Record (Required Before Publication)

- Reviewer: `TBD`
- Decision: `pending`
- Date: `TBD`
- Notes: `TBD`
