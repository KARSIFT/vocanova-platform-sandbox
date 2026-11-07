-- One explicitly budgeted signal-recovery job per original operator attempt.
-- Ordinary/public generation policy and original attempt receipts are unchanged.
ALTER TABLE pronunciation_catalog_jobs
 ADD COLUMN recovery_of_attempt uuid UNIQUE REFERENCES pronunciation_catalog_attempts(id),
 ADD CONSTRAINT pronunciation_catalog_recovery_single_attempt
 CHECK (recovery_of_attempt IS NULL OR max_generations = 1);
