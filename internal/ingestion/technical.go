package ingestion

import "strings"

// technicalWords are single tokens that place a role in technology. They are
// matched whole, never as substrings: "ai" inside "retail" and "go" inside
// "cargo" would otherwise let anything through.
var technicalWords = map[string]bool{
	// Roles
	"developer": true, "developers": true, "dev": true, "devs": true,
	"engineer": true, "engineers": true, "engineering": true,
	"programmer": true, "coder": true, "software": true,
	"backend": true, "frontend": true, "fullstack": true, "devops": true,
	"sre": true, "sysadmin": true, "architect": true, "architects": true,
	"cto": true, "ciso": true, "qa": true, "tester": true, "webmaster": true,
	"dba": true, "cybersecurity": true, "infosec": true,

	// Portuguese, for the community boards
	"programador": true, "programadora": true,
	"desenvolvedor": true, "desenvolvedora": true,
	"engenheiro": true, "engenheira": true, "arquiteto": true, "arquiteta": true,

	// Domains and stacks
	"api": true, "sdk": true, "saas": true, "blockchain": true, "crypto": true,
	"ios": true, "android": true, "react": true, "angular": true, "vue": true,
	"svelte": true, "node": true, "nodejs": true, "python": true, "java": true,
	"golang": true, "ruby": true, "rails": true, "php": true, "dotnet": true,
	"csharp": true, "rust": true, "kotlin": true, "swift": true, "scala": true,
	"elixir": true, "erlang": true, "typescript": true, "javascript": true,
	"sql": true, "nosql": true, "mongodb": true, "postgres": true,
	"postgresql": true, "mysql": true, "redis": true, "kubernetes": true,
	"docker": true, "terraform": true, "aws": true, "azure": true, "gcp": true,
	"ml": true, "ai": true, "llm": true, "nlp": true,
}

// technicalPhrases are multi-word signals, matched as substrings because their
// parts mean nothing alone: "data" and "lead" are not technical by themselves.
var technicalPhrases = []string{
	"back-end", "back end", "front-end", "front end", "full-stack", "full stack",
	"site reliability", "machine learning", "artificial intelligence",
	"data science", "data scientist", "data engineer", "data analyst",
	"tech lead", "technical lead", "solutions architect",
	"product manager", "product engineer", "technical writer",
	"quality assurance", "system administrator", "systems administrator",
	"service desk", "it support", "information technology",
	"c++", "c#", ".net", "smart contract", "web3", "game dev",
	"ux designer", "ui designer", "product designer", "web designer", "ui/ux",
	"platform engineer", "release engineer", "support engineer",
	"security engineer", "network engineer", "cloud engineer", "mobile developer",
}

// looksTechnical reports whether a role belongs on a technology board.
//
// This is an allowlist, not a blocklist, and that direction is deliberate. The
// list of jobs that are *not* technology has no end — the boards this project
// reads have published warehouse work, lifeguarding and cake recipes — so a
// blocklist would leak whatever nobody thought to name. An allowlist errs the
// other way: it drops a genuine role with an unusual title, which costs one
// missing posting, instead of admitting noise onto a board that advertises
// itself as technical.
//
// Seniority words are deliberately absent: "senior" and "junior" say nothing
// about the field, and including them admitted "Senior Graphic Designer".
func looksTechnical(role string) bool {
	lowered := strings.ToLower(role)

	for _, phrase := range technicalPhrases {
		if strings.Contains(lowered, phrase) {
			return true
		}
	}

	for _, token := range tokenize(lowered) {
		if technicalWords[token] {
			return true
		}
	}

	return false
}

// tokenize splits on anything that is not alphanumeric, keeping '+' and '#' so
// that "c++" and "c#" survive as tokens of their own.
func tokenize(value string) []string {
	return strings.FieldsFunc(value, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '+', r == '#':
			return false
		default:
			return true
		}
	})
}
