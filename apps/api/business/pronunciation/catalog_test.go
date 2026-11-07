package pronunciation

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCatalogBudgetRequiresExplicitFiniteBounds(t *testing.T) {
	for _, b := range []CatalogBudget{{}, {MaxGenerations: 1, MaxCharacters: 10}, {MaxGenerations: 1001, MaxCharacters: 10, MaxAudioSeconds: 60}} {
		require.Error(t, b.Validate())
	}
	require.NoError(t, (CatalogBudget{10, 1000, 600}).Validate())
}
func TestCatalogPostgreSQLResumePreservesReadyAndBudget(t *testing.T) {
	db := audioDB(t)
	ctx := t.Context()
	word, id := uuid.New(), uuid.New()
	_, err := db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) VALUES($1,'fixture','fixture','en','active',now(),now())`, word)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'noun','A synthetic fixture.',1,'active',now(),now())`, id, word)
	require.NoError(t, err)
	p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}}
	svc := NewCatalogOperator(db, p)
	before, err := svc.Audit(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, before.Missing)
	job := uuid.New()
	budget := CatalogBudget{1, 100, 60}
	report, err := svc.Generate(ctx, job, budget)
	require.NoError(t, err)
	require.Equal(t, 1, report.Attempted)
	require.Equal(t, 1, p.calls)
	_, err = svc.Generate(ctx, job, budget)
	require.NoError(t, err)
	require.Equal(t, 1, p.calls)
	after, err := svc.Audit(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, after.Ready)
	var publicBudget int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_audio_budget`).Scan(&publicBudget))
	require.Zero(t, publicBudget)
	budget.MaxGenerations = 2
	_, err = svc.Generate(ctx, job, budget)
	require.Error(t, err)
}
func TestCatalogPostgreSQLChangedMeaningAndProviderFailureConsumeAttempt(t *testing.T) {
	for _, mode := range []string{"changed", "failed", "cancelled"} {
		t.Run(mode, func(t *testing.T) {
			db := audioDB(t)
			word, id := uuid.New(), uuid.New()
			_, err := db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) VALUES($1,'fixture','fixture','en','active',now(),now())`, word)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'noun','A synthetic fixture.',1,'active',now(),now())`, id, word)
			require.NoError(t, err)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}, fn: func(context.Context, Utterance) ([]byte, error) {
				if mode == "failed" {
					return nil, ErrUnavailable
				}
				if mode == "cancelled" {
					cancel()
					return nil, context.Canceled
				}
				_, err := db.Exec(`UPDATE word_meanings SET short_definition='Changed fixture.' WHERE id=$1`, id)
				require.NoError(t, err)
				return testMP3, nil
			}}
			svc := NewCatalogOperator(db, p)
			job := uuid.New()
			budget := CatalogBudget{1, 100, 60}
			_, _ = svc.Generate(ctx, job, budget)
			_, err = svc.Generate(t.Context(), job, budget)
			require.NoError(t, err)
			require.Equal(t, 1, p.calls)
			var ready, attempts int
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_audio_cache WHERE status='ready'`).Scan(&ready))
			require.Zero(t, ready)
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_catalog_attempts WHERE job_id=$1`, job).Scan(&attempts))
			require.Equal(t, 1, attempts)
		})
	}
}

func seedCatalogFixture(t *testing.T) (*sql.DB, Utterance) {
	t.Helper()
	db := audioDB(t)
	word, id := uuid.New(), uuid.New()
	_, err := db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) VALUES($1,'fixture','fixture','en','active',now(),now())`, word)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'noun','A synthetic fixture.',1,'active',now(),now())`, id, word)
	require.NoError(t, err)
	return db, Utterance{MeaningID: id.String(), WordText: "fixture", Text: "fixture", Kind: "word", PartOfSpeech: "noun", Definition: "A synthetic fixture.", Locale: "en-US"}
}
func TestCatalogPostgreSQLSharesPublicLeaseAndSingleOperator(t *testing.T) {
	db, u := seedCatalogFixture(t)
	started, release := make(chan struct{}), make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}, fn: func(context.Context, Utterance) ([]byte, error) { close(started); <-release; return testMP3, nil }}
	svc := NewCatalogOperator(db, p)
	results := make(chan error, 1)
	go func() { _, err := svc.Generate(t.Context(), uuid.New(), CatalogBudget{1, 100, 60}); results <- err }()
	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("generation did not start")
	}
	_, err := svc.Generate(t.Context(), uuid.New(), CatalogBudget{1, 100, 60})
	require.ErrorIs(t, err, ErrBusy)
	key := utteranceCacheKey(p.identity, u)
	repo := &postgresRepository{db}
	_, err = repo.Reserve(t.Context(), key, 7, time.Now().UTC())
	require.ErrorIs(t, err, ErrBusy)
	close(release)
	require.NoError(t, <-results)
	cached, err := repo.Reserve(t.Context(), key, 7, time.Now().UTC())
	require.NoError(t, err)
	require.Equal(t, testMP3, cached.audio)
	require.Equal(t, 1, p.calls)
}
func TestCatalogPostgreSQLPublicReadyOrPendingNeverRegenerates(t *testing.T) {
	for _, ready := range []bool{true, false} {
		t.Run(fmt.Sprint(ready), func(t *testing.T) {
			db, u := seedCatalogFixture(t)
			p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}}
			key := utteranceCacheKey(p.identity, u)
			repo := &postgresRepository{db}
			reserved, err := repo.Reserve(t.Context(), key, 7, time.Now().UTC())
			require.NoError(t, err)
			if ready {
				require.NoError(t, repo.Complete(t.Context(), key, reserved.generation, testMP3, time.Now().UTC()))
			}
			svc := NewCatalogOperator(db, p)
			r, err := svc.Generate(t.Context(), uuid.New(), CatalogBudget{1, 100, 60})
			require.NoError(t, err)
			require.Zero(t, r.Attempted)
			require.Zero(t, p.calls)
		})
	}
}
func TestCatalogPostgreSQLInvalidAudioAndCharacterBudgetFailClosed(t *testing.T) {
	db, _ := seedCatalogFixture(t)
	p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}, fn: func(context.Context, Utterance) ([]byte, error) { return silentMP3, nil }}
	svc := NewCatalogOperator(db, p)
	_, err := svc.Generate(t.Context(), uuid.New(), CatalogBudget{1, 1, 60})
	require.NoError(t, err)
	require.Zero(t, p.calls)
	job := uuid.New()
	_, err = svc.Generate(t.Context(), job, CatalogBudget{1, 100, 60})
	require.Error(t, err)
	require.Equal(t, 1, p.calls)
	_, err = svc.Generate(t.Context(), job, CatalogBudget{1, 100, 60})
	require.NoError(t, err)
	require.Equal(t, 1, p.calls)
	var ready int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_audio_cache WHERE status='ready'`).Scan(&ready))
	require.Zero(t, ready)
}

func TestCatalogPostgreSQLHistoricalJobIdentityIsNotRelabelled(t *testing.T) {
	db, _ := seedCatalogFixture(t)
	p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}}
	svc := NewCatalogOperator(db, p)
	job := uuid.New()
	budget := CatalogBudget{1, 100, 60}
	_, err := svc.Generate(t.Context(), job, budget)
	require.NoError(t, err)
	old := p.identity
	old.InstructionVersion = "historical"
	raw, _ := json.Marshal(old)
	_, err = db.Exec(`UPDATE pronunciation_catalog_jobs SET provider_identity=$2 WHERE id=$1`, job, string(raw))
	require.NoError(t, err)
	report, err := svc.JobReport(t.Context(), job)
	require.NoError(t, err)
	require.Equal(t, old, report.Identity)
	require.Equal(t, p.identity, *report.CatalogIdentity)
	require.Len(t, report.Receipts, 1)
	_, err = svc.Generate(t.Context(), job, budget)
	require.Error(t, err)
	require.Equal(t, 1, p.calls)
}
func TestCatalogPostgreSQLPublicationSerializesCanonicalEdit(t *testing.T) {
	db, u := seedCatalogFixture(t)
	ctx := t.Context()
	var appName string
	require.NoError(t, db.QueryRow(`SELECT current_setting('application_name')`).Scan(&appName))
	// Hold completion after its canonical read, immediately before cache update.
	barrier, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer barrier.Rollback()
	_, err = barrier.Exec(`SELECT pg_advisory_xact_lock(7140711400)`)
	require.NoError(t, err)
	_, err = db.Exec(`CREATE FUNCTION catalog_publication_gate() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.status='ready' THEN PERFORM pg_advisory_xact_lock(7140711400); END IF; RETURN NEW; END $$; CREATE TRIGGER catalog_gate BEFORE UPDATE ON pronunciation_audio_cache FOR EACH ROW EXECUTE FUNCTION catalog_publication_gate()`)
	require.NoError(t, err)
	p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}}
	svc := NewCatalogOperator(db, p)
	result := make(chan error, 1)
	go func() { _, err := svc.Generate(ctx, uuid.New(), CatalogBudget{1, 100, 60}); result <- err }()
	waiting := func() int {
		var n int
		_ = db.QueryRow(`SELECT count(*) FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock'`, appName).Scan(&n)
		return n
	}
	require.Eventually(t, func() bool { return waiting() >= 1 }, 3*time.Second, 5*time.Millisecond)
	edit := make(chan error, 1)
	go func() {
		_, err := db.Exec(`UPDATE word_meanings SET short_definition='Edited after publication.' WHERE id=$1`, u.MeaningID)
		edit <- err
	}()
	require.Eventually(t, func() bool { return waiting() >= 2 }, 3*time.Second, 5*time.Millisecond, "editor must wait for the publication's shared canonical lock")
	require.NoError(t, barrier.Commit())
	require.NoError(t, <-result)
	require.NoError(t, <-edit)
	report, err := svc.Audit(ctx)
	require.NoError(t, err)
	require.Zero(t, report.Ready)
	require.Equal(t, 1, report.Missing, "a later edit cannot inherit the prior audio identity")
}

func TestCatalogPostgreSQLReviewedExampleChangesInvalidateWordAudio(t *testing.T) {
	for _, during := range []bool{true, false} {
		t.Run(fmt.Sprint(during), func(t *testing.T) {
			db := audioDB(t)
			word := uuid.New()
			id := "ac93d068-7c3e-55c9-9d85-3fd3518592dc"
			_, err := db.Exec(`INSERT INTO canonical_words(id,text,normalized_text,language_code,status,created_at,updated_at) VALUES($1,'greeting','greeting','en','active',now(),now())`, word)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO word_meanings(id,word_id,part_of_speech,short_definition,meaning_order,status,created_at,updated_at) VALUES($1,$2,'noun','A welcome.',1,'active',now(),now())`, id, word)
			require.NoError(t, err)
			_, err = db.Exec(`INSERT INTO word_examples(id,meaning_id,example_text,example_order,status,created_at,updated_at) VALUES(gen_random_uuid(),$1,'An original greeting.',1,'active',now(),now())`, id)
			require.NoError(t, err)
			edit := func() {
				_, err := db.Exec(`UPDATE word_examples SET example_text='A changed greeting.' WHERE meaning_id=$1`, id)
				require.NoError(t, err)
			}
			p := &fakeProvider{identity: ProviderIdentity{Provider: "openai", Model: OpenAIModel, Voice: OpenAIVoice, InstructionVersion: OpenAIInstructionVersion}, fn: func(_ context.Context, u Utterance) ([]byte, error) {
				require.Equal(t, "An original greeting.", u.Context)
				if during {
					edit()
				}
				return testMP3, nil
			}}
			svc := NewCatalogOperator(db, p)
			_, err = svc.Generate(t.Context(), uuid.New(), CatalogBudget{1, 100, 60})
			require.NoError(t, err)
			if !during {
				edit()
			}
			report, err := svc.Audit(t.Context())
			require.NoError(t, err)
			require.Zero(t, report.Ready)
			require.Equal(t, 1, report.Missing)
			var stored int
			require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_audio_cache WHERE status='ready'`).Scan(&stored))
			if during {
				require.Zero(t, stored)
			} else {
				require.Equal(t, 1, stored, "obsolete identity must not count as current coverage")
			}
		})
	}
}

func TestCatalogPostgreSQLValidatedAuditRejectsSilentReadyWithoutMutation(t *testing.T) {
	db, u := seedCatalogFixture(t)
	svc := NewCatalogOperator(db, nil)
	key := utteranceCacheKey(svc.identity(), u)
	repo := &postgresRepository{db}
	reserved, err := repo.Reserve(t.Context(), key, 7, time.Now().UTC())
	require.NoError(t, err)
	require.NoError(t, repo.Complete(t.Context(), key, reserved.generation, silentMP3, time.Now().UTC()))
	report, err := svc.AuditValidated(t.Context())
	require.NoError(t, err)
	require.Equal(t, 1, report.Ready)
	require.Equal(t, 1, report.Validation.Checked)
	require.Equal(t, 1, report.Validation.Invalid)
	require.Zero(t, report.Validation.Valid)
	var stored []byte
	require.NoError(t, db.QueryRow(`SELECT audio FROM pronunciation_audio_cache WHERE cache_key=$1`, key).Scan(&stored))
	require.Equal(t, silentMP3, stored)
	var jobs int
	require.NoError(t, db.QueryRow(`SELECT count(*) FROM pronunciation_catalog_jobs`).Scan(&jobs))
	require.Zero(t, jobs)
}
