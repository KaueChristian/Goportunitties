// Package ingestion fills the index from public job boards.
//
// A Source knows how to talk to one board and hand back openings; the Runner
// knows nothing about any board and only orchestrates. Adding a board means
// adding one file that implements Source.
package ingestion

import (
	"context"
	"fmt"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

// Source is one job board.
//
// Fetch returns openings already shaped as the domain expects them, with Source
// and ExternalID filled in — the identity is what makes a re-run update instead
// of duplicate, and only the adapter knows what the board considers stable.
type Source interface {
	// Slug is the value stored in Opening.Source. It is part of the record's
	// identity, so changing it orphans everything previously ingested.
	Slug() string
	// Name is how the interface spells the source out to a person.
	Name() string
	Fetch(ctx context.Context) ([]model.Opening, error)
}

// Result is what one source did on one run.
type Result struct {
	Source  string
	Fetched int
	Created int64
	Updated int64
	Err     error
}

// Summary aggregates a whole run.
type Summary struct {
	Results []Result
}

// Totals adds the run up.
func (s Summary) Totals() (fetched int, created, updated int64, failed int) {
	for _, result := range s.Results {
		if result.Err != nil {
			failed++
			continue
		}
		fetched += result.Fetched
		created += result.Created
		updated += result.Updated
	}
	return fetched, created, updated, failed
}

// Err returns a single error when every source failed, and nil otherwise: one
// board being down is not a reason to call the whole run a failure.
func (s Summary) Err() error {
	if len(s.Results) == 0 {
		return nil
	}

	failed := 0
	for _, result := range s.Results {
		if result.Err != nil {
			failed++
		}
	}
	if failed < len(s.Results) {
		return nil
	}

	return fmt.Errorf("ingestion: all %d sources failed", failed)
}
