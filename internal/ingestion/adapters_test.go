package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// serve stands in for a job board.
func serve(t *testing.T, status int, body string) (*httptest.Server, *http.Request) {
	t.Helper()

	var seen *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clone := *r
		seen = &clone

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return server, seen
}

func newTestClient() *Client {
	return NewClient(2*time.Second, "")
}

// The feed opens with a legal notice rather than a job; taking it as one would
// store a posting with no company and no link.
const remoteOKBody = `[
  {"legal": "See https://remoteok.com/api for the terms"},
  {"id": "1001", "position": "  Senior   Go   Engineer ", "company": "Acme", "location": "Worldwide", "url": "https://remoteok.com/l/1001"},
  {"id": 1002, "position": "SRE", "company": "Globex", "location": "", "url": "https://remoteok.com/l/1002"},
  {"id": "1003", "position": "", "company": "Sem cargo", "url": "https://remoteok.com/l/1003"},
  {"id": "1004", "position": "Sem link", "company": "Initech", "url": ""},
  {"id": "", "position": "Sem identidade", "company": "Initech", "url": "https://remoteok.com/l/x"}
]`

func TestRemoteOKMapsTheFeed(t *testing.T) {
	server, _ := serve(t, http.StatusOK, remoteOKBody)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	openings, err := source.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	// Only the two complete entries survive: the notice and the three broken
	// rows are dropped rather than stored half-filled.
	if len(openings) != 2 {
		t.Fatalf("got %d openings, want 2: %+v", len(openings), openings)
	}

	first := openings[0]
	if first.Role != "Senior Go Engineer" {
		t.Fatalf("role = %q — the runs of whitespace should be collapsed", first.Role)
	}
	if first.Company != "Acme" || first.Location != "Worldwide" || first.Link != "https://remoteok.com/l/1001" {
		t.Fatalf("mapping = %+v", first)
	}
	if first.Source != SlugRemoteOK || first.ExternalID != "1001" {
		t.Fatalf("identity = %q/%q", first.Source, first.ExternalID)
	}
	if !first.Remote {
		t.Fatal("every posting on this board is remote")
	}
	// The board publishes salaries in USD per year; this application stores BRL
	// per month. Inventing the conversion is worse than saying nothing.
	if first.Salary != 0 {
		t.Fatalf("salary = %d, want 0 (a combinar)", first.Salary)
	}
}

// The same board publishes ids as strings and as numbers.
func TestRemoteOKReadsNumericAndTextualIDs(t *testing.T) {
	server, _ := serve(t, http.StatusOK, remoteOKBody)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	openings, err := source.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if openings[1].ExternalID != "1002" {
		t.Fatalf("numeric id read as %q, want 1002", openings[1].ExternalID)
	}
}

// Location is required and the facet groups by exact value, so an empty one
// gets a label instead of becoming a nameless bucket.
func TestRemoteOKLabelsAnEmptyLocation(t *testing.T) {
	server, _ := serve(t, http.StatusOK, remoteOKBody)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	openings, _ := source.Fetch(context.Background())
	if openings[1].Location != "Remoto" {
		t.Fatalf("location = %q, want Remoto", openings[1].Location)
	}
}

func TestRemoteOKReportsABadStatus(t *testing.T) {
	server, _ := serve(t, http.StatusTooManyRequests, `{}`)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	_, err := source.Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "429") {
		t.Fatalf("got %v, want an error naming the status", err)
	}
}

func TestRemoteOKReportsAMalformedBody(t *testing.T) {
	server, _ := serve(t, http.StatusOK, `{"nao": "e um array"}`)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	if _, err := source.Fetch(context.Background()); err == nil {
		t.Fatal("a body that is not the expected shape should fail")
	}
}

const remotiveBody = `{
  "job-count": 2,
  "jobs": [
    {"id": 501, "title": "Backend Engineer", "company_name": "Acme", "candidate_required_location": "Brazil", "url": "https://remotive.com/j/501", "salary": "$50k - $70k"},
    {"id": 502, "title": "Data Analyst", "company_name": "Globex", "candidate_required_location": "", "url": "https://remotive.com/j/502", "salary": ""},
    {"id": 503, "title": "Sem empresa", "company_name": "", "url": "https://remotive.com/j/503"}
  ]
}`

func TestRemotiveMapsTheFeed(t *testing.T) {
	server, _ := serve(t, http.StatusOK, remotiveBody)

	source := NewRemotive(newTestClient(), 0)
	source.endpoint = server.URL

	openings, err := source.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if len(openings) != 2 {
		t.Fatalf("got %d openings, want 2: %+v", len(openings), openings)
	}

	first := openings[0]
	if first.Role != "Backend Engineer" || first.Company != "Acme" || first.Location != "Brazil" {
		t.Fatalf("mapping = %+v", first)
	}
	if first.Source != SlugRemotive || first.ExternalID != "501" {
		t.Fatalf("identity = %q/%q", first.Source, first.ExternalID)
	}
	// The board publishes salary as free text with no currency or period.
	if first.Salary != 0 {
		t.Fatalf("salary = %d, want 0", first.Salary)
	}
}

// The board supports a limit, so the adapter asks for what it will keep rather
// than downloading everything and discarding most of it.
func TestRemotiveAsksForOnlyWhatItKeeps(t *testing.T) {
	var query string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte(remotiveBody))
	}))
	t.Cleanup(server.Close)

	source := NewRemotive(newTestClient(), 25)
	source.endpoint = server.URL

	if _, err := source.Fetch(context.Background()); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if query != "limit=25" {
		t.Fatalf("query = %q, want limit=25", query)
	}
}

func TestClientIdentifiesItself(t *testing.T) {
	var agent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	if _, err := source.Fetch(context.Background()); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	// Several boards reject an anonymous caller, and an honest agent is what
	// lets them reach whoever is calling.
	if !strings.Contains(agent, "Goportunitties") {
		t.Fatalf("User-Agent = %q", agent)
	}
}

func TestClientStopsWhenTheContextIsCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)

	source := NewRemoteOK(newTestClient())
	source.endpoint = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := source.Fetch(ctx); err == nil {
		t.Fatal("a cancelled context should abort the fetch")
	}
}

func TestBuildResolvesConfiguredSlugs(t *testing.T) {
	sources, err := Build(newTestClient(), 10, []string{"remoteok", " REMOTIVE ", "remoteok", ""})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Case and padding are tolerated; a repeat is not a second source.
	if len(sources) != 2 {
		t.Fatalf("got %d sources, want 2", len(sources))
	}
	if sources[0].Slug() != SlugRemoteOK || sources[1].Slug() != SlugRemotive {
		t.Fatalf("slugs = %q, %q", sources[0].Slug(), sources[1].Slug())
	}
}

// A typo must not look like a board that returned nothing.
func TestBuildRejectsAnUnknownSlug(t *testing.T) {
	_, err := Build(newTestClient(), 10, []string{"remoteok", "linkedin"})

	if err == nil {
		t.Fatal("an unknown source should be an error")
	}
	if !strings.Contains(err.Error(), "linkedin") || !strings.Contains(err.Error(), "remoteok") {
		t.Fatalf("the error should name the typo and the alternatives, got %q", err)
	}
}

// Real feeds join a city with a state that is sometimes missing, and ship the
// separator anyway: "Three Hills," would otherwise be a facet entry of its own,
// separate from "Three Hills".
func TestCleanTextDropsDanglingPunctuation(t *testing.T) {
	tests := map[string]string{
		"Three Hills,":    "Three Hills",
		"  Amet,  ":       "Amet",
		", Belfast":       "Belfast",
		"Curitiba, PR":    "Curitiba, PR",
		"Remote - Brazil": "Remote - Brazil",
		",":               "",
	}

	for given, want := range tests {
		if got := cleanText(given); got != want {
			t.Fatalf("cleanText(%q) = %q, want %q", given, got, want)
		}
	}
}

func TestCleanTextCollapsesAndTruncatesOnRunes(t *testing.T) {
	if got := cleanText("  Desenvolvedor\n\tGo   Sênior  "); got != "Desenvolvedor Go Sênior" {
		t.Fatalf("cleanText = %q", got)
	}

	long := strings.Repeat("é", 200)
	got := cleanText(long)
	if len([]rune(got)) != 120 {
		t.Fatalf("truncated to %d runes, want 120", len([]rune(got)))
	}
	// Cutting on bytes would leave half of a multi-byte character behind.
	if !strings.HasPrefix(got, "éé") {
		t.Fatalf("truncation split a character: %q", got[:8])
	}
}
