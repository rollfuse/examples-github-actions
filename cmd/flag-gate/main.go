// Command flag-gate evaluates one rollfuse flag and exits 0 only if the
// resulting variation matches -want-variation — designed to sit between
// two steps of a CI/CD pipeline as a kill switch: a subsequent step that
// only runs `if: success()` (GitHub Actions) or equivalent never executes
// unless the gate passes. See the accompanying
// .github/workflows/deploy-with-flag-gate.yml for the pattern in
// context.
//
// This is deliberately a thin, single-purpose CLI rather than a GitHub
// composite Action: it's plain `go run`/a compiled binary, so the same
// pattern works unchanged in any CI system that can run a Go binary
// (GitLab CI, CircleCI, a Makefile target run from cron), not just GitHub
// Actions.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	rollfuse "github.com/rollfuse/go-sdk"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flag-gate:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		flagKey       = flag.String("flag", envOr("FLAG_KEY", ""), "the flag key to evaluate (or $FLAG_KEY)")
		subjectKey    = flag.String("subject", envOr("SUBJECT_KEY", "ci"), "the subject key to evaluate for (or $SUBJECT_KEY)")
		wantVariation = flag.String("want-variation", envOr("WANT_VARIATION", "on"), "the variation key that means \"pass\" (or $WANT_VARIATION)")
		baseURL       = flag.String("base-url", envOr("ROLLFUSE_API_BASE_URL", "http://localhost:8090"), "rollfuse API base URL (or $ROLLFUSE_API_BASE_URL)")
		credential    = flag.String("credential", envOr("ROLLFUSE_SERVICE_CREDENTIAL", ""), "Service Credential (or $ROLLFUSE_SERVICE_CREDENTIAL)")
	)
	flag.Parse()

	if *flagKey == "" {
		return fmt.Errorf("-flag (or $FLAG_KEY) is required")
	}

	client, err := rollfuse.NewClient(*baseURL, *credential)
	if err != nil {
		return fmt.Errorf("rollfuse.NewClient: %w", err)
	}
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Start(startCtx); err != nil {
		return fmt.Errorf("client.Start: %w (is the rollfuse API at %s reachable?)", err, *baseURL)
	}

	result, err := client.Evaluate(*subjectKey, *flagKey, rollfuse.WithFallback(false))
	if err != nil {
		return fmt.Errorf("evaluating %q for subject %q: %w", *flagKey, *subjectKey, err)
	}

	fmt.Printf("flag-gate: %s = %q for subject %q (reason: %s)\n", *flagKey, result.VariationKey, *subjectKey, result.Reason)

	if result.VariationKey != *wantVariation {
		fmt.Printf("flag-gate: BLOCKED — wanted variation %q, got %q\n", *wantVariation, result.VariationKey)

		return fmt.Errorf("gate failed")
	}

	fmt.Println("flag-gate: PASSED")

	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
