package knowledgecatalog

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/google/uuid"
)

//go:embed catalog.json.gz
var catalogCompressed []byte

type CatalogEntry struct {
	Rank           int       `json:"rank"`
	SourceRank     int       `json:"source_rank"`
	Count          int64     `json:"count"`
	Word           string    `json:"word"`
	WordID         uuid.UUID `json:"word_id"`
	MeaningID      uuid.UUID `json:"meaning_id"`
	PartOfSpeech   string    `json:"part_of_speech"`
	Definition     string    `json:"definition"`
	Example        string    `json:"example"`
	Source         string    `json:"source"`
	SourceSenseKey string    `json:"source_sense_key"`
	ExampleID      uuid.UUID `json:"example_id"`
	TagCount       *int      `json:"tag_count"`
}
type catalogData struct {
	Version        string         `json:"version"`
	WordNetLicense string         `json:"wordnet_license"`
	Items          []CatalogEntry `json:"items"`
}

var catalogOnce sync.Once
var loadedCatalog catalogData
var catalogLoadError error

func loadCatalog() (catalogData, error) {
	catalogOnce.Do(func() {
		reader, err := gzip.NewReader(bytes.NewReader(catalogCompressed))
		if err != nil {
			catalogLoadError = err
			return
		}
		defer reader.Close()
		raw, err := io.ReadAll(io.LimitReader(reader, 8*1024*1024+1))
		if err != nil {
			catalogLoadError = err
			return
		}
		if len(raw) > 8*1024*1024 {
			catalogLoadError = errors.New("catalog exceeds bound")
			return
		}
		if err = json.Unmarshal(raw, &loadedCatalog); err != nil {
			catalogLoadError = err
			return
		}
		if loadedCatalog.Version != CatalogVersion || len(loadedCatalog.Items) != TotalWords || !strings.Contains(loadedCatalog.WordNetLicense, "Copyright 2006 by Princeton University") {
			catalogLoadError = errors.New("catalog metadata or count invalid")
			return
		}
		words := map[string]bool{}
		wordIDs := map[uuid.UUID]bool{}
		meanings := map[uuid.UUID]bool{}
		for i := range loadedCatalog.Items {
			e := &loadedCatalog.Items[i]
			if e.Rank != i+1 || e.Word == "" || strings.Contains(e.Word, " ") || e.WordID == uuid.Nil || e.MeaningID == uuid.Nil || e.Definition == "" || e.SourceRank < 1 || e.Count < 1 || words[e.Word] || wordIDs[e.WordID] || meanings[e.MeaningID] {
				catalogLoadError = fmt.Errorf("invalid catalog entry %d", i+1)
				return
			}
			if e.Source == "WordNet 3.0" && e.Example != "" {
				catalogLoadError = fmt.Errorf("unreviewed source example for %s", e.Word)
				return
			}
			words[e.Word] = true
			wordIDs[e.WordID] = true
			meanings[e.MeaningID] = true
			if e.Example != "" {
				e.ExampleID = uuid.NewSHA1(e.MeaningID, []byte("source-example-v1"))
			}
		}
	})
	return loadedCatalog, catalogLoadError
}

// Seed installs the public catalog in the caller's canonical seed transaction.
// Existing authored rows are never overwritten, and no learner rows are touched.
func Seed(ctx context.Context, tx *sql.Tx) error {
	data, err := loadCatalog()
	if err != nil {
		return err
	}
	raw, err := json.Marshal(data.Items)
	if err != nil {
		return err
	}
	queries := []string{
		`INSERT INTO canonical_words(id,text,normalized_text,word_type,language_code,status,difficulty_level,frequency_rank,created_at,updated_at)
 SELECT word_id,CASE WHEN word='i' THEN 'I' ELSE word END,word,'word','en','active','unknown',source_rank,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
 FROM jsonb_to_recordset($1::jsonb) AS x(word_id uuid,word text,source_rank integer)
 ON CONFLICT(id) DO NOTHING`,
		`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,difficulty_level,created_at,updated_at)
 SELECT meaning_id,word_id,part_of_speech,definition,1,'active','unknown',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
 FROM jsonb_to_recordset($1::jsonb) AS x(meaning_id uuid,word_id uuid,part_of_speech text,definition text,source text)
 WHERE source<>'Vocanova authored'
 ON CONFLICT(id) DO UPDATE SET part_of_speech=EXCLUDED.part_of_speech,short_definition=EXCLUDED.short_definition,updated_at=CURRENT_TIMESTAMP
 WHERE word_meanings.part_of_speech IS DISTINCT FROM EXCLUDED.part_of_speech OR word_meanings.short_definition IS DISTINCT FROM EXCLUDED.short_definition`,
		`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at)
 SELECT example_id,meaning_id,example,1,'active',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
 FROM jsonb_to_recordset($1::jsonb) AS x(example_id uuid,meaning_id uuid,example text,source text)
 WHERE example<>'' AND source<>'Vocanova authored' ON CONFLICT(id) DO UPDATE SET example_text=EXCLUDED.example_text,updated_at=CURRENT_TIMESTAMP
 WHERE word_examples.example_text IS DISTINCT FROM EXCLUDED.example_text`,
		`INSERT INTO knowledge_catalog_entries(rank,word_id,meaning_id,source_rank,source_count,source_sense_key,source_name,catalog_version)
 SELECT rank,word_id,meaning_id,source_rank,count,source_sense_key,source,'english-10000-v1'
 FROM jsonb_to_recordset($1::jsonb) AS x(rank integer,word_id uuid,meaning_id uuid,source_rank integer,count bigint,source_sense_key text,source text)
 ON CONFLICT(rank) DO UPDATE SET word_id=EXCLUDED.word_id,meaning_id=EXCLUDED.meaning_id,source_rank=EXCLUDED.source_rank,source_count=EXCLUDED.source_count,source_sense_key=EXCLUDED.source_sense_key,source_name=EXCLUDED.source_name,catalog_version=EXCLUDED.catalog_version`,
	}
	for _, q := range queries {
		if _, err = tx.ExecContext(ctx, q, string(raw)); err != nil {
			return fmt.Errorf("seed knowledge catalog: %w", err)
		}
	}
	return nil
}
