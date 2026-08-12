package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultUserAgent identifies this project to the boards it reads. Several of
// them reject requests without one, and an honest agent is what lets an
// operator contact whoever is calling.
const DefaultUserAgent = "Goportunitties/1.0 (+https://github.com/KaueChristian/Goportunitties)"

// maxResponseBytes caps what a board can make this process allocate.
const maxResponseBytes = 16 << 20 // 16 MiB

// Client is the HTTP client the adapters share.
type Client struct {
	http      *http.Client
	userAgent string
}

// NewClient builds a Client with an overall timeout per request.
func NewClient(timeout time.Duration, userAgent string) *Client {
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	return &Client{
		http: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				// Boards are contacted once per run, so a pool per host of one
				// is plenty and keeps idle connections from lingering.
				MaxIdleConnsPerHost: 1,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		userAgent: userAgent,
	}
}

// GetJSON fetches url and decodes the body into target.
func (c *Client) GetJSON(ctx context.Context, url string, target any) error {
	return c.GetJSONWith(ctx, url, nil, target)
}

// GetJSONWith is GetJSON with extra headers — an API key or a different Accept,
// which some boards require and others reject.
func (c *Client) GetJSONWith(ctx context.Context, url string, headers map[string]string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building the request for %s: %w", url, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	for name, value := range headers {
		req.Header.Set(name, value)
	}

	res, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("requesting %s: %w", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s answered %s", url, res.Status)
	}

	// A board is not trusted to bound its own response.
	body := io.LimitReader(res.Body, maxResponseBytes)
	if err := json.NewDecoder(body).Decode(target); err != nil {
		return fmt.Errorf("decoding the response from %s: %w", url, err)
	}

	return nil
}

// flexibleID reads an id that a feed may publish as either a string or a
// number — the same board has been seen doing both across entries.
type flexibleID string

func (f *flexibleID) UnmarshalJSON(data []byte) error {
	var asString string
	if err := json.Unmarshal(data, &asString); err == nil {
		*f = flexibleID(asString)
		return nil
	}

	var asNumber json.Number
	if err := json.Unmarshal(data, &asNumber); err == nil {
		*f = flexibleID(asNumber.String())
		return nil
	}

	// An id this adapter cannot read is not a reason to drop the whole feed;
	// the entry is discarded later for having no identity.
	*f = ""
	return nil
}

func (f flexibleID) String() string {
	return string(f)
}
