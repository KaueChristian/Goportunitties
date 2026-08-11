package ingestion

import (
	"fmt"
	"sort"
	"strings"
)

// builders maps a configured slug to the adapter that serves it. Adding a board
// means adding one entry here and one file next to this one.
var builders = map[string]func(client *Client, maxPerSource int) Source{
	SlugRemoteOK: func(client *Client, _ int) Source { return NewRemoteOK(client) },
	SlugRemotive: func(client *Client, maxPerSource int) Source { return NewRemotive(client, maxPerSource) },
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
func Build(client *Client, maxPerSource int, slugs []string) ([]Source, error) {
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
		sources = append(sources, build(client, maxPerSource))
	}

	return sources, nil
}
