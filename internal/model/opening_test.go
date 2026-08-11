package model_test

import (
	"strings"
	"testing"

	"github.com/KaueChristian/Goportunitties/internal/model"
)

func TestBeforeCreateMintsAnIdentityForManualOpenings(t *testing.T) {
	opening := model.Opening{Role: "Desenvolvedor Go"}

	if err := opening.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate: %v", err)
	}

	if opening.Source != model.SourceManual {
		t.Fatalf("source = %q, want %q", opening.Source, model.SourceManual)
	}
	if opening.ExternalID == "" {
		t.Fatal("a manual opening must leave with an identity, or it collides on the unique index")
	}
	if !opening.IsManual() {
		t.Fatal("IsManual should agree with the assigned source")
	}
}

func TestBeforeCreateKeepsAnIdentityItWasGiven(t *testing.T) {
	opening := model.Opening{Source: "remoteok", ExternalID: "abc-123"}

	if err := opening.BeforeCreate(nil); err != nil {
		t.Fatalf("BeforeCreate: %v", err)
	}

	if opening.Source != "remoteok" || opening.ExternalID != "abc-123" {
		t.Fatalf("provenance was rewritten: %q/%q", opening.Source, opening.ExternalID)
	}
	if opening.IsManual() {
		t.Fatal("an ingested opening is not manual")
	}
}

// Minting an id here would turn a worker bug into silent duplication: every run
// would insert the same posting again under a fresh identity.
func TestBeforeCreateRejectsAnIngestedOpeningWithoutAnIdentity(t *testing.T) {
	opening := model.Opening{Source: "remoteok"}

	err := opening.BeforeCreate(nil)
	if err == nil {
		t.Fatal("an ingested opening with no external id must fail loudly")
	}
	if !strings.Contains(err.Error(), "remoteok") {
		t.Fatalf("the error should name the source, got %q", err)
	}
}

func TestNewIdentityIsUnique(t *testing.T) {
	seen := make(map[string]bool, 1000)

	for i := 0; i < 1000; i++ {
		identity := model.NewIdentity()
		if identity == "" {
			t.Fatal("NewIdentity returned an empty string")
		}
		if seen[identity] {
			t.Fatalf("NewIdentity repeated %q", identity)
		}
		seen[identity] = true
	}
}
