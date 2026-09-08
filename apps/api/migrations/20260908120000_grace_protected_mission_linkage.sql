-- atlas:txmode file
-- DOC-05 §§2,10,12: a protected daily mission is supported by one exact,
-- same-learner grace-ledger debit. Keep the snapshot's denormalized grace
-- fields coherent with that append-only source of truth at the database
-- boundary, not only in the reconciliation service.

-- PostgreSQL requires a unique parent key for a composite foreign key. The
-- primary key already makes id unique; this adds the user-scoped parent key
-- needed to prove a snapshot cannot reference another learner's ledger row.
ALTER TABLE grace_day_ledger
  ADD CONSTRAINT grace_day_ledger_id_user_id_key UNIQUE (id, user_id);

ALTER TABLE daily_mission_snapshots
  ADD CONSTRAINT daily_mission_snapshots_grace_protection_fields_consistent
    CHECK (
      (status = 'protected' AND grace_applied AND grace_day_id IS NOT NULL)
      OR
      (status <> 'protected' AND NOT grace_applied AND grace_day_id IS NULL)
    ),
  ADD CONSTRAINT daily_mission_snapshots_grace_day_same_user_fkey
    FOREIGN KEY (grace_day_id, user_id)
    REFERENCES grace_day_ledger (id, user_id)
    ON DELETE RESTRICT;
