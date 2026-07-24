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

type Session struct{ ent.Schema }

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "sessions"}}
}
func (Session) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Bytes("token_hash").Unique().Immutable(),
		field.Time("created_at").Immutable(),
		field.Time("expires_at").Immutable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}
func (Session) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("sessions").Field("user_id").Unique().Required()}
}
func (Session) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "expires_at"), index.Fields("expires_at").Annotations(entsql.IndexWhere("revoked_at IS NULL"))}
}
