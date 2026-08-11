// Package model holds the structs persisted by GORM.
package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"gorm.io/gorm"
)

// SourceManual marks an opening someone typed into the form, as opposed to one
// brought in by ingestion. Every other value is a provider slug ("remoteok").
const SourceManual = "manual"

// Opening is a job posting as stored in the database.
//
// It deliberately carries no json tags: what the API returns is
// dto.OpeningResponse, so the storage layout can change without breaking any
// client.
//
// The indexes exist because every one of these columns is filtered or sorted on
// by the listing endpoint.
type Opening struct {
	gorm.Model

	Role     string `gorm:"size:120;not null;index"`
	Company  string `gorm:"size:120;not null;index"`
	Location string `gorm:"size:120;not null;index"`
	Remote   bool   `gorm:"not null;index"`
	Link     string `gorm:"size:500;not null"`
	Salary   int64  `gorm:"not null;index"`

	// Source and ExternalID together identify a posting, and carry the unique
	// index created in database.Migrate. Re-running ingestion updates the
	// matching row instead of adding a copy.
	//
	// A manual opening has no external system to borrow an id from, so one is
	// minted for it — which is also what keeps two identical hand-typed
	// postings from colliding on the same key.
	Source     string `gorm:"size:40;not null;default:manual;index"`
	ExternalID string `gorm:"size:200;not null;default:''"`
}

// IsManual reports whether the opening was entered by hand.
func (o Opening) IsManual() bool {
	return o.Source == SourceManual
}

// BeforeCreate settles provenance before the row reaches the unique index.
//
// The rule lives here rather than in a caller because it is an invariant of the
// record: an opening without an identity would collide with every other one
// that lacks it, and any code path that forgot to assign one would fail at the
// database with a constraint error instead of at the point of the mistake.
//
// The two cases are deliberately not symmetric. A manual opening has no
// external system to borrow an id from, so one is minted. An ingested opening
// without an id is a bug in the worker, and minting a fresh one each run would
// turn that bug into silent duplication — the loud failure is the useful one.
func (o *Opening) BeforeCreate(*gorm.DB) error {
	if o.Source == "" {
		o.Source = SourceManual
	}
	if o.ExternalID != "" {
		return nil
	}
	if o.Source != SourceManual {
		return fmt.Errorf("model: an opening from %q must carry an external id", o.Source)
	}

	o.ExternalID = NewIdentity()
	return nil
}

// NewIdentity mints an identity for a posting whose source provides none.
//
// Ingestion normally reuses the provider's own id; this covers manual openings
// and feeds that publish nothing stable to key on.
func NewIdentity() string {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		// crypto/rand only fails if the system entropy source is broken, in
		// which case there is nothing sensible left to fall back to.
		panic("model: cannot read random bytes for an opening identity: " + err.Error())
	}
	return hex.EncodeToString(buf)
}
