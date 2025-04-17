package models

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

type Job struct {
	gorm.Model
	Title        string    `gorm:"not null" json:"title"`
	Description  string    `gorm:"not null" json:"description"`
	Requirements string    `gorm:"not null" json:"requirements"`
	Location     string    `json:"location"`
	Status       JobStatus `gorm:"not null" json:"status"`
	PostedDate   time.Time `gorm:"not null" json:"postedDate"`
	RecruiterId  uint      `json:"recruiterId"`
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
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_PORT"))

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	PostgresDB = db
	InitPostgres(db)
}
