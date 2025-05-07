package tests

import (
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
)

type MockJobRepo struct {
	jobs map[uint]*models.Job
}

func NewMockJobRepo() repos.JobRepoInterface {
	return &MockJobRepo{
		jobs: make(map[uint]*models.Job),
	}
}

func (m *MockJobRepo) CreateJob(job *models.Job) error {
	if job.ID == 0 {
		job.ID = uint(len(m.jobs) + 1)
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *MockJobRepo) GetJobByID(id uint) (*models.Job, error) {
	if job, exists := m.jobs[id]; exists {
		return job, nil
	}
	return nil, nil
}

func (m *MockJobRepo) GetJobs() (*[]models.Job, error) {
	jobs := make([]models.Job, 0, len(m.jobs))
	for _, job := range m.jobs {
		jobs = append(jobs, *job)
	}
	return &jobs, nil
}

func (m *MockJobRepo) GetJobsByRecruiterID(recruiterID uint) (*[]models.Job, error) {
	jobs := make([]models.Job, 0)
	for _, job := range m.jobs {
		if job.RecruiterId == recruiterID {
			jobs = append(jobs, *job)
		}
	}
	return &jobs, nil
}

func (m *MockJobRepo) UpdateJob(job *models.Job) error {
	if _, exists := m.jobs[job.ID]; exists {
		m.jobs[job.ID] = job
		return nil
	}
	return nil
}

func (m *MockJobRepo) DeleteJob(id uint) error {
	delete(m.jobs, id)
	return nil
}

func (m *MockJobRepo) GetJobsByIDs(ids []uint) (*[]models.Job, error) {
	jobs := make([]models.Job, 0)
	for _, id := range ids {
		if job, exists := m.jobs[id]; exists {
			jobs = append(jobs, *job)
		}
	}
	return &jobs, nil
}
