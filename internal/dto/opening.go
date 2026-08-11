package dto

import (
	"time"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

// Field limits are shared with the validation tags below so the frontend and
// the tests can quote the same numbers.
const (
	MaxTextLength = 120
	MaxLinkLength = 500
	MaxSalary     = 10_000_000
)

// OpeningRequest is the complete representation of an opening, accepted by
// POST (create) and PUT (full replacement).
//
// Remote is a pointer so that `false` is distinguishable from "absent": with a
// plain bool, `required` would reject every on-site opening.
type OpeningRequest struct {
	Role     string `json:"role" binding:"required,min=2,max=120"`
	Company  string `json:"company" binding:"required,min=2,max=120"`
	Location string `json:"location" binding:"required,min=2,max=120"`
	Remote   *bool  `json:"remote" binding:"required"`
	Link     string `json:"link" binding:"required,httpurl,max=500"`
	// Salary accepts 0, which the UI shows as "a combinar".
	Salary int64 `json:"salary" binding:"gte=0,lte=10000000"`
}

// PatchOpeningRequest is the partial counterpart, used by PATCH. Every field is
// a pointer: nil means "leave as is", which is what lets a client clear a value
// or set a salary back to zero — impossible when absence and zero look alike.
type PatchOpeningRequest struct {
	Role     *string `json:"role" binding:"omitempty,min=2,max=120"`
	Company  *string `json:"company" binding:"omitempty,min=2,max=120"`
	Location *string `json:"location" binding:"omitempty,min=2,max=120"`
	Remote   *bool   `json:"remote"`
	Link     *string `json:"link" binding:"omitempty,httpurl,max=500"`
	Salary   *int64  `json:"salary" binding:"omitempty,gte=0,lte=10000000"`
}

// IsEmpty reports whether the patch would change nothing.
func (r PatchOpeningRequest) IsEmpty() bool {
	return r.Role == nil && r.Company == nil && r.Location == nil &&
		r.Remote == nil && r.Link == nil && r.Salary == nil
}

// Sort options accepted by the listing endpoint.
const (
	SortRecent     = "recent"
	SortSalaryDesc = "salary-desc"
	SortSalaryAsc  = "salary-asc"
	SortRole       = "role"
	// SortUpdated orders by last edit, which is what an activity feed shows.
	SortUpdated = "updated"
)

// ListOpeningsQuery is the filtering, sorting and paging surface of
// GET /openings. Everything is optional.
//
// Filters are validated — a bad one means the caller asked a question the API
// cannot answer — while paging is clamped by Normalize instead. There is no way
// to tell an omitted `page` from `page=0` on a plain int, so rejecting the
// second would also reject the first.
type ListOpeningsQuery struct {
	Search    string `form:"search" binding:"max=120"`
	Location  string `form:"location" binding:"max=120"`
	Remote    *bool  `form:"remote"`
	MinSalary int64  `form:"minSalary" binding:"gte=0"`
	Sort      string `form:"sort" binding:"omitempty,oneof=recent salary-desc salary-asc role updated"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

// Normalize fills in the defaults the caller omitted and clamps the page size
// so a client cannot ask for the entire table in one request. The window the
// caller actually got comes back in the response's pagination meta.
func (q *ListOpeningsQuery) Normalize(defaultPageSize, maxPageSize int) {
	if q.Sort == "" {
		q.Sort = SortRecent
	}
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 {
		q.PageSize = defaultPageSize
	}
	if q.PageSize > maxPageSize {
		q.PageSize = maxPageSize
	}
}

// Offset is the row offset the normalized page window starts at.
func (q ListOpeningsQuery) Offset() int {
	return (q.Page - 1) * q.PageSize
}

// OpeningResponse is what every opening endpoint returns.
//
// Source is exposed so the interface can credit an opening to where it came
// from; the identity within that source stays internal.
type OpeningResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Source    string    `json:"source"`

	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

// NewOpeningResponse maps a persisted opening to its API representation.
func NewOpeningResponse(opening *model.Opening) OpeningResponse {
	return OpeningResponse{
		ID:        opening.ID,
		CreatedAt: opening.CreatedAt,
		UpdatedAt: opening.UpdatedAt,
		Source:    opening.Source,
		Role:      opening.Role,
		Company:   opening.Company,
		Location:  opening.Location,
		Remote:    opening.Remote,
		Link:      opening.Link,
		Salary:    opening.Salary,
	}
}

// NewOpeningResponseList maps a page of openings. It never returns nil, so the
// client always receives a JSON array.
func NewOpeningResponseList(openings []model.Opening) []OpeningResponse {
	responses := make([]OpeningResponse, 0, len(openings))
	for i := range openings {
		responses = append(responses, NewOpeningResponse(&openings[i]))
	}
	return responses
}

// Suggestion limits for the autocomplete endpoint.
const (
	DefaultSuggestionLimit = 8
	MaxSuggestionLimit     = 20
	// Below two characters almost everything matches, which is noise rather
	// than a suggestion.
	MinSuggestionTerm = 2
)

// SuggestOpeningsQuery is the input of GET /openings/suggestions.
type SuggestOpeningsQuery struct {
	Search string `form:"search" binding:"max=120"`
	Limit  int    `form:"limit"`
}

// Normalize clamps the limit into range.
func (q *SuggestOpeningsQuery) Normalize() {
	if q.Limit < 1 {
		q.Limit = DefaultSuggestionLimit
	}
	if q.Limit > MaxSuggestionLimit {
		q.Limit = MaxSuggestionLimit
	}
}

// Suggestion is one autocomplete entry. Kind lets the interface label the row
// as a role or a company instead of showing a bare string.
type Suggestion struct {
	Value string `json:"value"`
	Kind  string `json:"kind"`
	Count int64  `json:"count"`
}

// LocationCount is one entry of the location facet.
type LocationCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// RemoteCounts holds how many openings each modality option would return.
type RemoteCounts struct {
	All    int64 `json:"all"`
	Remote int64 `json:"remote"`
	Onsite int64 `json:"onsite"`
}

// OpeningFacets powers the filter sidebar: each facet is counted against the
// query with that same facet removed, so choosing an option never leaves its
// siblings advertising a total the click cannot produce.
type OpeningFacets struct {
	Remote        RemoteCounts    `json:"remote"`
	Locations     []LocationCount `json:"locations"`
	SalaryCeiling int64           `json:"salaryCeiling"`
}

// MonthCount is one bar of the dashboard's publications-per-month chart.
type MonthCount struct {
	Month string `json:"month"` // YYYY-MM
	Count int64  `json:"count"`
}

// OpeningStats is the dashboard's aggregate view of the whole index.
//
// The salary figures ignore openings saved with salary 0 ("a combinar").
// MedianSalary is what the interface shows: a single high posting skews the
// mean into a number that describes no real opening.
type OpeningStats struct {
	Total         int64        `json:"total"`
	Remote        int64        `json:"remote"`
	Onsite        int64        `json:"onsite"`
	Companies     int64        `json:"companies"`
	MedianSalary  int64        `json:"medianSalary"`
	AverageSalary int64        `json:"averageSalary"`
	MaxSalary     int64        `json:"maxSalary"`
	Monthly       []MonthCount `json:"monthly"`
}
