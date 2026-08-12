package ingestion

import (
	"context"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

// SlugRemoteOK is stored in Opening.Source for everything this adapter brings
// in. It is part of the record's identity and must not change.
const SlugRemoteOK = "remoteok"

// remoteOKEndpoint is the board's public JSON feed.
const remoteOKEndpoint = "https://remoteok.com/api"

// RemoteOK reads remoteok.com.
type RemoteOK struct {
	client   *Client
	endpoint string
}

// NewRemoteOK builds the adapter.
func NewRemoteOK(client *Client) *RemoteOK {
	return &RemoteOK{client: client, endpoint: remoteOKEndpoint}
}

func (s *RemoteOK) Slug() string { return SlugRemoteOK }
func (s *RemoteOK) Name() string { return "RemoteOK" }

// remoteOKEntry is the subset of the feed this project uses.
//
// The feed also carries salary_min/salary_max in US dollars per year. They are
// not read: this application stores one salary in BRL per month, and converting
// would mean inventing both an exchange rate and a working-hours assumption.
// Ingested openings are stored as "a combinar" instead — which is also why the
// salary statistics exclude zeros.
type remoteOKEntry struct {
	ID       flexibleID `json:"id"`
	Position string     `json:"position"`
	Company  string     `json:"company"`
	Location string     `json:"location"`
	URL      string     `json:"url"`
	// The first element of the feed is a legal notice rather than a job; it is
	// the only entry carrying this field.
	Legal string `json:"legal"`
}

// Fetch reads the feed and maps it onto openings.
func (s *RemoteOK) Fetch(ctx context.Context) ([]model.Opening, error) {
	entries := []remoteOKEntry{}
	if err := s.client.GetJSON(ctx, s.endpoint, &entries); err != nil {
		return nil, err
	}

	openings := make([]model.Opening, 0, len(entries))
	for _, entry := range entries {
		if entry.Legal != "" {
			continue
		}

		role := cleanText(entry.Position)
		company := cleanText(entry.Company)
		link := cleanLink(entry.URL)
		identity := entry.ID.String()

		if !isUsable(role, company, link, identity) {
			continue
		}

		// This board mixes technology with everything else — retail, logistics
		// and, in one snapshot, a cake recipe. The filter is applied here and
		// not to the community boards, whose whole repository is technical: a
		// posting there with an unusual title should not be dropped.
		if !looksTechnical(role) {
			continue
		}

		openings = append(openings, model.Opening{
			Role:     role,
			Company:  company,
			Location: cleanLocation(entry.Location),
			// The whole board is remote; the location field only says where the
			// company will accept someone working from.
			Remote:     true,
			Link:       link,
			Salary:     0,
			Source:     SlugRemoteOK,
			ExternalID: identity,
		})
	}

	return openings, nil
}
