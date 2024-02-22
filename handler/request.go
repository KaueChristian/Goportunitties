package handler

import "fmt"

// errparamIsRequired returns an error indicating that a parameter is required.
func errparamIsRequired(name, typ string) error {
	return fmt.Errorf("parameter: %s (type: %s) is required", name, typ)
}

// Create opening request structure
type crateOpeningRequest struct {
	Role      string `json:"role"`
	Company   string `json:"company"`
	Location  string `json:"location"`
	Remote    *bool  `json:"remote"`
	Link      string `json:"link"`
	Salary    int64  `json:"salary"`
}

// Validate validates the crateOpeningRequest structure.
func (r *crateOpeningRequest) Validate() error {
	if r.Role == "" && r.Company == "" && r.Location == "" && r.Remote == nil && r.Salary <= 0 {
		return fmt.Errorf("request body is empty or malformed")
	}
	if r.Role == "" {
		return errparamIsRequired("role", "string")
	}
	if r.Company == "" {
		return errparamIsRequired("company", "string")
	}
	if r.Location == "" {
		return errparamIsRequired("location", "string")
	}
	if r.Link == "" {
		return errparamIsRequired("link", "string")
	}
	if r.Remote == nil {
		return errparamIsRequired("remote", "bool")
	}
	if r.Salary <= 0 {
		return errparamIsRequired("salary", "int64")
	}
	return nil
}
