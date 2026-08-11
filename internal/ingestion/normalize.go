package ingestion

import (
	"strings"
	"unicode/utf8"

	"github.com/KaueChristian/Goportunitties/internal/dto"
)

// danglingPunctuation is what feeds leave at the edges of a value when a field
// they meant to join turned out to be empty — "Three Hills," for a city with no
// state, for example. It groups as a separate facet entry from "Three Hills",
// so it has to come off.
const danglingPunctuation = " \t,;:·-–—/|"

// cleanText repairs mojibake, trims a value, collapses runs of whitespace,
// drops punctuation left dangling at either end, and cuts the result to the
// length the column accepts.
//
// Feeds arrive with newlines, tabs and double spaces inside titles; leaving
// them in would make two spellings of the same role look like different roles
// to the search and to the facets.
func cleanText(value string) string {
	collapsed := strings.Join(strings.Fields(repairMojibake(value)), " ")
	return truncate(strings.Trim(collapsed, danglingPunctuation), dto.MaxTextLength)
}

// repairMojibake undoes a specific, common corruption seen in the wild: a
// board's own pipeline decodes its UTF-8 text as Latin-1 and re-encodes that as
// UTF-8, turning one character (’, é, —, ...) into several. RemoteOK ships
// titles this way — "Don't" arrives as "Donâ<C1>t", never as "Don't".
//
// The tell is C1 control characters (U+0080–U+009F): real text never contains
// them, but they are exactly what a byte in that range becomes when Latin-1
// stands in for UTF-8. Their presence is treated as a strong, low-false-positive
// signal rather than a guess.
//
// The repair reads each rune's own code point back as a single byte and
// decodes that byte sequence as UTF-8. It only trusts that round-trip when it
// is itself valid UTF-8 — a board is occasionally seen truncating a title mid
// multi-byte character (an emoji cut in half), which turns the reconstructed
// bytes into an incomplete sequence no correct repair can recover.
//
// When the round-trip fails, the C1 controls are stripped instead of restored:
// the exact character is unrecoverable, but the raw control bytes are what
// render as garbled glyphs, and dropping them is still strictly better than
// showing the corruption.
func repairMojibake(value string) string {
	if !strings.ContainsFunc(value, isC1Control) {
		return value
	}

	raw := make([]byte, 0, len(value))
	for _, r := range value {
		if r > 0xFF {
			// A genuine wide character (emoji, CJK, ...) alongside a C1 control
			// means this is not the single-layer corruption this function
			// repairs; only the controls are stripped, nothing is guessed.
			return stripC1Controls(value)
		}
		raw = append(raw, byte(r))
	}

	if !utf8.Valid(raw) {
		return stripC1Controls(value)
	}
	return string(raw)
}

func stripC1Controls(value string) string {
	return strings.Map(func(r rune) rune {
		if isC1Control(r) {
			return -1
		}
		return r
	}, value)
}

func isC1Control(r rune) bool {
	return r >= 0x80 && r <= 0x9F
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
