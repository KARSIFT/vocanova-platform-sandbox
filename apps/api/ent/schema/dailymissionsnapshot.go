package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DailyMissionSnapshot is the per-user-per-local-date daily-mission state.
// One row per (user_id, local_date). Owned by the missions module.
type DailyMissionSnapshot struct{ ent.Schema }

func (DailyMissionSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "daily_mission_snapshots",
			Checks: map[string]string{
				"review_target_in_range":        "review_target >= 5 AND review_target <= 100",
				"reviews_completed_in_range":    "reviews_completed >= 0 AND reviews_completed <= review_target",
				"new_word_target_in_range":      "new_word_target IS NULL OR (new_word_target >= 1 AND new_word_target <= 100)",
				"new_words_completed_in_range":  "new_word_target IS NULL OR (new_words_completed >= 0 AND new_words_completed <= new_word_target)",
				"new_word_goal_pair_consistent": "(new_word_target IS NULL AND new_words_completed IS NULL) OR (new_word_target IS NOT NULL AND new_words_completed IS NOT NULL AND new_words_completed >= 0 AND new_words_completed <= new_word_target)",
				"sentence_target_in_range":      "sentence_practice_target IS NULL OR (sentence_practice_target >= 1 AND sentence_practice_target <= 100)",
				"sentence_completed_in_range":   "sentence_practice_target IS NULL OR (sentence_practices_completed >= 0 AND sentence_practices_completed <= sentence_practice_target)",
				"sentence_goal_pair_consistent": "(sentence_practice_target IS NULL AND sentence_practices_completed IS NULL) OR (sentence_practice_target IS NOT NULL AND sentence_practices_completed IS NOT NULL AND sentence_practices_completed >= 0 AND sentence_practices_completed <= sentence_practice_target)",
				"completed_at_required_on_done": "status <> 'completed' OR completed_at IS NOT NULL",
				"completed_at_only_on_done":     "status = 'completed' OR completed_at IS NULL",
				// A protected day is the persisted result of consuming one
				// grace day. The cross-table, same-user foreign key is owned
				// by the Atlas migration because Ent cannot express composite
				// foreign keys; keep the local status/link invariant here too.
				"grace_protection_fields_consistent": "(status = 'protected' AND grace_applied AND grace_day_id IS NOT NULL) OR (status <> 'protected' AND NOT grace_applied AND grace_day_id IS NULL)",
			},
		},
	}
}

func (DailyMissionSnapshot) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (DailyMissionSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Time("local_date"),
		field.String("timezone").NotEmpty(),
		field.Int("review_target").Min(5).Max(100),
		field.Int("reviews_completed").Default(0).Min(0),
		field.Int("new_word_target").Optional().Nillable().Min(1).Max(100),
		field.Int("new_words_completed").Optional().Nillable().Min(0),
		field.Int("sentence_practice_target").Optional().Nillable().Min(1).Max(100),
		field.Int("sentence_practices_completed").Optional().Nillable().Min(0),
		field.String("policy_version").NotEmpty(),
		field.Enum("status").
			Values("open", "completed", "missed", "protected").
			Default("open"),
		field.Time("completed_at").Optional().Nillable(),
		field.Bool("grace_applied").Default(false),
		field.UUID("grace_day_id", uuid.UUID{}).Optional().Nillable(),
	}
}

func (DailyMissionSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("daily_mission_snapshots").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (DailyMissionSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "local_date").Unique(),
		index.Fields("user_id", "status"),
	}
}
