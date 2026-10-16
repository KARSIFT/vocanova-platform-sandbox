-- atlas:txmode file
ALTER TABLE review_attempts ADD CONSTRAINT review_attempts_id_user_meaning_key UNIQUE(id,user_id,meaning_id);
-- A successful meaning is counted once on the learner's local calendar day.
CREATE TABLE daily_review_successes(
 user_id uuid NOT NULL REFERENCES users(id),
 meaning_id uuid NOT NULL REFERENCES word_meanings(id),
 local_date date NOT NULL,
 review_attempt_id uuid NOT NULL,
 daily_session_id uuid NOT NULL,
 created_at timestamptz NOT NULL,
 PRIMARY KEY(user_id,meaning_id,local_date),
 UNIQUE(review_attempt_id),
 FOREIGN KEY(review_attempt_id,user_id,meaning_id) REFERENCES review_attempts(id,user_id,meaning_id)
);
CREATE TABLE daily_lesson_sessions(
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id),
 state jsonb NOT NULL CHECK(jsonb_typeof(state)='object'),
 status text NOT NULL CHECK(status IN('in_progress','completed')),
 revision integer NOT NULL DEFAULT 0 CHECK(revision>=0),
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL,
 completed_at timestamptz,
 CONSTRAINT daily_lesson_completion_valid CHECK((status='completed')=(completed_at IS NOT NULL)),
 UNIQUE(id,user_id)
);
ALTER TABLE daily_review_successes ADD CONSTRAINT daily_review_successes_session_owner_fk FOREIGN KEY(daily_session_id,user_id) REFERENCES daily_lesson_sessions(id,user_id);
CREATE UNIQUE INDEX daily_lesson_one_active ON daily_lesson_sessions(user_id) WHERE status='in_progress';
CREATE INDEX daily_lesson_user_updated ON daily_lesson_sessions(user_id,updated_at DESC,id);
CREATE TABLE daily_lesson_actions(
 id uuid PRIMARY KEY,
 session_id uuid NOT NULL,
 user_id uuid NOT NULL REFERENCES users(id),
 operation text NOT NULL CHECK(operation IN('start','action')),
 idempotency_key text NOT NULL CHECK(char_length(idempotency_key) BETWEEN 1 AND 128),
 client_action_id text NOT NULL CHECK(char_length(client_action_id) BETWEEN 1 AND 128),
 fingerprint text NOT NULL CHECK(fingerprint ~ '^[0-9a-f]{64}$'),
 action jsonb NOT NULL CHECK(jsonb_typeof(action)='object'),
 result jsonb NOT NULL CHECK(jsonb_typeof(result)='object'),
 created_at timestamptz NOT NULL,
 FOREIGN KEY(session_id,user_id) REFERENCES daily_lesson_sessions(id,user_id) ON DELETE CASCADE,
 UNIQUE(user_id,operation,idempotency_key),
 UNIQUE(session_id,operation,client_action_id)
);
