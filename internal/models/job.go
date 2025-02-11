package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

type Job struct {
	gorm.Model
	Title        string `gorm:"not null"`
	Description  string `gorm:"not null"`
	Requirements string `gorm:"not null"`
	Location     string
	Status       JobStatus `gorm:"not null"`
	PostedDate   time.Time `gorm:"not null"`
	RecruiterId  uint
}

type Recruiter struct {
	gorm.Model
	Name  string `gorm:"not null"`
	Email string `gorm:"not null"`
	Jobs  []Job
}

type JobStatus int

const (
	Open JobStatus = iota
	Closed
)
