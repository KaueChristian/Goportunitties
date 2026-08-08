package handler

import (
	"github.com/KaueChristian/Goportunitties/config"
	"github.com/KaueChristian/Goportunitties/repository"
	"github.com/KaueChristian/Goportunitties/service"
)

var logger = config.GetLogger("handler")

// OpeningHandler handles HTTP requests for openings.
type OpeningHandler struct {
	service service.OpeningService
}

// NewOpeningHandler builds an OpeningHandler wired to the SQLite-backed service.
func NewOpeningHandler() *OpeningHandler {
	repo := repository.NewOpeningRepository(config.GetSQlite())
	svc := service.NewOpeningService(repo)
	return &OpeningHandler{service: svc}
}
