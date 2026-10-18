package knowledgecatalog

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/revieweligibility"
	"github.com/google/uuid"
)

type PostgreSQLRepository struct{ db *sql.DB }

func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository { return &PostgreSQLRepository{db} }

// latest word is included even when removed: stale tabs must see its revision.
const joins = ` FROM knowledge_catalog_entries kc JOIN canonical_words cw ON cw.id=kc.word_id AND cw.status='active' JOIN word_meanings wm ON wm.id=kc.meaning_id AND wm.word_id=cw.id AND wm.status='active'
 LEFT JOIN user_word_knowledge k ON k.meaning_id=wm.id AND k.user_id=$1
 LEFT JOIN LATERAL (SELECT * FROM user_words WHERE user_id=$1 AND meaning_id=wm.id ORDER BY deleted_at ASC NULLS FIRST,updated_at DESC,id LIMIT 1) uw ON true `

var selected = `SELECT kc.rank,wm.id,cw.id,cw.text,replace(cw.normalized_text,' ','-'),wm.part_of_speech,wm.short_definition,
 COALESCE((SELECT example_text FROM word_examples WHERE meaning_id=wm.id AND status='active' ORDER BY example_order,id LIMIT 1),''),
 COALESCE(k.self_reported_known,false),COALESCE(uw.deleted_at IS NULL AND uw.id IS NOT NULL AND uw.status NOT IN ('ignored','archived'),false),
 COALESCE(` + revieweligibility.Due("uw", "$2", true) + `,false),
 COALESCE(uw.id::text,'')||':'||COALESCE(uw.status,'')||':'||COALESCE(extract(epoch FROM uw.updated_at)::text,'')||':'||COALESCE(extract(epoch FROM uw.deleted_at)::text,'')||':'||COALESCE(extract(epoch FROM uw.next_review_at)::text,'')||':'||COALESCE(uw.review_step::text,'')||':'||COALESCE(uw.total_review_count::text,'')||':'||COALESCE(extract(epoch FROM k.updated_at)::text,'') `

type scanner interface{ Scan(...any) error }

func scanItem(row scanner) (Item, error) {
	var item Item
	var token string
	err := row.Scan(&item.Rank, &item.MeaningID, &item.WordID, &item.WordText, &item.WordSlug, &item.PartOfSpeech, &item.ShortDefinition, &item.ExampleSentence, &item.SelfReportedKnown, &item.Saved, &item.Due, &token)
	sum := sha256.Sum256([]byte(item.MeaningID.String() + ":" + fmt.Sprint(item.SelfReportedKnown) + ":" + token))
	item.Revision = hex.EncodeToString(sum[:])
	return item, err
}
func active(ctx context.Context, tx *sql.Tx, user uuid.UUID, lock bool) error {
	query := `SELECT id FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL`
	if lock {
		query += ` FOR UPDATE`
	}
	var id uuid.UUID
	err := tx.QueryRowContext(ctx, query, user).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
func (r *PostgreSQLRepository) Read(ctx context.Context, user uuid.UUID, band int, now time.Time) (*Map, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if err = active(ctx, tx, user, false); err != nil {
		return nil, err
	}
	out := &Map{Attribution: SourceAttribution(), CatalogVersion: CatalogVersion, BandSize: BandSize, Ordering: ordering(), Bands: []Band{}, CurrentBand: CurrentBand{Number: band, Items: []Item{}}}
	rows, err := tx.QueryContext(ctx, `SELECT (kc.rank-1)/100+1,count(*),count(*) FILTER(WHERE COALESCE(k.self_reported_known,false)),count(*) FILTER(WHERE NOT COALESCE(k.self_reported_known,false) AND uw.id IS NOT NULL AND uw.deleted_at IS NULL AND uw.status NOT IN ('ignored','archived'))`+joins+` GROUP BY (kc.rank-1)/100 ORDER BY (kc.rank-1)/100`, user)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var b Band
		if err = rows.Scan(&b.Number, &b.TotalWords, &b.KnownCount, &b.LearningCount); err != nil {
			rows.Close()
			return nil, err
		}
		b.StartRank = (b.Number-1)*100 + 1
		b.EndRank = b.Number * 100
		b.UnseenCount = b.TotalWords - b.KnownCount - b.LearningCount
		out.TotalWords += b.TotalWords
		out.Bands = append(out.Bands, b)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if out.TotalWords != TotalWords || len(out.Bands) != TotalWords/BandSize {
		return nil, errors.New("knowledge catalog is not fully seeded")
	}
	rows, err = tx.QueryContext(ctx, selected+joins+` WHERE kc.rank BETWEEN $3 AND $4 ORDER BY kc.rank`, user, now, (band-1)*BandSize+1, band*BandSize)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		item, e := scanItem(rows)
		if e != nil {
			rows.Close()
			return nil, e
		}
		out.CurrentBand.Items = append(out.CurrentBand.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}
func readItem(ctx context.Context, tx *sql.Tx, user, meaning uuid.UUID, now time.Time) (Item, error) {
	item, err := scanItem(tx.QueryRowContext(ctx, selected+joins+` WHERE wm.id=$3`, user, now, meaning))
	if errors.Is(err, sql.ErrNoRows) {
		return item, ErrNotFound
	}
	return item, err
}
func result(item Item, decision string) *DecisionResult {
	return &DecisionResult{MeaningID: item.MeaningID, Decision: decision, Revision: item.Revision, SelfReportedKnown: item.SelfReportedKnown, Saved: item.Saved, Due: item.Due}
}
func (r *PostgreSQLRepository) Decide(ctx context.Context, req DecisionRequest, now time.Time) (*DecisionResult, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// Shared order with daily lessons, reviews, assessment and account deletion:
	// active learner first, then user-word rows. Never acquire a learning session.
	if err = active(ctx, tx, req.UserID, true); err != nil {
		return nil, err
	}
	fp := fingerprint(req)
	var prior string
	err = tx.QueryRowContext(ctx, `SELECT fingerprint FROM word_knowledge_actions WHERE user_id=$1 AND idempotency_key=$2`, req.UserID, req.IdempotencyKey).Scan(&prior)
	if err == nil {
		if fp != prior {
			return nil, ErrConflict
		}
		item, e := readItem(ctx, tx, req.UserID, req.MeaningID, now)
		if e != nil {
			return nil, e
		}
		if e = tx.Commit(); e != nil {
			return nil, e
		}
		return result(item, req.Decision), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	// Lock existing rows before computing the revision, including soft removals.
	rows, err := tx.QueryContext(ctx, `SELECT id FROM user_words WHERE user_id=$1 AND meaning_id=$2 ORDER BY id FOR UPDATE`, req.UserID, req.MeaningID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id uuid.UUID
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	item, err := readItem(ctx, tx, req.UserID, req.MeaningID, now)
	if err != nil {
		return nil, err
	}
	if item.Revision != req.ExpectedRevision {
		return nil, ErrStale
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO user_word_knowledge(user_id,meaning_id,self_reported_known,note,updated_at) VALUES($1,$2,$3,'',$4) ON CONFLICT(user_id,meaning_id) DO UPDATE SET self_reported_known=EXCLUDED.self_reported_known,updated_at=EXCLUDED.updated_at`, req.UserID, req.MeaningID, req.Decision == "known", now)
	if err != nil {
		return nil, err
	}
	if req.Decision == "known" {
		_, err = tx.ExecContext(ctx, `UPDATE user_words SET status='ignored',ignored_at=$3,updated_at=$3 WHERE user_id=$1 AND meaning_id=$2 AND deleted_at IS NULL`, req.UserID, req.MeaningID, now)
	} else {
		// Restore the same historical row; do not erase review attempts/counters.
		var wid uuid.UUID
		e := tx.QueryRowContext(ctx, `SELECT id FROM user_words WHERE user_id=$1 AND meaning_id=$2 ORDER BY deleted_at ASC NULLS FIRST,updated_at DESC,id LIMIT 1`, req.UserID, req.MeaningID).Scan(&wid)
		if e == nil {
			_, err = tx.ExecContext(ctx, `UPDATE user_words SET deleted_at=NULL,status=CASE WHEN status IN ('ignored','archived') OR deleted_at IS NOT NULL THEN CASE WHEN total_review_count>0 THEN 'learning' ELSE 'new' END ELSE status END,ignored_at=NULL,next_review_at=CASE WHEN status IN ('ignored','archived') OR deleted_at IS NOT NULL THEN CASE WHEN next_review_at>$2 THEN next_review_at ELSE $2 END ELSE next_review_at END,updated_at=$2 WHERE id=$1`, wid, now)
		} else if errors.Is(e, sql.ErrNoRows) {
			_, err = tx.ExecContext(ctx, `INSERT INTO user_words(id,user_id,meaning_id,status,source,review_step,next_review_at,added_at,created_at,updated_at) VALUES($1,$2,$3,'new','manual',0,$4,$4,$4,$4) ON CONFLICT(user_id,meaning_id) WHERE deleted_at IS NULL DO NOTHING`, uuid.New(), req.UserID, req.MeaningID, now)
		} else {
			return nil, e
		}
	}
	if err != nil {
		return nil, err
	}
	// Reuse the existing privacy-safe fingerprint receipt and account lifecycle.
	_, err = tx.ExecContext(ctx, `INSERT INTO word_knowledge_actions(user_id,idempotency_key,fingerprint,created_at) VALUES($1,$2,$3,$4)`, req.UserID, req.IdempotencyKey, fp, now)
	if err != nil {
		return nil, err
	}
	item, err = readItem(ctx, tx, req.UserID, req.MeaningID, now)
	if err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return result(item, req.Decision), nil
}
