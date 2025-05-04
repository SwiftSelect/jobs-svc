package repos

import (
	//"github.com/jinzhu/gorm"
	"jobs-svc/internal/models"

	"gorm.io/gorm"
)

type JobRepo struct {
	DB *gorm.DB
}

func (repo *JobRepo) CreateJob(job *models.Job) error {
	return repo.DB.Create(job).Error
}

func (repo *JobRepo) GetJobByID(id uint) (*models.Job, error) {
	var job models.Job
	err := repo.DB.First(&job, id).Error
	return &job, err
}

func (repo *JobRepo) GetJobs() (*[]models.Job, error) {
	var jobs []models.Job
	err := repo.DB.Find(&jobs).Error
	return &jobs, err
}

func (repo *JobRepo) GetJobsByRecruiterID(recruiterID uint) (*[]models.Job, error) {
	var jobs []models.Job
	err := repo.DB.Where("recruiter_id = ?", recruiterID).Find(&jobs).Error
	return &jobs, err
}

func (repo *JobRepo) UpdateJob(job *models.Job) error {
	return repo.DB.Save(job).Error
}

func (repo *JobRepo) DeleteJob(id uint) error {
	return repo.DB.Delete(&models.Job{}, id).Error
}

func (repo *JobRepo) GetJobsByIDs(ids []uint) (*[]models.Job, error) {
	var jobs []models.Job
	err := repo.DB.Where("id IN ?", ids).Find(&jobs).Error
	return &jobs, err
}
