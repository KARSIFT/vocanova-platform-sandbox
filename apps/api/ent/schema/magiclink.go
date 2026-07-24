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

type MagicLink struct{ ent.Schema }

func (MagicLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "magic_links"}}
}
func (MagicLink) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }
func (MagicLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.String("email").NotEmpty().Immutable(),
		field.Bytes("token_hash").Unique().Immutable(),
		field.String("environment").NotEmpty().Immutable(),
		field.Time("created_at").Immutable(),
		field.Time("expires_at").Immutable(),
		field.Time("consumed_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}
func (MagicLink) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("magic_links").Field("user_id").Unique()}
}
func (MagicLink) Indexes() []ent.Index {
	return []ent.Index{index.Fields("expires_at").Annotations(entsql.IndexWhere("consumed_at IS NULL AND revoked_at IS NULL"))}
}
