package repos

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

var ErrDuplicateApplication = errors.New("candidate has already applied for this job")

type ApplicationRepoInterface interface {
	CreateApplication(application bson.M) error
	GetApplicationsByJobID(jobID uint) ([]bson.M, error)
	GetApplicationByID(applicationID string) (bson.M, error)
	GetApplicationsByCandidateID(candidateID uint) ([]bson.M, error)
	GetApplicationByCandidateID(candidateID uint) (bson.M, error)
	CreateUniqueIndex() error
}
