package ingestion

import (
	"context"
	"fmt"
	"strings"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

// Slugs for the Brazilian community boards. They are part of each record's
// identity and must not change.
const (
	SlugBackendBR  = "backend-br"
	SlugFrontendBR = "frontend-br"
)

// githubAPI is where the issues of a repository are read from.
const githubAPI = "https://api.github.com"

// GitHubVagas reads a Brazilian community job board.
//
// Several of them run on GitHub Issues rather than on a job platform: one issue
// is one posting. That makes the GitHub API a free, public, keyless feed of
// real Brazilian openings — which the international boards this project also
// reads do not offer.
//
// The cost is that a title is prose, not fields. Everything below is a best
// effort at reading a human convention, and a posting whose company cannot be
// identified is skipped rather than stored under a guess.
type GitHubVagas struct {
	client   *Client
	slug     string
	name     string
	repo     string
	endpoint string
	limit    int
	token    string
}

// NewGitHubVagas builds an adapter for one repository, e.g. "backend-br/vagas".
func NewGitHubVagas(client *Client, slug, name, repo, token string, limit int) *GitHubVagas {
	return &GitHubVagas{
		client:   client,
		slug:     slug,
		name:     name,
		repo:     repo,
		endpoint: githubAPI,
		limit:    limit,
		token:    token,
	}
}

func (s *GitHubVagas) Slug() string { return s.slug }
func (s *GitHubVagas) Name() string { return s.name }

// githubIssue is the subset of the issue payload a posting needs.
type githubIssue struct {
	Number  int    `json:"number"`
	Title   string `json:"title"`
	HTMLURL string `json:"html_url"`
	Labels  []struct {
		Name string `json:"name"`
	} `json:"labels"`
	// Pull requests come back from the same endpoint and are not postings; this
	// field is the only thing that tells them apart.
	PullRequest *struct{} `json:"pull_request"`
}

// Fetch reads the open issues and maps them onto openings.
func (s *GitHubVagas) Fetch(ctx context.Context) ([]model.Opening, error) {
	perPage := s.limit
	if perPage <= 0 || perPage > 100 {
		// The API caps a page at 100; more than that would need pagination,
		// which one pass of a job board does not justify.
		perPage = 100
	}

	url := fmt.Sprintf(
		"%s/repos/%s/issues?state=open&sort=created&direction=desc&per_page=%d",
		s.endpoint, s.repo, perPage,
	)

	headers := map[string]string{"Accept": "application/vnd.github+json"}
	if s.token != "" {
		// Unauthenticated GitHub allows 60 requests an hour per address, which
		// a few sources on a shared address can exhaust. A token raises it to
		// 5000 and is the only reason this adapter takes one.
		headers["Authorization"] = "Bearer " + s.token
	}

	issues := []githubIssue{}
	if err := s.client.GetJSONWith(ctx, url, headers, &issues); err != nil {
		return nil, err
	}

	openings := make([]model.Opening, 0, len(issues))
	for _, issue := range issues {
		if issue.PullRequest != nil {
			continue
		}

		posting, ok := s.toOpening(issue)
		if !ok {
			continue
		}
		openings = append(openings, posting)
	}

	return openings, nil
}

func (s *GitHubVagas) toOpening(issue githubIssue) (model.Opening, bool) {
	bracket, rest := splitBracket(issue.Title)

	role, company, ok := roleAndCompany(rest)
	if !ok {
		return model.Opening{}, false
	}

	labels := make([]string, 0, len(issue.Labels))
	for _, label := range issue.Labels {
		labels = append(labels, label.Name)
	}

	role = cleanText(role)
	company = cleanText(company)
	link := cleanLink(issue.HTMLURL)
	identity := fmt.Sprintf("%d", issue.Number)

	if !isUsable(role, company, link, identity) {
		return model.Opening{}, false
	}

	// The arrangement is decided first and the place is derived from it, so the
	// two cannot contradict each other. Deciding them independently produced
	// openings located at "Remoto, Brasil" while flagged as on-site.
	remote := isRemote(bracket, labels)

	return model.Opening{
		Role:       role,
		Company:    company,
		Location:   locationFrom(bracket, remote),
		Remote:     remote,
		Link:       link,
		Salary:     0,
		Source:     s.slug,
		ExternalID: identity,
	}, true
}

// splitBracket peels the "[...]" prefix these boards use for the arrangement
// and the city, returning it and whatever follows.
func splitBracket(title string) (bracket, rest string) {
	trimmed := strings.TrimSpace(title)
	if !strings.HasPrefix(trimmed, "[") {
		return "", trimmed
	}

	end := strings.Index(trimmed, "]")
	if end < 0 {
		return "", trimmed
	}

	return strings.TrimSpace(trimmed[1:end]), strings.TrimSpace(trimmed[end+1:])
}

// companySeparators are what the boards put between the role and the company,
// most specific first.
var companySeparators = []string{" @ ", " | ", " – ", " — ", " - ", " na ", " no "}

// arrangementWords are not company names. They appear where a company would sit
// when a title ends with the arrangement instead ("Back-end developer - Híbrido").
var arrangementWords = []string{"remoto", "remote", "híbrido", "hibrido", "presencial", "hybrid", "onsite"}

// roleAndCompany splits "Cargo - Empresa" into its two halves.
//
// The split is taken at the last separator that sits outside parentheses: a
// role like "Full-stack Engineer (JAVA | KOTLIN | GOLANG) - Strider" carries a
// separator inside the parentheses that has nothing to do with the company.
//
// It reports false when no company can be identified, which is the honest
// outcome for a title that does not follow the convention — storing the role as
// its own employer would be worse than skipping the posting.
func roleAndCompany(title string) (role, company string, ok bool) {
	if index, separator := lastSeparatorOutsideParens(title); index >= 0 {
		role = strings.TrimSpace(title[:index])
		company = strings.TrimSpace(title[index+len(separator):])

		if role != "" && !isArrangement(company) {
			return role, company, true
		}
	}

	// "IT jobs (Jcal Consultoria)" — the company is the trailing parenthesis.
	if strings.HasSuffix(title, ")") {
		if open := strings.LastIndex(title, "("); open > 0 {
			role = strings.TrimSpace(title[:open])
			company = strings.TrimSpace(title[open+1 : len(title)-1])

			if role != "" && company != "" && !isArrangement(company) {
				return role, company, true
			}
		}
	}

	return "", "", false
}

// lastSeparatorOutsideParens finds where the company starts, ignoring anything
// nested in parentheses.
func lastSeparatorOutsideParens(title string) (int, string) {
	depth := 0
	bestIndex, bestSeparator := -1, ""

	for i, r := range title {
		switch r {
		case '(':
			depth++
			continue
		case ')':
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth > 0 {
			continue
		}

		for _, separator := range companySeparators {
			if strings.HasPrefix(title[i:], separator) {
				bestIndex, bestSeparator = i, separator
				break
			}
		}
	}

	return bestIndex, bestSeparator
}

func isArrangement(value string) bool {
	lowered := strings.ToLower(strings.TrimSpace(value))
	for _, word := range arrangementWords {
		if lowered == word {
			return true
		}
	}
	return false
}

// locationFrom reads the place out of the "[...]" prefix.
//
// The prefix holds either an arrangement ("Remoto"), a city ("São Paulo"), or
// both ("Híbrido - São Paulo/SP"). Everything on these boards is in Brazil, so
// a prefix that names no city still says more than an empty string would.
//
// It takes the arrangement already decided by isRemote rather than reading it
// again: a prefix like "Remoto logo após Híbrido" mentions both, and two
// independent readings of it disagreed.
func locationFrom(bracket string, remote bool) string {
	bracket = strings.TrimSpace(bracket)
	if bracket == "" {
		return "Brasil"
	}

	// "Híbrido - São Paulo/SP" -> the city is what follows the dash.
	for _, separator := range []string{" - ", " – ", " — ", " | ", "/"} {
		if index := strings.Index(bracket, separator); index > 0 {
			city := strings.TrimSpace(bracket[index+len(separator):])
			if city != "" && !isArrangement(city) {
				return cleanText(city)
			}
		}
	}

	if isArrangementish(bracket) {
		// "[Remoto]" names a place; "[Híbrido]" and "[Presencial]" name only an
		// arrangement, and reading them as remote would put an opening that
		// expects someone in an office under a location that says otherwise.
		if remote {
			return "Remoto, Brasil"
		}
		return "Brasil"
	}
	return cleanText(bracket)
}

// isArrangementish is the loose form of isArrangement: it matches "100% Remoto"
// and "Fully Remote" as well as the bare words.
func isArrangementish(value string) bool {
	lowered := strings.ToLower(value)
	for _, word := range arrangementWords {
		if strings.Contains(lowered, word) {
			return true
		}
	}
	return false
}

// isRemote decides the arrangement from the title's prefix, falling back to the
// issue's labels.
//
// On-site wins over remote when a posting mentions both: "[Remoto logo após
// Híbrido]" ends up in an office, and claiming otherwise would send someone to
// a posting that does not match the filter they chose.
func isRemote(bracket string, labels []string) bool {
	haystack := strings.ToLower(bracket + " " + strings.Join(labels, " "))

	switch {
	case strings.Contains(haystack, "presencial"), strings.Contains(haystack, "onsite"):
		return false
	case strings.Contains(haystack, "híbrido"), strings.Contains(haystack, "hibrido"),
		strings.Contains(haystack, "hybrid"):
		return false
	case strings.Contains(haystack, "remot"):
		return true
	default:
		return false
	}
}
