package ingestion

import (
	"strings"

	"github.com/KaueChristian/Goportunitties/internal/dto"
)

// danglingPunctuation is what feeds leave at the edges of a value when a field
// they meant to join turned out to be empty — "Three Hills," for a city with no
// state, for example. It groups as a separate facet entry from "Three Hills",
// so it has to come off.
const danglingPunctuation = " \t,;:·-–—/|"

// cleanText trims a value, collapses runs of whitespace, drops punctuation left
// dangling at either end and cuts the result to the length the column accepts.
//
// Feeds arrive with newlines, tabs and double spaces inside titles; leaving
// them in would make two spellings of the same role look like different roles
// to the search and to the facets.
func cleanText(value string) string {
	collapsed := strings.Join(strings.Fields(value), " ")
	return truncate(strings.Trim(collapsed, danglingPunctuation), dto.MaxTextLength)
}

// cleanLink trims a URL and cuts it to the column width.
func cleanLink(value string) string {
	return truncate(strings.TrimSpace(value), dto.MaxLinkLength)
}

// truncate cuts on a rune boundary, so a multi-byte character is never split
// into an invalid fragment.
func truncate(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return strings.TrimSpace(string(runes[:max]))
}

// cleanLocation normalises the place, falling back to a label rather than an
// empty string: location is required, and the filter groups by exact value.
func cleanLocation(value string) string {
	location := cleanText(value)
	if location == "" {
		return "Remoto"
	}
	return location
}

// isUsable reports whether an opening carries enough to be worth storing.
//
// Feeds do publish rows with an empty title or no link. Dropping them here is
// cheaper than letting the database reject the whole batch over one bad row.
func isUsable(role, company, link, externalID string) bool {
	return externalID != "" &&
		len(role) >= 2 &&
		len(company) >= 2 &&
		(strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://"))
}
