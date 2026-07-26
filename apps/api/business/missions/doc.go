// Package missions implements the per-user daily-mission and daily-activity
// persistence boundaries. It is transaction-scoped: every method that writes
// takes the caller's *sql.Tx (DOC-06 §3). The missions module owns
// daily_mission_snapshots, daily_activity_summaries, the lazy per-user
// per-local-date snapshot creation, and the real MissionUpdater
// implementation that the aifeedback module already declares an interface
// seam for.
package missions
