package meaningmedia

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func examplePublication() Publication {
	id := uuid.NewString()
	return Publication{MeaningID: id, MeaningRevision: meaningidentity.Revision(id, "light", "adjective", "Having little weight."), SHA256: strings.Repeat("a", 64), Alt: "A child easily lifting a small bag.", Width: 960, Height: 640, Approved: true}
}
func TestPublicationRejectsUnreviewedOrUnsafeMetadata(t *testing.T) {
	valid := examplePublication()
	require.NoError(t, valid.Validate())
	for _, tc := range []struct {
		name   string
		mutate func(*Publication)
	}{
		{"nil identity", func(p *Publication) { p.MeaningID = uuid.Nil.String() }},
		{"identity url", func(p *Publication) { p.MeaningID = "https://example.com" }},
		{"uppercase hash", func(p *Publication) { p.SHA256 = strings.Repeat("A", 64) }},
		{"short revision", func(p *Publication) { p.MeaningRevision = "abc" }},
		{"unapproved", func(p *Publication) { p.Approved = false }},
		{"zero width", func(p *Publication) { p.Width = 0 }},
		{"oversized height", func(p *Publication) { p.Height = 641 }},
		{"empty alt", func(p *Publication) { p.Alt = "" }},
		{"long alt", func(p *Publication) { p.Alt = strings.Repeat("x", 501) }},
		{"control alt", func(p *Publication) { p.Alt = "bag\nroom" }},
		{"padded alt", func(p *Publication) { p.Alt = " image " }},
	} {
		t.Run(tc.name, func(t *testing.T) { p := valid; tc.mutate(&p); require.ErrorIs(t, p.Validate(), ErrInvalid) })
	}
	require.Equal(t, PublicOrigin+"/meanings/"+valid.MeaningID+"/"+valid.MeaningRevision+"/"+valid.SHA256+".webp", valid.image().ImageURL)
	// Preserving source aspect ratio is intentional, not a permission to upscale.
	valid.Width = 640
	valid.Height = 640
	require.NoError(t, valid.Validate())
}
func TestReadUsesOneBoundedQueryAndNeverServesStaleMeaning(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	service := NewService(db)
	p := examplePublication()
	stale := examplePublication()
	mock.ExpectQuery("SELECT a.meaning_id").WillReturnRows(sqlmock.NewRows([]string{"id", "revision", "sha256", "alt", "width", "height", "word", "pos", "definition"}).
		AddRow(p.MeaningID, p.MeaningRevision, p.SHA256, p.Alt, p.Width, p.Height, "light", "adjective", "Having little weight.").
		AddRow(stale.MeaningID, stale.MeaningRevision, stale.SHA256, stale.Alt, stale.Width, stale.Height, "light", "verb", "To make a place bright."))
	images, err := service.Read(t.Context(), []string{p.MeaningID, stale.MeaningID, p.MeaningID})
	require.NoError(t, err)
	require.Equal(t, []Image{p.image()}, images)
	require.NoError(t, mock.ExpectationsWereMet())
	for _, ids := range [][]string{nil, {"bad"}, {uuid.Nil.String()}, strings.Split(strings.Repeat(p.MeaningID+",", 101), ",")} {
		_, err := service.Read(t.Context(), ids)
		require.ErrorIs(t, err, ErrInvalid)
	}
}

func TestOptionalReadStopsWaitingForBusyDatabasePool(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	// Hold the only connection: QueryContext must wait in database/sql itself,
	// rather than receiving an immediate error from a mock query expectation.
	connection, err := db.Conn(t.Context())
	require.NoError(t, err)
	defer connection.Close()
	caller, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()
	started := time.Now()
	images, err := NewService(db).Read(caller, []string{uuid.NewString()})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Empty(t, images)
	require.NoError(t, caller.Err(), "Optional image metadata must stop before exhausting the main request budget")
	require.Less(t, time.Since(started), 2*time.Second)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOptionalReadPreservesShorterCallerDeadline(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxOpenConns(1)
	connection, err := db.Conn(t.Context())
	require.NoError(t, err)
	defer connection.Close()
	caller, cancel := context.WithTimeout(t.Context(), 25*time.Millisecond)
	defer cancel()
	started := time.Now()
	images, err := NewService(db).Read(caller, []string{uuid.NewString()})
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Empty(t, images)
	require.ErrorIs(t, caller.Err(), context.DeadlineExceeded)
	require.Less(t, time.Since(started), 500*time.Millisecond)
	require.NoError(t, mock.ExpectationsWereMet())
}
func TestInvalidManifestCannotStartPublication(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	s := NewService(db)
	p := examplePublication()
	for _, manifest := range []Manifest{{SchemaVersion: 2, Images: []Publication{p}}, {SchemaVersion: 1}, {SchemaVersion: 1, Images: []Publication{p, p}}} {
		_, err := s.Import(t.Context(), manifest)
		require.ErrorIs(t, err, ErrInvalid)
	}
	require.NoError(t, mock.ExpectationsWereMet())
}
