package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newVagas(t *testing.T, endpoint string) *GitHubVagas {
	t.Helper()

	source := NewGitHubVagas(newTestClient(), SlugBackendBR, "Vagas Back-end BR", "backend-br/vagas", "", 100)
	source.endpoint = endpoint
	return source
}

// Every title below was taken from the live board, including the awkward ones.
const githubVagasBody = `[
  {"number": 12376, "title": "[Híbrido - São Paulo/SP] Backend Developer - Rentbrella",
   "html_url": "https://github.com/backend-br/vagas/issues/12376",
   "labels": [{"name": "PJ"}, {"name": "Híbrido"}]},

  {"number": 12374, "title": "[Remoto] IT jobs (Jcal Consultoria)",
   "html_url": "https://github.com/backend-br/vagas/issues/12374",
   "labels": [{"name": "Remoto"}]},

  {"number": 12360, "title": "[Remoto] Full-stack Engineer (JAVA | KOTLIN | GOLANG) - Strider",
   "html_url": "https://github.com/backend-br/vagas/issues/12360",
   "labels": [{"name": "Remoto"}]},

  {"number": 12350, "title": "[Remoto] Desenvolvedora Backend Plena (Afirmativa para Mulheres) | Grupo RBS",
   "html_url": "https://github.com/backend-br/vagas/issues/12350",
   "labels": [{"name": "CLT"}, {"name": "Remoto"}]},

  {"number": 12340, "title": "[REMOTO] Desenvolvedor Pleno @ Eloverde",
   "html_url": "https://github.com/backend-br/vagas/issues/12340",
   "labels": [{"name": "PJ"}]},

  {"number": 12330, "title": "[Presencial - Belo Horizonte/MG] Desenvolvedor TOTVS - BHS",
   "html_url": "https://github.com/backend-br/vagas/issues/12330",
   "labels": [{"name": "CLT"}, {"name": "Presencial"}]},

  {"number": 12320, "title": "Back-end developer (Chatbot) - Híbrido",
   "html_url": "https://github.com/backend-br/vagas/issues/12320",
   "labels": [{"name": "Híbrido"}]},

  {"number": 12310, "title": "[Remoto] Fullstack para controle de tráfego aéreo na Fundação SDTP",
   "html_url": "https://github.com/backend-br/vagas/issues/12310",
   "labels": [{"name": "CLT"}]},

  {"number": 99999, "title": "[Remoto] Uma PR, não uma vaga - Empresa",
   "html_url": "https://github.com/backend-br/vagas/pull/99999",
   "labels": [], "pull_request": {"url": "x"}}
]`

func TestGitHubVagasMapsRealTitles(t *testing.T) {
	server, _ := serve(t, http.StatusOK, githubVagasBody)

	openings, err := newVagas(t, server.URL).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	byID := map[string]int{}
	for i, opening := range openings {
		byID[opening.ExternalID] = i
	}

	tests := []struct {
		id       string
		role     string
		company  string
		location string
		remote   bool
	}{
		{"12376", "Backend Developer", "Rentbrella", "São Paulo/SP", false},
		// The company is the trailing parenthesis when no separator precedes it.
		{"12374", "IT jobs", "Jcal Consultoria", "Remoto, Brasil", true},
		// The pipes inside the parentheses belong to the role, not the company.
		{"12360", "Full-stack Engineer (JAVA | KOTLIN | GOLANG)", "Strider", "Remoto, Brasil", true},
		// Here the pipe outside the parentheses is the real separator.
		{"12350", "Desenvolvedora Backend Plena (Afirmativa para Mulheres)", "Grupo RBS", "Remoto, Brasil", true},
		{"12340", "Desenvolvedor Pleno", "Eloverde", "Remoto, Brasil", true},
		{"12330", "Desenvolvedor TOTVS", "BHS", "Belo Horizonte/MG", false},
		{"12310", "Fullstack para controle de tráfego aéreo", "Fundação SDTP", "Remoto, Brasil", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			index, ok := byID[tt.id]
			if !ok {
				t.Fatalf("issue %s was dropped", tt.id)
			}

			got := openings[index]
			if got.Role != tt.role {
				t.Errorf("role = %q, want %q", got.Role, tt.role)
			}
			if got.Company != tt.company {
				t.Errorf("company = %q, want %q", got.Company, tt.company)
			}
			if got.Location != tt.location {
				t.Errorf("location = %q, want %q", got.Location, tt.location)
			}
			if got.Remote != tt.remote {
				t.Errorf("remote = %t, want %t", got.Remote, tt.remote)
			}
			if got.Source != SlugBackendBR {
				t.Errorf("source = %q", got.Source)
			}
		})
	}
}

// A pull request comes back from the same endpoint and is not a posting.
func TestGitHubVagasSkipsPullRequests(t *testing.T) {
	server, _ := serve(t, http.StatusOK, githubVagasBody)

	openings, err := newVagas(t, server.URL).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	for _, opening := range openings {
		if opening.ExternalID == "99999" {
			t.Fatal("a pull request was stored as an opening")
		}
	}
}

// "Back-end developer (Chatbot) - Híbrido" ends with the arrangement where a
// company would sit. Storing "Híbrido" as the employer would be worse than not
// storing the posting.
func TestGitHubVagasSkipsAPostingWithNoIdentifiableCompany(t *testing.T) {
	server, _ := serve(t, http.StatusOK, githubVagasBody)

	openings, err := newVagas(t, server.URL).Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}

	for _, opening := range openings {
		if opening.ExternalID == "12320" {
			t.Fatalf("stored a posting with %q as the company", opening.Company)
		}
	}
}

func TestGitHubVagasStoresNoSalary(t *testing.T) {
	server, _ := serve(t, http.StatusOK, githubVagasBody)

	openings, _ := newVagas(t, server.URL).Fetch(context.Background())
	for _, opening := range openings {
		if opening.Salary != 0 {
			t.Fatalf("issue %s carried a salary of %d; these boards state none in a readable form",
				opening.ExternalID, opening.Salary)
		}
	}
}

func TestGitHubVagasSendsTheTokenWhenItHasOne(t *testing.T) {
	var auth, accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, accept = r.Header.Get("Authorization"), r.Header.Get("Accept")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	source := NewGitHubVagas(newTestClient(), SlugBackendBR, "n", "backend-br/vagas", "s3cr3t", 100)
	source.endpoint = server.URL

	if _, err := source.Fetch(context.Background()); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if auth != "Bearer s3cr3t" {
		t.Fatalf("Authorization = %q", auth)
	}
	if accept != "application/vnd.github+json" {
		t.Fatalf("Accept = %q, want GitHub's own media type", accept)
	}
}

// Without a token the request must still go out — these boards are readable
// anonymously, just at a lower rate limit.
func TestGitHubVagasWorksWithoutAToken(t *testing.T) {
	var auth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	if _, err := newVagas(t, server.URL).Fetch(context.Background()); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if auth != "" {
		t.Fatalf("Authorization = %q, want it absent", auth)
	}
}

func TestGitHubVagasReportsARateLimit(t *testing.T) {
	server, _ := serve(t, http.StatusForbidden, `{"message": "API rate limit exceeded"}`)

	_, err := newVagas(t, server.URL).Fetch(context.Background())
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("got %v, want an error naming the status", err)
	}
}

func TestRoleAndCompany(t *testing.T) {
	tests := []struct {
		title   string
		role    string
		company string
		ok      bool
	}{
		{"Backend Developer - Rentbrella", "Backend Developer", "Rentbrella", true},
		{"Desenvolvedor Pleno @ Eloverde", "Desenvolvedor Pleno", "Eloverde", true},
		{"Dev Backend | Grupo RBS", "Dev Backend", "Grupo RBS", true},
		{"Fullstack na Fundação SDTP", "Fullstack", "Fundação SDTP", true},
		{"IT jobs (Jcal Consultoria)", "IT jobs", "Jcal Consultoria", true},
		// The last separator wins: the company sits at the end.
		{"Dev Back-end - Pleno - Acme", "Dev Back-end - Pleno", "Acme", true},
		// Nothing identifiable.
		{"Desenvolvedor Backend", "", "", false},
		{"Back-end developer (Chatbot) - Híbrido", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			role, company, ok := roleAndCompany(tt.title)

			if ok != tt.ok {
				t.Fatalf("ok = %t, want %t (role %q, company %q)", ok, tt.ok, role, company)
			}
			if ok && (role != tt.role || company != tt.company) {
				t.Fatalf("got %q / %q, want %q / %q", role, company, tt.role, tt.company)
			}
		})
	}
}

func TestLocationFrom(t *testing.T) {
	tests := []struct {
		bracket string
		remote  bool
		want    string
	}{
		{"Híbrido - São Paulo/SP", false, "São Paulo/SP"},
		{"Presencial - Belo Horizonte/MG", false, "Belo Horizonte/MG"},
		{"Remoto", true, "Remoto, Brasil"},
		{"100% Remoto", true, "Remoto, Brasil"},
		{"Fully Remote", true, "Remoto, Brasil"},
		{"São Paulo", false, "São Paulo"},
		{"", false, "Brasil"},
		// An arrangement that is not remote names no place at all; calling it
		// remote would contradict the modality stored alongside it.
		{"Híbrido", false, "Brasil"},
		{"Presencial", false, "Brasil"},
	}

	for _, tt := range tests {
		if got := locationFrom(tt.bracket, tt.remote); got != tt.want {
			t.Errorf("locationFrom(%q, %t) = %q, want %q", tt.bracket, tt.remote, got, tt.want)
		}
	}
}

// The place and the arrangement are two readings of the same prefix, and a
// prefix naming both ("Remoto logo após Híbrido") made them disagree: the
// opening was stored at "Remoto, Brasil" while flagged as on-site.
func TestLocationAndArrangementNeverContradictEachOther(t *testing.T) {
	brackets := []string{
		"Remoto", "Híbrido", "Presencial", "100% Remoto", "Fully Remote",
		"Remoto logo após Híbrido", "Híbrido - São Paulo/SP", "São Paulo", "",
	}

	for _, bracket := range brackets {
		remote := isRemote(bracket, nil)
		location := locationFrom(bracket, remote)

		if !remote && strings.Contains(strings.ToLower(location), "remoto") {
			t.Errorf("bracket %q: on-site opening located at %q", bracket, location)
		}
	}
}

// On-site wins when a posting mentions both: someone filtering for remote must
// not land on a job that ends up in an office.
func TestIsRemotePrefersTheStricterReading(t *testing.T) {
	tests := []struct {
		bracket string
		labels  []string
		want    bool
	}{
		{"Remoto", nil, true},
		{"100% Remoto", nil, true},
		{"Híbrido - São Paulo", nil, false},
		{"Presencial", nil, false},
		{"Remoto logo após Híbrido", nil, false},
		{"", []string{"Remoto"}, true},
		{"", []string{"CLT", "Presencial"}, false},
		{"", nil, false},
	}

	for _, tt := range tests {
		if got := isRemote(tt.bracket, tt.labels); got != tt.want {
			t.Errorf("isRemote(%q, %v) = %t, want %t", tt.bracket, tt.labels, got, tt.want)
		}
	}
}
