package repository_test

import (
	"context"
	"errors"
	"path/filepath"
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
	return []model.Opening{
		{Role: "Desenvolvedor Go", Company: "Acme", Location: "São Paulo, SP", Remote: true, Link: "https://acme.com/1", Salary: 15000},
		{Role: "Desenvolvedor React", Company: "Globex", Location: "São Paulo, SP", Remote: false, Link: "https://globex.com/2", Salary: 9000},
		{Role: "SRE", Company: "Initech", Location: "Curitiba, PR", Remote: true, Link: "https://initech.com/3", Salary: 21500},
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
