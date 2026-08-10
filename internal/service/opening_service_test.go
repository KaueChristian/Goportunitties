package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/repository"
	"github.com/KaueChristian/Goportunitties/internal/service"
	"gorm.io/gorm"
)

// fakeRepo is an in-memory stand-in for the database. The service layer exists
// precisely so its rules can be exercised without one.
type fakeRepo struct {
	openings map[uint]*model.Opening
	nextID   uint

	lastFilter   repository.Filter
	total        int64
	remote       int64
	onsite       int64
	locations    []repository.LocationCount
	maxSalary    int64
	medianSalary int64
	suggestions  []repository.Suggestion
	lastTerm     string
	lastLimit    int
	aggregates   repository.Aggregates
	monthly      []repository.MonthCount
	months       int

	err error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{openings: map[uint]*model.Opening{}, nextID: 1}
}

func (f *fakeRepo) Create(_ context.Context, opening *model.Opening) error {
	if f.err != nil {
		return f.err
	}
	opening.ID = f.nextID
	f.nextID++
	stored := *opening
	f.openings[opening.ID] = &stored
	return nil
}

func (f *fakeRepo) FindByID(_ context.Context, id uint) (*model.Opening, error) {
	if f.err != nil {
		return nil, f.err
	}
	opening, ok := f.openings[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	found := *opening
	return &found, nil
}

func (f *fakeRepo) List(_ context.Context, filter repository.Filter) ([]model.Opening, int64, error) {
	if f.err != nil {
		return nil, 0, f.err
	}
	f.lastFilter = filter

	openings := make([]model.Opening, 0, len(f.openings))
	for _, opening := range f.openings {
		openings = append(openings, *opening)
	}
	return openings, f.total, nil
}

func (f *fakeRepo) Suggest(_ context.Context, term string, limit int) ([]repository.Suggestion, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.lastTerm, f.lastLimit = term, limit
	return f.suggestions, nil
}

func (f *fakeRepo) CountByRemote(_ context.Context, filter repository.Filter) (int64, int64, error) {
	if f.err != nil {
		return 0, 0, f.err
	}
	f.lastFilter = filter
	return f.remote, f.onsite, nil
}

func (f *fakeRepo) CountByLocation(_ context.Context, filter repository.Filter) ([]repository.LocationCount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.lastFilter = filter
	return f.locations, nil
}

func (f *fakeRepo) MaxSalary(context.Context) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.maxSalary, nil
}

func (f *fakeRepo) MedianSalary(context.Context) (int64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.medianSalary, nil
}

func (f *fakeRepo) Aggregates(context.Context) (repository.Aggregates, error) {
	if f.err != nil {
		return repository.Aggregates{}, f.err
	}
	return f.aggregates, nil
}

func (f *fakeRepo) MonthlyCounts(_ context.Context, months int) ([]repository.MonthCount, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.months = months
	return f.monthly, nil
}

func (f *fakeRepo) Update(_ context.Context, opening *model.Opening) error {
	if f.err != nil {
		return f.err
	}
	stored := *opening
	f.openings[opening.ID] = &stored
	return nil
}

func (f *fakeRepo) Delete(_ context.Context, opening *model.Opening) error {
	if f.err != nil {
		return f.err
	}
	delete(f.openings, opening.ID)
	return nil
}

func boolPtr(v bool) *bool    { return &v }
func strPtr(v string) *string { return &v }
func int64Ptr(v int64) *int64 { return &v }
func ctx() context.Context    { return context.Background() }
func errBoom() error          { return errors.New("boom") }
func fullRequest() dto.OpeningRequest {
	return dto.OpeningRequest{
		Role:     "Desenvolvedor Go",
		Company:  "Acme",
		Location: "São Paulo, SP",
		Remote:   boolPtr(true),
		Link:     "https://acme.com/vagas/1",
		Salary:   15000,
	}
}

func TestCreateTrimsTextAndStores(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	req := fullRequest()
	req.Role = "  Desenvolvedor Go  "
	req.Company = " Acme "

	opening, err := svc.Create(ctx(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if opening.Role != "Desenvolvedor Go" || opening.Company != "Acme" {
		t.Fatalf("stored %q at %q, want the trimmed values", opening.Role, opening.Company)
	}
	if opening.ID == 0 {
		t.Fatal("Create should return the persisted opening, with its id")
	}
}

func TestCreatePropagatesRepositoryErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errBoom()

	if _, err := service.NewOpeningService(repo).Create(ctx(), fullRequest()); err == nil {
		t.Fatal("a repository failure must reach the caller")
	}
}

func TestFindByIDTranslatesTheGormError(t *testing.T) {
	svc := service.NewOpeningService(newFakeRepo())

	_, err := svc.FindByID(ctx(), 99)
	if !errors.Is(err, service.ErrOpeningNotFound) {
		t.Fatalf("got %v, want service.ErrOpeningNotFound — GORM's error must not leak upward", err)
	}
}

func TestReplaceOverwritesEveryField(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	created, err := svc.Create(ctx(), fullRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	replacement := dto.OpeningRequest{
		Role:     "SRE",
		Company:  "Initech",
		Location: "Curitiba, PR",
		Remote:   boolPtr(false),
		Link:     "https://initech.com/vagas/9",
		Salary:   0,
	}

	updated, err := svc.Replace(ctx(), created.ID, replacement)
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if updated.Role != "SRE" || updated.Company != "Initech" || updated.Remote {
		t.Fatalf("Replace left stale values: %+v", updated)
	}
	// The point of PUT: a zero salary is a value, not an omission.
	if updated.Salary != 0 {
		t.Fatalf("salary = %d, want 0", updated.Salary)
	}
	if updated.ID != created.ID {
		t.Fatalf("Replace changed the id: %d -> %d", created.ID, updated.ID)
	}
}

func TestReplaceOnMissingOpening(t *testing.T) {
	svc := service.NewOpeningService(newFakeRepo())

	if _, err := svc.Replace(ctx(), 99, fullRequest()); !errors.Is(err, service.ErrOpeningNotFound) {
		t.Fatalf("got %v, want service.ErrOpeningNotFound", err)
	}
}

func TestPatchOnlyTouchesTheFieldsThatWereSent(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	created, err := svc.Create(ctx(), fullRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Patch(ctx(), created.ID, dto.PatchOpeningRequest{Company: strPtr("Globex")})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}

	if updated.Company != "Globex" {
		t.Fatalf("company = %q, want Globex", updated.Company)
	}
	if updated.Role != created.Role || updated.Salary != created.Salary || updated.Remote != created.Remote {
		t.Fatalf("Patch changed untouched fields: %+v", updated)
	}
}

// The bug the old pointer-less DTO could not express: "set this to zero/false".
func TestPatchCanSetZeroValues(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	created, err := svc.Create(ctx(), fullRequest())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Patch(ctx(), created.ID, dto.PatchOpeningRequest{
		Salary: int64Ptr(0),
		Remote: boolPtr(false),
	})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if updated.Salary != 0 {
		t.Fatalf("salary = %d, want 0 — a sent zero must be applied", updated.Salary)
	}
	if updated.Remote {
		t.Fatal("remote = true, want false — a sent false must be applied")
	}
}

func TestPatchTrimsText(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	created, _ := svc.Create(ctx(), fullRequest())

	updated, err := svc.Patch(ctx(), created.ID, dto.PatchOpeningRequest{Role: strPtr("  SRE  ")})
	if err != nil {
		t.Fatalf("Patch: %v", err)
	}
	if updated.Role != "SRE" {
		t.Fatalf("role = %q, want %q", updated.Role, "SRE")
	}
}

func TestPatchRequestIsEmpty(t *testing.T) {
	if !(dto.PatchOpeningRequest{}).IsEmpty() {
		t.Fatal("an all-nil patch should report itself as empty")
	}
	if (dto.PatchOpeningRequest{Remote: boolPtr(false)}).IsEmpty() {
		t.Fatal("a patch carrying remote=false is not empty")
	}
	if (dto.PatchOpeningRequest{Salary: int64Ptr(0)}).IsEmpty() {
		t.Fatal("a patch carrying salary=0 is not empty")
	}
}

func TestDeleteReturnsTheRemovedOpening(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	created, _ := svc.Create(ctx(), fullRequest())

	deleted, err := svc.Delete(ctx(), created.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if deleted.ID != created.ID {
		t.Fatalf("Delete returned opening %d, want %d", deleted.ID, created.ID)
	}
	if _, err := svc.FindByID(ctx(), created.ID); !errors.Is(err, service.ErrOpeningNotFound) {
		t.Fatalf("opening still readable after delete: %v", err)
	}
}

func TestDeleteOnMissingOpening(t *testing.T) {
	svc := service.NewOpeningService(newFakeRepo())

	if _, err := svc.Delete(ctx(), 99); !errors.Is(err, service.ErrOpeningNotFound) {
		t.Fatalf("got %v, want service.ErrOpeningNotFound", err)
	}
}

func TestListBuildsThePageWindowAndPagination(t *testing.T) {
	repo := newFakeRepo()
	repo.total = 57
	svc := service.NewOpeningService(repo)

	query := dto.ListOpeningsQuery{Page: 3, PageSize: 12, Sort: dto.SortSalaryDesc, Search: "  go  "}

	_, pagination, err := svc.List(ctx(), query)
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if repo.lastFilter.Offset != 24 || repo.lastFilter.Limit != 12 {
		t.Fatalf("filter window = offset %d limit %d, want 24/12", repo.lastFilter.Offset, repo.lastFilter.Limit)
	}
	if repo.lastFilter.Search != "go" {
		t.Fatalf("search = %q, want the trimmed term", repo.lastFilter.Search)
	}
	if pagination.Total != 57 || pagination.TotalPages != 5 || pagination.Page != 3 {
		t.Fatalf("pagination = %+v, want page 3 of 5 over 57 items", pagination)
	}
}

// "all" is how the UI spells "no location filter"; it must not reach SQL.
func TestListTreatsAllAsNoLocationFilter(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	query := dto.ListOpeningsQuery{Location: "all"}
	query.Normalize(12, 100)

	if _, _, err := svc.List(ctx(), query); err != nil {
		t.Fatalf("List: %v", err)
	}
	if repo.lastFilter.Location != "" {
		t.Fatalf("location = %q, want it dropped", repo.lastFilter.Location)
	}
}

func TestFacetsCombineCountsAndRoundTheCeiling(t *testing.T) {
	repo := newFakeRepo()
	repo.remote, repo.onsite = 6, 4
	repo.maxSalary = 21500
	repo.locations = []repository.LocationCount{{Location: "Curitiba, PR", Count: 4}}
	svc := service.NewOpeningService(repo)

	facets, err := svc.Facets(ctx(), dto.ListOpeningsQuery{})
	if err != nil {
		t.Fatalf("Facets: %v", err)
	}

	if facets.Remote.All != 10 || facets.Remote.Remote != 6 || facets.Remote.Onsite != 4 {
		t.Fatalf("remote counts = %+v, want 10/6/4", facets.Remote)
	}
	if facets.SalaryCeiling != 22000 {
		t.Fatalf("salary ceiling = %d, want 22000 — the slider max should be a round figure", facets.SalaryCeiling)
	}
	if len(facets.Locations) != 1 || facets.Locations[0].Value != "Curitiba, PR" {
		t.Fatalf("locations = %+v", facets.Locations)
	}
}

func TestFacetsCeilingHasAFloorOnAnEmptyIndex(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	facets, err := svc.Facets(ctx(), dto.ListOpeningsQuery{})
	if err != nil {
		t.Fatalf("Facets: %v", err)
	}
	if facets.SalaryCeiling != 1000 {
		t.Fatalf("salary ceiling = %d, want 1000 so the slider still has a range", facets.SalaryCeiling)
	}
}

func TestStatsDerivesOnsiteAndKeepsTheSeries(t *testing.T) {
	repo := newFakeRepo()
	repo.aggregates = repository.Aggregates{Total: 10, Remote: 6, Companies: 7, AverageSalary: 12000, MaxSalary: 30000}
	repo.medianSalary = 11000
	repo.monthly = []repository.MonthCount{{Month: "2026-07", Count: 4}, {Month: "2026-08", Count: 6}}
	svc := service.NewOpeningService(repo)

	stats, err := svc.Stats(ctx())
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	if stats.Onsite != 4 {
		t.Fatalf("onsite = %d, want 4 (total - remote)", stats.Onsite)
	}
	if stats.Companies != 7 || stats.AverageSalary != 12000 || stats.MaxSalary != 30000 {
		t.Fatalf("stats = %+v", stats)
	}
	// The median comes from its own query, not from the aggregate row.
	if stats.MedianSalary != 11000 {
		t.Fatalf("median = %d, want 11000", stats.MedianSalary)
	}
	if len(stats.Monthly) != 2 || stats.Monthly[0].Month != "2026-07" {
		t.Fatalf("monthly series = %+v", stats.Monthly)
	}
	if repo.months != 9 {
		t.Fatalf("asked for %d months, want 9", repo.months)
	}
}

func TestSuggestIgnoresTermsTooShortToBeUseful(t *testing.T) {
	repo := newFakeRepo()
	repo.suggestions = []repository.Suggestion{{Value: "SRE", Kind: "role", Count: 1}}
	svc := service.NewOpeningService(repo)

	for _, term := range []string{"", " ", "g", "  g  "} {
		got, err := svc.Suggest(ctx(), dto.SuggestOpeningsQuery{Search: term, Limit: 8})
		if err != nil {
			t.Fatalf("Suggest(%q): %v", term, err)
		}
		if len(got) != 0 {
			t.Fatalf("Suggest(%q) returned %+v, want nothing — the database is never touched", term, got)
		}
		if repo.lastTerm != "" {
			t.Fatalf("Suggest(%q) reached the repository with %q", term, repo.lastTerm)
		}
	}
}

func TestSuggestTrimsTheTermAndMapsTheResult(t *testing.T) {
	repo := newFakeRepo()
	repo.suggestions = []repository.Suggestion{
		{Value: "Desenvolvedor Go", Kind: "role", Count: 3},
		{Value: "Globex", Kind: "company", Count: 1},
	}
	svc := service.NewOpeningService(repo)

	got, err := svc.Suggest(ctx(), dto.SuggestOpeningsQuery{Search: "  go  ", Limit: 5})
	if err != nil {
		t.Fatalf("Suggest: %v", err)
	}

	if repo.lastTerm != "go" || repo.lastLimit != 5 {
		t.Fatalf("repository received %q with limit %d", repo.lastTerm, repo.lastLimit)
	}
	if len(got) != 2 || got[0].Value != "Desenvolvedor Go" || got[0].Kind != "role" || got[0].Count != 3 {
		t.Fatalf("suggestions = %+v", got)
	}
}

// An accented two-letter term is two runes, not four bytes.
func TestSuggestCountsRunesNotBytes(t *testing.T) {
	repo := newFakeRepo()
	svc := service.NewOpeningService(repo)

	if _, err := svc.Suggest(ctx(), dto.SuggestOpeningsQuery{Search: "çã", Limit: 8}); err != nil {
		t.Fatalf("Suggest: %v", err)
	}
	if repo.lastTerm != "çã" {
		t.Fatalf("term %q was rejected as too short", "çã")
	}
}

func TestSuggestNormalizeClampsTheLimit(t *testing.T) {
	tests := []struct {
		given int
		want  int
	}{
		{0, dto.DefaultSuggestionLimit},
		{-3, dto.DefaultSuggestionLimit},
		{5, 5},
		{999, dto.MaxSuggestionLimit},
	}

	for _, tt := range tests {
		query := dto.SuggestOpeningsQuery{Limit: tt.given}
		query.Normalize()

		if query.Limit != tt.want {
			t.Fatalf("limit %d normalized to %d, want %d", tt.given, query.Limit, tt.want)
		}
	}
}

func TestStatsPropagatesRepositoryErrors(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errBoom()

	if _, err := service.NewOpeningService(repo).Stats(ctx()); err == nil {
		t.Fatal("a repository failure must reach the caller")
	}
}

func TestNormalizeFillsDefaultsAndClamps(t *testing.T) {
	tests := []struct {
		name         string
		query        dto.ListOpeningsQuery
		wantPage     int
		wantPageSize int
		wantSort     string
		wantOffset   int
	}{
		{"empty query", dto.ListOpeningsQuery{}, 1, 12, dto.SortRecent, 0},
		{"page size above the cap", dto.ListOpeningsQuery{PageSize: 5000}, 1, 100, dto.SortRecent, 0},
		{"negative page", dto.ListOpeningsQuery{Page: -3}, 1, 12, dto.SortRecent, 0},
		{"explicit window", dto.ListOpeningsQuery{Page: 4, PageSize: 20, Sort: dto.SortRole}, 4, 20, dto.SortRole, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := tt.query
			query.Normalize(12, 100)

			if query.Page != tt.wantPage || query.PageSize != tt.wantPageSize || query.Sort != tt.wantSort {
				t.Fatalf("normalized to %+v", query)
			}
			if query.Offset() != tt.wantOffset {
				t.Fatalf("offset = %d, want %d", query.Offset(), tt.wantOffset)
			}
		})
	}
}

func TestNewPagination(t *testing.T) {
	tests := []struct {
		total     int64
		pageSize  int
		wantPages int
	}{
		{0, 12, 0},
		{1, 12, 1},
		{12, 12, 1},
		{13, 12, 2},
		{57, 12, 5},
	}

	for _, tt := range tests {
		got := dto.NewPagination(1, tt.pageSize, tt.total)
		if got.TotalPages != tt.wantPages {
			t.Fatalf("%d items in pages of %d = %d pages, want %d", tt.total, tt.pageSize, got.TotalPages, tt.wantPages)
		}
	}
}
