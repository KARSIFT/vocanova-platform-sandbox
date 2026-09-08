-- atlas:txmode file
-- DOC-05 §§2,10: these daily aggregate fields represent observed learner
-- activity. Keep their immutable source records authoritative by rejecting a
-- negative cached projection before it reaches Home, streak, or Progress.

ALTER TABLE daily_activity_summaries
  ADD CONSTRAINT daily_activity_summaries_remaining_counters_nonnegative
    CHECK (
      words_discovered >= 0
      AND words_added >= 0
      AND sentences_submitted >= 0
      AND ai_feedback_received >= 0
    ) NOT VALID;

-- NOT VALID keeps this forward-only migration deployable where a legacy
-- aggregate is already corrupt. PostgreSQL nevertheless enforces the check
-- for every later INSERT or UPDATE; reconciliation and validation of legacy
-- summaries remain a separate, deliberate operation.
