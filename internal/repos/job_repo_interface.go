package repos

import (
	"jobs-svc/internal/models"
)

type JobRepoInterface interface {
	CreateJob(job *models.Job) error
	GetJobByID(id uint) (*models.Job, error)
	GetJobs() (*[]models.Job, error)
	GetJobsByRecruiterID(recruiterID uint) (*[]models.Job, error)
	UpdateJob(job *models.Job) error
	DeleteJob(id uint) error
	GetJobsByIDs(ids []uint) (*[]models.Job, error)
}
