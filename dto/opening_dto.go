package dto

import (
	"fmt"
	"time"

	"github.com/KaueChristian/Goportunitties/schemas"
)

// errParamIsRequired returns an error indicating that a parameter is required.
func errParamIsRequired(name, typ string) error {
	return fmt.Errorf("parameter: %s (type: %s) is required", name, typ)
}

// CreateOpeningRequest represents the payload to create an Opening.
type CreateOpeningRequest struct {
	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   *bool  `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

// Validate validates the CreateOpeningRequest structure.
func (r *CreateOpeningRequest) Validate() error {
	if r.Role == "" {
		return errParamIsRequired("role", "string")
	}
	if r.Company == "" {
		return errParamIsRequired("company", "string")
	}
	if r.Location == "" {
		return errParamIsRequired("location", "string")
	}
	if r.Link == "" {
		return errParamIsRequired("link", "string")
	}
	if r.Remote == nil {
		return errParamIsRequired("remote", "bool")
	}
	if r.Salary <= 0 {
		return errParamIsRequired("salary", "int64")
	}
	return nil
}

// UpdateOpeningRequest represents the payload to partially update an Opening.
type UpdateOpeningRequest struct {
	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   *bool  `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

// Validate validates that at least one field was provided.
func (r *UpdateOpeningRequest) Validate() error {
	if r.Role != "" || r.Company != "" || r.Location != "" || r.Remote != nil || r.Link != "" || r.Salary > 0 {
		return nil
	}
	return fmt.Errorf("at least one valid field must be provided")
}

// OpeningResponse represents the response schema for an opening.
type OpeningResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	Role     string `json:"role"`
	Company  string `json:"company"`
	Location string `json:"location"`
	Remote   bool   `json:"remote"`
	Link     string `json:"link"`
	Salary   int64  `json:"salary"`
}

// NewOpeningResponse maps an Opening schema to its API response representation.
func NewOpeningResponse(o *schemas.Opening) OpeningResponse {
	return OpeningResponse{
		ID:        o.ID,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
		Role:      o.Role,
		Company:   o.Company,
		Location:  o.Location,
		Remote:    o.Remote,
		Link:      o.Link,
		Salary:    o.Salary,
	}
}

// NewOpeningResponseList maps a slice of Opening schemas to their API response representation.
func NewOpeningResponseList(openings []schemas.Opening) []OpeningResponse {
	responses := make([]OpeningResponse, 0, len(openings))
	for i := range openings {
		responses = append(responses, NewOpeningResponse(&openings[i]))
	}
	return responses
}
