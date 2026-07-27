// Package accounts implements the learner-owned account-lifecycle flows
// introduced by P5 (VOC-031). It currently owns the email-change
// verification flow (T03, DOC-06 §6, VOC-031-D05) and will own the
// account-deletion flow (T04, DOC-05 §16, VOC-031-D07) in a future PR.
//
// The package depends on the auth module for token generation,
// rate-limiting primitives, and the cross-module user/session
// operations (lookup, session revocation, etc.) it must not
// reimplement. auth's own magic-link/session code is never modified
// from here; the email-change flow is built strictly on top of the
// existing primitives.
package accounts
