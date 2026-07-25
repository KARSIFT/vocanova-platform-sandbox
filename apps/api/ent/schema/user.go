package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct{ ent.Schema }

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "users"}}
}
func (User) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").Optional().Nillable(),
		field.String("display_name").Optional().Nillable(),
		field.String("avatar_url").Optional().Nillable(),
		field.Enum("status").Values("active", "disabled", "deleted").Default("active"),
		field.Enum("onboarding_status").Values("not_started", "in_progress", "completed").Default("not_started"),
		field.Time("email_verified_at").Optional().Nillable(),
		field.Time("last_login_at").Optional().Nillable(),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("external_identities", ExternalIdentity.Type),
		edge.To("sessions", Session.Type),
		edge.To("magic_links", MagicLink.Type),
		edge.To("user_words", UserWord.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{index.Fields("email").Annotations(entsql.IndexWhere("email IS NOT NULL AND deleted_at IS NULL"))}
}
