// pronunciation-catalog audits public word audio by default. Paid generation is
// an explicit operator job with durable immutable budgets, never an HTTP route.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/KARSIFT/vocanova-platform/apps/api/business/pronunciation"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func run(args []string) error {
	fs := flag.NewFlagSet("pronunciation-catalog", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	validateAudio := fs.Bool("validate-audio", false, "Read-only streaming decode and signal validation of current ready word clips")
	generate := fs.Bool("generate", false, "Explicitly generate missing word clips (paid)")
	jobID := fs.String("job", "", "Stable operator job UUID; reuse to resume")
	attempts := fs.Int("max-generations", 0, "Maximum attempted calls for entire job, including failures")
	characters := fs.Int("max-characters", 0, "Maximum spoken input characters for entire job")
	seconds := fs.Float64("max-audio-seconds", 0, "Stop before another call when observed duration reaches this value")
	if err := fs.Parse(args); err != nil || fs.NArg() != 0 {
		return fmt.Errorf("invalid catalog arguments")
	}
	if !*generate && (*attempts != 0 || *characters != 0 || *seconds != 0) {
		return fmt.Errorf("job budgets require explicit --generate")
	}
	if *generate && *validateAudio {
		return fmt.Errorf("--validate-audio is read-only and cannot combine with --generate")
	}
	var job uuid.UUID
	if *jobID != "" {
		var err error
		job, err = uuid.Parse(*jobID)
		if err != nil || job == uuid.Nil {
			return fmt.Errorf("invalid job UUID")
		}
	}
	budget := pronunciation.CatalogBudget{MaxGenerations: *attempts, MaxCharacters: *characters, MaxAudioSeconds: *seconds}
	if *generate {
		if job == uuid.Nil {
			return fmt.Errorf("--generate requires a stable --job UUID")
		}
		if err := budget.Validate(); err != nil {
			return err
		}
		if os.Getenv("PRONUNCIATION_PROVIDER") != "openai" || strings.TrimSpace(os.Getenv("PRONUNCIATION_API_KEY")) == "" || !speechEnabled() {
			return fmt.Errorf("dedicated OpenAI speech configuration unavailable")
		}
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is required in the operator environment")
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("database unavailable")
	}
	defer db.Close()
	db.SetMaxOpenConns(4)
	var provider pronunciation.Provider
	if *generate {
		provider = pronunciation.NewOpenAIProvider(pronunciation.OpenAIConfig{APIKey: strings.TrimSpace(os.Getenv("PRONUNCIATION_API_KEY"))})
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 4*time.Hour)
	defer cancel()
	operator := pronunciation.NewCatalogOperator(db, provider)
	var report pronunciation.CatalogReport
	if *generate {
		report, err = operator.Generate(ctx, job, budget)
		receiptCtx, cancelReceipt := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		if receipt, receiptErr := operator.JobReport(receiptCtx, job); receiptErr == nil {
			report = receipt
		}
		cancelReceipt()
	} else if job != uuid.Nil {
		report, err = operator.JobReport(ctx, job)
		if err == nil && *validateAudio {
			validated, validationErr := operator.AuditValidated(ctx)
			report.Validation = validated.Validation
			err = validationErr
		}
	} else {
		if *validateAudio {
			report, err = operator.AuditValidated(ctx)
		} else {
			report, err = operator.Audit(ctx)
		}
	}
	// The database ledger is authoritative; JSON stdout can be saved as a receipt.
	if encodeErr := json.NewEncoder(os.Stdout).Encode(report); encodeErr != nil {
		return fmt.Errorf("report output unavailable")
	}
	if err != nil {
		return fmt.Errorf("catalog operation stopped; inspect the job ledger and sanitized report before resuming")
	}
	return nil
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func speechEnabled() bool {
	value := os.Getenv("AI_FEATURES_ENABLED")
	if value == "" {
		return true
	}
	enabled, err := strconv.ParseBool(value)
	return err == nil && enabled
}
