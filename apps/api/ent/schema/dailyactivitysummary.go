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

// DailyActivitySummary is the per-user-per-local-date aggregate counter table.
// The record of truth for individual events remains review_attempts,
// learner_sentences, ai_feedback_attempts, and confidence_point_ledger; this
// table is a fast aggregate read for Home/streak/Progress. Owned by missions.
type DailyActivitySummary struct{ ent.Schema }

func (DailyActivitySummary) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "daily_activity_summaries"}}
}

func (DailyActivitySummary) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (DailyActivitySummary) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Time("local_date"),
		field.String("timezone").NotEmpty(),
		field.Int("reviews_attempted").Default(0).Min(0),
		field.Int("reviews_correct").Default(0).Min(0),
		field.Int("reviews_skipped").Default(0).Min(0),
		field.Int("words_discovered").Default(0).Min(0),
		field.Int("words_added").Default(0).Min(0),
		field.Int("sentences_submitted").Default(0).Min(0),
		field.Int("ai_feedback_received").Default(0).Min(0),
		field.Int("confidence_points_earned").Default(0),
		field.Int("confidence_points_spent").Default(0).Min(0),
	}
}

func (DailyActivitySummary) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("daily_activity_summaries").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (DailyActivitySummary) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "local_date").Unique(),
	}
}
