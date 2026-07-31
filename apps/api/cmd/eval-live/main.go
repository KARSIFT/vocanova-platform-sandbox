// Command eval-live runs the T10 live-provider AI
// evaluation pass against a real provider in staging. It is
// the single command the founder runs once, after
// VOC-032-DEP-03 (staging AI-provider credentials) is
// resolved, to record EV-22's R1-gate evidence.
//
// The command is intentionally NOT invoked from CI. Normal
// CI uses T08's evaluation-gate command (which runs against
// the deterministic mock provider) - this command is
// staging-only and one-shot, per DOC-12 §9 ("protected
// live-provider evaluation outside CI with explicit cost
// limits"). The exit codes follow the same pass/fail
// semantics as the in-CI gate (0 = pass, 1 = release-
// blocking finding) so the founder can pipe the output into
// staging-evidence.md and the R1 gate-readiness summary
// without translation.
//
// Required environment variables:
//
//	AI_PROVIDER_API_KEY    - provider API key (never logged)
//
// Required only when provider is OpenCode:
//
//	AI_PROVIDER_BASE_URL   - the OpenCode server's base URL
//
// Required only when provider is Cloudflare:
//
//	AI_PROVIDER_ACCOUNT_ID - Cloudflare account identifier
//
// Optional environment variables (overridden by flags):
//
//	AI_PROVIDER           - provider selector ("opencode",
//	                       "gemini", or "cloudflare";
//	                       default: "opencode")
//	AI_PROVIDER_MODEL     - "providerID/modelID" (default:
//	                        provider-specific)
//	AI_PROVIDER_TIMEOUT   - per-request timeout (default: 8s)
//	EVAL_LIVE_REQUEST_INTERVAL - optional delay inserted
//	                              before each provider call
//	                              (default: 0, disabled).
//	                              Added so live runs can pace
//	                              requests on free-tier limits
//	                              (especially Cloudflare).
//	EVAL_LIVE_COST_USD          - post-run billed cost in USD
//	EVAL_LIVE_COST_CEILING_USD  - pre-agreed cost ceiling
//	                              (negative = no ceiling)
//	EVAL_LIVE_OUTPUT            - file path the report is
//	                              also written to
//
// Exit codes:
//
//	0  every tracked DOC-09 §23 threshold met and cost
//	   ceiling not exceeded
//	1  release-blocking finding: at least one threshold
//	   violation OR cost ceiling exceeded
//	2  configuration error: required env var missing, bad
//	   flag, or output path not writable
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/KARSIFT/vocanova-platform/apps/api/business/aifeedback"
)

// exitSuccess / exitReleaseBlocking / exitUsageError are
// the three exit codes main returns. They are named rather
// than numeric so the reader of main does not have to
// remember which number means what.
const (
	exitSuccess         = 0
	exitReleaseBlocking = 1
	exitUsageError      = 2

	providerOpenCode   = string(aifeedback.ProviderOpenCode)
	providerGemini     = "gemini"
	providerCloudflare = "cloudflare"
)

// newProvider is the constructor the command uses to build
// the FeedbackProvider. It is a package-level variable
// rather than a hard-coded call to
// NewOpenCodeFeedbackProvider so the test suite can swap it
// for a fake provider without needing real staging
// credentials or a live OpenCode server. Production code
// always sees the default (a real OpenCodeFeedbackProvider);
// the test override is the only other assignment.
//
// Swapping this variable from a test is a test seam, not a
// production-affecting indirection: the variable is read
// exactly once per runEvalLive call, not on every
// GenerateFeedback, so a concurrent test cannot race with a
// real run.
var newProvider = func(cfg aifeedback.OpenCodeConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewOpenCodeFeedbackProvider(cfg)
}

var newGeminiProvider = func(cfg aifeedback.GeminiConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewGeminiFeedbackProvider(cfg)
}

var newCloudflareProvider = func(cfg aifeedback.CloudflareConfig) aifeedback.FeedbackProvider {
	return aifeedback.NewCloudflareFeedbackProvider(cfg)
}

var runLiveEvaluation = func(ctx context.Context, provider aifeedback.FeedbackProvider, opts aifeedback.LiveEvaluationOptions) aifeedback.LiveEvaluationReport {
	return aifeedback.RunLiveEvaluation(ctx, provider, opts)
}

type pacedFeedbackProvider struct {
	wrapped  aifeedback.FeedbackProvider
	interval time.Duration
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

// stringFlagOrEnv returns the value of the named env var
// when set and non-empty, else the supplied default. It is
// the shared resolution rule the flags below apply to their
// default values, so a flag's default reflects the env
// var's value when set and the hard-coded default only
// when the env var is unset.
func stringFlagOrEnv(envName, fallback string) string {
	if v := os.Getenv(envName); v != "" {
		return v
	}
	return fallback
}

// floatFlagOrEnv returns the value of the named env var
// parsed as a float64 when set and non-empty, else the
// supplied default. An unparseable value is treated as
// "unset" - the caller's flag default applies - and the
// error is dropped; a noisy warning is worse than a quiet
// fallback to a sensible default for an operator who
// mistyped a value.
func floatFlagOrEnv(envName string, fallback float64) float64 {
	v := os.Getenv(envName)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

// durationFlagOrEnv returns the value of the named env var
// parsed as a time.Duration when set and non-empty, else
// the supplied default. An unparseable value falls back to
// the default; see floatFlagOrEnv for the rationale.
func durationFlagOrEnv(envName string, fallback time.Duration) time.Duration {
	v := os.Getenv(envName)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

// runEvalLive is the testable entry point. It returns the
// exit code main should return; the rendered report is
// written to stdout and (when set) to the output file. All
// side effects (file writes, stderr diagnostics) are
// explicit parameters so the test suite can swap them.
//
// The function is split out of main for the same reason
// cmd/api/main.go's run() is: the existing TestRun /
// command-pipeline tests can exercise the wiring without
// spawning a real process.
func runEvalLive(args []string, stdout, stderr io.Writer, now func() time.Time) int {
	fs := flag.NewFlagSet("eval-live", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Defaults pull from env vars so a CI-style invocation
	// (env file) works without re-specifying the same value
	// on the command line. The flag is the per-invocation
	// override; the env var is the per-environment default.
	var (
		provider        = fs.String("provider", stringFlagOrEnv("AI_PROVIDER", providerOpenCode), "provider to evaluate: opencode, gemini, or cloudflare (env: AI_PROVIDER)")
		baseURL         = fs.String("base-url", stringFlagOrEnv("AI_PROVIDER_BASE_URL", ""), "OpenCode server base URL (env: AI_PROVIDER_BASE_URL)")
		accountID       = fs.String("account-id", stringFlagOrEnv("AI_PROVIDER_ACCOUNT_ID", ""), "Cloudflare account ID (env: AI_PROVIDER_ACCOUNT_ID)")
		apiKey          = fs.String("api-key", stringFlagOrEnv("AI_PROVIDER_API_KEY", ""), "provider API key (env: AI_PROVIDER_API_KEY; never logged)")
		model           = fs.String("model", stringFlagOrEnv("AI_PROVIDER_MODEL", ""), "model identifier (env: AI_PROVIDER_MODEL)")
		timeout         = fs.Duration("timeout", durationFlagOrEnv("AI_PROVIDER_TIMEOUT", 8*time.Second), "per-request timeout (env: AI_PROVIDER_TIMEOUT)")
		requestInterval = fs.Duration("request-interval", durationFlagOrEnv("EVAL_LIVE_REQUEST_INTERVAL", 0), "delay before each provider call (env: EVAL_LIVE_REQUEST_INTERVAL)")
		costUSD         = fs.Float64("cost", floatFlagOrEnv(aifeedback.LiveEvaluationCostUSDEnv, -1), "post-run billed cost in USD (env: EVAL_LIVE_COST_USD)")
		ceilingUSD      = fs.Float64("cost-ceiling", floatFlagOrEnv(aifeedback.LiveEvaluationCostCeilingUSDEnv, -1), "pre-agreed cost ceiling in USD; negative disables (env: EVAL_LIVE_COST_CEILING_USD)")
		output          = fs.String("output", stringFlagOrEnv(aifeedback.LiveEvaluationOutputEnv, ""), "optional output file path; report is also written here in addition to stdout (env: EVAL_LIVE_OUTPUT)")
		notes           = fs.String("notes", stringFlagOrEnv("EVAL_LIVE_OPERATOR_NOTES", ""), "free-text notes appended to the report (env: EVAL_LIVE_OPERATOR_NOTES)")
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
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintf(stderr, "eval-live: cannot create output file %s: %v\n", *output, err)
			return exitUsageError
		}
		_ = f.Close()
	}
	var feedbackProvider aifeedback.FeedbackProvider
	switch *provider {
	case providerGemini:
		geminiModel := *model
		feedbackProvider = newGeminiProvider(aifeedback.GeminiConfig{
			BaseURL:    *baseURL,
			APIKey:     *apiKey,
			Model:      geminiModel,
			Timeout:    *timeout,
			MaxRetries: 1,
		})
	case providerCloudflare:
		feedbackProvider = newCloudflareProvider(aifeedback.CloudflareConfig{
			APIToken:   *apiKey,
			AccountID:  *accountID,
			Model:      *model,
			BaseURL:    *baseURL,
			Timeout:    *timeout,
			MaxRetries: 1,
		})
	default:
		openCodeModel := *model
		if openCodeModel == "" {
			openCodeModel = aifeedback.DefaultOpenCodeModel
		}
		feedbackProvider = newProvider(aifeedback.OpenCodeConfig{
			BaseURL:    *baseURL,
			APIKey:     *apiKey,
			Model:      openCodeModel,
			Timeout:    *timeout,
			MaxRetries: 1,
		})
	}
	if *requestInterval > 0 {
		feedbackProvider = &pacedFeedbackProvider{
			wrapped:  feedbackProvider,
			interval: *requestInterval,
		}
	}
	// Use the supplied now() for testability; main passes
	// time.Now. RunLiveEvaluation measures its own
	// startedAt/finishedAt internally, so now() is not
	// strictly required - the parameter is kept for
	// future use (e.g. pinning a clock in tests).
	_ = now
	report := runLiveEvaluation(context.Background(), feedbackProvider, aifeedback.LiveEvaluationOptions{
		CostCeilingUSD: *ceilingUSD,
		CostUSD:        *costUSD,
		OperatorNotes:  *notes,
	})
	rendered := aifeedback.FormatLiveEvaluationReport(report)
	if _, err := io.WriteString(stdout, rendered); err != nil {
		fmt.Fprintf(stderr, "eval-live: write stdout: %v\n", err)
		return exitUsageError
	}
	if *output != "" {
		if err := os.WriteFile(*output, []byte(rendered), 0o600); err != nil {
			fmt.Fprintf(stderr, "eval-live: write output file %s: %v\n", *output, err)
			return exitUsageError
		}
		fmt.Fprintf(stderr, "eval-live: report also written to %s (mode 0600)\n", *output)
	}
	// Release-blocking: any tracked threshold violation OR
	// a cost-ceiling-exceeded. The latter is procedural
	// (the operator agreed to the ceiling in advance; a
	// run that exceeded it is a finding, not a pass), and
	// the report's CostCeilingExceeded field makes it
	// explicit which one tripped.
	if len(report.Violations) > 0 || report.CostCeilingExceeded {
		return exitReleaseBlocking
	}
	return exitSuccess
}

func main() {
	code := runEvalLive(os.Args[1:], os.Stdout, os.Stderr, time.Now)
	os.Exit(code)
}
