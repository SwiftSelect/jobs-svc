package tests

import (
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/models"

	"go.mongodb.org/mongo-driver/bson"
)

type MockKafkaPublisher struct {
	publishedJobs         []*models.Job
	publishedApplications []bson.M
}

func NewMockKafkaPublisher() kafka.PublisherInterface {
	return &MockKafkaPublisher{
		publishedJobs:         make([]*models.Job, 0),
		publishedApplications: make([]bson.M, 0),
	}
}

func (m *MockKafkaPublisher) PublishJob(job *models.Job) error {
	m.publishedJobs = append(m.publishedJobs, job)
	return nil
}

func (m *MockKafkaPublisher) PublishApplication(application bson.M) error {
	m.publishedApplications = append(m.publishedApplications, application)
	return nil
}

func (m *MockKafkaPublisher) Close() error {
	return nil
}

func (m *MockKafkaPublisher) GetPublishedJobs() []*models.Job {
	return m.publishedJobs
}

func (m *MockKafkaPublisher) GetPublishedApplications() []bson.M {
	return m.publishedApplications
}
