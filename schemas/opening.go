package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Opening represents the schema of an opening in the database.
type Opening struct {
	gorm.Model // Embedded struct for basic fields like ID, CreatedAt, UpdatedAt, DeletedAt

	// Custom fields for an opening
	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

// OpeningResponse represents the response schema for an opening.
type OpeningResponse struct {
	// Basic fields
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt,omitempty"` // This field will be omitted if it's not set

	// Custom fields for an opening
	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}
