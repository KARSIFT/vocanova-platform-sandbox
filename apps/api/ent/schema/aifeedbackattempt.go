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

// AIFeedbackAttempt is the immutable history of every AI feedback generation
// request for a learner sentence. It is intentionally append-only.
type AIFeedbackAttempt struct{ ent.Schema }

func (AIFeedbackAttempt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "ai_feedback_attempts",
			Checks: map[string]string{
				"completed_at_required_on_success": "status <> 'succeeded' OR completed_at IS NOT NULL",
				"error_code_required_on_failure":   "status <> 'failed' OR error_code IS NOT NULL",
			},
		},
	}
}

func (AIFeedbackAttempt) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}

func (AIFeedbackAttempt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("learner_sentence_id", uuid.UUID{}),
		field.Enum("status").
			Values("pending", "succeeded", "failed", "cancelled").
			Default("pending"),
		field.String("provider").NotEmpty(),
		field.String("model").NotEmpty(),
		field.String("prompt_version").NotEmpty(),
		field.String("request_hash").NotEmpty(),
		field.JSON("feedback_json", map[string]any{}).Optional(),
		field.String("feedback_text").Optional().Nillable(),
		field.String("error_code").Optional().Nillable(),
		field.String("error_message").Optional().Nillable(),
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
	}
}

func (AIFeedbackAttempt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("learner_sentence", LearnerSentence.Type).
			Ref("feedback_attempts").
			Field("learner_sentence_id").
			Unique().
			Required(),
	}
}

func (AIFeedbackAttempt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("request_hash").Unique().Annotations(entsql.IndexWhere("status IN ('pending', 'succeeded')")),
		index.Fields("learner_sentence_id", "started_at"),
		index.Fields("status", "started_at"),
	}
}
