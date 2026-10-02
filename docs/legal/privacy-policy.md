# Privacy Policy (Draft for Founder Review)

**Status:** Draft, not yet approved for publication  
**Prepared under:** `VOC-037-T02`  
**Last updated:** 2026-10-02 (technical amendments awaiting review)

## 1. Scope

This Privacy Policy describes how VocaNova ("we", "our", "us") handles personal
data when you use the VocaNova web application and related services.

This document is a draft prepared for founder review and may be revised before
publication. The conditional founder approval recorded on 2026-08-02 applies to
the earlier draft; the technical amendments dated 2026-10-02 have not been
approved. See the review records below and the [legal document status](README.md).

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
  - Unsubmitted or unresolved sentence drafts and retry information stored in the browser tab
  - Sentences you submit for learning practice
  - AI feedback results associated with those sentences
  - Feedback-quality reports you submit about AI results
- Settings data:
  - Timezone and selected app-learning settings
  - Notification and marketing preference flags
  - The device's selected Light, Dark or System appearance preference
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

## 5. Cookies and Browser Storage

VocaNova uses server-managed authenticated sessions and security mechanisms,
including HttpOnly cookies and CSRF protections, to keep accounts secure and
maintain signed-in state.

The web application also uses browser storage for these implemented purposes:

- **Sentence recovery:** Optional tab-scoped session storage holds unsent or
  unresolved sentence text, the practice source, attempt identifier and save
  timestamp. An unresolved submission can also retain an idempotency key so a
  retry can identify the same request rather than create a duplicate. Storage
  keys include the authenticated learner's identifier, practice source and
  attempt identifier. These are identifying learning/retry data, not login
  credentials; draft values do not contain session cookies or CSRF tokens.
- **Review-to-sentence recovery:** A separate learner-scoped session-storage
  entry holds a completed review's attempt identifier, target word, optional
  short definition and timestamp, so sentence practice can reopen without
  repeating the completed review.
- **Sign-in continuation:** Session storage can hold an allowed internal app
  destination and creation time to return to after Google sign-in. It is
  consumed once; the application removes it when consuming it and rejects it
  if more than 15 minutes old. It is also cleared after successful explicit
  sign-out or account deactivation, subject to browser storage availability.
- **Appearance:** Local storage and the `vocanova_theme` cookie remember Light,
  Dark or System on this device, including the appearance used on later page
  loads. Local storage has no application expiry; the cookie is set with a
  one-year maximum age. This preference is separate from authentication and is
  not cleared by the implemented sign-out or account-deactivation actions.

Sentence drafts and review context are eligible for recovery for two hours from
their latest stored timestamp. Expiry is checked when the application reads the
entry; an expired entry is rejected and removal is attempted. The sign-out draft
check also attempts stale-draft cleanup. There is no background timer that
guarantees erasure at two hours: unread entries may remain in browser session
storage until cleanup or the browser ends or clears that storage session.

Draft removal is attempted after a completed successful feedback response, an
explicit discard or starting another sentence; empty draft edits also remove
the stored draft. Successful explicit sign-out or account deactivation attempts
to clear all sentence drafts and review context in the current tab. An expired
authentication session alone does not clear them, so re-authentication can
recover a still-eligible draft. Browser storage restrictions or failures can
prevent recovery or cleanup; these controls do not erase separately retained
server-side learning records.

Technical sources: [draft and review-context storage](<../../apps/web/src/app/(app)/_components/sentence-feedback-drafts.ts>),
[practice lifecycle](<../../apps/web/src/app/(app)/_components/sentence-feedback.tsx>),
[sign-in continuation](../../apps/web/src/lib/oauth-continuation.ts) and
[appearance storage](../../apps/web/src/app/_components/theme-preference.tsx).
Account-exit cleanup is implemented by [sign-out](<../../apps/web/src/app/(app)/_components/app-header.tsx>)
and [account deactivation](<../../apps/web/src/app/(app)/settings/account/_components/account-deletion-form.tsx>).
These describe implemented behavior, not approval of final retention policy.

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

VocaNova requires users to be at least **13 years old** (founder decision,
2026-08-02). VocaNova does not knowingly collect personal data from anyone
under 13. If we learn an account belongs to a user under 13, we will take
steps to delete the associated data.

## 11. International Processing

The draft reviewed on 2026-08-02 listed these processing locations: application
hosting in Turkey; error-monitoring (Sentry) in the EU (Germany); Cloudflare's global network
for DNS/CDN/proxying. The technical amendments dated 2026-10-02 do not verify
current vendor locations; this inventory must be checked before publication as
vendors and hosting evolve. **Final cross-border transfer legal
basis (e.g. GDPR standard contractual clauses, if applicable) depends on
VocaNova's registered legal jurisdiction, which is not yet finalized (see the
Founder Review Record below) — this section must be revisited once that is
decided, before publication.**

## 12. Changes to This Policy

We may update this policy as the product evolves. Material updates should include
an updated effective date and, where required, additional notice.

## 13. Contact

Contact: **mr.groom.verge@gmail.com** (founder-designated support/privacy contact,
2026-08-02). A dedicated `support@vocanova.site` address is planned once
Cloudflare Email Routing is configured for the domain; this document should be
updated to that address once it is live and verified receiving mail.

---

## Founder Review Record (Required Before Publication)

- Reviewer: Founder (m-e-h-r-d-a-a-d)
- Decision: **Reviewed and approved, with one item still open before publication**
- Date: 2026-08-02
- Notes: Data-collection description (§2) confirmed accurate against the actual
  implemented product. Founder decisions made and applied to this draft:
  minimum age 13 (§10); contact email `mr.groom.verge@gmail.com`, to be
  upgraded to `support@vocanova.site` once Cloudflare Email Routing is
  verified live (§13). **Still open, blocking publication:** VocaNova's
  registered legal jurisdiction/governing law is not yet decided (pending
  incorporation status) — §11's cross-border transfer language and the
  parallel item in `terms-of-service.md` §14 both depend on it. This
  version was founder-approved in substance but not cleared to publish until
  that item resolved. This is the historical 2026-08-02 record; it does not
  approve the later technical amendments.

## Technical Amendment Review Record

- Amendment date: 2026-10-02
- Scope: Browser-local drafts, retry identity, review context, sign-in
  continuation and appearance storage (§2, §5); historical infrastructure
  inventory qualification (§11)
- Reviewer and review date: Not yet recorded
- Decision: Pending; no publication clearance
- Remaining conditions: Review these amendments and resolve the historical
  jurisdiction/cross-border transfer publication gate. The existing minimum-age
  and contact decisions are preserved, not newly decided or verified here.
