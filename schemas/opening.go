package schemas

import "gorm.io/gorm"

type Opening struct {
	gorm.Model
	
	ID       int64
	Role     string
	Company  string
	Location string
	Remote   bool
	Link     string
	Salary   int64
}