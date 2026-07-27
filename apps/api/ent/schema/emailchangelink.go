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

// EmailChangeLink is the per-user, single-use email-change verification
// token. VOC-031-T03 / DOC-06 §6 / VOC-031-D05. Mirrors magic_links in
// every discipline that matters (32-random-byte token, only its SHA-256
// hash persisted, 15-minute expiry, single-use, environment-scoped),
// with three deliberate differences because this is not a login
// mechanism:
//
//  1. user_id is NOT NULL: requesting an email change requires an
//     already-authenticated session (no anonymous flow).
//  2. new_email replaces email: the destination the requester wants
//     to switch to. The current sign-in address stays on the users
//     row and is overwritten atomically at confirm time, with
//     users_active_email_key as the authoritative uniqueness guard
//     (VOC-031-R02).
//  3. no email-uniqueness constraint here: the same new_email can be
//     requested by multiple learners; only one confirm will win, the
//     rest receive a stable conflict response.
type EmailChangeLink struct{ ent.Schema }

func (EmailChangeLink) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "email_change_links"}}
}

func (EmailChangeLink) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}} }

func (EmailChangeLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.String("new_email").NotEmpty().Immutable(),
		field.Bytes("token_hash").Unique().Immutable(),
		field.String("environment").NotEmpty().Immutable(),
		field.Time("created_at").Immutable(),
		field.Time("expires_at").Immutable(),
		field.Time("consumed_at").Optional().Nillable(),
		field.Time("revoked_at").Optional().Nillable(),
	}
}

func (EmailChangeLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("email_change_links").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (EmailChangeLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at").
			Annotations(entsql.IndexWhere("consumed_at IS NULL AND revoked_at IS NULL")),
		index.Fields("user_id"),
	}
}
