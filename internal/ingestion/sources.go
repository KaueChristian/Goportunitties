package ingestion

import (
	"fmt"
	"sort"
	"strings"
)

// Options carries what an adapter may need beyond the client.
type Options struct {
	MaxPerSource int
	// GitHubToken is optional. Without it GitHub allows 60 requests an hour per
	// address, which is enough for a few sources on a schedule but not for a
	// shared address running them often.
	GitHubToken string
}

// builders maps a configured slug to the adapter that serves it. Adding a board
// means adding one entry here and one file next to this one.
var builders = map[string]func(client *Client, opts Options) Source{
	SlugRemoteOK: func(client *Client, _ Options) Source { return NewRemoteOK(client) },
	SlugRemotive: func(client *Client, opts Options) Source {
		return NewRemotive(client, opts.MaxPerSource)
	},
	SlugBackendBR: func(client *Client, opts Options) Source {
		return NewGitHubVagas(client, SlugBackendBR, "Vagas Back-end BR",
			"backend-br/vagas", opts.GitHubToken, opts.MaxPerSource)
	},
	SlugFrontendBR: func(client *Client, opts Options) Source {
		return NewGitHubVagas(client, SlugFrontendBR, "Vagas Front-end BR",
			"frontendbr/vagas", opts.GitHubToken, opts.MaxPerSource)
	},
}

// Available lists the slugs that can be configured, sorted for a stable message.
func Available() []string {
	slugs := make([]string, 0, len(builders))
	for slug := range builders {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

// Build resolves configured slugs into adapters.
//
// An unknown slug is an error rather than a silent skip: a typo in the
// configuration would otherwise look exactly like a board that returned
// nothing, and the run would report success while ingesting nothing.
func Build(client *Client, opts Options, slugs []string) ([]Source, error) {
	sources := make([]Source, 0, len(slugs))
	seen := make(map[string]bool, len(slugs))

	for _, slug := range slugs {
		slug = strings.ToLower(strings.TrimSpace(slug))
		if slug == "" || seen[slug] {
			continue
		}

		build, ok := builders[slug]
		if !ok {
			return nil, fmt.Errorf(
				"unknown ingestion source %q; available: %s",
				slug, strings.Join(Available(), ", "),
			)
		}

		seen[slug] = true
		sources = append(sources, build(client, opts))
	}

	return sources, nil
}
