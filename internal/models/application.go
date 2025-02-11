package models

import (
	"github.com/jinzhu/gorm"
	"time"
)

// using postgres jsonb for simulation nosql like data

type Application struct {
	gorm.Model
	CandidateId uint
	JobId       uint
	Resume      Resume `gorm:"type:jsonb"`
	Status      Status `gorm:"type:jsonb"`
}

type Resume struct {
	Text       string       `json:"text"`
	Skills     []string     `json:"skills"`
	Experience []Experience `json:"experience"`
}

type Experience struct {
	Company  string `json:"company"`
	Role     string `json:"role"`
	Duration string `json:"duration"`
}

type Status struct {
	CurrentStage ApplicationStage `json:"current_stage"`
	LastUpdated  time.Time        `json:"last_updated"`
}

type ApplicationStage int

const (
	Pending ApplicationStage = iota
	UnderReview
	Interviewing
	Hired
	Declined
)
