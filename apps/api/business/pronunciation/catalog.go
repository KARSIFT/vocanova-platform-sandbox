package pronunciation

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KARSIFT/vocanova-platform/apps/api/foundation/meaningidentity"
	"github.com/google/uuid"
	"math"
	"time"
	"unicode/utf8"
)

// CatalogBudget bounds attempts, input and observed output, not a provider bill.
// A response may cross the observed-output threshold; no further request starts.
type CatalogBudget struct {
	MaxGenerations  int     `json:"maxGenerations"`
	MaxCharacters   int     `json:"maxCharacters"`
	MaxAudioSeconds float64 `json:"maxAudioSeconds"`
}

func (b CatalogBudget) Validate() error {
	if b.MaxGenerations < 1 || b.MaxGenerations > 1000 || b.MaxCharacters < 1 || b.MaxCharacters > 500000 || b.MaxAudioSeconds <= 0 || b.MaxAudioSeconds > 60000 || math.IsNaN(b.MaxAudioSeconds) || math.IsInf(b.MaxAudioSeconds, 0) {
		return fmt.Errorf("explicit bounded catalog budgets are required")
	}
	return nil
}

type CatalogReceipt struct {
	ID              string  `json:"id"`
	MeaningID       string  `json:"meaningId"`
	MeaningRevision string  `json:"meaningRevision"`
	CacheKey        string  `json:"cacheKey"`
	Status          string  `json:"status"`
	Characters      int     `json:"characters"`
	AudioSeconds    float64 `json:"audioSeconds"`
}
type CatalogValidation struct {
	Checked int `json:"checked"`
	Valid   int `json:"valid"`
	Invalid int `json:"invalid"`
}
type CatalogReport struct {
	Validation      *CatalogValidation `json:"validation,omitempty"`
	CatalogIdentity *ProviderIdentity  `json:"catalogIdentity,omitempty"`
	Budget          *CatalogBudget     `json:"budget,omitempty"`
	Receipts        []CatalogReceipt   `json:"receipts,omitempty"`
	JobID           string             `json:"jobId,omitempty"`
	Identity        ProviderIdentity   `json:"identity"`
	Total           int                `json:"total"`
	Ready           int                `json:"ready"`
	Missing         int                `json:"missing"`
	Pending         int                `json:"pending"`
	Failed          int                `json:"failed"`
	Attempted       int                `json:"attempted"`
	Characters      int                `json:"characters"`
	AudioSeconds    float64            `json:"audioSeconds"`
}
type CatalogOperator struct {
	db       *sql.DB
	provider Provider
}

func NewCatalogOperator(db *sql.DB, p Provider) *CatalogOperator { return &CatalogOperator{db, p} }
func (o *CatalogOperator) identity() ProviderIdentity {
	return ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}
}
func (o *CatalogOperator) utterances(ctx context.Context) ([]Utterance, error) {
	rows, err := o.db.QueryContext(ctx, `SELECT wm.id::text,cw.text,wm.part_of_speech,wm.short_definition,COALESCE((SELECT example_text FROM word_examples WHERE meaning_id=wm.id AND status='active' ORDER BY example_order,id LIMIT 1),'') FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.status='active' AND cw.status='active' AND cw.language_code='en' ORDER BY wm.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Utterance
	for rows.Next() {
		var u Utterance
		var word, example string
		if err = rows.Scan(&u.MeaningID, &word, &u.PartOfSpeech, &u.Definition, &example); err != nil {
			return nil, err
		}
		u, ok := canonicalUtterance(u, word, word, example)
		if !ok {
			return nil, ErrInvalid
		}
		result = append(result, u)
	}
	return result, rows.Err()
}
func (o *CatalogOperator) Audit(ctx context.Context) (CatalogReport, error) {
	report, _, err := o.auditSnapshot(ctx, false)
	return report, err
}

// auditSnapshot optionally retains identities already hashed for this report.
// Ordinary metadata audits need no retained key set; validation reuses one view
// of canonical content rather than rereading it before streaming cached clips.
func (o *CatalogOperator) auditSnapshot(ctx context.Context, retainKeys bool) (CatalogReport, map[string]struct{}, error) {
	r := CatalogReport{Identity: o.identity()}
	us, err := o.utterances(ctx)
	if err != nil {
		return r, nil, err
	}
	rows, err := o.db.QueryContext(ctx, `SELECT cache_key,status FROM pronunciation_audio_cache`)
	if err != nil {
		return r, nil, err
	}
	defer rows.Close()
	statuses := map[string]string{}
	for rows.Next() {
		var k, s string
		if err = rows.Scan(&k, &s); err != nil {
			return r, nil, err
		}
		statuses[k] = s
	}
	if err = rows.Err(); err != nil {
		return r, nil, err
	}
	var keys map[string]struct{}
	if retainKeys {
		keys = make(map[string]struct{}, len(us))
	}
	for _, u := range us {
		r.Total++
		key := utteranceCacheKey(r.Identity, u)
		if retainKeys {
			keys[key] = struct{}{}
		}
		switch statuses[key] {
		case "ready":
			r.Ready++
		case "pending":
			r.Pending++
		case "failed":
			r.Failed++
		default:
			r.Missing++
		}
	}
	return r, keys, nil
}
func (o *CatalogOperator) Generate(ctx context.Context, job uuid.UUID, b CatalogBudget) (CatalogReport, error) {
	r := CatalogReport{JobID: job.String(), Identity: o.identity()}
	if err := b.Validate(); err != nil {
		return r, err
	}
	if job == uuid.Nil || o.provider == nil || o.provider.Identity() != r.Identity {
		return r, ErrInvalid
	}
	// Session ownership allows only one operator process, across all job IDs.
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
		if _, err := conn.ExecContext(c, `SELECT pg_advisory_unlock(hashtextextended('pronunciation:catalog-operator',0))`); err != nil {
			// Do not return a connection retaining an advisory lock to the pool.
			_ = conn.Raw(func(any) error { return driver.ErrBadConn })
		}
	}()
	identity, _ := json.Marshal(r.Identity)
	_, err = o.db.ExecContext(ctx, `INSERT INTO pronunciation_catalog_jobs(id,provider_identity,max_generations,max_characters,max_audio_seconds) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, job, string(identity), b.MaxGenerations, b.MaxCharacters, b.MaxAudioSeconds)
	if err != nil {
		return r, err
	}
	var saved CatalogBudget
	var savedIdentity string
	if err = o.db.QueryRowContext(ctx, `SELECT provider_identity,max_generations,max_characters,max_audio_seconds FROM pronunciation_catalog_jobs WHERE id=$1`, job).Scan(&savedIdentity, &saved.MaxGenerations, &saved.MaxCharacters, &saved.MaxAudioSeconds); err != nil {
		return r, err
	}
	if saved != b || savedIdentity != string(identity) {
		return r, fmt.Errorf("existing job identity and budgets cannot change")
	}
	us, err := o.utterances(ctx)
	if err != nil {
		return r, err
	}
	r.Total = len(us)
	for _, u := range us {
		if err = ctx.Err(); err != nil {
			return r, err
		}
		if err = o.totals(ctx, job, &r); err != nil {
			return r, err
		}
		characters := utf8.RuneCountInString(u.Text)
		if r.Attempted >= b.MaxGenerations || r.Characters+characters > b.MaxCharacters || r.AudioSeconds >= b.MaxAudioSeconds {
			break
		}
		key := utteranceCacheKey(r.Identity, u)
		generation, claimed, claimErr := o.claim(ctx, job, u, key, characters)
		if claimErr != nil {
			return r, claimErr
		}
		if !claimed {
			continue
		}
		// The persisted attempt precedes every provider call, including cancellations.
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
		if providerErr != nil {
			totalsCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
			_ = o.totals(totalsCtx, job, &r)
			cancel()
			return r, ErrUnavailable
		}
	}
	err = o.totals(ctx, job, &r)
	return r, err
}
func (o *CatalogOperator) totals(ctx context.Context, job uuid.UUID, r *CatalogReport) error {
	return o.db.QueryRowContext(ctx, `SELECT count(*),COALESCE(sum(characters),0),COALESCE(sum(audio_seconds),0) FROM pronunciation_catalog_attempts WHERE job_id=$1`, job).Scan(&r.Attempted, &r.Characters, &r.AudioSeconds)
}
func (o *CatalogOperator) current(ctx context.Context, tx *sql.Tx, id string) (Utterance, error) {
	var u Utterance
	var word, example string
	err := tx.QueryRowContext(ctx, `SELECT wm.id::text,cw.text,wm.part_of_speech,wm.short_definition,COALESCE((SELECT example_text FROM word_examples WHERE meaning_id=wm.id AND status='active' ORDER BY example_order,id LIMIT 1),'') FROM word_meanings wm JOIN canonical_words cw ON cw.id=wm.word_id WHERE wm.id=$1 AND wm.status='active' AND cw.status='active' AND cw.language_code='en' FOR SHARE OF wm,cw`, id).Scan(&u.MeaningID, &word, &u.PartOfSpeech, &u.Definition, &example)
	if err != nil {
		return u, err
	}
	u, ok := canonicalUtterance(u, word, word, example)
	if !ok {
		return u, ErrInvalid
	}
	return u, nil
}
func (o *CatalogOperator) claim(ctx context.Context, job uuid.UUID, u Utterance, key string, characters int) (uuid.UUID, bool, error) {
	generation := uuid.New()
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return generation, false, err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('pronunciation:' || $1,0))`, key)
	if err != nil {
		return generation, false, err
	}
	// Enforce immutable job budgets again under a row lock, so even a lost
	// operator session cannot make two reservers overspend the same job.
	var budget CatalogBudget
	if err = tx.QueryRowContext(ctx, `SELECT max_generations,max_characters,max_audio_seconds FROM pronunciation_catalog_jobs WHERE id=$1 FOR UPDATE`, job).Scan(&budget.MaxGenerations, &budget.MaxCharacters, &budget.MaxAudioSeconds); err != nil {
		return generation, false, err
	}
	var spent CatalogReport
	if err = tx.QueryRowContext(ctx, `SELECT count(*),COALESCE(sum(characters),0),COALESCE(sum(audio_seconds),0) FROM pronunciation_catalog_attempts WHERE job_id=$1`, job).Scan(&spent.Attempted, &spent.Characters, &spent.AudioSeconds); err != nil {
		return generation, false, err
	}
	if spent.Attempted >= budget.MaxGenerations || spent.Characters+characters > budget.MaxCharacters || spent.AudioSeconds >= budget.MaxAudioSeconds {
		return generation, false, nil
	}
	var exists bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pronunciation_catalog_attempts WHERE job_id=$1 AND cache_key=$2)`, job, key).Scan(&exists); err != nil {
		return generation, false, err
	}
	if exists {
		return generation, false, nil
	}
	var status string
	var sameDay, leased bool
	err = tx.QueryRowContext(ctx, `SELECT status,attempted_day=(now() AT TIME ZONE 'UTC')::date,lease_until>now() FROM pronunciation_audio_cache WHERE cache_key=$1 FOR UPDATE`, key).Scan(&status, &sameDay, &leased)
	if err == nil && (status == "ready" || sameDay || leased) {
		return generation, false, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return generation, false, err
	}
	if characters < 1 || characters > 500 {
		return generation, false, ErrInvalid
	}
	revision := meaningidentity.Revision(u.MeaningID, u.WordText, u.PartOfSpeech, u.Definition)
	_, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_catalog_attempts(id,job_id,cache_key,meaning_id,meaning_revision,characters,status) VALUES($1,$2,$3,$4,$5,$6,'reserved')`, generation, job, key, u.MeaningID, revision, characters)
	if err != nil {
		return generation, false, err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO pronunciation_audio_cache(cache_key,status,generation_id,attempted_day,lease_until,created_at,updated_at) VALUES($1,'pending',$2,(now() AT TIME ZONE 'UTC')::date,now()+interval '20 seconds',now(),now()) ON CONFLICT(cache_key) DO UPDATE SET status='pending',generation_id=EXCLUDED.generation_id,attempted_day=EXCLUDED.attempted_day,lease_until=EXCLUDED.lease_until,audio=NULL,updated_at=now()`, key, generation)
	if err != nil {
		return generation, false, err
	}
	err = tx.Commit()
	return generation, err == nil, err
}
func (o *CatalogOperator) finish(ctx context.Context, job uuid.UUID, meaningID, key string, generation uuid.UUID, status string, data []byte, seconds float64) error {
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if status == "ready" {
		current, resolveErr := o.current(ctx, tx, meaningID)
		if resolveErr != nil || utteranceCacheKey(o.identity(), current) != key {
			status = "changed"
			data = nil
		}
	}
	cacheStatus := "failed"
	if status == "ready" {
		cacheStatus = "ready"
	}
	result, err := tx.ExecContext(ctx, `UPDATE pronunciation_audio_cache SET status=$3,audio=$4,updated_at=now() WHERE cache_key=$1 AND generation_id=$2 AND status='pending'`, key, generation, cacheStatus, data)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil || n != 1 {
		return ErrBusy
	}
	_, err = tx.ExecContext(ctx, `UPDATE pronunciation_catalog_attempts SET status=$3,audio_seconds=$4,finished_at=now() WHERE id=$1 AND job_id=$2 AND status='reserved'`, generation, job, status, seconds)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// JobReport is a read-only resumable manifest; the database remains authoritative.
func (o *CatalogOperator) JobReport(ctx context.Context, job uuid.UUID) (CatalogReport, error) {
	r, err := o.Audit(ctx)
	if err != nil {
		return r, err
	}
	r.JobID = job.String()
	r.Budget = &CatalogBudget{}
	currentIdentity := r.Identity
	r.CatalogIdentity = &currentIdentity
	var storedIdentity string
	if err = o.db.QueryRowContext(ctx, `SELECT provider_identity,max_generations,max_characters,max_audio_seconds FROM pronunciation_catalog_jobs WHERE id=$1`, job).Scan(&storedIdentity, &r.Budget.MaxGenerations, &r.Budget.MaxCharacters, &r.Budget.MaxAudioSeconds); err != nil {
		return r, err
	}
	if err = json.Unmarshal([]byte(storedIdentity), &r.Identity); err != nil {
		return r, err
	}
	if err = o.totals(ctx, job, &r); err != nil {
		return r, err
	}
	rows, err := o.db.QueryContext(ctx, `SELECT id::text,meaning_id::text,meaning_revision,cache_key,status,characters,audio_seconds FROM pronunciation_catalog_attempts WHERE job_id=$1 ORDER BY created_at,id`, job)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var receipt CatalogReceipt
		if err = rows.Scan(&receipt.ID, &receipt.MeaningID, &receipt.MeaningRevision, &receipt.CacheKey, &receipt.Status, &receipt.Characters, &receipt.AudioSeconds); err != nil {
			return r, err
		}
		r.Receipts = append(r.Receipts, receipt)
	}
	return r, rows.Err()
}

// AuditValidated streams one bounded MP3 at a time. It does not retain catalog
// blobs, mutate invalid cache rows, invoke a provider, or claim correct speech.
func (o *CatalogOperator) AuditValidated(ctx context.Context) (CatalogReport, error) {
	r, keys, err := o.auditSnapshot(ctx, true)
	if err != nil {
		return r, err
	}
	r.Validation = &CatalogValidation{}
	rows, err := o.db.QueryContext(ctx, `SELECT cache_key,audio FROM pronunciation_audio_cache WHERE status='ready' ORDER BY cache_key`)
	if err != nil {
		return r, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		var data []byte
		if err = rows.Scan(&key, &data); err != nil {
			return r, err
		}
		if _, current := keys[key]; !current {
			continue
		}
		r.Validation.Checked++
		if _, err := ValidateAudio(data); err != nil {
			r.Validation.Invalid++
		} else {
			r.Validation.Valid++
		}
	}
	return r, rows.Err()
}
