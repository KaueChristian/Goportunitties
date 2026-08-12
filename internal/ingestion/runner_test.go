package ingestion_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/database"
	"github.com/KaueChristian/Goportunitties/internal/ingestion"
	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/repository"
)

// fakeSource stands in for a job board.
type fakeSource struct {
	slug     string
	openings []model.Opening
	err      error
	delay    time.Duration
	calls    atomic.Int32
}

func (f *fakeSource) Slug() string { return f.slug }
func (f *fakeSource) Name() string { return f.slug }

func (f *fakeSource) Fetch(ctx context.Context) ([]model.Opening, error) {
	f.calls.Add(1)

	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}

	// Hand back a copy: the runner stamps the slug on what it receives, and a
	// real adapter would not be sharing its own backing array either.
	out := make([]model.Opening, len(f.openings))
	copy(out, f.openings)
	return out, nil
}

func opening(externalID, role string) model.Opening {
	return model.Opening{
		Role:       role,
		Company:    "Acme",
		Location:   "Remoto",
		Remote:     true,
		Link:       "https://acme.com/" + externalID,
		ExternalID: externalID,
	}
}

func newRepo(t *testing.T) repository.OpeningRepository {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	return repository.NewOpeningRepository(db)
}

func newRunner(t *testing.T, repo repository.OpeningRepository, max int, sources ...ingestion.Source) *ingestion.Runner {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return ingestion.NewRunner(repo, log, 2*time.Second, max, sources...)
}

func TestRunWritesEverySource(t *testing.T) {
	repo := newRepo(t)
	runner := newRunner(t, repo, 0,
		&fakeSource{slug: "boardA", openings: []model.Opening{opening("1", "Dev Go"), opening("2", "SRE")}},
		&fakeSource{slug: "boardB", openings: []model.Opening{opening("1", "Dev Rust")}},
	)

	summary := runner.Run(context.Background())

	fetched, created, updated, failed := summary.Totals()
	if fetched != 3 || created != 3 || updated != 0 || failed != 0 {
		t.Fatalf("totals = %d fetched, %d created, %d updated, %d failed", fetched, created, updated, failed)
	}
	if err := summary.Err(); err != nil {
		t.Fatalf("Err = %v, want nil", err)
	}

	_, total, err := repo.List(context.Background(), repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	// The same external id on two boards is two openings: the key is the pair.
	if total != 3 {
		t.Fatalf("%d openings stored, want 3", total)
	}
}

// Running the worker twice is the normal case, not the exception.
func TestRunIsIdempotentAcrossPasses(t *testing.T) {
	repo := newRepo(t)
	source := &fakeSource{slug: "boardA", openings: []model.Opening{opening("1", "Dev Go"), opening("2", "SRE")}}
	runner := newRunner(t, repo, 0, source)
	ctx := context.Background()

	runner.Run(ctx)
	summary := runner.Run(ctx)

	_, created, updated, _ := summary.Totals()
	if created != 0 || updated != 2 {
		t.Fatalf("second pass created %d and updated %d, want 0 and 2", created, updated)
	}

	_, total, err := repo.List(ctx, repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Fatalf("%d openings after two passes, want 2", total)
	}
}

// A board being down must not cost the run the boards that answered.
func TestRunKeepsGoingWhenOneSourceFails(t *testing.T) {
	repo := newRepo(t)
	runner := newRunner(t, repo, 0,
		&fakeSource{slug: "broken", err: errors.New("502 bad gateway")},
		&fakeSource{slug: "healthy", openings: []model.Opening{opening("1", "Dev Go")}},
	)

	summary := runner.Run(context.Background())

	fetched, created, _, failed := summary.Totals()
	if fetched != 1 || created != 1 || failed != 1 {
		t.Fatalf("totals = %d fetched, %d created, %d failed", fetched, created, failed)
	}
	if err := summary.Err(); err != nil {
		t.Fatalf("Err = %v, want nil while at least one source worked", err)
	}

	for _, result := range summary.Results {
		if result.Source == "broken" && result.Err == nil {
			t.Fatal("the failing source should carry its error")
		}
	}
}

func TestRunFailsWhenEverySourceFails(t *testing.T) {
	runner := newRunner(t, newRepo(t), 0,
		&fakeSource{slug: "a", err: errors.New("down")},
		&fakeSource{slug: "b", err: errors.New("down")},
	)

	if err := runner.Run(context.Background()).Err(); err == nil {
		t.Fatal("a run where nothing succeeded should report an error")
	}
}

// The slug is authority: an adapter cannot write under another source's name,
// which is what protects manual openings and every other board's rows.
func TestRunStampsTheSourceSlug(t *testing.T) {
	repo := newRepo(t)

	liar := opening("1", "Vaga forjada")
	liar.Source = model.SourceManual

	runner := newRunner(t, repo, 0, &fakeSource{slug: "boardA", openings: []model.Opening{liar}})
	runner.Run(context.Background())

	openings, _, err := repo.List(context.Background(), repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(openings) != 1 || openings[0].Source != "boardA" {
		t.Fatalf("stored source = %+v, want boardA", openings)
	}
}

func TestRunCapsWhatOneSourceContributes(t *testing.T) {
	repo := newRepo(t)

	many := make([]model.Opening, 0, 10)
	for i := 0; i < 10; i++ {
		many = append(many, opening(string(rune('a'+i)), "Vaga"))
	}

	runner := newRunner(t, repo, 4, &fakeSource{slug: "boardA", openings: many})
	summary := runner.Run(context.Background())

	fetched, created, _, _ := summary.Totals()
	if fetched != 4 || created != 4 {
		t.Fatalf("%d fetched and %d created, want 4 and 4", fetched, created)
	}
}

// One board hanging must not hold the run: the per-source timeout cancels it.
func TestRunTimesOutASlowSource(t *testing.T) {
	repo := newRepo(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	slow := &fakeSource{slug: "slow", delay: time.Second, openings: []model.Opening{opening("1", "Dev Go")}}
	fast := &fakeSource{slug: "fast", openings: []model.Opening{opening("1", "SRE")}}

	runner := ingestion.NewRunner(repo, log, 50*time.Millisecond, 0, slow, fast)

	started := time.Now()
	summary := runner.Run(context.Background())
	elapsed := time.Since(started)

	if elapsed > 900*time.Millisecond {
		t.Fatalf("the run waited %s on a hanging source", elapsed)
	}

	fetched, _, _, failed := summary.Totals()
	if failed != 1 {
		t.Fatalf("%d sources failed, want the slow one", failed)
	}
	if fetched != 1 {
		t.Fatalf("%d openings fetched, want the fast source's", fetched)
	}
}

// The sources are fetched at the same time, not one after the other.
func TestRunFetchesSourcesConcurrently(t *testing.T) {
	repo := newRepo(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	const delay = 120 * time.Millisecond
	sources := make([]ingestion.Source, 0, 4)
	for i := 0; i < 4; i++ {
		sources = append(sources, &fakeSource{
			slug:     "board" + string(rune('A'+i)),
			delay:    delay,
			openings: []model.Opening{opening("1", "Vaga")},
		})
	}

	runner := ingestion.NewRunner(repo, log, 2*time.Second, 0, sources...)

	started := time.Now()
	runner.Run(context.Background())
	elapsed := time.Since(started)

	// Sequentially this would take at least 4 × delay.
	if elapsed >= 3*delay {
		t.Fatalf("the run took %s; the sources look sequential", elapsed)
	}
}

func TestRunWithNoSources(t *testing.T) {
	summary := newRunner(t, newRepo(t), 0).Run(context.Background())

	if len(summary.Results) != 0 {
		t.Fatalf("results = %+v, want none", summary.Results)
	}
	if err := summary.Err(); err != nil {
		t.Fatalf("Err = %v, want nil — nothing was asked for", err)
	}
}

func TestRunHandlesASourceWithNothingToOffer(t *testing.T) {
	summary := newRunner(t, newRepo(t), 0, &fakeSource{slug: "empty"}).Run(context.Background())

	fetched, created, _, failed := summary.Totals()
	if fetched != 0 || created != 0 || failed != 0 {
		t.Fatalf("totals = %d/%d/%d, want zeros without a failure", fetched, created, failed)
	}
}
