package ingestion

import (
	"context"
	"fmt"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

// SlugRemotive is stored in Opening.Source for everything this adapter brings
// in. It is part of the record's identity and must not change.
const SlugRemotive = "remotive"

// remotiveEndpoint is the board's public JSON feed. It accepts a limit, so this
// adapter asks for what it will keep instead of downloading the whole board.
const remotiveEndpoint = "https://remotive.com/api/remote-jobs"

// Remotive reads remotive.com.
type Remotive struct {
	client   *Client
	endpoint string
	limit    int
}

// NewRemotive builds the adapter. A limit of zero asks for the whole feed.
func NewRemotive(client *Client, limit int) *Remotive {
	return &Remotive{client: client, endpoint: remotiveEndpoint, limit: limit}
}

func (s *Remotive) Slug() string { return SlugRemotive }
func (s *Remotive) Name() string { return "Remotive" }

// remotiveResponse wraps the feed's array.
//
// The feed publishes `salary` as free text ("$50k - $70k", "competitive"), with
// no currency or period to key on. It is not parsed: a number invented from
// that string would sit next to real ones in the salary statistics and be
// indistinguishable from them.
type remotiveResponse struct {
	Jobs []remotiveEntry `json:"jobs"`
}

type remotiveEntry struct {
	ID          flexibleID `json:"id"`
	Title       string     `json:"title"`
	CompanyName string     `json:"company_name"`
	Location    string     `json:"candidate_required_location"`
	URL         string     `json:"url"`
}

// Fetch reads the feed and maps it onto openings.
func (s *Remotive) Fetch(ctx context.Context) ([]model.Opening, error) {
	url := s.endpoint
	if s.limit > 0 {
		url = fmt.Sprintf("%s?limit=%d", s.endpoint, s.limit)
	}

	response := remotiveResponse{}
	if err := s.client.GetJSON(ctx, url, &response); err != nil {
		return nil, err
	}

	openings := make([]model.Opening, 0, len(response.Jobs))
	for _, entry := range response.Jobs {
		role := cleanText(entry.Title)
		company := cleanText(entry.CompanyName)
		link := cleanLink(entry.URL)
		identity := entry.ID.String()

		if !isUsable(role, company, link, identity) {
			continue
		}

		openings = append(openings, model.Opening{
			Role:       role,
			Company:    company,
			Location:   cleanLocation(entry.Location),
			Remote:     true,
			Link:       link,
			Salary:     0,
			Source:     SlugRemotive,
			ExternalID: identity,
		})
	}

	return openings, nil
}
