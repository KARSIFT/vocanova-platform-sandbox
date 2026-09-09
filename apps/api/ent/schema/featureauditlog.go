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

// FeatureAuditLog is immutable operational history. Account deletion uses an
// explicit raw-SQL de-identification path for the nullable linkage fields.
type FeatureAuditLog struct{ ent.Schema }

func (FeatureAuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:    "feature_audit_logs",
			OnDelete: entsql.Restrict,
			Checks: map[string]string{
				"action_check":      "action <> ''",
				"entity_type_check": "entity_type <> ''",
				"actor_type_check":  "actor_type IN ('user', 'system', 'admin', 'ai')",
			},
		},
	}
}

func (FeatureAuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{UUIDMixin{}, ImmutableTimeMixin{}}
}

func (FeatureAuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.String("action").NotEmpty().Immutable(),
		field.String("entity_type").NotEmpty().Immutable(),
		field.UUID("entity_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.String("request_id").Optional().Nillable().Immutable(),
		field.Enum("actor_type").Values("user", "system", "admin", "ai").Immutable(),
		field.UUID("actor_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.JSON("metadata", map[string]any{}).Default(emptyFeatureAuditMetadata).Immutable(),
	}
}

// emptyFeatureAuditMetadata allocates a fresh JSON object for each Ent create
// operation; audit metadata must not share mutable map state between records.
func emptyFeatureAuditMetadata() map[string]any { return map[string]any{} }

func (FeatureAuditLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("feature_audit_logs").
			Field("user_id").
			Unique(),
	}
}

func (FeatureAuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "created_at").
			Annotations(entsql.DescColumns("created_at")),
		index.Fields("entity_type", "entity_id"),
	}
}
