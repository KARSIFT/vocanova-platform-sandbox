package migrations_test

import (
	"os"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	entschema "github.com/KARSIFT/vocanova-platform/apps/api/ent/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureAuditLogEntSchemaMatchesCommittedMigration(t *testing.T) {
	schema := entschema.FeatureAuditLog{}
	mixins := schema.Mixin()
	require.Len(t, mixins, 2)
	assert.IsType(t, entschema.UUIDMixin{}, mixins[0])
	assert.IsType(t, entschema.ImmutableTimeMixin{}, mixins[1])

	annotations := schema.Annotations()
	require.Len(t, annotations, 1)
	annotation, ok := annotations[0].(entsql.Annotation)
	require.True(t, ok)
	assert.Equal(t, "feature_audit_logs", annotation.Table)
	assert.Equal(t, entsql.Restrict, annotation.OnDelete)
	assert.Equal(t, map[string]string{
		"action_check":      "action <> ''",
		"entity_type_check": "entity_type <> ''",
		"actor_type_check":  "actor_type IN ('user', 'system', 'admin', 'ai')",
	}, annotation.Checks)

	fields := make(map[string]*field.Descriptor)
	for _, item := range schema.Fields() {
		descriptor := item.Descriptor()
		require.NoError(t, descriptor.Err)
		fields[descriptor.Name] = descriptor
	}
	require.ElementsMatch(t, []string{
		"user_id", "action", "entity_type", "entity_id", "request_id",
		"actor_type", "actor_id", "metadata",
	}, featureAuditFieldNames(fields))

	for _, name := range []string{"user_id", "entity_id", "request_id", "actor_id"} {
		assert.True(t, fields[name].Optional, name)
		assert.True(t, fields[name].Nillable, name)
	}
	for _, name := range []string{"action", "entity_type", "actor_type", "metadata"} {
		assert.False(t, fields[name].Optional, name)
	}
	assert.Equal(t, field.TypeUUID, fields["user_id"].Info.Type)
	assert.Equal(t, field.TypeString, fields["action"].Info.Type)
	assert.Equal(t, field.TypeString, fields["entity_type"].Info.Type)
	assert.Equal(t, field.TypeUUID, fields["entity_id"].Info.Type)
	assert.Equal(t, field.TypeString, fields["request_id"].Info.Type)
	assert.Equal(t, field.TypeEnum, fields["actor_type"].Info.Type)
	assert.Equal(t, field.TypeUUID, fields["actor_id"].Info.Type)
	assert.Equal(t, field.TypeJSON, fields["metadata"].Info.Type)
	for name, descriptor := range fields {
		assert.True(t, descriptor.Immutable, name)
	}
	metadataDefault, ok := fields["metadata"].Default.(func() map[string]any)
	require.True(t, ok, "metadata must allocate a fresh default object")
	firstMetadata, secondMetadata := metadataDefault(), metadataDefault()
	firstMetadata["isolation"] = true
	assert.Empty(t, secondMetadata, "metadata defaults must not share mutable map state")
	assert.ElementsMatch(t, []string{"user", "system", "admin", "ai"}, featureAuditEnumValues(fields["actor_type"]))

	timestamps := entschema.ImmutableTimeMixin{}.Fields()
	require.Len(t, timestamps, 2)
	for _, item := range timestamps {
		descriptor := item.Descriptor()
		require.NoError(t, descriptor.Err)
		assert.True(t, descriptor.Immutable, descriptor.Name)
		assert.Nil(t, descriptor.UpdateDefault, descriptor.Name)
	}

	edges := schema.Edges()
	require.Len(t, edges, 1)
	edge := edges[0].Descriptor()
	assert.Equal(t, "user", edge.Name)
	assert.Equal(t, "user_id", edge.Field)
	assert.True(t, edge.Unique)
	assert.False(t, edge.Required, "de-identified audit rows must allow a null user")
	userEdges := entschema.User{}.Edges()
	require.Contains(t, featureAuditEdgeNames(userEdges), "feature_audit_logs")

	indexes := schema.Indexes()
	require.Len(t, indexes, 2)
	first := indexes[0].Descriptor()
	assert.Equal(t, []string{"user_id", "created_at"}, first.Fields)
	require.Len(t, first.Annotations, 1)
	descending, ok := first.Annotations[0].(*entsql.IndexAnnotation)
	require.True(t, ok)
	assert.True(t, descending.DescColumns["created_at"])
	assert.Equal(t, []string{"entity_type", "entity_id"}, indexes[1].Descriptor().Fields)

	migration, err := os.ReadFile("20260908020000_voc1352_feature_audit_logs.sql")
	require.NoError(t, err)
	text := string(migration)
	for _, invariant := range []string{
		"CREATE TABLE feature_audit_logs",
		"user_id uuid REFERENCES users(id) ON DELETE RESTRICT",
		"action text NOT NULL CHECK (action <> '')",
		"entity_type text NOT NULL CHECK (entity_type <> '')",
		"entity_id uuid",
		"request_id text",
		"actor_type text NOT NULL CHECK (actor_type IN ('user', 'system', 'admin', 'ai'))",
		"actor_id uuid",
		"metadata jsonb NOT NULL DEFAULT '{}'::jsonb",
		"ON feature_audit_logs (user_id, created_at DESC)",
		"ON feature_audit_logs (entity_type, entity_id)",
	} {
		assert.Contains(t, text, invariant)
	}
}

func featureAuditFieldNames(fields map[string]*field.Descriptor) []string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	return keys
}

func featureAuditEnumValues(descriptor *field.Descriptor) []string {
	values := make([]string, 0, len(descriptor.Enums))
	for _, value := range descriptor.Enums {
		values = append(values, value.V)
	}
	return values
}

func featureAuditEdgeNames(edges []ent.Edge) []string {
	names := make([]string, 0, len(edges))
	for _, item := range edges {
		names = append(names, item.Descriptor().Name)
	}
	return names
}
