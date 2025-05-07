package services

import (
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
)

type JobService struct {
	JobRepo repos.JobRepoInterface
}

func (s *JobService) CreateJob(job *models.Job) error {
	return s.JobRepo.CreateJob(job)
}

func (s *JobService) GetJobByID(id uint) (*models.Job, error) {
	return s.JobRepo.GetJobByID(id)
}

func (s *JobService) GetJobs() (*[]models.Job, error) {
	return s.JobRepo.GetJobs()
}

func (s *JobService) GetJobsByRecruiterID(recruiterID uint) (*[]models.Job, error) {
	return s.JobRepo.GetJobsByRecruiterID(recruiterID)
}

func (s *JobService) UpdateJob(job *models.Job) error {
	return s.JobRepo.UpdateJob(job)
}

func (s *JobService) DeleteJob(id uint) error {
	return s.JobRepo.DeleteJob(id)
}

func (s *JobService) GetJobsByIDs(ids []uint) (*[]models.Job, error) {
	return s.JobRepo.GetJobsByIDs(ids)
}
