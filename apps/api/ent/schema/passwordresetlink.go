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

type PasswordResetLink struct{ ent.Schema }

func (PasswordResetLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "password_reset_links"}}
}
func (PasswordResetLink) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (PasswordResetLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.String("email_at_issue").NotEmpty().Immutable(),
		field.Bytes("token_hash").Unique().Immutable(),
		field.String("environment").NotEmpty(),
		field.Time("created_at").Immutable(),
		field.Time("expires_at").Immutable(),
		field.Time("consumed_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}
func (PasswordResetLink) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("password_reset_links").Field("user_id").Unique().Required()}
}
func (PasswordResetLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at").Annotations(entsql.IndexWhere("consumed_at IS NULL AND revoked_at IS NULL")),
		index.Fields("user_id"),
	}
}
