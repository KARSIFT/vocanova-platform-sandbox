package pronunciation

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"time"
	"unicode/utf8"
)

// RecoverFailed makes at most one explicitly budgeted call for an ordinary
// operator attempt rejected after decoding. The unique durable linkage is spent
// before the call, including cancellation; neither this job nor another UUID can
// retry it. Read-only JobReport remains available after any interruption.
func (o *CatalogOperator) RecoverFailed(ctx context.Context, job, original uuid.UUID, b CatalogBudget) (CatalogReport, error) {
	r := CatalogReport{JobID: job.String(), Identity: o.identity(), RecoveryOfAttempt: original.String()}
	if err := b.Validate(); err != nil {
		return r, err
	}
	if b.MaxGenerations != 1 || job == uuid.Nil || original == uuid.Nil || job == original || o.provider == nil || o.provider.Identity() != r.Identity {
		return r, ErrInvalid
	}
	conn, err := o.db.Conn(ctx)
	if err != nil {
		return r, err
	}
	defer conn.Close()
	var locked bool
	if err = conn.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(hashtextextended('pronunciation:catalog-operator',0))`).Scan(&locked); err != nil {
		return r, err
	}
	if !locked {
		return r, ErrBusy
	}
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if _, e := conn.ExecContext(c, `SELECT pg_advisory_unlock(hashtextextended('pronunciation:catalog-operator',0))`); e != nil {
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	u, key, generation, err := o.claimRecovery(ctx, job, original, b)
	if err != nil {
		return r, err
	}
	// Same bounded provider, signal gate and completion CAS as ordinary generation.
	providerCtx, cancel := context.WithTimeout(ctx, 12*time.Second)
	data, providerErr := o.provider.Generate(providerCtx, u)
	cancel()
	quality := AudioQuality{}
	if providerErr == nil {
		quality, providerErr = ValidateAudio(data)
	}
	status := "ready"
	if providerErr != nil {
		status = "failed"
		data = nil
	}
	finishCtx, finishCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	err = o.finish(finishCtx, job, u.MeaningID, key, generation, status, data, quality.DurationSeconds)
	finishCancel()
	if err != nil {
		return r, err
	}
	totalsCtx, totalsCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	defer totalsCancel()
	if err = o.totals(totalsCtx, job, &r); err != nil {
		return r, err
	}
	if providerErr != nil {
		return r, ErrUnavailable
	}
	return r, nil
}

func (o *CatalogOperator) claimRecovery(ctx context.Context, job, original uuid.UUID, b CatalogBudget) (Utterance, string, uuid.UUID, error) {
	var u Utterance
	var key string
	generation := uuid.New()
	// Receipt identity is reread under lock below; this first lookup only selects
	// the same per-key serialization namespace used by ordinary/public requests.
	if err := o.db.QueryRowContext(ctx, `SELECT cache_key FROM pronunciation_catalog_attempts WHERE id=$1`, original).Scan(&key); err != nil {
		return u, key, generation, err
	}
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return u, key, generation, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('pronunciation:' || $1,0))`, key); err != nil {
		return u, key, generation, err
	}
	var meaningID, revision string
	var ordinaryOriginal bool
	identity, _ := json.Marshal(o.identity())
	// A positive decoded duration with failed completion identifies the existing
	// post-decode rejection path, rather than an HTTP/provider/quota failure.
	err = tx.QueryRowContext(ctx, `SELECT a.meaning_id::text,a.meaning_revision,
 j.recovery_of_attempt IS NULL AND j.id<>$2 AND j.provider_identity=$3
 FROM pronunciation_catalog_attempts a JOIN pronunciation_catalog_jobs j ON j.id=a.job_id
 WHERE a.id=$1 AND a.cache_key=$4 AND a.status='failed' AND a.audio_seconds>0 AND a.finished_at IS NOT NULL
 FOR UPDATE OF a FOR SHARE OF j`, original, job, string(identity), key).Scan(&meaningID, &revision, &ordinaryOriginal)
	if err != nil {
		return u, key, generation, err
	}
	if !ordinaryOriginal {
		return u, key, generation, ErrInvalid
	}
	// Match completion's canonical-before-cache row lock order. Example context is
	// reread and hashed by current(), with the same documented context-edit limits.
	u, err = o.current(ctx, tx, meaningID)
	if err != nil {
		return u, key, generation, err
	}
	if utteranceCacheKey(o.identity(), u) != key || meaningidentity.Revision(u.MeaningID, u.WordText, u.PartOfSpeech, u.Definition) != revision {
		return u, key, generation, ErrInvalid
	}
	var eligible bool
	err = tx.QueryRowContext(ctx, `SELECT c.status='failed' AND c.generation_id=$2 AND c.audio IS NULL
 AND c.lease_until<=clock_timestamp() AND c.attempted_day=(a.created_at AT TIME ZONE 'UTC')::date
 AND c.attempted_day=(clock_timestamp() AT TIME ZONE 'UTC')::date
 FROM pronunciation_audio_cache c JOIN pronunciation_catalog_attempts a ON a.id=$2 WHERE c.cache_key=$1 FOR UPDATE OF c`, key, original).Scan(&eligible)
	if err != nil {
		return u, key, generation, err
	}
	if !eligible {
		return u, key, generation, ErrInvalid
	}
	characters := utf8.RuneCountInString(u.Text)
	if characters < 1 || characters > 500 || characters > b.MaxCharacters {
		return u, key, generation, ErrInvalid
	}
	// No ON CONFLICT retry: an existing job (ordinary or recovery), prior linked
	// recovery, or changed budget/target all fail before any paid attempt.
	_, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_catalog_jobs(id,provider_identity,max_generations,max_characters,max_audio_seconds,recovery_of_attempt) VALUES($1,$2,1,$3,$4,$5)`, job, string(identity), b.MaxCharacters, b.MaxAudioSeconds, original)
	if err != nil {
		return u, key, generation, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_catalog_attempts(id,job_id,cache_key,meaning_id,meaning_revision,characters,status) VALUES($1,$2,$3,$4,$5,$6,'reserved')`, generation, job, key, meaningID, revision, characters)
	if err != nil {
		return u, key, generation, err
	}
	// Preserve attempted_day: recovery does not reset the public same-day policy.
	// Wall-clock time starts the lease after any per-key lock wait.
	result, err := tx.ExecContext(ctx, `UPDATE pronunciation_audio_cache SET status='pending',generation_id=$2,lease_until=clock_timestamp()+interval '20 seconds',updated_at=clock_timestamp() WHERE cache_key=$1 AND generation_id=$3 AND status='failed' AND audio IS NULL`, key, generation, original)
	if err != nil {
		return u, key, generation, err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return u, key, generation, ErrBusy
	}
	if err = tx.Commit(); err != nil {
		return u, key, generation, err
	}
	return u, key, generation, nil
}
