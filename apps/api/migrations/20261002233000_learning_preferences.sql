-- Current direction is separate from immutable historical onboarding answers.
CREATE TABLE user_learning_preferences (
 id uuid PRIMARY KEY,
 user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE RESTRICT,
 learning_goal text NOT NULL CHECK (learning_goal IN ('general','work','travel','study','conversation','exam')),
 main_use_case text NOT NULL CHECK (main_use_case IN ('daily_life','work','travel','study','social')),
 revision bigint NOT NULL CHECK (revision > 0),
 created_at timestamptz NOT NULL,
 updated_at timestamptz NOT NULL
);
