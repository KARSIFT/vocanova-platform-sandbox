package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AIFeedbackQualityReviewReport is the internal, learner-initiated quality
// signal attached to one feedback result. It is intentionally not exportable.
type AIFeedbackQualityReviewReport struct{ ent.Schema }

func (AIFeedbackQualityReviewReport) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "ai_feedback_quality_review_reports", Checks: map[string]string{
		"reason_valid":         "reason IN ('already_correct', 'correction_changed_meaning', 'explanation_unclear', 'inappropriate', 'something_else')",
		"state_valid":          "state IN ('open', 'reviewing', 'confirmed_issue', 'no_issue_found', 'duplicate', 'resolved')",
		"classification_valid": "classification IS NULL OR classification IN ('incorrect_judgment', 'unnecessary_correction', 'meaning_changed', 'unclear_explanation', 'inappropriate_tone', 'unsafe_response', 'regional_variant_error', 'provider_failure', 'other')",
	}}}
}
func (AIFeedbackQualityReviewReport) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, TimeMixin{}}
}
func (AIFeedbackQualityReviewReport) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("ai_feedback_attempt_id", uuid.UUID{}).Unique().Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.String("reason").NotEmpty().Immutable(), field.String("state").Default("open"),
		field.String("classification").Optional().Nillable(),
	}
}
func (AIFeedbackQualityReviewReport) Indexes() []ent.Index {
	return []ent.Index{index.Fields("state", "created_at"), index.Fields("user_id")}
}
