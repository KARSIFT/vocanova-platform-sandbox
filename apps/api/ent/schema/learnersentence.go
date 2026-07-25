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

// LearnerSentence is a learner-generated original sentence submitted for AI
// feedback. It is soft-deleted because it contains learner-generated content.
type LearnerSentence struct{ ent.Schema }

func (LearnerSentence) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "learner_sentences",
			Checks: map[string]string{
				"sentence_text_length": "char_length(sentence_text) <= 1000",
			},
		},
	}
}

func (LearnerSentence) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}, SoftDeleteMixin{}}
}

func (LearnerSentence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("meaning_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_word_id", uuid.UUID{}).Optional().Nillable(),
		field.String("sentence_text").NotEmpty(),
		field.String("normalized_sentence_text").NotEmpty(),
		field.Enum("source").
			Values("word_detail", "review", "daily_mission", "free_practice"),
		field.Enum("status").
			Values("submitted", "feedback_ready", "feedback_failed", "archived").
			Default("submitted"),
		field.Time("submitted_at").Default(time.Now),
	}
}

func (LearnerSentence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("learner_sentences").
			Field("user_id").
			Unique().
			Required(),
		edge.From("meaning", WordMeaning.Type).
			Ref("learner_sentences").
			Field("meaning_id").
			Unique(),
		edge.From("user_word", UserWord.Type).
			Ref("learner_sentences").
			Field("user_word_id").
			Unique(),
		edge.To("feedback_attempts", AIFeedbackAttempt.Type),
	}
}

func (LearnerSentence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "submitted_at"),
		index.Fields("user_id", "status"),
		index.Fields("meaning_id", "submitted_at"),
		index.Fields("user_word_id", "submitted_at"),
	}
}
