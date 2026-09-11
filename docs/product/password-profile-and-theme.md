# Password authentication, profile and appearance

## Scope

The founder requested explicit Google or email/password registration, a useful
profile, and Light/Dark/System appearance after the learning-continuity delivery
in PR #1464. This extends the original passwordless MVP; existing Google and
email-link accounts retain their identity and learning records.

Context: authentication service and account lifecycle, production capability
wiring, versioned migrations, current-user API contract, sign-in/account pages,
global design tokens and browser/accessibility fixtures. The baseline is
`f4d9d183`; main was pulled before implementation.

## Acceptance

- Sign-in offers configured Google and email/password methods, a clear create
  account path, password recovery, and the existing emailed-link alternative.
- Registration verifies email ownership before creating an account/credential.
  Registering an existing email never replaces its password or attaches a new
  credential. Existing Google/email-link users can establish a password only
  after proving ownership through the recovery email.
- Passwords allow passphrases and password managers, use bounded modern hashing,
  and never appear in logs, URLs, browser persistence or account exports.
- Verification/reset tokens are expiring, single-use, purpose-bound and hashed
  at rest. Invalid links show an error. Email-request responses avoid account
  enumeration. Reset revokes existing sessions and competing recovery links.
- Existing controlled-signup, disabled-account and reserved synthetic-identity
  restrictions remain enforced for every new path.
- Profile shows the learner's identity and supports changing the display name.
  Email changes retain their confirmation flow; profile and security links are
  discoverable from Settings without adding another primary learning tab.
- Appearance defaults to System, supports explicit Light/Dark overrides, persists
  on this device, reacts to OS changes in System mode and applies before first
  paint. Public, authentication and learning screens remain readable in either
  theme, including controls, errors and focus states.
- Settings About and public `/version` identify version, commit, environment and
  build time for both API and web; deployment workflows inject matching identity.

## Security and operational boundaries

Use the existing session/CSRF and account ownership boundaries. Password setup,
reset and session issuance require transaction-level protection against replay
and competing updates. New tables follow the versioned migration history and
must be included in account cleanup; schema changes are not applied at API startup.

The existing rate limiter is process-local, despite an older interface comment
describing a future shared backend. Password actions must apply IP and normalized
email limits before expensive work. Horizontal scaling needs a shared limiter;
this delivery must not claim a Redis implementation that the repository lacks.

Provider availability is configuration-driven. A Google button is offered only
when the API reports a configured provider. Registration/recovery capability must
not report success through a fake production email sender. Changing production
credentials or opening the controlled-signup cohort is not implied by UI changes.

## References

- [OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)
  informs passphrase-friendly password rules and session lifecycle.
- [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
  informs Argon2id hashing and bounded parameter validation.
- [OWASP Forgot Password](https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html)
  informs generic requests, single-use recovery tokens and session invalidation.

## Verification

Verify service/HTTP behavior and real PostgreSQL transactions, including replay,
existing-account registration, session revocation and account cleanup. Exercise
registration, sign-in, recovery, profile ownership and theme persistence in the
browser. Run accessibility and contrast checks in both themes at supported widths.
CI must apply the complete migration set and execute the new password lifecycle
against disposable PostgreSQL without silently skipping it.

Keep the PR draft for independent review and resolve findings before merge.
Record final verification and release evidence here when available.

Operational setup: [email activation and release identity](../development/account-and-release-operations.md).
