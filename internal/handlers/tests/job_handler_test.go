package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"jobs-svc/internal/clients"
	"jobs-svc/internal/handlers"
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"jobs-svc/internal/tests"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func setupTestRouter(handler *handlers.JobHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/jobs", handler.GetJobs).Methods("GET")
	router.HandleFunc("/jobs/{id}", handler.GetJobByID).Methods("GET")
	router.HandleFunc("/jobs/recruiter/{id}", handler.GetJobsByRecruiterID).Methods("GET")
	router.HandleFunc("/jobs", handler.CreateJob).Methods("POST")
	router.HandleFunc("/jobs/{id}", handler.UpdateJob).Methods("PUT")
	router.HandleFunc("/jobs/{id}", handler.DeleteJob).Methods("DELETE")
	router.HandleFunc("/jobs/summary", handler.GetJobsByIDs).Methods("POST")
	return router
}

func TestJobHandler_GetJobs(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

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

	router := setupTestRouter(&handler)
	req := httptest.NewRequest("GET", "/jobs", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var jobs []models.Job
	if err := json.NewDecoder(rr.Body).Decode(&jobs); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(jobs) != 2 {
		t.Errorf("Expected 2 jobs, got %d", len(jobs))
	}
}

func TestJobHandler_GetJobByID(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

	// Create a test job
	job := &models.Job{
		Title:    "Test Job",
		Overview: "Test Overview",
		Company:  "Test Company",
	}
	mockRepo.CreateJob(job)

	router := setupTestRouter(&handler)
	req := httptest.NewRequest("GET", "/jobs/1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response["title"] != "Test Job" {
		t.Errorf("Expected title 'Test Job', got '%v'", response["title"])
	}
}

func TestJobHandler_CreateJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

	newJob := models.Job{
		Title:       "New Job",
		Overview:    "New Overview",
		Description: "New Description",
		Company:     "New Company",
		CompanyID:   1,
		Skills:      "Go, Python",
		Experience:  "5+ years",
		Location:    "Remote",
		Status:      models.Open,
		PostedDate:  time.Now(),
		RecruiterId: 1,
	}

	jobJSON, _ := json.Marshal(newJob)
	router := setupTestRouter(&handler)
	req := httptest.NewRequest("POST", "/jobs", bytes.NewBuffer(jobJSON))
	req.Header.Set("Content-Type", "application/json")

	// Add user info to context
	userInfo := &clients.UserResponse{
		ID:    1,
		Email: "test@example.com",
		Org: &clients.Org{
			ID:   1,
			Name: "Test Company",
		},
	}
	ctx := context.WithValue(req.Context(), "userInfo", userInfo)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var response models.Job
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Title != "New Job" {
		t.Errorf("Expected title 'New Job', got '%v'", response.Title)
	}

	// Verify job was published to Kafka
	mockPub := mockKafka.(*tests.MockKafkaPublisher)
	publishedJobs := mockPub.GetPublishedJobs()
	if len(publishedJobs) != 1 {
		t.Errorf("Expected 1 job to be published to Kafka, got %d", len(publishedJobs))
	}
}

func TestJobHandler_UpdateJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

	// Create a test job
	job := &models.Job{
		Title:    "Original Title",
		Overview: "Original Overview",
		Company:  "Original Company",
	}
	mockRepo.CreateJob(job)

	// Update the job
	updatedJob := models.Job{
		Title:    "Updated Title",
		Overview: "Updated Overview",
		Company:  "Updated Company",
	}

	jobJSON, _ := json.Marshal(updatedJob)
	router := setupTestRouter(&handler)
	req := httptest.NewRequest("PUT", "/jobs/1", bytes.NewBuffer(jobJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response models.Job
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got '%v'", response.Title)
	}
}

func TestJobHandler_DeleteJob(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

	// Create a test job
	job := &models.Job{
		Title:    "Test Job",
		Overview: "Test Overview",
	}
	mockRepo.CreateJob(job)

	router := setupTestRouter(&handler)
	req := httptest.NewRequest("DELETE", "/jobs/1", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Verify job was deleted
	deletedJob, _ := service.GetJobByID(1)
	if deletedJob != nil {
		t.Error("Expected job to be deleted but it still exists")
	}
}

func TestJobHandler_GetJobsByIDs(t *testing.T) {
	mockRepo := tests.NewMockJobRepo()
	mockKafka := tests.NewMockKafkaPublisher()
	service := services.JobService{JobRepo: mockRepo}
	handler := handlers.JobHandler{
		JobService:     service,
		KafkaPublisher: mockKafka,
	}

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

	// Request jobs by IDs
	jobIDs := []string{"1", "3"}
	jobIDsJSON, _ := json.Marshal(jobIDs)

	router := setupTestRouter(&handler)
	req := httptest.NewRequest("POST", "/jobs/summary", bytes.NewBuffer(jobIDsJSON))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response []models.JobSummary
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 job summaries, got %d", len(response))
	}
}
