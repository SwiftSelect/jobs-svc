package services

import (
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
)

type ApplicationsService struct {
	AppRepo repos.AppRepo
}

func (s *ApplicationsService) CreateApplication(app *models.Application) error {
	return s.AppRepo.CreateApplication(app)
}

func (s *ApplicationsService) GetApplicationsByJobID(jobID string) ([]models.Application, error) {
	return s.AppRepo.GetApplicationsByJobID(jobID)
}
