package kafka

import (
	"jobs-svc/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

type PublisherInterface interface {
	PublishJob(job *models.Job) error
	PublishApplication(application bson.M) error
	Close() error
}
