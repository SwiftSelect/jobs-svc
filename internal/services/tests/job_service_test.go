package tests

import (
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"jobs-svc/internal/tests"
	"testing"
	"time"
)

func TestJobService_CreateJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	job := &models.Job{
		Title:       "Software Engineer",
		Overview:    "Looking for a software engineer",
		Description: "Detailed job description",
		Company:     "Test Company",
		CompanyID:   1,
		Skills:      "Go, Python",
		Experience:  "5+ years",
		Location:    "Remote",
		Status:      models.Open,
		PostedDate:  time.Now(),
		RecruiterId: 1,
	}

	err := service.CreateJob(job)
	if err != nil {
		t.Errorf("CreateJob failed: %v", err)
	}

	// Verify job was created
	retrievedJob, err := service.GetJobByID(job.ID)
	if err != nil {
		t.Errorf("GetJobByID failed: %v", err)
	}
	if retrievedJob == nil {
		t.Error("Expected job to be created but it wasn't found")
	}
}

func TestJobService_GetJobs(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	// Create test jobs
	job1 := &models.Job{
		Title:    "Job 1",
		Overview: "Overview 1",
		Company:  "Company 1",
	}
	job2 := &models.Job{
		Title:    "Job 2",
		Overview: "Overview 2",
		Company:  "Company 2",
	}

	mockRepo.CreateJob(job1)
	mockRepo.CreateJob(job2)

	jobs, err := service.GetJobs()
	if err != nil {
		t.Errorf("GetJobs failed: %v", err)
	}

	if len(*jobs) != 2 {
		t.Errorf("Expected 2 jobs, got %d", len(*jobs))
	}
}

func TestJobService_GetJobsByRecruiterID(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	// Create test jobs
	job1 := &models.Job{
		Title:       "Job 1",
		Overview:    "Overview 1",
		RecruiterId: 1,
	}
	job2 := &models.Job{
		Title:       "Job 2",
		Overview:    "Overview 2",
		RecruiterId: 1,
	}
	job3 := &models.Job{
		Title:       "Job 3",
		Overview:    "Overview 3",
		RecruiterId: 2,
	}

	mockRepo.CreateJob(job1)
	mockRepo.CreateJob(job2)
	mockRepo.CreateJob(job3)

	jobs, err := service.GetJobsByRecruiterID(1)
	if err != nil {
		t.Errorf("GetJobsByRecruiterID failed: %v", err)
	}

	if len(*jobs) != 2 {
		t.Errorf("Expected 2 jobs for recruiter 1, got %d", len(*jobs))
	}
}

func TestJobService_UpdateJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	// Create a job
	job := &models.Job{
		Title:    "Original Title",
		Overview: "Original Overview",
		Company:  "Original Company",
	}
	mockRepo.CreateJob(job)

	// Update the job
	job.Title = "Updated Title"
	err := service.UpdateJob(job)
	if err != nil {
		t.Errorf("UpdateJob failed: %v", err)
	}

	// Verify the update
	updatedJob, err := service.GetJobByID(job.ID)
	if err != nil {
		t.Errorf("GetJobByID failed: %v", err)
	}
	if updatedJob.Title != "Updated Title" {
		t.Errorf("Expected title to be 'Updated Title', got '%s'", updatedJob.Title)
	}
}

func TestJobService_DeleteJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	// Create a job
	job := &models.Job{
		Title:    "Test Job",
		Overview: "Test Overview",
	}
	mockRepo.CreateJob(job)

	// Delete the job
	err := service.DeleteJob(job.ID)
	if err != nil {
		t.Errorf("DeleteJob failed: %v", err)
	}

	// Verify the job is deleted
	deletedJob, err := service.GetJobByID(job.ID)
	if err != nil {
		t.Errorf("GetJobByID failed: %v", err)
	}
	if deletedJob != nil {
		t.Error("Expected job to be deleted but it still exists")
	}
}

func TestJobService_GetJobsByIDs(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	service := services.JobService{JobRepo: mockRepo}

	// Create test jobs
	job1 := &models.Job{
		Title:    "Job 1",
		Overview: "Overview 1",
	}
	job2 := &models.Job{
		Title:    "Job 2",
		Overview: "Overview 2",
	}
	job3 := &models.Job{
		Title:    "Job 3",
		Overview: "Overview 3",
	}

	mockRepo.CreateJob(job1)
	mockRepo.CreateJob(job2)
	mockRepo.CreateJob(job3)

	// Get jobs by IDs
	ids := []uint{job1.ID, job3.ID}
	jobs, err := service.GetJobsByIDs(ids)
	if err != nil {
		t.Errorf("GetJobsByIDs failed: %v", err)
	}

	if len(*jobs) != 2 {
		t.Errorf("Expected 2 jobs, got %d", len(*jobs))
	}
}
