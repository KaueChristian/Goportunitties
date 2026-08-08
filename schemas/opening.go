package schemas

import (
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
