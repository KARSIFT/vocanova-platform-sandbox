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

// ReviewAttempt is immutable learner history. Ent declares its fields
// create-only and the forward migration enforces update/delete protection at
// PostgreSQL's final integrity layer.
type ReviewAttempt struct{ ent.Schema }

func (ReviewAttempt) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "review_attempts"}}
}

func (ReviewAttempt) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, ImmutableTimeMixin{}} }

func (ReviewAttempt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.UUID("user_word_id", uuid.UUID{}).Immutable(),
		field.UUID("meaning_id", uuid.UUID{}).Immutable(),
		field.String("attempt_type").NotEmpty().Immutable(),
		field.Enum("prompt_type").
			Values("multiple_choice", "self_check").Immutable(),
		field.Enum("result").
			Values("correct", "incorrect", "skipped").Immutable(),
		field.Enum("rating").
			Values("again", "hard", "good", "easy").
			Optional().Nillable().Immutable(),
		field.Int("review_step_before").Min(0).Max(7).Immutable(),
		field.Int("review_step_after").Min(0).Max(7).Immutable(),
		field.Time("answered_at").Immutable(),
		field.Int("response_time_ms").Min(0).Default(0).Immutable(),
		field.UUID("selected_option_meaning_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.String("typed_answer").Optional().Nillable().Immutable(),
		field.Bool("was_hint_used").Default(false).Immutable(),
		field.String("source").NotEmpty().Immutable(),
		field.String("client_attempt_id").Optional().Nillable().Immutable(),
		field.JSON("metadata", map[string]any{}).Optional().Immutable(),
	}
}

func (ReviewAttempt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("review_attempts").
			Field("user_id").
			Unique().
			Required(),
		edge.From("user_word", UserWord.Type).
			Ref("review_attempts").
			Field("user_word_id").
			Unique().
			Required(),
		edge.From("meaning", WordMeaning.Type).
			Ref("review_attempts").
			Field("meaning_id").
			Unique().
			Required(),
		edge.From("selected_option_meaning", WordMeaning.Type).
			Ref("selected_option_attempts").
			Field("selected_option_meaning_id").
			Unique(),
	}
}

func (ReviewAttempt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "client_attempt_id").
			Unique().
			Annotations(entsql.IndexWhere("client_attempt_id IS NOT NULL")),
		index.Fields("user_id", "answered_at"),
		index.Fields("user_word_id", "answered_at"),
		index.Fields("meaning_id", "answered_at"),
	}
}
