// Package model holds the structs persisted by GORM.
package model

import "gorm.io/gorm"

// Opening is a job posting as stored in the database.
//
// It deliberately carries no json tags: what the API returns is
// dto.OpeningResponse, so the storage layout can change (new columns, ingestion
// metadata) without breaking any client.
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
}
