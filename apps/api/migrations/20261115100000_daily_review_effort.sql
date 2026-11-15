-- atlas:txmode file
-- Effort and successful recall are separate facts. Canonical identity enforces
-- at most one goal credit across sibling meanings on a learner-local day.
ALTER TABLE word_meanings ADD CONSTRAINT word_meanings_id_word_key UNIQUE(id,word_id);
CREATE TABLE daily_review_effort (
 user_id uuid NOT NULL REFERENCES users(id),
 word_id uuid NOT NULL REFERENCES canonical_words(id),
 meaning_id uuid NOT NULL,
 local_date date NOT NULL,
 review_attempt_id uuid NOT NULL UNIQUE,
 daily_session_id uuid NOT NULL,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(user_id,word_id,local_date),
 FOREIGN KEY(meaning_id,word_id) REFERENCES word_meanings(id,word_id),
 FOREIGN KEY(review_attempt_id,user_id,meaning_id) REFERENCES review_attempts(id,user_id,meaning_id),
 FOREIGN KEY(daily_session_id,user_id) REFERENCES daily_lesson_sessions(id,user_id)
);
-- No bulk historical backfill: current-day reconciliation runs under the user
-- lock alongside mission completion/rewards. Completed past days stay immutable.
