-- Daily review eligibility uses the server-written attempt timestamp, including
-- incorrect and skipped outcomes. Keep the learner/day lookup indexable before
-- joining the meaning to its canonical word.
CREATE INDEX review_attempts_user_id_created_at_meaning_id_idx
    ON review_attempts (user_id, created_at, meaning_id);
