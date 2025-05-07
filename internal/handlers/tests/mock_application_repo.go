package tests

import (
	"jobs-svc/internal/repos"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type MockApplicationRepo struct {
	applications map[string]bson.M
	mu           sync.RWMutex
}

func NewMockApplicationRepo() repos.ApplicationRepoInterface {
	return &MockApplicationRepo{
		applications: make(map[string]bson.M),
	}
}

func (m *MockApplicationRepo) CreateApplication(application bson.M) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for duplicate application
	jobID := application["job_id"]
	candidateID := application["candidate_id"]
	for _, app := range m.applications {
		if app["job_id"] == jobID && app["candidate_id"] == candidateID {
			return repos.ErrDuplicateApplication
		}
	}

	appID := application["application_id"].(string)
	m.applications[appID] = application
	return nil
}

func (m *MockApplicationRepo) GetApplicationsByJobID(jobID uint) ([]bson.M, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var applications []bson.M
	for _, app := range m.applications {
		if app["job_id"] == jobID {
			applications = append(applications, app)
		}
	}
	return applications, nil
}

func (m *MockApplicationRepo) GetApplicationByID(applicationID string) (bson.M, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if app, exists := m.applications[applicationID]; exists {
		return app, nil
	}
	return nil, nil
}

func (m *MockApplicationRepo) GetApplicationsByCandidateID(candidateID uint) ([]bson.M, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var applications []bson.M
	for _, app := range m.applications {
		if app["candidate_id"] == candidateID {
			applications = append(applications, app)
		}
	}
	return applications, nil
}

func (m *MockApplicationRepo) GetApplicationByCandidateID(candidateID uint) (bson.M, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, app := range m.applications {
		if app["candidate_id"] == candidateID {
			return app, nil
		}
	}
	return nil, nil
}

func (m *MockApplicationRepo) CreateUniqueIndex() error {
	return nil
}
