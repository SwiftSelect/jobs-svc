package services

import (
	"jobs-svc/internal/repos"

	"go.mongodb.org/mongo-driver/bson"
)

type ApplicationsService struct {
	AppRepo repos.AppRepo
}

func (s *ApplicationsService) CreateApplication(app bson.M) error {
	return s.AppRepo.CreateApplication(app)
}

func (s *ApplicationsService) GetApplicationsByJobID(jobID uint) ([]bson.M, error) {
	return s.AppRepo.GetApplicationsByJobID(jobID)
}

func (s *ApplicationsService) GetApplicationByID(applicationID string) (bson.M, error) {
	return s.AppRepo.GetApplicationByID(applicationID)
}

func (s *ApplicationsService) GetApplicationByCandidateID(candidateID uint) (bson.M, error) {
	return s.AppRepo.GetApplicationByCandidateID(candidateID)
}

func (s *ApplicationsService) GetApplicationsByCandidateID(candidateID uint) ([]bson.M, error) {
	return s.AppRepo.GetApplicationsByCandidateID(candidateID)
}

// CreateUniqueIndex creates a unique compound index on candidate_id and job_id
func (s *ApplicationsService) CreateUniqueIndex() error {
	return s.AppRepo.CreateUniqueIndex()
}
