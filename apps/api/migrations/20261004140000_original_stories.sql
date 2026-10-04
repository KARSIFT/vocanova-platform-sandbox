-- atlas:txmode file
-- Original stories preserve immutable private authored/grading snapshots.
-- Story completion grants no SRS, lesson, mastery, mission or reward credit.
CREATE TABLE story_sessions(
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id),
 story_key text NOT NULL CHECK(char_length(story_key) BETWEEN 1 AND 100),
 content_version text NOT NULL,
 grading_version text NOT NULL,
 snapshot jsonb NOT NULL CHECK(jsonb_typeof(snapshot)='object'),
 current_step integer NOT NULL DEFAULT 0,
 total_steps integer NOT NULL CHECK(total_steps BETWEEN 1 AND 100),
 revision integer NOT NULL DEFAULT 0 CHECK(revision>=0),
 feedback jsonb CHECK(feedback IS NULL OR jsonb_typeof(feedback)='object'),
 first_answers_correct integer NOT NULL DEFAULT 0,
 questions_answered integer NOT NULL DEFAULT 0,
 status text NOT NULL DEFAULT 'in_progress' CHECK(status IN ('in_progress','completed')),
 completed_at timestamptz,
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL,
 CONSTRAINT story_progress_valid CHECK(current_step>=0 AND current_step<=total_steps AND first_answers_correct>=0 AND first_answers_correct<=questions_answered AND questions_answered<=total_steps),
 CONSTRAINT story_completion_valid CHECK((status='completed' AND completed_at IS NOT NULL AND current_step=total_steps) OR(status='in_progress' AND completed_at IS NULL AND current_step<total_steps)),
 UNIQUE(id,user_id)
);
CREATE INDEX story_sessions_user_updated_idx ON story_sessions(user_id,story_key,updated_at DESC,id);
CREATE TABLE story_actions(
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
 FOREIGN KEY(session_id,user_id) REFERENCES story_sessions(id,user_id) ON DELETE CASCADE,
 UNIQUE(user_id,operation,idempotency_key),
 UNIQUE(session_id,operation,client_action_id)
);
CREATE INDEX story_actions_user_created_idx ON story_actions(user_id,created_at,id);
