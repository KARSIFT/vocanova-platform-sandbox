package meaningmedia

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func mediaDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("VOCANOVA_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("VOCANOVA_TEST_POSTGRES_DSN unset")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		var err error
		dsn, err = pq.ParseURL(dsn)
		require.NoError(t, err)
	}
	admin, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	schema := "media_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.Exec("CREATE SCHEMA " + schema)
	require.NoError(t, err)
	db, err := sql.Open("postgres", dsn+" search_path="+schema)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
		_, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE")
		require.NoError(t, err)
		admin.Close()
	})
	paths, err := filepath.Glob("../../migrations/*.sql")
	require.NoError(t, err)
	for _, p := range paths {
		raw, err := os.ReadFile(p)
		require.NoError(t, err)
		_, err = db.Exec(string(raw))
		require.NoError(t, err, p)
	}
	return db
}
func TestMeaningImagesPostgreSQLPublicationRevisionAndIdempotence(t *testing.T) {
	db := mediaDB(t)
	s := NewService(db)
	ctx := t.Context()
	p := examplePublication()
	word := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO canonical_words(id,text,normalized_text,language_code,word_type,status,created_at,updated_at) VALUES($1,'light','light','en','word','active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, word)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'adjective','Having little weight.',1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, p.MeaningID, word)
	require.NoError(t, err)
	manifest := Manifest{SchemaVersion: 1, Images: []Publication{p}}
	n, err := s.Import(ctx, manifest)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = s.Import(ctx, manifest)
	require.NoError(t, err)
	require.Zero(t, n)
	images, err := s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Equal(t, []Image{p.image()}, images)
	changed := p
	changed.SHA256 = strings.Repeat("b", 64)
	_, err = s.Import(ctx, Manifest{SchemaVersion: 1, Images: []Publication{changed}})
	require.ErrorIs(t, err, ErrConflict)
	images, err = s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Equal(t, []Image{p.image()}, images)
	// The word identity alone is insufficient once its grammatical meaning changes.
	_, err = db.ExecContext(ctx, `UPDATE word_meanings SET part_of_speech='verb',short_definition='To make a place bright.' WHERE id=$1`, p.MeaningID)
	require.NoError(t, err)
	images, err = s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Empty(t, images)
	_, err = s.Import(ctx, manifest)
	require.ErrorIs(t, err, ErrInvalid)
	fresh := p
	fresh.MeaningRevision = meaningidentity.Revision(p.MeaningID, "light", "verb", "To make a place bright.")
	fresh.Alt = "A person switching on a lamp in a dark room."
	n, err = s.Import(ctx, Manifest{SchemaVersion: 1, Images: []Publication{fresh}})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	images, err = s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Equal(t, []Image{fresh.image()}, images)
	_, err = db.ExecContext(ctx, `UPDATE meaning_image_assets SET approved=false WHERE meaning_revision=$1`, fresh.MeaningRevision)
	require.NoError(t, err)
	images, err = s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Empty(t, images)
	_, err = db.ExecContext(ctx, `UPDATE meaning_image_assets SET approved=true WHERE meaning_revision=$1;`, fresh.MeaningRevision)
	require.NoError(t, err)
	// A late conflicting publication rolls back earlier new rows in the same batch.
	second := fresh
	second.MeaningID = uuid.NewString()
	second.MeaningRevision = meaningidentity.Revision(second.MeaningID, "light", "verb", "To make a place bright.")
	_, err = db.ExecContext(ctx, `INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'verb','To make a place bright.',2,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, second.MeaningID, word)
	require.NoError(t, err)
	conflict := fresh
	conflict.SHA256 = strings.Repeat("b", 64)
	_, err = s.Import(ctx, Manifest{SchemaVersion: 1, Images: []Publication{second, conflict}})
	require.ErrorIs(t, err, ErrConflict)
	images, err = s.Read(ctx, []string{second.MeaningID})
	require.NoError(t, err)
	require.Empty(t, images)
	// Concurrent operator retries converge to exactly one immutable publication.
	var wg sync.WaitGroup
	counts := make(chan int, 8)
	errs := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n, err := s.Import(ctx, Manifest{SchemaVersion: 1, Images: []Publication{second}})
			counts <- n
			errs <- err
		}()
	}
	wg.Wait()
	close(counts)
	close(errs)
	total := 0
	for n := range counts {
		total += n
	}
	for err := range errs {
		require.NoError(t, err)
	}
	require.Equal(t, 1, total)
	images, err = s.Read(ctx, []string{second.MeaningID})
	require.NoError(t, err)
	require.Equal(t, []Image{second.image()}, images)
	_, err = db.ExecContext(ctx, `UPDATE canonical_words SET status='archived' WHERE id=$1`, word)
	require.NoError(t, err)
	images, err = s.Read(ctx, []string{p.MeaningID})
	require.NoError(t, err)
	require.Empty(t, images)
}

func TestMeaningImagesPostgreSQLBatchedImportRollsBackAcrossBatchBoundary(t *testing.T) {
	db := mediaDB(t)
	ctx := t.Context()
	s := NewService(db)
	word := uuid.New()
	_, err := db.ExecContext(ctx, `INSERT INTO canonical_words(id,text,normalized_text,language_code,word_type,status,created_at,updated_at) VALUES($1,'light','light','en','word','active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)`, word)
	require.NoError(t, err)
	manifest := Manifest{SchemaVersion: 1}
	for range ImportBatchSize + 1 {
		manifest.Images = append(manifest.Images, examplePublication())
	}
	sort.Slice(manifest.Images, func(i, j int) bool { return manifest.Images[i].MeaningID < manifest.Images[j].MeaningID })
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("word_meanings", "id", "word_id", "part_of_speech", "short_definition", "meaning_order", "status", "created_at", "updated_at"))
	require.NoError(t, err)
	now := time.Now()
	for i, p := range manifest.Images {
		_, err = stmt.ExecContext(ctx, p.MeaningID, word, "adjective", "Having little weight.", i+1, "active", now, now)
		require.NoError(t, err)
	}
	_, err = stmt.ExecContext(ctx)
	require.NoError(t, err)
	require.NoError(t, stmt.Close())
	require.NoError(t, tx.Commit())
	// Publish only the last meaning with conflicting approved bytes. The later
	// conflict must roll back 500 valid new rows from the preceding SQL batch.
	conflict := manifest.Images[len(manifest.Images)-1]
	conflict.SHA256 = strings.Repeat("b", 64)
	n, err := s.Import(ctx, Manifest{SchemaVersion: 1, Images: []Publication{conflict}})
	require.NoError(t, err)
	require.Equal(t, 1, n)
	n, err = s.Import(ctx, manifest)
	require.ErrorIs(t, err, ErrConflict)
	require.Zero(t, n)
	var count int
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM meaning_image_assets`).Scan(&count))
	require.Equal(t, 1, count)
	manifest.Images[len(manifest.Images)-1] = conflict
	n, err = s.Import(ctx, manifest)
	require.NoError(t, err)
	require.Equal(t, ImportBatchSize, n)
	n, err = s.Import(ctx, manifest)
	require.NoError(t, err)
	require.Zero(t, n)
	require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM meaning_image_assets`).Scan(&count))
	require.Equal(t, ImportBatchSize+1, count)
}
