// Package gamification implements the Confidence Points, streak, and grace-day
// domain. It exposes pure, unit-tested domain functions (no Huma/chi, no
// direct Ent writes inside the domain core) plus transaction-scoped write
// helpers that operate against a caller's existing *sql.Tx. The transaction
// layer is responsible for opening the transaction (DOC-06 §3).
package gamification
