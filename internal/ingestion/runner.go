package ingestion

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/repository"
)

// Runner fetches from every source and writes what comes back.
type Runner struct {
	sources []Source
	repo    repository.OpeningRepository
	log     *slog.Logger

	// timeout bounds one source. A board that hangs must not hold the run.
	timeout time.Duration
	// maxPerSource caps how much of a board is taken in one pass.
	maxPerSource int
}

// NewRunner builds a Runner.
func NewRunner(
	repo repository.OpeningRepository,
	log *slog.Logger,
	timeout time.Duration,
	maxPerSource int,
	sources ...Source,
) *Runner {
	return &Runner{
		sources:      sources,
		repo:         repo,
		log:          log,
		timeout:      timeout,
		maxPerSource: maxPerSource,
	}
}

// batch is one source's answer travelling from a fetcher to the writer.
type batch struct {
	source   Source
	openings []model.Opening
	err      error
}

// Run fetches every source concurrently and writes the batches one at a time.
//
// The split is deliberate. Fetching is I/O against several independent boards,
// so it fans out; writing goes to SQLite, which serialises writers anyway, so a
// second writer would only trade parallelism for lock contention. One funnel
// also keeps each source's batch in its own transaction.
//
// A source that fails is recorded and skipped: the others still land.
func (r *Runner) Run(ctx context.Context) Summary {
	batches := make(chan batch)

	var wg sync.WaitGroup
	for _, source := range r.sources {
		wg.Add(1)

		go func(source Source) {
			defer wg.Done()

			fetchCtx, cancel := context.WithTimeout(ctx, r.timeout)
			defer cancel()

			started := time.Now()
			openings, err := safeFetch(fetchCtx, source)

			r.log.LogAttrs(ctx, levelFor(err), "source fetched",
				slog.String("source", source.Slug()),
				slog.Int("openings", len(openings)),
				slog.Duration("took", time.Since(started)),
				slog.Any("error", err),
			)

			select {
			case batches <- batch{source: source, openings: openings, err: err}:
			case <-ctx.Done():
			}
		}(source)
	}

	go func() {
		wg.Wait()
		close(batches)
	}()

	summary := Summary{Results: make([]Result, 0, len(r.sources))}
	for received := range batches {
		summary.Results = append(summary.Results, r.write(ctx, received))
	}

	return summary
}

// safeFetch isolates one adapter's Fetch from the rest of the run.
//
// A source is third-party code this process does not control, and an
// unrecovered panic in any goroutine kills the entire process — not just the
// goroutine it occurred in. Recovering here turns that into an ordinary
// failed Result, giving a misbehaving adapter the same "skipped, others still
// land" treatment an error already gets.
func safeFetch(ctx context.Context, source Source) (openings []model.Opening, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("source %s panicked: %v", source.Slug(), r)
		}
	}()
	return source.Fetch(ctx)
}

// write persists one source's batch, stamping the slug on every opening so a
// misbehaving adapter cannot write under someone else's name.
func (r *Runner) write(ctx context.Context, received batch) Result {
	slug := received.source.Slug()
	result := Result{Source: slug, Err: received.err}

	if received.err != nil {
		return result
	}

	openings := received.openings
	if r.maxPerSource > 0 && len(openings) > r.maxPerSource {
		openings = openings[:r.maxPerSource]
	}
	for i := range openings {
		openings[i].Source = slug
	}

	result.Fetched = len(openings)
	if len(openings) == 0 {
		return result
	}

	written, err := r.repo.UpsertBatch(ctx, openings)
	if err != nil {
		result.Err = err
		r.log.ErrorContext(ctx, "source not written",
			slog.String("source", slug),
			slog.String("error", err.Error()),
		)
		return result
	}

	result.Created, result.Updated = written.Created, written.Updated
	r.log.InfoContext(ctx, "source written",
		slog.String("source", slug),
		slog.Int64("created", written.Created),
		slog.Int64("updated", written.Updated),
	)
	return result
}

func levelFor(err error) slog.Level {
	if err != nil {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}
