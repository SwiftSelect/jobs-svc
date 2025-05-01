package services

import (
	"go.mongodb.org/mongo-driver/bson"
	"jobs-svc/internal/repos"
)

type ApplicationsService struct {
	AppRepo repos.AppRepo
}

func (s *ApplicationsService) CreateApplication(app bson.M) error {
	return s.AppRepo.CreateApplication(app)
}

func (s *ApplicationsService) GetApplicationsByJobID(jobID string) ([]bson.M, error) {
	return s.AppRepo.GetApplicationsByJobID(jobID)
}

func (s *ApplicationsService) GetApplicationByCandidateID(candidateID string) (bson.M, error) {
	return s.AppRepo.GetApplicationByCandidateID(candidateID)
}

// CreateUniqueIndex creates a unique compound index on candidate_id and job_id
func (s *ApplicationsService) CreateUniqueIndex() error {
	return s.AppRepo.CreateUniqueIndex()
}
