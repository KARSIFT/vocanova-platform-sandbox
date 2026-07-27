// Package users implements the learner-owned onboarding profile and
// settings/onboarding domain logic. It exposes pure, unit-tested domain
// functions (no Huma/chi, no direct Ent writes inside the domain core).
// The transaction layer is responsible for materializing domain decisions
// into database writes inside the caller's existing *sql.Tx, mirroring the
// pattern established by gamification and missions.
package users
