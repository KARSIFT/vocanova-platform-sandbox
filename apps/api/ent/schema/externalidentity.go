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

type ExternalIdentity struct{ ent.Schema }

func (ExternalIdentity) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "external_identities"}}
}
func (ExternalIdentity) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }
func (ExternalIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("provider").Values("google", "email"),
		field.String("provider_subject").NotEmpty(),
		field.String("provider_email").Optional().Nillable(),
		field.Bool("provider_email_verified").Default(false),
		field.Time("deleted_at").Optional().Nillable(),
	}
}
func (ExternalIdentity) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("external_identities").Field("user_id").Unique().Required()}
}
func (ExternalIdentity) Indexes() []ent.Index {
	return []ent.Index{index.Fields("provider", "provider_subject").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL"))}
}
