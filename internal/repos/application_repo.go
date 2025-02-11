package repos

import (
	"github.com/jinzhu/gorm"
	"jobs-svc/internal/models"
)

type AppRepo struct {
	DB *gorm.DB
}

func (repo *AppRepo) CreateApplication(application *models.Application) error {
	return repo.DB.Create(application).Error
}

// get all apps for a specific job listing
func (repo *AppRepo) GetApplicationsByJobID(jobID uint) ([]models.Application, error) {
	var applications []models.Application
	err := repo.DB.Where("job_id = ?", jobID).Find(&applications).Error
	return applications, err
}
