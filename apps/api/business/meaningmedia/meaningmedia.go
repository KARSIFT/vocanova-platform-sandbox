// Package meaningmedia publishes approved illustrations for exact canonical meanings.
// Original files stay private; only fixed-origin immutable WebP assets are public.
package meaningmedia

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const PublicOrigin = "https://images.vocanova.site"
const MaxBatch = 100
const MaxImport = 25000

// Media is optional to the main learning read; pool waits and SQL execution must
// not consume its full request budget. A shorter caller deadline still wins.
const readTimeout = time.Second

// ImportBatchSize bounds each database statement; the full manifest remains
// atomic. Three statements per batch replace three network round trips per row.
const ImportBatchSize = 500
const publicationRecords = `jsonb_to_recordset($1::jsonb) AS p("meaningId" uuid,"meaningRevision" text,sha256 text,alt text,width integer,height integer)`

var ErrInvalid = errors.New("invalid meaning image request")
var ErrConflict = errors.New("meaning image publication conflicts with existing metadata")

type Image struct {
	MeaningID       string `json:"meaningId"`
	MeaningRevision string `json:"meaningRevision"`
	SHA256          string `json:"sha256"`
	ImageURL        string `json:"imageUrl"`
	Alt             string `json:"alt"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
}

type Publication struct {
	MeaningID       string `json:"meaningId"`
	MeaningRevision string `json:"meaningRevision"`
	SHA256          string `json:"sha256"`
	Alt             string `json:"alt"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	Approved        bool   `json:"approved"`
}
type Manifest struct {
	SchemaVersion int           `json:"schemaVersion"`
	Images        []Publication `json:"images"`
}

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func hash(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && value == strings.ToLower(value)
}
func (p Publication) Validate() error {
	id, err := uuid.Parse(p.MeaningID)
	if err != nil || id == uuid.Nil || id.String() != p.MeaningID || !hash(p.MeaningRevision) || !hash(p.SHA256) || !p.Approved || p.Width < 1 || p.Width > 960 || p.Height < 1 || p.Height > 640 || !utf8.ValidString(p.Alt) || strings.TrimSpace(p.Alt) != p.Alt || p.Alt == "" || utf8.RuneCountInString(p.Alt) > 500 || strings.ContainsAny(p.Alt, "\x00\r\n") {
		return ErrInvalid
	}
	return nil
}
func (p Publication) ObjectKey() string {
	return "meanings/" + p.MeaningID + "/" + p.MeaningRevision + "/" + p.SHA256 + ".webp"
}
func (p Publication) image() Image {
	return Image{MeaningID: p.MeaningID, MeaningRevision: p.MeaningRevision, SHA256: p.SHA256, ImageURL: PublicOrigin + "/" + p.ObjectKey(), Alt: p.Alt, Width: p.Width, Height: p.Height}
}

// Read makes one bounded catalog query, omitting missing, unapproved and stale
// publications. Caller authentication is enforced by the HTTP route. No fetch is
// performed against the image origin or any client-provided URL.
func (s *Service) Read(ctx context.Context, ids []string) ([]Image, error) {
	if len(ids) < 1 || len(ids) > MaxBatch {
		return nil, ErrInvalid
	}
	unique := make([]string, 0, len(ids))
	seen := map[string]bool{}
	for _, raw := range ids {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil || raw != id.String() {
			return nil, ErrInvalid
		}
		if !seen[raw] {
			seen[raw] = true
			unique = append(unique, raw)
		}
	}
	out := make([]Image, 0, len(unique))
	if s == nil || s.db == nil {
		return out, nil
	}
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()
	rows, err := s.db.QueryContext(ctx, `SELECT a.meaning_id,a.meaning_revision,a.sha256,a.alt,a.width,a.height,c.text,m.part_of_speech,m.short_definition
 FROM meaning_image_assets a JOIN word_meanings m ON m.id=a.meaning_id JOIN canonical_words c ON c.id=m.word_id
 WHERE a.meaning_id=ANY($1::uuid[]) AND a.approved=true AND m.status='active' AND c.status='active'
 ORDER BY a.meaning_id`, pq.Array(unique))
	if err != nil {
		return nil, fmt.Errorf("read meaning images: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p Publication
		var word, pos, definition string
		if err := rows.Scan(&p.MeaningID, &p.MeaningRevision, &p.SHA256, &p.Alt, &p.Width, &p.Height, &word, &pos, &definition); err != nil {
			return nil, err
		}
		p.Approved = true
		if p.Validate() != nil || meaningidentity.Revision(p.MeaningID, word, pos, definition) != p.MeaningRevision {
			continue
		}
		out = append(out, p.image())
	}
	return out, rows.Err()
}

// Import is operator-only: the CLI supplies a privileged DB connection and an
// approved manifest after verifying the uploaded bytes. There is no HTTP write
// route. The transaction locks canonical rows so a concurrent content edit cannot
// publish an image against an outdated revision. Exact retries are no-ops;
// replacing a publication requires an explicit later editorial operation.
func (s *Service) Import(ctx context.Context, manifest Manifest) (int, error) {
	if s == nil || s.db == nil || manifest.SchemaVersion != 1 || len(manifest.Images) < 1 || len(manifest.Images) > MaxImport {
		return 0, ErrInvalid
	}
	seen := map[string]bool{}
	for _, p := range manifest.Images {
		if err := p.Validate(); err != nil {
			return 0, err
		}
		key := p.MeaningID
		if seen[key] {
			return 0, ErrInvalid
		}
		seen[key] = true
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	// Stable lock order avoids deadlocks between overlapping operator batches.
	publications := append([]Publication(nil), manifest.Images...)
	sort.Slice(publications, func(i, j int) bool { return publications[i].MeaningID < publications[j].MeaningID })
	inserted := 0
	for start := 0; start < len(publications); start += ImportBatchSize {
		batch := publications[start:min(start+ImportBatchSize, len(publications))]
		payload, err := json.Marshal(batch)
		if err != nil {
			return 0, err
		}
		rows, err := tx.QueryContext(ctx, `SELECT m.id,c.text,m.part_of_speech,m.short_definition FROM `+publicationRecords+` JOIN word_meanings m ON m.id=p."meaningId" JOIN canonical_words c ON c.id=m.word_id WHERE m.status='active' AND c.status='active' ORDER BY m.id FOR SHARE OF m,c`, string(payload))
		if err != nil {
			return 0, err
		}
		revisions := make(map[string]string, len(batch))
		for rows.Next() {
			var id, word, pos, definition string
			if err := rows.Scan(&id, &word, &pos, &definition); err != nil {
				rows.Close()
				return 0, err
			}
			revisions[id] = meaningidentity.Revision(id, word, pos, definition)
		}
		readErr := rows.Err()
		rows.Close()
		if readErr != nil {
			return 0, readErr
		}
		for _, p := range batch {
			if revisions[p.MeaningID] != p.MeaningRevision {
				return 0, fmt.Errorf("meaning %s is absent or its revision changed: %w", p.MeaningID, ErrInvalid)
			}
		}
		result, err := tx.ExecContext(ctx, `INSERT INTO meaning_image_assets(id,meaning_id,meaning_revision,sha256,alt,width,height,approved,created_at) SELECT gen_random_uuid(),p."meaningId",p."meaningRevision",p.sha256,p.alt,p.width,p.height,true,CURRENT_TIMESTAMP FROM `+publicationRecords+` ON CONFLICT(meaning_id,meaning_revision) DO NOTHING`, string(payload))
		if err != nil {
			return 0, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		inserted += int(n)
		var matching int
		err = tx.QueryRowContext(ctx, `SELECT count(*) FROM `+publicationRecords+` JOIN meaning_image_assets a ON a.meaning_id=p."meaningId" AND a.meaning_revision=p."meaningRevision" WHERE a.sha256=p.sha256 AND a.alt=p.alt AND a.width=p.width AND a.height=p.height AND a.approved=true`, string(payload)).Scan(&matching)
		if err != nil {
			return 0, err
		}
		if matching != len(batch) {
			return 0, ErrConflict
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return inserted, nil
}
