package main

import (
	"testing"

	"github.com/KaueChristian/Goportunitties/internal/config"
)

func TestValidateForProductionRequiresAnAdminKey(t *testing.T) {
	settings := config.Settings{Env: "production", AdminKey: ""}

	if err := validateForProduction(settings); err == nil {
		t.Fatal("production with no admin key should refuse to start")
	}
}

func TestValidateForProductionPassesOnceAKeyIsSet(t *testing.T) {
	settings := config.Settings{Env: "production", AdminKey: "s3cr3t"}

	if err := validateForProduction(settings); err != nil {
		t.Fatalf("validateForProduction: %v", err)
	}
}

// Development never needed configuration before this, and an admin key must
// not become the exception — an empty key there is what keeps the write gate
// a no-op.
func TestValidateForProductionAllowsDevelopmentWithNoKey(t *testing.T) {
	settings := config.Settings{Env: "development", AdminKey: ""}

	if err := validateForProduction(settings); err != nil {
		t.Fatalf("validateForProduction: %v", err)
	}
}
