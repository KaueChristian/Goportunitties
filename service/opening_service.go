package service

import (
	"errors"

	"github.com/KaueChristian/Goportunitties/dto"
	"github.com/KaueChristian/Goportunitties/repository"
	"github.com/KaueChristian/Goportunitties/schemas"
	"gorm.io/gorm"
)

// ErrOpeningNotFound is returned when an opening cannot be found.
var ErrOpeningNotFound = errors.New("opening not found")

// OpeningService concentrates the business rules around Opening.
type OpeningService interface {
	Create(req *dto.CreateOpeningRequest) (*schemas.Opening, error)
	FindByID(id string) (*schemas.Opening, error)
	FindAll() ([]schemas.Opening, error)
	Update(id string, req *dto.UpdateOpeningRequest) (*schemas.Opening, error)
	Delete(id string) (*schemas.Opening, error)
}

type openingService struct {
	repo repository.OpeningRepository
}

// NewOpeningService creates a new OpeningService.
func NewOpeningService(repo repository.OpeningRepository) OpeningService {
	return &openingService{repo: repo}
}

func (s *openingService) Create(req *dto.CreateOpeningRequest) (*schemas.Opening, error) {
	opening := &schemas.Opening{
		Role:     req.Role,
		Company:  req.Company,
		Location: req.Location,
		Remote:   *req.Remote,
		Link:     req.Link,
		Salary:   req.Salary,
	}
	if err := s.repo.Create(opening); err != nil {
		return nil, err
	}
	return opening, nil
}

func (s *openingService) FindByID(id string) (*schemas.Opening, error) {
	opening, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOpeningNotFound
		}
		return nil, err
	}
	return opening, nil
}

func (s *openingService) FindAll() ([]schemas.Opening, error) {
	return s.repo.FindAll()
}

func (s *openingService) Update(id string, req *dto.UpdateOpeningRequest) (*schemas.Opening, error) {
	opening, err := s.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Role != "" {
		opening.Role = req.Role
	}
	if req.Company != "" {
		opening.Company = req.Company
	}
	if req.Location != "" {
		opening.Location = req.Location
	}
	if req.Remote != nil {
		opening.Remote = *req.Remote
	}
	if req.Link != "" {
		opening.Link = req.Link
	}
	if req.Salary > 0 {
		opening.Salary = req.Salary
	}

	if err := s.repo.Update(opening); err != nil {
		return nil, err
	}
	return opening, nil
}

func (s *openingService) Delete(id string) (*schemas.Opening, error) {
	opening, err := s.FindByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Delete(opening); err != nil {
		return nil, err
	}
	return opening, nil
}
