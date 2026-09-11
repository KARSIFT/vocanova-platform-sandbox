package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// PasswordCredential stores only an Argon2id encoding. Its primary key is also
// the owning user foreign key, matching password_credentials.user_id.
type PasswordCredential struct{ ent.Schema }

func (PasswordCredential) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "password_credentials"}}
}
func (PasswordCredential) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).StorageKey("user_id").Immutable(),
		field.String("password_hash").Sensitive().NotEmpty(),
		field.Time("created_at").Immutable(),
		field.Time("updated_at"),
	}
}
func (PasswordCredential) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("password_credential").Field("id").Unique().Required()}
}
