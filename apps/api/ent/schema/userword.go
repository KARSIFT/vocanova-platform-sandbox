package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type UserWord struct{ ent.Schema }

func (UserWord) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "user_words"}}
}
func (UserWord) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}, SoftDeleteMixin{}} }

func (UserWord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("meaning_id", uuid.UUID{}),
		field.Enum("status").
			Values("new", "learning", "reviewing", "mastered", "ignored", "archived").
			Default("new"),
		field.Enum("source").
			Values("journey", "search", "ai_suggestion", "manual", "seed"),
		field.Int("review_step").Min(0).Max(7).Default(0),
		field.Time("next_review_at").Optional().Nillable(),
		field.Time("last_reviewed_at").Optional().Nillable(),
		field.Enum("last_result").
			Values("correct", "incorrect", "skipped").
			Optional().Nillable(),
		field.Enum("last_rating").
			Values("again", "hard", "good", "easy").
			Optional().Nillable(),
		field.Int("consecutive_correct_count").Default(0),
		field.Int("consecutive_incorrect_count").Default(0),
		field.Int("total_review_count").Default(0),
		field.Int("correct_review_count").Default(0),
		field.Time("added_at").Default(time.Now),
		field.Time("mastered_at").Optional().Nillable(),
		field.Time("ignored_at").Optional().Nillable(),
	}
}

func (UserWord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("user_words").Field("user_id").Unique().Required(),
		edge.From("meaning", WordMeaning.Type).Ref("user_words").Field("meaning_id").Unique().Required(),
		edge.To("review_attempts", ReviewAttempt.Type),
		edge.To("learner_sentences", LearnerSentence.Type),
	}
}

func (UserWord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "meaning_id").
			Unique().
			Annotations(entsql.IndexWhere("deleted_at IS NULL")),
		index.Fields("user_id", "status"),
		index.Fields("user_id", "next_review_at"),
	}
}
