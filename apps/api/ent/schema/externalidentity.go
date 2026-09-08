package schema

import (
	"errors"
	"unicode"

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
	return []schema.Annotation{
		entsql.Annotation{
			Table: "external_identities",
			Checks: map[string]string{
				"provider_subject_nonblank": authRecordIdentifierNonblankCheck("provider_subject"),
			},
		},
	}
}
func (ExternalIdentity) Mixin() []ent.Mixin { return []ent.Mixin{UUIDMixin{}, TimeMixin{}} }
func (ExternalIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.Enum("provider").Values("google", "email"),
		field.String("provider_subject").NotEmpty().Validate(validateAuthRecordIdentifierNonblank),
		field.String("provider_email").Optional().Nillable(),
		field.Bool("provider_email_verified").Default(false),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

const authRecordIdentifierNonblankPattern = `U&'[^\0009-\000D\0020\0085\00A0\1680\2000-\200A\2028\2029\202F\205F\3000]'`

func authRecordIdentifierNonblankCheck(column string) string {
	return column + " ~ " + authRecordIdentifierNonblankPattern
}

func validateAuthRecordIdentifierNonblank(value string) error {
	for _, character := range value {
		if !unicode.IsSpace(character) {
			return nil
		}
	}
	return errors.New("authentication record identifier must contain a non-whitespace character")
}
func (ExternalIdentity) Edges() []ent.Edge {
	return []ent.Edge{edge.From("user", User.Type).Ref("external_identities").Field("user_id").Unique().Required()}
}
func (ExternalIdentity) Indexes() []ent.Index {
	return []ent.Index{index.Fields("provider", "provider_subject").Unique().Annotations(entsql.IndexWhere("deleted_at IS NULL"))}
}
