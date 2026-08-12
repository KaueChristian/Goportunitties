// Command worker fills the index from public job boards.
//
// It runs once and exits when INGESTION_INTERVAL is 0 — the shape a cron job or
// a manual invocation wants — and otherwise keeps running on that interval.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/config"
	"github.com/KaueChristian/Goportunitties/internal/database"
	"github.com/KaueChristian/Goportunitties/internal/ingestion"
	"github.com/KaueChristian/Goportunitties/internal/logger"
	"github.com/KaueChristian/Goportunitties/internal/repository"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("worker exited with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run() error {
	settings := config.Load()
	log := logger.New(settings.Env).With(slog.String("component", "worker"))

	db, err := database.Open(settings.DBPath, !settings.IsProduction())
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("closing database", slog.String("error", err.Error()))
		}
	}()

	client := ingestion.NewClient(settings.Ingestion.Timeout, settings.Ingestion.UserAgent)
	sources, err := ingestion.Build(client, ingestion.Options{
		MaxPerSource: settings.Ingestion.MaxPerSource,
		GitHubToken:  settings.Ingestion.GitHubToken,
	}, settings.Ingestion.Sources)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return fmt.Errorf("no ingestion source configured; available: %v", ingestion.Available())
	}

	runner := ingestion.NewRunner(
		repository.NewOpeningRepository(db),
		log,
		settings.Ingestion.Timeout,
		settings.Ingestion.MaxPerSource,
		sources...,
	)

	// The signal cancels the context the fetches run under, so a run in flight
	// stops at its next boundary instead of being killed mid-transaction.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("worker started",
		slog.String("version", version),
		slog.Any("sources", settings.Ingestion.Sources),
		slog.Duration("interval", settings.Ingestion.Interval),
	)

	if settings.Ingestion.Interval == 0 {
		return once(ctx, runner, log)
	}

	return loop(ctx, runner, log, settings.Ingestion.Interval)
}

// once performs a single pass and reports whether it was worth anything.
func once(ctx context.Context, runner *ingestion.Runner, log *slog.Logger) error {
	summary := runner.Run(ctx)
	report(ctx, log, summary)

	// Only a run where every source failed is an exit code: one board being
	// down should not fail a cron job that got everything else.
	return summary.Err()
}

// loop runs on a ticker until the context is cancelled.
func loop(ctx context.Context, runner *ingestion.Runner, log *slog.Logger, interval time.Duration) error {
	// The first pass happens immediately; waiting a full interval would leave a
	// freshly started worker looking broken.
	report(ctx, log, runner.Run(ctx))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("shutdown signal received; worker stopped")
			return nil
		case <-ticker.C:
			report(ctx, log, runner.Run(ctx))
		}
	}
}

func report(ctx context.Context, log *slog.Logger, summary ingestion.Summary) {
	fetched, created, updated, failed := summary.Totals()

	log.LogAttrs(ctx, slog.LevelInfo, "ingestion run finished",
		slog.Int("fetched", fetched),
		slog.Int64("created", created),
		slog.Int64("updated", updated),
		slog.Int("failedSources", failed),
	)
}
