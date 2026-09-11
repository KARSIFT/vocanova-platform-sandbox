package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PasswordRegistrationLink is an unauthenticated, expiring proof record. It
// intentionally has no user edge because an account does not exist yet.
type PasswordRegistrationLink struct{ ent.Schema }

func (PasswordRegistrationLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "password_registration_links"}}
}
func (PasswordRegistrationLink) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (PasswordRegistrationLink) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty(),
		field.String("display_name").Optional(),
		field.String("password_hash").Sensitive().NotEmpty(),
		field.Bytes("token_hash").Unique().Immutable(),
		field.String("environment").NotEmpty(),
		field.Time("created_at").Immutable(),
		field.Time("expires_at").Immutable(),
		field.Time("consumed_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}
func (PasswordRegistrationLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at").Annotations(entsql.IndexWhere("consumed_at IS NULL AND revoked_at IS NULL")),
		index.Fields("email"),
	}
}
