package models

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	Title       string `gorm:"not null" json:"title"`
	Overview    string `gorm:"not null" json:"overview"`
	Description string `gorm:"not null" json:"description"`
	Company     string `gorm:"not null;default:'Engineering'" json:"company"`
	CompanyID   uint   `gorm:"not null" json:"companyId"`
	//CompanyDescription string    `json:"companyDescription"`
	//Size               string    `json:"size"`
	//Industry           string    `json:"industry"`
	Skills           string    `gorm:"not null;default:'React, Node.js, TypeScript, AWS, MongoDB'" json:"skills"`
	Experience       string    `gorm:"not null;default:'5+ yrs React development, Team leadership'" json:"experience"`
	Location         string    `json:"location"`
	Status           JobStatus `gorm:"not null" json:"status"`
	PostedDate       time.Time `gorm:"not null" json:"postedDate"`
	SalaryRange      string    `gorm:"not null;default:'$120,000 - $160,000'" json:"salaryRange"`
	RecruiterId      uint      `json:"recruiterId"`
	BenefitsAndPerks string    `gorm:"not null;default:'Health, Dental, Vision, 401k'" json:"benefitsAndPerks"`
}

func (j Job) DaysPostedAgo() int {
	return int(time.Since(j.PostedDate).Hours() / 24)
}

type Recruiter struct {
	gorm.Model
	Name  string `gorm:"not null" json:"name"`
	Email string `gorm:"not null" json:"email"`
	Jobs  []Job  `json:"jobs"`
}

type JobStatus int

const (
	Open JobStatus = iota
	Closed
)

func InitPostgres(db *gorm.DB) {
	db.AutoMigrate(&Job{})
}

var PostgresDB *gorm.DB

func ConnectPostgres() {
	dsn := os.Getenv("POSTGRES_URI")
	if dsn == "" {
		log.Fatal("POSTGRES_URI environment variable is not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	PostgresDB = db
	InitPostgres(db)
}
