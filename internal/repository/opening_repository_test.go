package repository_test

import (
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/KaueChristian/Goportunitties/internal/database"
	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/repository"
	"gorm.io/gorm"
)

// newRepo builds a repository over a throwaway SQLite file, migrated through
// the same code path production uses.
func newRepo(t *testing.T) (repository.OpeningRepository, *gorm.DB) {
	t.Helper()

	db, err := database.Open(filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })

	return repository.NewOpeningRepository(db), db
}

func seed(t *testing.T, repo repository.OpeningRepository, openings ...model.Opening) []model.Opening {
	t.Helper()

	stored := make([]model.Opening, 0, len(openings))
	for i := range openings {
		if err := repo.Create(context.Background(), &openings[i]); err != nil {
			t.Fatalf("seeding opening %d: %v", i, err)
		}
		stored = append(stored, openings[i])
	}
	return stored
}

func sample() []model.Opening {
	// Each row carries its own identity: the unique index does not accept two
	// openings sharing one, which is exactly the point of it.
	return []model.Opening{
		{Role: "Desenvolvedor Go", Company: "Acme", Location: "São Paulo, SP", Remote: true, Link: "https://acme.com/1", Salary: 15000, Source: model.SourceManual, ExternalID: "sample-1"},
		{Role: "Desenvolvedor React", Company: "Globex", Location: "São Paulo, SP", Remote: false, Link: "https://globex.com/2", Salary: 9000, Source: model.SourceManual, ExternalID: "sample-2"},
		{Role: "SRE", Company: "Initech", Location: "Curitiba, PR", Remote: true, Link: "https://initech.com/3", Salary: 21500, Source: model.SourceManual, ExternalID: "sample-3"},
	}
}

func boolPtr(v bool) *bool { return &v }

func TestCreateAndFindByID(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()

	opening := model.Opening{Role: "Desenvolvedor Go", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com/1", Salary: 15000}
	if err := repo.Create(ctx, &opening); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if opening.ID == 0 {
		t.Fatal("Create should populate the generated id")
	}

	found, err := repo.FindByID(ctx, opening.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Role != opening.Role || found.Salary != opening.Salary {
		t.Fatalf("FindByID returned %+v, want %+v", found, opening)
	}
}

func TestFindByIDMissingReturnsGormNotFound(t *testing.T) {
	repo, _ := newRepo(t)

	_, err := repo.FindByID(context.Background(), 4242)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestDeleteIsSoftAndHidesTheRecord(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	stored := seed(t, repo, sample()...)
	if err := repo.Delete(ctx, &stored[0]); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, err := repo.FindByID(ctx, stored[0].ID); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("deleted opening should be invisible, got %v", err)
	}

	var rows int64
	db.Unscoped().Model(&model.Opening{}).Where("id = ?", stored[0].ID).Count(&rows)
	if rows != 1 {
		t.Fatalf("soft delete should keep the row, found %d", rows)
	}

	_, total, err := repo.List(ctx, repository.Filter{Limit: 10})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 2 {
		t.Fatalf("total after delete = %d, want 2", total)
	}
}

// ingested builds an opening as a worker would hand it over.
func ingested(source, externalID, role string, salary int64) model.Opening {
	return model.Opening{
		Role:       role,
		Company:    "Acme",
		Location:   "Remoto",
		Remote:     true,
		Link:       "https://acme.com/" + externalID,
		Salary:     salary,
		Source:     source,
		ExternalID: externalID,
	}
}

func countRows(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	var rows int64
	if err := db.Model(&model.Opening{}).Count(&rows).Error; err != nil {
		t.Fatalf("counting openings: %v", err)
	}
	return rows
}

// The property the whole ingestion step rests on: running a source twice must
// not duplicate anything.
func TestUpsertBatchIsIdempotent(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	batch := []model.Opening{
		ingested("remoteok", "1", "Desenvolvedor Go", 15000),
		ingested("remoteok", "2", "SRE", 21500),
	}

	first, err := repo.UpsertBatch(ctx, batch)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if first.Created != 2 || first.Updated != 0 {
		t.Fatalf("first run = %+v, want 2 created", first)
	}

	second, err := repo.UpsertBatch(ctx, batch)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if second.Created != 0 || second.Updated != 2 {
		t.Fatalf("second run = %+v, want 2 updated", second)
	}

	if rows := countRows(t, db); rows != 2 {
		t.Fatalf("%d rows after two runs, want 2", rows)
	}
}

func TestUpsertBatchRefreshesChangedFields(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()

	if _, err := repo.UpsertBatch(ctx, []model.Opening{ingested("remoteok", "1", "Dev Go", 15000)}); err != nil {
		t.Fatalf("first run: %v", err)
	}

	// The source raised the salary and renamed the role.
	changed := ingested("remoteok", "1", "Desenvolvedor Go Sênior", 19000)
	if _, err := repo.UpsertBatch(ctx, []model.Opening{changed}); err != nil {
		t.Fatalf("second run: %v", err)
	}

	openings, _, err := repo.List(ctx, repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(openings) != 1 {
		t.Fatalf("%d openings, want 1", len(openings))
	}
	if openings[0].Role != "Desenvolvedor Go Sênior" || openings[0].Salary != 19000 {
		t.Fatalf("row was not refreshed: %+v", openings[0])
	}
}

// Ingestion must never touch what the user typed. It cannot, because a manual
// row's source never matches an ingested batch's — this pins that down.
func TestUpsertBatchLeavesManualOpeningsAlone(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	manual := model.Opening{
		Role: "Vaga minha", Company: "Acme", Location: "Remoto", Remote: true,
		Link: "https://acme.com/1", Salary: 30000,
		Source: model.SourceManual, ExternalID: "manual-1",
	}
	if err := repo.Create(ctx, &manual); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Same company and link, but coming from a provider.
	if _, err := repo.UpsertBatch(ctx, []model.Opening{ingested("remoteok", "1", "Vaga da fonte", 1000)}); err != nil {
		t.Fatalf("UpsertBatch: %v", err)
	}

	stored, err := repo.FindByID(ctx, manual.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if stored.Role != "Vaga minha" || stored.Salary != 30000 {
		t.Fatalf("ingestion overwrote a manual opening: %+v", stored)
	}
	if rows := countRows(t, db); rows != 2 {
		t.Fatalf("%d rows, want 2 — the ingested one is a separate record", rows)
	}
}

// An opening the user dismissed must not come back on the next run.
func TestUpsertBatchDoesNotResurrectDeletedOpenings(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	batch := []model.Opening{ingested("remoteok", "1", "Dev Go", 15000)}
	if _, err := repo.UpsertBatch(ctx, batch); err != nil {
		t.Fatalf("first run: %v", err)
	}

	openings, _, err := repo.List(ctx, repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if err := repo.Delete(ctx, &openings[0]); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	result, err := repo.UpsertBatch(ctx, batch)
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	// A soft-deleted row still holds its slot in the unique index, so it counts
	// as known rather than new.
	if result.Created != 0 || result.Updated != 1 {
		t.Fatalf("second run = %+v, want 1 updated", result)
	}

	_, total, err := repo.List(ctx, repository.Filter{})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 0 {
		t.Fatalf("the dismissed opening came back: %d visible", total)
	}
	if rows := countRows(t, db.Unscoped()); rows != 1 {
		t.Fatalf("%d rows stored, want 1 — no copy was inserted", rows)
	}
}

// Two sources can carry the same id without colliding: the key is the pair.
func TestUpsertBatchKeepsSourcesApart(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	if _, err := repo.UpsertBatch(ctx, []model.Opening{ingested("remoteok", "1", "Dev Go", 15000)}); err != nil {
		t.Fatalf("remoteok: %v", err)
	}
	if _, err := repo.UpsertBatch(ctx, []model.Opening{ingested("weworkremotely", "1", "Dev Rust", 16000)}); err != nil {
		t.Fatalf("weworkremotely: %v", err)
	}

	if rows := countRows(t, db); rows != 2 {
		t.Fatalf("%d rows, want 2", rows)
	}
}

func TestUpsertBatchRejectsMixedSources(t *testing.T) {
	repo, db := newRepo(t)

	_, err := repo.UpsertBatch(context.Background(), []model.Opening{
		ingested("remoteok", "1", "Dev Go", 15000),
		ingested("weworkremotely", "2", "Dev Rust", 16000),
	})

	if !errors.Is(err, repository.ErrMixedSources) {
		t.Fatalf("got %v, want repository.ErrMixedSources", err)
	}
	if rows := countRows(t, db); rows != 0 {
		t.Fatalf("%d rows written, want none", rows)
	}
}

// The batch is one transaction: a failure halfway leaves nothing behind.
func TestUpsertBatchIsAtomic(t *testing.T) {
	repo, db := newRepo(t)

	// The second entry repeats the first's identity, which the unique index
	// rejects mid-statement.
	_, err := repo.UpsertBatch(context.Background(), []model.Opening{
		ingested("remoteok", "1", "Dev Go", 15000),
		{Role: "Sem link", Company: "Acme", Location: "Remoto", Source: "remoteok", ExternalID: "1", Salary: -1},
	})
	if err == nil {
		t.Skip("this database accepted the batch; nothing to assert about a rollback")
	}

	if rows := countRows(t, db.Unscoped()); rows != 0 {
		t.Fatalf("%d rows survived a failed batch, want none", rows)
	}
}

func TestUpsertBatchHandlesAnEmptyBatch(t *testing.T) {
	repo, _ := newRepo(t)

	result, err := repo.UpsertBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("UpsertBatch: %v", err)
	}
	if (result != repository.UpsertResult{}) {
		t.Fatalf("result = %+v, want the zero value", result)
	}
}

func TestUpsertBatchWritesBeyondOneChunk(t *testing.T) {
	repo, db := newRepo(t)

	batch := make([]model.Opening, 0, 250)
	for i := 0; i < 250; i++ {
		batch = append(batch, ingested("remoteok", strconv.Itoa(i), "Vaga", int64(1000+i)))
	}

	result, err := repo.UpsertBatch(context.Background(), batch)
	if err != nil {
		t.Fatalf("UpsertBatch: %v", err)
	}
	if result.Created != 250 {
		t.Fatalf("created %d, want 250", result.Created)
	}
	if rows := countRows(t, db); rows != 250 {
		t.Fatalf("%d rows, want 250", rows)
	}
}

func TestListFilters(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	tests := []struct {
		name   string
		filter repository.Filter
		want   int64
	}{
		{"no filter", repository.Filter{}, 3},
		{"search matches role", repository.Filter{Search: "desenvolvedor"}, 2},
		{"search is case insensitive", repository.Filter{Search: "GLOBEX"}, 1},
		{"search matches company", repository.Filter{Search: "initech"}, 1},
		{"search with no match", repository.Filter{Search: "kubernetes"}, 0},
		{"location", repository.Filter{Location: "Curitiba, PR"}, 1},
		{"remote only", repository.Filter{Remote: boolPtr(true)}, 2},
		{"onsite only", repository.Filter{Remote: boolPtr(false)}, 1},
		{"minimum salary", repository.Filter{MinSalary: 15000}, 2},
		{"combined", repository.Filter{Remote: boolPtr(true), MinSalary: 20000}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openings, total, err := repo.List(context.Background(), tt.filter)
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if total != tt.want {
				t.Fatalf("total = %d, want %d", total, tt.want)
			}
			if int64(len(openings)) != tt.want {
				t.Fatalf("returned %d openings, want %d", len(openings), tt.want)
			}
		})
	}
}

func TestListSorting(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	tests := []struct {
		sort  string
		first string
	}{
		{"salary-desc", "SRE"},
		{"salary-asc", "Desenvolvedor React"},
		{"role", "Desenvolvedor Go"},
	}

	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			openings, _, err := repo.List(context.Background(), repository.Filter{Sort: tt.sort})
			if err != nil {
				t.Fatalf("List: %v", err)
			}
			if openings[0].Role != tt.first {
				t.Fatalf("first opening = %q, want %q", openings[0].Role, tt.first)
			}
		})
	}
}

func TestListPagination(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)
	ctx := context.Background()

	page, total, err := repo.List(ctx, repository.Filter{Sort: "role", Limit: 2, Offset: 0})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 || len(page) != 2 {
		t.Fatalf("first page: %d items of %d total, want 2 of 3", len(page), total)
	}

	page, total, err = repo.List(ctx, repository.Filter{Sort: "role", Limit: 2, Offset: 2})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if total != 3 || len(page) != 1 {
		t.Fatalf("second page: %d items of %d total, want 1 of 3", len(page), total)
	}

	// A window past the end is an empty page, not an error.
	page, total, err = repo.List(ctx, repository.Filter{Limit: 2, Offset: 99})
	if err != nil {
		t.Fatalf("List past the end: %v", err)
	}
	if len(page) != 0 || total != 3 {
		t.Fatalf("page past the end: %d items of %d total, want 0 of 3", len(page), total)
	}
}

func TestSuggest(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)
	ctx := context.Background()

	tests := []struct {
		name  string
		term  string
		want  []repository.Suggestion
		limit int
	}{
		{
			name:  "matches a role",
			term:  "sre",
			limit: 8,
			want:  []repository.Suggestion{{Value: "SRE", Kind: "role", Count: 1}},
		},
		{
			name:  "matches a company",
			term:  "globex",
			limit: 8,
			want:  []repository.Suggestion{{Value: "Globex", Kind: "company", Count: 1}},
		},
		{
			name:  "is case insensitive",
			term:  "INITECH",
			limit: 8,
			want:  []repository.Suggestion{{Value: "Initech", Kind: "company", Count: 1}},
		},
		{
			name:  "matches in the middle of a word",
			term:  "envolvedor",
			limit: 8,
			want: []repository.Suggestion{
				{Value: "Desenvolvedor Go", Kind: "role", Count: 1},
				{Value: "Desenvolvedor React", Kind: "role", Count: 1},
			},
		},
		{name: "no match", term: "kubernetes", limit: 8, want: []repository.Suggestion{}},
		{name: "empty term", term: "", limit: 8, want: []repository.Suggestion{}},
		{name: "limit of zero", term: "sre", limit: 0, want: []repository.Suggestion{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.Suggest(ctx, tt.term, tt.limit)
			if err != nil {
				t.Fatalf("Suggest: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("entry %d = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// The same company on several openings is one suggestion, not one per row.
func TestSuggestGroupsAndRanksByFrequency(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo,
		model.Opening{Role: "Dev Go", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com/1", Salary: 1},
		model.Opening{Role: "Dev Rust", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com/2", Salary: 2},
		model.Opening{Role: "Dev Go", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com/3", Salary: 3},
		model.Opening{Role: "Dev Elixir", Company: "Acmezinha", Location: "Remoto", Remote: true, Link: "https://acmez.com/4", Salary: 4},
	)

	got, err := repo.Suggest(context.Background(), "acme", 8)
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}

	want := []repository.Suggestion{
		{Value: "Acme", Kind: "company", Count: 3},
		{Value: "Acmezinha", Kind: "company", Count: 1},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("entry %d = %+v, want %+v — most frequent first", i, got[i], want[i])
		}
	}
}

func TestSuggestHonoursTheLimitAcrossBothColumns(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	got, err := repo.Suggest(context.Background(), "e", 2)
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	// "e" matches roles and companies alike; the cap applies to the merged list.
	if len(got) != 2 {
		t.Fatalf("got %d suggestions, want 2", len(got))
	}
}

// Raw SQL does not get GORM's soft-delete condition for free.
func TestSuggestIgnoresDeletedOpenings(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	stored := seed(t, repo, sample()...)

	if err := repo.Delete(ctx, &stored[2]); err != nil { // SRE @ Initech
		t.Fatalf("Delete: %v", err)
	}

	got, err := repo.Suggest(ctx, "initech", 8)
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("a deleted opening still suggests: %+v", got)
	}
}

func TestCountByRemoteIgnoresTheRemoteFilter(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	// Asking only for remote openings must still report how many on-site ones a
	// click on the other option would return.
	remote, onsite, err := repo.CountByRemote(context.Background(), repository.Filter{Remote: boolPtr(true)})
	if err != nil {
		t.Fatalf("CountByRemote: %v", err)
	}
	if remote != 2 || onsite != 1 {
		t.Fatalf("counts = (remote %d, onsite %d), want (2, 1)", remote, onsite)
	}
}

func TestCountByRemoteHonoursTheOtherFilters(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	remote, onsite, err := repo.CountByRemote(context.Background(), repository.Filter{
		Location: "São Paulo, SP",
		Remote:   boolPtr(false),
	})
	if err != nil {
		t.Fatalf("CountByRemote: %v", err)
	}
	if remote != 1 || onsite != 1 {
		t.Fatalf("counts = (remote %d, onsite %d), want (1, 1)", remote, onsite)
	}
}

func TestCountByLocationIgnoresTheLocationFilter(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	counts, err := repo.CountByLocation(context.Background(), repository.Filter{Location: "Curitiba, PR"})
	if err != nil {
		t.Fatalf("CountByLocation: %v", err)
	}
	if len(counts) != 2 {
		t.Fatalf("got %d locations, want 2", len(counts))
	}
	if counts[0].Location != "Curitiba, PR" || counts[0].Count != 1 {
		t.Fatalf("first bucket = %+v, want Curitiba with 1", counts[0])
	}
	if counts[1].Location != "São Paulo, SP" || counts[1].Count != 2 {
		t.Fatalf("second bucket = %+v, want São Paulo with 2", counts[1])
	}
}

func TestAggregates(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	aggregates, err := repo.Aggregates(context.Background())
	if err != nil {
		t.Fatalf("Aggregates: %v", err)
	}

	want := repository.Aggregates{
		Total:         3,
		Remote:        2,
		Companies:     3,
		AverageSalary: (15000 + 9000 + 21500) / 3,
		MaxSalary:     21500,
	}
	if aggregates != want {
		t.Fatalf("aggregates = %+v, want %+v", aggregates, want)
	}
}

// The salary figures describe the openings that state a salary. A posting saved
// as "a combinar" is an absence, not a salary of zero.
func TestSalaryFiguresIgnoreUnstatedSalaries(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()

	openings := sample() // 15000, 9000, 21500
	openings = append(openings, model.Opening{
		Role: "Estágio", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com/4", Salary: 0,
	})
	seed(t, repo, openings...)

	aggregates, err := repo.Aggregates(ctx)
	if err != nil {
		t.Fatalf("Aggregates: %v", err)
	}
	if aggregates.Total != 4 {
		t.Fatalf("total = %d, want 4 — the opening still counts", aggregates.Total)
	}

	wantAverage := int64((15000 + 9000 + 21500) / 3)
	if aggregates.AverageSalary != wantAverage {
		t.Fatalf("average = %d, want %d — a salary of 0 must not be averaged in",
			aggregates.AverageSalary, wantAverage)
	}

	median, err := repo.MedianSalary(ctx)
	if err != nil {
		t.Fatalf("MedianSalary: %v", err)
	}
	if median != 15000 {
		t.Fatalf("median = %d, want 15000", median)
	}
}

func TestMedianSalary(t *testing.T) {
	tests := []struct {
		name     string
		salaries []int64
		want     int64
	}{
		{"empty index", nil, 0},
		{"single opening", []int64{15000}, 15000},
		{"odd count takes the middle", []int64{9000, 15000, 21500}, 15000},
		{"even count averages the two middle", []int64{9000, 12000, 16000, 21500}, 14000},
		// The point of the median: one outlier moves it far less than the mean.
		{"an outlier barely moves it", []int64{9000, 12000, 15000, 16000, 500000}, 15000},
		{"every salary unstated", []int64{0, 0}, 0},
		{"order of insertion is irrelevant", []int64{21500, 9000, 15000}, 15000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _ := newRepo(t)

			openings := make([]model.Opening, 0, len(tt.salaries))
			for i, salary := range tt.salaries {
				openings = append(openings, model.Opening{
					Role:     "Vaga",
					Company:  "Empresa",
					Location: "Remoto",
					Remote:   true,
					Link:     "https://empresa.com/" + strconv.Itoa(i),
					Salary:   salary,
				})
			}
			seed(t, repo, openings...)

			got, err := repo.MedianSalary(context.Background())
			if err != nil {
				t.Fatalf("MedianSalary: %v", err)
			}
			if got != tt.want {
				t.Fatalf("median = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMedianSalaryIgnoresDeletedOpenings(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	stored := seed(t, repo, sample()...) // 15000, 9000, 21500 -> median 15000

	if err := repo.Delete(ctx, &stored[0]); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	median, err := repo.MedianSalary(ctx)
	if err != nil {
		t.Fatalf("MedianSalary: %v", err)
	}
	// Left with 9000 and 21500.
	if median != 15250 {
		t.Fatalf("median = %d, want 15250", median)
	}
}

func TestAggregatesOnEmptyTable(t *testing.T) {
	repo, _ := newRepo(t)

	aggregates, err := repo.Aggregates(context.Background())
	if err != nil {
		t.Fatalf("Aggregates: %v", err)
	}
	if (aggregates != repository.Aggregates{}) {
		t.Fatalf("aggregates on empty table = %+v, want zero value", aggregates)
	}
}

func TestMonthlyCountsGroupsByMonth(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()
	stored := seed(t, repo, sample()...)

	// Backdate two openings so the series spans three distinct months.
	now := time.Now().UTC()
	backdate := func(id uint, months int) {
		when := now.AddDate(0, months, 0)
		if err := db.Model(&model.Opening{}).Where("id = ?", id).UpdateColumn("created_at", when).Error; err != nil {
			t.Fatalf("backdating opening %d: %v", id, err)
		}
	}
	backdate(stored[1].ID, -1)
	backdate(stored[2].ID, -2)

	counts, err := repo.MonthlyCounts(ctx, 9)
	if err != nil {
		t.Fatalf("MonthlyCounts: %v", err)
	}
	if len(counts) != 3 {
		t.Fatalf("got %d months, want 3: %+v", len(counts), counts)
	}

	// Oldest first, so the chart can draw left to right.
	wantOldest := now.AddDate(0, -2, 0).Format("2006-01")
	wantNewest := now.Format("2006-01")
	if counts[0].Month != wantOldest {
		t.Fatalf("first month = %q, want %q", counts[0].Month, wantOldest)
	}
	if counts[2].Month != wantNewest {
		t.Fatalf("last month = %q, want %q", counts[2].Month, wantNewest)
	}
	for _, count := range counts {
		if count.Count != 1 {
			t.Fatalf("month %s has %d openings, want 1", count.Month, count.Count)
		}
	}
}

func TestMonthlyCountsKeepsTheMostRecentMonths(t *testing.T) {
	repo, db := newRepo(t)
	ctx := context.Background()

	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		opening := model.Opening{Role: "Vaga", Company: "Acme", Location: "Remoto", Remote: true, Link: "https://acme.com", Salary: 1000}
		if err := repo.Create(ctx, &opening); err != nil {
			t.Fatalf("Create: %v", err)
		}
		when := now.AddDate(0, -i, 0)
		if err := db.Model(&model.Opening{}).Where("id = ?", opening.ID).UpdateColumn("created_at", when).Error; err != nil {
			t.Fatalf("backdating: %v", err)
		}
	}

	counts, err := repo.MonthlyCounts(ctx, 3)
	if err != nil {
		t.Fatalf("MonthlyCounts: %v", err)
	}
	if len(counts) != 3 {
		t.Fatalf("got %d months, want 3", len(counts))
	}
	if counts[2].Month != now.Format("2006-01") {
		t.Fatalf("newest month = %q, want %q", counts[2].Month, now.Format("2006-01"))
	}
}

func TestMaxSalary(t *testing.T) {
	repo, _ := newRepo(t)

	empty, err := repo.MaxSalary(context.Background())
	if err != nil {
		t.Fatalf("MaxSalary on empty table: %v", err)
	}
	if empty != 0 {
		t.Fatalf("MaxSalary on empty table = %d, want 0", empty)
	}

	seed(t, repo, sample()...)
	max, err := repo.MaxSalary(context.Background())
	if err != nil {
		t.Fatalf("MaxSalary: %v", err)
	}
	if max != 21500 {
		t.Fatalf("MaxSalary = %d, want 21500", max)
	}
}

func TestUpdatePersistsChanges(t *testing.T) {
	repo, _ := newRepo(t)
	ctx := context.Background()
	stored := seed(t, repo, sample()...)

	stored[0].Salary = 18000
	stored[0].Remote = false
	if err := repo.Update(ctx, &stored[0]); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, err := repo.FindByID(ctx, stored[0].ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if found.Salary != 18000 || found.Remote {
		t.Fatalf("after update: salary %d remote %t, want 18000 false", found.Salary, found.Remote)
	}
}

func TestContextCancellationStopsTheQuery(t *testing.T) {
	repo, _ := newRepo(t)
	seed(t, repo, sample()...)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, _, err := repo.List(ctx, repository.Filter{}); err == nil {
		t.Fatal("a cancelled context should abort the query")
	}
}
