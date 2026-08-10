// Package repository is the only layer that talks to the database.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/KaueChristian/Goportunitties/internal/model"
	"gorm.io/gorm"
)

// Filter is the set of conditions the listing endpoint can express. The zero
// value means "everything, unsorted, unpaged".
type Filter struct {
	Search    string
	Location  string
	Remote    *bool
	MinSalary int64
	Sort      string
	Offset    int
	Limit     int
}

// facet names a criterion that can be excluded when applying a Filter, so a
// facet count can be taken against every condition but its own.
type facet string

const (
	facetNone     facet = ""
	facetRemote   facet = "remote"
	facetLocation facet = "location"
)

// LocationCount is how many openings share one location.
type LocationCount struct {
	Location string
	Count    int64
}

// MonthCount is how many openings were published in one YYYY-MM month.
type MonthCount struct {
	Month string
	Count int64
}

// Suggestion is one autocomplete entry: a role or a company that matches what
// the user has typed, with how many openings carry it.
type Suggestion struct {
	Value string
	Kind  string // "role" or "company"
	Count int64
}

// Aggregates is the whole-index summary behind the dashboard.
//
// The salary figures cover only the openings that state one: a posting saved
// with salary 0 means "a combinar", and counting it as zero would drag the
// numbers down without any real salary having changed.
type Aggregates struct {
	Total         int64
	Remote        int64
	Companies     int64
	AverageSalary int64
	MaxSalary     int64
}

// OpeningRepository defines data access for openings.
type OpeningRepository interface {
	Create(ctx context.Context, opening *model.Opening) error
	FindByID(ctx context.Context, id uint) (*model.Opening, error)
	List(ctx context.Context, filter Filter) ([]model.Opening, int64, error)
	Suggest(ctx context.Context, term string, limit int) ([]Suggestion, error)
	CountByRemote(ctx context.Context, filter Filter) (remote, onsite int64, err error)
	CountByLocation(ctx context.Context, filter Filter) ([]LocationCount, error)
	MaxSalary(ctx context.Context) (int64, error)
	MedianSalary(ctx context.Context) (int64, error)
	Aggregates(ctx context.Context) (Aggregates, error)
	MonthlyCounts(ctx context.Context, months int) ([]MonthCount, error)
	Update(ctx context.Context, opening *model.Opening) error
	Delete(ctx context.Context, opening *model.Opening) error
}

type gormOpeningRepository struct {
	db *gorm.DB
}

// NewOpeningRepository creates a GORM-backed OpeningRepository.
func NewOpeningRepository(db *gorm.DB) OpeningRepository {
	return &gormOpeningRepository{db: db}
}

func (r *gormOpeningRepository) Create(ctx context.Context, opening *model.Opening) error {
	return r.db.WithContext(ctx).Create(opening).Error
}

func (r *gormOpeningRepository) FindByID(ctx context.Context, id uint) (*model.Opening, error) {
	opening := &model.Opening{}
	if err := r.db.WithContext(ctx).First(opening, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return opening, nil
}

func (r *gormOpeningRepository) List(ctx context.Context, filter Filter) ([]model.Opening, int64, error) {
	var total int64
	if err := r.query(ctx, filter, facetNone).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("counting openings: %w", err)
	}

	// A page past the end costs nothing to answer once the total is known.
	if total == 0 || int64(filter.Offset) >= total {
		return []model.Opening{}, total, nil
	}

	query := r.query(ctx, filter, facetNone).Order(orderClause(filter.Sort))
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit).Offset(filter.Offset)
	}

	openings := []model.Opening{}
	if err := query.Find(&openings).Error; err != nil {
		return nil, 0, fmt.Errorf("listing openings: %w", err)
	}
	return openings, total, nil
}

// Suggest returns the roles and companies matching term, most frequent first.
//
// Both columns are searched in one statement so the limit applies to the merged
// list — querying them separately would give roles and companies a fixed share
// of the slots regardless of how well either actually matches. The soft-delete
// condition is spelled out because raw SQL bypasses the one GORM adds.
func (r *gormOpeningRepository) Suggest(ctx context.Context, term string, limit int) ([]Suggestion, error) {
	term = strings.TrimSpace(term)
	if term == "" || limit <= 0 {
		return []Suggestion{}, nil
	}

	pattern := "%" + strings.ToLower(term) + "%"
	suggestions := []Suggestion{}

	err := r.db.WithContext(ctx).Raw(`
		SELECT value, kind, COUNT(*) AS count
		FROM (
			SELECT role AS value, 'role' AS kind FROM openings
			WHERE deleted_at IS NULL AND LOWER(role) LIKE @pattern
			UNION ALL
			SELECT company AS value, 'company' AS kind FROM openings
			WHERE deleted_at IS NULL AND LOWER(company) LIKE @pattern
		)
		GROUP BY value, kind
		ORDER BY count DESC, value ASC
		LIMIT @limit`,
		sql.Named("pattern", pattern),
		sql.Named("limit", limit),
	).Scan(&suggestions).Error
	if err != nil {
		return nil, fmt.Errorf("reading suggestions: %w", err)
	}

	return suggestions, nil
}

func (r *gormOpeningRepository) CountByRemote(ctx context.Context, filter Filter) (int64, int64, error) {
	rows := []struct {
		Remote bool
		Count  int64
	}{}

	err := r.query(ctx, filter, facetRemote).
		Select("remote, COUNT(*) AS count").
		Group("remote").
		Scan(&rows).Error
	if err != nil {
		return 0, 0, fmt.Errorf("counting openings by modality: %w", err)
	}

	var remote, onsite int64
	for _, row := range rows {
		if row.Remote {
			remote = row.Count
			continue
		}
		onsite = row.Count
	}
	return remote, onsite, nil
}

func (r *gormOpeningRepository) CountByLocation(ctx context.Context, filter Filter) ([]LocationCount, error) {
	counts := []LocationCount{}

	err := r.query(ctx, filter, facetLocation).
		Select("location, COUNT(*) AS count").
		Group("location").
		Order("location ASC").
		Scan(&counts).Error
	if err != nil {
		return nil, fmt.Errorf("counting openings by location: %w", err)
	}
	return counts, nil
}

func (r *gormOpeningRepository) MaxSalary(ctx context.Context) (int64, error) {
	var max int64
	err := r.db.WithContext(ctx).
		Model(&model.Opening{}).
		Select("COALESCE(MAX(salary), 0)").
		Scan(&max).Error
	if err != nil {
		return 0, fmt.Errorf("reading highest salary: %w", err)
	}
	return max, nil
}

// MedianSalary returns the middle salary among the openings that state one.
//
// SQLite has no median aggregate, so the value is read positionally: count the
// rows, then fetch the one (odd count) or two (even count) in the middle. That
// keeps the work in SQL instead of pulling every salary into memory — which
// matters once the index is filled by ingestion rather than by hand.
func (r *gormOpeningRepository) MedianSalary(ctx context.Context) (int64, error) {
	var stated int64
	err := r.db.WithContext(ctx).
		Model(&model.Opening{}).
		Where("salary > 0").
		Count(&stated).Error
	if err != nil {
		return 0, fmt.Errorf("counting stated salaries: %w", err)
	}
	if stated == 0 {
		return 0, nil
	}

	middle := []int64{}
	err = r.db.WithContext(ctx).
		Model(&model.Opening{}).
		Where("salary > 0").
		Order("salary ASC").
		Limit(int(2-stated%2)).
		Offset(int((stated-1)/2)).
		Pluck("salary", &middle).Error
	if err != nil {
		return 0, fmt.Errorf("reading median salary: %w", err)
	}
	if len(middle) == 0 {
		return 0, nil
	}

	// An even count has no single middle row: the median is the mean of the two.
	var sum int64
	for _, salary := range middle {
		sum += salary
	}
	return sum / int64(len(middle)), nil
}

func (r *gormOpeningRepository) Aggregates(ctx context.Context) (Aggregates, error) {
	row := struct {
		Total         int64
		Remote        int64
		Companies     int64
		AverageSalary int64
		MaxSalary     int64
	}{}

	err := r.db.WithContext(ctx).
		Model(&model.Opening{}).
		// CASE returns NULL for an unstated salary, and AVG skips NULLs — that
		// is what keeps "a combinar" out of the average.
		Select(`COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN remote = 1 THEN 1 ELSE 0 END), 0) AS remote,
			COUNT(DISTINCT company) AS companies,
			CAST(COALESCE(AVG(CASE WHEN salary > 0 THEN salary END), 0) AS INTEGER) AS average_salary,
			COALESCE(MAX(salary), 0) AS max_salary`).
		Scan(&row).Error
	if err != nil {
		return Aggregates{}, fmt.Errorf("reading aggregates: %w", err)
	}

	return Aggregates(row), nil
}

func (r *gormOpeningRepository) MonthlyCounts(ctx context.Context, months int) ([]MonthCount, error) {
	counts := []MonthCount{}

	err := r.db.WithContext(ctx).
		Model(&model.Opening{}).
		Select("strftime('%Y-%m', created_at) AS month, COUNT(*) AS count").
		Group("month").
		Order("month DESC").
		Limit(months).
		Scan(&counts).Error
	if err != nil {
		return nil, fmt.Errorf("counting openings by month: %w", err)
	}

	// The query reads newest-first so LIMIT keeps the most recent months; the
	// chart draws oldest-first.
	for i, j := 0, len(counts)-1; i < j; i, j = i+1, j-1 {
		counts[i], counts[j] = counts[j], counts[i]
	}
	return counts, nil
}

func (r *gormOpeningRepository) Update(ctx context.Context, opening *model.Opening) error {
	return r.db.WithContext(ctx).Save(opening).Error
}

func (r *gormOpeningRepository) Delete(ctx context.Context, opening *model.Opening) error {
	return r.db.WithContext(ctx).Delete(opening).Error
}

// query builds the base statement for a Filter, optionally leaving one facet
// out of the conditions.
func (r *gormOpeningRepository) query(ctx context.Context, filter Filter, except facet) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.Opening{})

	if term := strings.TrimSpace(filter.Search); term != "" {
		pattern := "%" + strings.ToLower(term) + "%"
		query = query.Where("LOWER(role) LIKE ? OR LOWER(company) LIKE ?", pattern, pattern)
	}
	if except != facetLocation && filter.Location != "" {
		query = query.Where("location = ?", filter.Location)
	}
	if except != facetRemote && filter.Remote != nil {
		query = query.Where("remote = ?", *filter.Remote)
	}
	if filter.MinSalary > 0 {
		query = query.Where("salary >= ?", filter.MinSalary)
	}

	return query
}

// orderClause maps a sort option to SQL. Unknown values fall back to newest
// first; the caller validates the option, this is only a safety net against
// interpolating anything unexpected into the statement.
func orderClause(sort string) string {
	switch sort {
	case "salary-desc":
		return "salary DESC, created_at DESC"
	case "salary-asc":
		return "salary ASC, created_at DESC"
	case "role":
		return "role ASC, created_at DESC"
	case "updated":
		return "updated_at DESC, id DESC"
	default:
		return "created_at DESC, id DESC"
	}
}
