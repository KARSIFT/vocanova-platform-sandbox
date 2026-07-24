# Versioned database migrations

Migrations are explicit operational artifacts and are never executed by API
startup. Apply the timestamped forward SQL with approved Atlas/PostgreSQL tooling
against a disposable database before promotion. The matching
`.down.sql.example` is deliberately not executable by Atlas and exists only for
disposable recovery rehearsal: production recovery must be separately approved
and must preserve integrity and invalidate unsafe sessions.
