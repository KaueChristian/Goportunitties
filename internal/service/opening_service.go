// Package service holds the business rules around openings. It knows nothing
// about Gin or GORM: it takes DTOs, returns models, and speaks its own errors.
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/KaueChristian/Goportunitties/internal/dto"
	"github.com/KaueChristian/Goportunitties/internal/model"
	"github.com/KaueChristian/Goportunitties/internal/repository"
	"gorm.io/gorm"
)

// ErrOpeningNotFound is returned when no opening matches the requested id. The
// handler turns it into a 404; nothing else in the stack needs to know about
// GORM's own not-found error.
var ErrOpeningNotFound = errors.New("opening not found")

// monthsOnChart is how far back the dashboard's monthly series goes.
const monthsOnChart = 9

// OpeningService is the business surface the handlers depend on.
type OpeningService interface {
	Create(ctx context.Context, req dto.OpeningRequest) (*model.Opening, error)
	FindByID(ctx context.Context, id uint) (*model.Opening, error)
	List(ctx context.Context, query dto.ListOpeningsQuery) ([]model.Opening, dto.Pagination, error)
	Facets(ctx context.Context, query dto.ListOpeningsQuery) (dto.OpeningFacets, error)
	Suggest(ctx context.Context, query dto.SuggestOpeningsQuery) ([]dto.Suggestion, error)
	Stats(ctx context.Context) (dto.OpeningStats, error)
	Replace(ctx context.Context, id uint, req dto.OpeningRequest) (*model.Opening, error)
	Patch(ctx context.Context, id uint, req dto.PatchOpeningRequest) (*model.Opening, error)
	Delete(ctx context.Context, id uint) (*model.Opening, error)
}

type openingService struct {
	repo repository.OpeningRepository
}

// NewOpeningService creates an OpeningService backed by repo.
func NewOpeningService(repo repository.OpeningRepository) OpeningService {
	return &openingService{repo: repo}
}

func (s *openingService) Create(ctx context.Context, req dto.OpeningRequest) (*model.Opening, error) {
	// Provenance is settled by the model on the way in, and never travels in
	// the request: there is no field for a client to claim an opening came from
	// somewhere it did not.
	opening := &model.Opening{}
	applyRequest(opening, req)

	if err := s.repo.Create(ctx, opening); err != nil {
		return nil, err
	}
	return opening, nil
}

func (s *openingService) FindByID(ctx context.Context, id uint) (*model.Opening, error) {
	opening, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOpeningNotFound
		}
		return nil, err
	}
	return opening, nil
}

func (s *openingService) List(ctx context.Context, query dto.ListOpeningsQuery) ([]model.Opening, dto.Pagination, error) {
	filter := toFilter(query)
	filter.Offset = query.Offset()
	filter.Limit = query.PageSize

	openings, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, dto.Pagination{}, err
	}

	return openings, dto.NewPagination(query.Page, query.PageSize, total), nil
}

func (s *openingService) Facets(ctx context.Context, query dto.ListOpeningsQuery) (dto.OpeningFacets, error) {
	filter := toFilter(query)

	remote, onsite, err := s.repo.CountByRemote(ctx, filter)
	if err != nil {
		return dto.OpeningFacets{}, err
	}

	locations, err := s.repo.CountByLocation(ctx, filter)
	if err != nil {
		return dto.OpeningFacets{}, err
	}

	sources, err := s.repo.CountBySource(ctx, filter)
	if err != nil {
		return dto.OpeningFacets{}, err
	}

	ceiling, err := s.repo.MaxSalary(ctx)
	if err != nil {
		return dto.OpeningFacets{}, err
	}

	entries := make([]dto.LocationCount, 0, len(locations))
	for _, location := range locations {
		entries = append(entries, dto.LocationCount{Value: location.Location, Count: location.Count})
	}

	origins := make([]dto.SourceCount, 0, len(sources))
	for _, source := range sources {
		origins = append(origins, dto.SourceCount{Value: source.Source, Count: source.Count})
	}

	return dto.OpeningFacets{
		Remote: dto.RemoteCounts{
			All:    remote + onsite,
			Remote: remote,
			Onsite: onsite,
		},
		Locations:     entries,
		Sources:       origins,
		SalaryCeiling: roundUpToThousand(ceiling),
	}, nil
}

func (s *openingService) Suggest(ctx context.Context, query dto.SuggestOpeningsQuery) ([]dto.Suggestion, error) {
	term := strings.TrimSpace(query.Search)

	// A term this short matches nearly everything; answering with an empty list
	// costs one comparison instead of a table scan per keystroke.
	if len([]rune(term)) < dto.MinSuggestionTerm {
		return []dto.Suggestion{}, nil
	}

	found, err := s.repo.Suggest(ctx, term, query.Limit)
	if err != nil {
		return nil, err
	}

	suggestions := make([]dto.Suggestion, 0, len(found))
	for _, suggestion := range found {
		suggestions = append(suggestions, dto.Suggestion{
			Value: suggestion.Value,
			Kind:  suggestion.Kind,
			Count: suggestion.Count,
		})
	}
	return suggestions, nil
}

func (s *openingService) Stats(ctx context.Context) (dto.OpeningStats, error) {
	aggregates, err := s.repo.Aggregates(ctx)
	if err != nil {
		return dto.OpeningStats{}, err
	}

	median, err := s.repo.MedianSalary(ctx)
	if err != nil {
		return dto.OpeningStats{}, err
	}

	monthly, err := s.repo.MonthlyCounts(ctx, monthsOnChart)
	if err != nil {
		return dto.OpeningStats{}, err
	}

	months := make([]dto.MonthCount, 0, len(monthly))
	for _, month := range monthly {
		months = append(months, dto.MonthCount{Month: month.Month, Count: month.Count})
	}

	return dto.OpeningStats{
		Total:         aggregates.Total,
		Remote:        aggregates.Remote,
		Onsite:        aggregates.Total - aggregates.Remote,
		Companies:     aggregates.Companies,
		MedianSalary:  median,
		AverageSalary: aggregates.AverageSalary,
		MaxSalary:     aggregates.MaxSalary,
		Monthly:       months,
	}, nil
}

func (s *openingService) Replace(ctx context.Context, id uint, req dto.OpeningRequest) (*model.Opening, error) {
	opening, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	applyRequest(opening, req)

	if err := s.repo.Update(ctx, opening); err != nil {
		return nil, err
	}
	return opening, nil
}

func (s *openingService) Patch(ctx context.Context, id uint, req dto.PatchOpeningRequest) (*model.Opening, error) {
	opening, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Only non-nil fields were sent, so each assignment is an explicit request
	// to change that value — including to an empty string or a zero salary.
	if req.Role != nil {
		opening.Role = strings.TrimSpace(*req.Role)
	}
	if req.Company != nil {
		opening.Company = strings.TrimSpace(*req.Company)
	}
	if req.Location != nil {
		opening.Location = strings.TrimSpace(*req.Location)
	}
	if req.Remote != nil {
		opening.Remote = *req.Remote
	}
	if req.Link != nil {
		opening.Link = strings.TrimSpace(*req.Link)
	}
	if req.Salary != nil {
		opening.Salary = *req.Salary
	}

	if err := s.repo.Update(ctx, opening); err != nil {
		return nil, err
	}
	return opening, nil
}

func (s *openingService) Delete(ctx context.Context, id uint) (*model.Opening, error) {
	opening, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Delete(ctx, opening); err != nil {
		return nil, err
	}
	return opening, nil
}

// applyRequest copies a full representation onto an opening, trimming the text
// so " Go Developer " and "Go Developer" never become two distinct roles.
func applyRequest(opening *model.Opening, req dto.OpeningRequest) {
	opening.Role = strings.TrimSpace(req.Role)
	opening.Company = strings.TrimSpace(req.Company)
	opening.Location = strings.TrimSpace(req.Location)
	opening.Remote = req.Remote != nil && *req.Remote
	opening.Link = strings.TrimSpace(req.Link)
	opening.Salary = req.Salary
}

func toFilter(query dto.ListOpeningsQuery) repository.Filter {
	// "all" is how the interface spells "no filter"; it never reaches SQL.
	location := query.Location
	if location == "all" {
		location = ""
	}
	source := query.Source
	if source == "all" {
		source = ""
	}

	return repository.Filter{
		Search:    strings.TrimSpace(query.Search),
		Location:  location,
		Source:    source,
		Remote:    query.Remote,
		MinSalary: query.MinSalary,
		Sort:      query.Sort,
	}
}

// roundUpToThousand keeps the salary slider's maximum a round number instead of
// whichever odd figure happens to be the highest salary on record.
func roundUpToThousand(value int64) int64 {
	const step = 1000
	if value <= step {
		return step
	}
	return ((value + step - 1) / step) * step
}
