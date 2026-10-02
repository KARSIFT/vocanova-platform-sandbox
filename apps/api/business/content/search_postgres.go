package content

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// The total and page share one PostgreSQL snapshot. EXISTS avoids duplicate
// meanings when several situations match a category. strpos treats '%' and '_'
// literally, rather than turning learner input into SQL wildcard patterns.
func (r *PostgreSQLRepository) SearchMeanings(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	req, cursor, err := prepareSearch(req)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `WITH matches AS (
	 SELECT m.id AS meaning_id, cw.id AS word_id, cw.text, cw.normalized_text,
	 m.part_of_speech, m.short_definition,
	 COALESCE(m.difficulty_level, cw.difficulty_level, 'unknown') AS difficulty_level
	 FROM word_meanings m JOIN canonical_words cw ON cw.id=m.word_id
	 WHERE m.status='active' AND cw.status='active'
	 AND ($1='' OR strpos(lower(cw.text),$1)>0 OR strpos(lower(m.short_definition),$1)>0 OR strpos(lower(COALESCE(m.learner_definition,'')),$1)>0)
	 AND ($2='' OR EXISTS (SELECT 1 FROM journey_words jw JOIN journey_situations js ON js.id=jw.journey_situation_id WHERE jw.meaning_id=m.id AND js.status='active' AND js.category=$2))
	 AND ($7='' OR
 ($7='known' AND EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=$8 AND k.meaning_id=m.id AND k.self_reported_known)) OR
 ($7='saved' AND EXISTS(SELECT 1 FROM user_words uw WHERE uw.user_id=$8 AND uw.meaning_id=m.id AND uw.deleted_at IS NULL)) OR
 ($7='unexplored' AND NOT EXISTS(SELECT 1 FROM user_word_knowledge k WHERE k.user_id=$8 AND k.meaning_id=m.id AND k.self_reported_known) AND NOT EXISTS(SELECT 1 FROM user_words uw WHERE uw.user_id=$8 AND uw.meaning_id=m.id AND uw.deleted_at IS NULL)))
 AND ($3='' OR COALESCE(m.difficulty_level,cw.difficulty_level,'unknown')=$3)
	), page AS (
	 SELECT * FROM matches WHERE ($4::uuid='00000000-0000-0000-0000-000000000000'::uuid OR (normalized_text COLLATE "C",meaning_id)>($5::text COLLATE "C",$4::uuid))
	 ORDER BY normalized_text COLLATE "C",meaning_id LIMIT $6
	)
	SELECT totals.total_count,page.meaning_id,page.word_id,page.text,page.normalized_text,page.part_of_speech,page.short_definition,page.difficulty_level
	FROM (SELECT count(*) AS total_count FROM matches) totals LEFT JOIN page ON true
	ORDER BY page.normalized_text COLLATE "C",page.meaning_id`, req.Query, req.Category, req.Level, cursor.ID, cursor.Text, req.Limit+1, req.Knowledge, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("search meanings: %w", err)
	}
	defer rows.Close()
	resp := &SearchResponse{Items: []SearchMeaning{}}
	for rows.Next() {
		var meaningID, wordID uuid.NullUUID
		var text, normalized, pos, definition, level sql.NullString
		if err := rows.Scan(&resp.TotalCount, &meaningID, &wordID, &text, &normalized, &pos, &definition, &level); err != nil {
			return nil, fmt.Errorf("scan search meaning: %w", err)
		}
		if !meaningID.Valid {
			continue
		}
		resp.Items = append(resp.Items, SearchMeaning{MeaningSummary: MeaningSummary{MeaningID: meaningID.UUID, WordID: wordID.UUID, WordText: text.String, WordSlug: wordSlug(normalized.String), PartOfSpeech: pos.String, ShortDefinition: definition.String}, DifficultyLevel: level.String, normalizedText: normalized.String})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search meaning rows: %w", err)
	}
	if len(resp.Items) > req.Limit {
		resp.Items = resp.Items[:req.Limit]
		resp.NextCursor = searchNextCursor(req, resp.Items[len(resp.Items)-1])
	}
	return resp, nil
}
