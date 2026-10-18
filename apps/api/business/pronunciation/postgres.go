package pronunciation

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/meaninguses"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/practice"
	"github.com/google/uuid"
)

type postgresRepository struct{ db *sql.DB }

// ResolveUtterance binds the selected sense to its exact public spelling/example.
// Ambiguous legacy text requests fail closed rather than cache a random sense.
func (r *postgresRepository) ResolveUtterance(ctx context.Context, user uuid.UUID, text, meaningID string) (Utterance, error) {
	var active bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1 AND status='active' AND deleted_at IS NULL)`, user).Scan(&active); err != nil {
		return Utterance{}, err
	}
	if !active {
		return Utterance{}, ErrNotFound
	}
	rows, err := r.db.QueryContext(ctx, `SELECT wm.id::text,cw.text,wm.part_of_speech,wm.short_definition,COALESCE((SELECT we.example_text FROM word_examples we WHERE we.meaning_id=wm.id AND we.status='active' ORDER BY we.example_order,we.id LIMIT 1),'') FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id AND cw.status='active' AND cw.language_code='en' WHERE wm.status='active' AND ($2='' OR wm.id::text=$2) AND (cw.text=$1 OR (cw.text='resume' AND wm.part_of_speech='noun' AND $1='résumé') OR EXISTS(SELECT 1 FROM word_examples we WHERE we.meaning_id=wm.id AND we.status='active' AND we.example_text=$1)) ORDER BY wm.meaning_order,wm.id LIMIT 2`, text, meaningID)
	if err != nil {
		return Utterance{}, err
	}
	defer rows.Close()
	var found []Utterance
	for rows.Next() {
		var u Utterance
		var word, example string
		if err = rows.Scan(&u.MeaningID, &word, &u.PartOfSpeech, &u.Definition, &example); err != nil {
			return Utterance{}, err
		}
		u.Text = text
		u.Locale = "en-US"
		u.WordText = word
		u.Kind = "word"
		// Both accepted spellings name the same reviewed CV noun. Normalize only
		// spoken word text, keeping canonical identity and all example text intact.
		if word == "resume" && u.PartOfSpeech == "noun" && (text == "resume" || text == "résumé") {
			u.Text = "résumé"
		}
		curatedContext, curated := meaninguses.ReviewedContext(u.MeaningID, word, u.Definition)
		reviewedPractice := practice.SupportsMeaning(u.MeaningID, word, u.Definition)
		reviewedMeaning := reviewedPractice || meaninguses.SupportsMeaning(u.MeaningID, word, u.Definition)
		if text != word && !(word == "resume" && u.PartOfSpeech == "noun" && text == "résumé") {
			if curated {
				if text != curatedContext {
					continue
				}
			} else if !reviewedMeaning {
				continue
			}
			u.Kind = "example"
		}
		// Exact reviewed contexts disambiguate heteronyms, including present/past
		// read. A random database example cannot replace this authored guidance.
		if curated {
			u.Context = boundedReference(curatedContext, 750)
		} else if reviewedPractice {
			u.Context = boundedReference(example, 750)
		}

		found = append(found, u)
	}
	if err = rows.Err(); err != nil {
		return Utterance{}, err
	}
	if len(found) != 1 {
		return Utterance{}, ErrNotFound
	}
	return found[0], nil
}

func (r *postgresRepository) Reserve(ctx context.Context, key string, characters int, now time.Time) (reservation, error) {
	var cached []byte
	err := r.db.QueryRowContext(ctx, `SELECT audio FROM pronunciation_audio_cache WHERE cache_key=$1 AND status='ready'`, key).Scan(&cached)
	if err == nil {
		if !validMP3(cached) {
			return reservation{}, ErrUnavailable
		}
		return reservation{audio: cached}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return reservation{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return reservation{}, err
	}
	defer tx.Rollback()
	// Serialize the cache identity before its date-specific budget. Two callers
	// straddling UTC midnight otherwise lock different budget rows, both observe
	// a missing cache row and overwrite the same pending generation lease.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('pronunciation:' || $1,0))`, key); err != nil {
		return reservation{}, err
	}
	day := now.UTC().Format(time.DateOnly)
	if _, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_audio_budget(utc_day,reserved_characters)VALUES($1::date,0) ON CONFLICT DO NOTHING`, day); err != nil {
		return reservation{}, err
	}
	var used, generations int
	if err = tx.QueryRowContext(ctx, `SELECT reserved_characters,reserved_generations FROM pronunciation_audio_budget WHERE utc_day=$1::date FOR UPDATE`, day).Scan(&used, &generations); err != nil {
		return reservation{}, err
	}
	var status, attemptedDay string
	var lease time.Time
	err = tx.QueryRowContext(ctx, `SELECT status,attempted_day::text,lease_until,audio FROM pronunciation_audio_cache WHERE cache_key=$1 FOR UPDATE`, key).Scan(&status, &attemptedDay, &lease, &cached)
	if err == nil {
		if status == "ready" {
			if !validMP3(cached) {
				return reservation{}, ErrUnavailable
			}
			return reservation{audio: cached}, tx.Commit()
		}
		if status == "pending" && (attemptedDay == day || lease.After(now)) {
			return reservation{}, ErrBusy
		}
		if status == "failed" && attemptedDay == day {
			return reservation{}, ErrUnavailable
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return reservation{}, err
	}
	if characters < 1 || characters > 500 {
		return reservation{}, ErrInvalid
	}
	if used+characters > DailyCharacterLimit || generations >= DailyGenerationLimit {
		return reservation{}, ErrRateLimited
	}
	generation := uuid.New()
	if _, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_audio_cache(cache_key,status,generation_id,attempted_day,lease_until,created_at,updated_at)VALUES($1,'pending',$2,$3::date,$4,$5,$5) ON CONFLICT(cache_key) DO UPDATE SET status='pending',generation_id=EXCLUDED.generation_id,attempted_day=EXCLUDED.attempted_day,lease_until=EXCLUDED.lease_until,audio=NULL,updated_at=EXCLUDED.updated_at`, key, generation, day, now.Add(20*time.Second), now); err != nil {
		return reservation{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE pronunciation_audio_budget SET reserved_characters=reserved_characters+$2,reserved_generations=reserved_generations+1 WHERE utc_day=$1::date`, day, characters); err != nil {
		return reservation{}, err
	}
	if err = tx.Commit(); err != nil {
		return reservation{}, err
	}
	return reservation{generation: generation}, nil
}

func (r *postgresRepository) Complete(ctx context.Context, key string, generation uuid.UUID, audio []byte, now time.Time) error {
	status := "ready"
	if len(audio) == 0 {
		status = "failed"
		audio = nil
	}
	result, err := r.db.ExecContext(ctx, `UPDATE pronunciation_audio_cache SET status=$3,audio=$4,updated_at=$5 WHERE cache_key=$1 AND generation_id=$2 AND status='pending'`, key, generation, status, audio, now)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrBusy
	}
	return nil
}

func boundedReference(text string, limit int) string {
	runes := []rune(text)
	if len(runes) > limit {
		return string(runes[:limit])
	}
	return text
}
