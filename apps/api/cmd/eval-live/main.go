// Command eval-live collects synthetic provider-adapter evidence outside CI.
// It does not exercise the learner service or replace human feedback review.
// Billing options compare recorded costs; they do not cap live spending.
// Exit codes: 0 = explicit PASS, 1 = measured failure, 2 = configuration/output
// error, 3 = incomplete evidence.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime/debug"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
)

const (
	exitSuccess         = 0
	exitReleaseBlocking = 1
	exitUsageError      = 2
	exitIncomplete      = 3
	providerOpenCode    = string(aifeedback.ProviderOpenCode)
	providerGemini      = "gemini"
	providerOpenAI      = "openai"
	providerCloudflare  = "cloudflare"
)

// Constructor seams keep command tests independent of credentials and networks.
var newProvider = func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewOpenCodeFeedbackProvider(cfg)
}
var newGeminiProvider = func(cfg aifeedback.GeminiConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewGeminiFeedbackProvider(cfg)
}
var newCloudflareProvider = func(cfg aifeedback.CloudflareConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewCloudflareFeedbackProvider(cfg)
}
var newOpenAIProvider = func(cfg aifeedback.OpenAIConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewOpenAIFeedbackProvider(cfg)
}
var runLiveEvaluation = aifeedback.RunLiveEvaluation

type pacedFeedbackProvider struct {
	wrapped  aifeedback.FeedbackProvider
	interval time.Duration
}

func (p *pacedFeedbackProvider) EvaluationIdentity() (string, string) {
	return aifeedback.ProviderEvaluationIdentity(p.wrapped)
}
func (p *pacedFeedbackProvider) GenerateFeedback(ctx context.Context, task aifeedback.ProviderTask) (*aifeedback.ProviderFeedback, error) {
	if p.interval > 0 {
		timer := time.NewTimer(p.interval)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	return p.wrapped.GenerateFeedback(ctx, task)
}
func buildCommit() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	revision, modified := "", false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}
	if revision == "" {
		return "unknown"
	}
	if modified {
		return revision + "-dirty"
	}
	return revision
}

// Resolve environment values after parsing/help. The flag package must never
// print credentials, private endpoints or notes as default help values.
// Explicit flags take precedence, including explicitly empty strings.
func resolveEnvironment(fs *flag.FlagSet) error {
	provided := make(map[string]bool)
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	for _, binding := range [][2]string{
		{"provider", "AI_PROVIDER"}, {"base-url", "AI_PROVIDER_BASE_URL"},
		{"account-id", "AI_PROVIDER_ACCOUNT_ID"}, {"api-key", "AI_PROVIDER_API_KEY"},
		{"model", "AI_PROVIDER_MODEL"}, {"timeout", "AI_PROVIDER_TIMEOUT"},
		{"request-interval", "EVAL_LIVE_REQUEST_INTERVAL"},
		{"cost", aifeedback.LiveEvaluationCostUSDEnv},
		{"cost-ceiling", aifeedback.LiveEvaluationCostCeilingUSDEnv},
		{"output", aifeedback.LiveEvaluationOutputEnv},
		{"notes", "EVAL_LIVE_OPERATOR_NOTES"},
	} {
		if provided[binding[0]] {
			continue
		}
		if value := os.Getenv(binding[1]); value != "" {
			if err := fs.Set(binding[0], value); err != nil {
				return fmt.Errorf("invalid %s; use --%s to override it", binding[1], binding[0])
			}
		}
	}
	return nil
}
func validOptionalCost(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && (value == -1 || value >= 0)
}
func runEvalLive(args []string, stdout, stderr io.Writer, _ func() time.Time) int {
	fs := flag.NewFlagSet("eval-live", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		provider        = fs.String("provider", providerOpenCode, "evaluation provider: opencode, gemini, cloudflare, openai (env: AI_PROVIDER)")
		baseURL         = fs.String("base-url", "", "provider base URL (env: AI_PROVIDER_BASE_URL)")
		accountID       = fs.String("account-id", "", "Cloudflare account ID (env: AI_PROVIDER_ACCOUNT_ID)")
		apiKey          = fs.String("api-key", "", "provider API key; prefer AI_PROVIDER_API_KEY (OpenAI shell mapping: export AI_PROVIDER_API_KEY=\"$OPENAI_API_KEY\")")
		model           = fs.String("model", "", "model identifier; OpenAI defaults to gpt-5-nano (env: AI_PROVIDER_MODEL)")
		timeout         = fs.Duration("timeout", 8*time.Second, "positive per-request timeout (env: AI_PROVIDER_TIMEOUT)")
		requestInterval = fs.Duration("request-interval", 0, "nonnegative delay before each call (env: EVAL_LIVE_REQUEST_INTERVAL)")
		costUSD         = fs.Float64("cost", -1, "recorded billed cost in USD; -1 means unknown (env: EVAL_LIVE_COST_USD)")
		ceilingUSD      = fs.Float64("cost-ceiling", -1, "billing comparison only, NOT a spending cap; -1 means unknown (env: EVAL_LIVE_COST_CEILING_USD)")
		output          = fs.String("output", "", "new private report file; existing paths are refused (env: EVAL_LIVE_OUTPUT)")
		format          = fs.String("format", "text", "report format: text or json; JSON retains complete case evidence")
		commit          = fs.String("commit", buildCommit(), "source revision (default from Go build metadata)")
		notes           = fs.String("notes", "", "non-secret operator notes (env: EVAL_LIVE_OPERATOR_NOTES)")
		help            = fs.Bool("help", false, "show usage information and exit")
	)
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return exitSuccess
		}
		return exitUsageError
	}
	if *help {
		fs.Usage()
		return exitSuccess
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(stderr, "eval-live: positional arguments are not supported")
		return exitUsageError
	}
	if err := resolveEnvironment(fs); err != nil {
		fmt.Fprintln(stderr, "eval-live:", err)
		return exitUsageError
	}
	switch *provider {
	case providerOpenCode, providerGemini, providerCloudflare, providerOpenAI:
	default:
		fmt.Fprintln(stderr, "eval-live: --provider must be opencode, gemini, cloudflare, or openai")
		return exitUsageError
	}
	if *timeout <= 0 || *requestInterval < 0 {
		fmt.Fprintln(stderr, "eval-live: --timeout must be positive and --request-interval nonnegative")
		return exitUsageError
	}
	if !validOptionalCost(*costUSD) || !validOptionalCost(*ceilingUSD) {
		fmt.Fprintln(stderr, "eval-live: --cost and --cost-ceiling must be finite, nonnegative USD amounts or -1 for unknown")
		return exitUsageError
	}
	if *format != "text" && *format != "json" {
		fmt.Fprintln(stderr, "eval-live: --format must be text or json")
		return exitUsageError
	}
	if *provider == providerOpenCode && *baseURL == "" {
		fmt.Fprintln(stderr, "eval-live: --base-url (or AI_PROVIDER_BASE_URL) is required")
		return exitUsageError
	}
	if *provider == providerCloudflare && *accountID == "" {
		fmt.Fprintln(stderr, "eval-live: --account-id (or AI_PROVIDER_ACCOUNT_ID) is required")
		return exitUsageError
	}
	if *apiKey == "" {
		fmt.Fprintln(stderr, "eval-live: --api-key (or AI_PROVIDER_API_KEY) is required")
		return exitUsageError
	}
	// Reserve a private exclusive file before any potentially paid call. Keep the
	// descriptor so a replaced path is never reopened after the evaluation.
	var outputFile *os.File
	if *output != "" {
		var err error
		outputFile, err = os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			fmt.Fprintln(stderr, "eval-live: cannot create a new private output file; choose an unused writable path")
			return exitUsageError
		}
		defer func() {
			if outputFile != nil {
				_ = outputFile.Close()
			}
		}()
	}
	var feedbackProvider aifeedback.FeedbackProvider
	switch *provider {
	case providerOpenAI:
		if *model == "" {
			*model = aifeedback.DefaultOpenAIModel
		}
		feedbackProvider = newOpenAIProvider(aifeedback.OpenAIConfig{
			BaseURL: *baseURL, APIKey: *apiKey, Model: *model, Timeout: *timeout, MaxRetries: 1,
		})
	case providerGemini:
		feedbackProvider = newGeminiProvider(aifeedback.GeminiConfig{
			BaseURL: *baseURL, APIKey: *apiKey, Model: *model, Timeout: *timeout, MaxRetries: 1,
		})
	case providerCloudflare:
		feedbackProvider = newCloudflareProvider(aifeedback.CloudflareConfig{
			BaseURL: *baseURL, APIToken: *apiKey, AccountID: *accountID,
			Model: *model, Timeout: *timeout, MaxRetries: 1,
		})
	case providerOpenCode:
		if *model == "" {
			*model = aifeedback.DefaultOpenCodeModel
		}
		feedbackProvider = newProvider(aifeedback.OpenCodeConfig{
			BaseURL: *baseURL, APIKey: *apiKey, Model: *model, Timeout: *timeout, MaxRetries: 1,
		})
	}
	if *requestInterval > 0 {
		feedbackProvider = &pacedFeedbackProvider{wrapped: feedbackProvider, interval: *requestInterval}
	}
	report := runLiveEvaluation(context.Background(), feedbackProvider, aifeedback.LiveEvaluationOptions{
		CostCeilingUSD: *ceilingUSD, CostUSD: *costUSD, OperatorNotes: *notes, Commit: *commit,
		Timeout: *timeout, RequestInterval: *requestInterval, MaxRetries: 1,
	})
	rendered := aifeedback.FormatLiveEvaluationReport(report)
	if *format == "json" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(stderr, "eval-live: cannot encode report as JSON")
			return exitUsageError
		}
		rendered = string(data) + "\n"
	}
	// Persist evidence first, so a closed stdout pipe cannot discard a paid run.
	if outputFile != nil {
		if _, err := io.WriteString(outputFile, rendered); err != nil {
			fmt.Fprintln(stderr, "eval-live: cannot write output file")
			return exitUsageError
		}
		if err := outputFile.Sync(); err != nil {
			fmt.Fprintln(stderr, "eval-live: cannot sync output file")
			return exitUsageError
		}
		closeErr := outputFile.Close()
		outputFile = nil
		if closeErr != nil {
			fmt.Fprintln(stderr, "eval-live: cannot close output file")
			return exitUsageError
		}
		fmt.Fprintln(stderr, "eval-live: report also written to the requested private output file")
	}
	if _, err := io.WriteString(stdout, rendered); err != nil {
		fmt.Fprintln(stderr, "eval-live: cannot write stdout")
		return exitUsageError
	}
	if report.AcceptanceState == "FAIL" || len(report.Violations) > 0 || report.CostCeilingExceeded {
		return exitReleaseBlocking
	}
	if report.AcceptanceState == "PASS" {
		return exitSuccess
	}
	return exitIncomplete
}
func main() { os.Exit(runEvalLive(os.Args[1:], os.Stdout, os.Stderr, time.Now)) }
