package tests

import (
	"bytes"
	"encoding/json"
	"jobs-svc/internal/handlers"
	"jobs-svc/internal/services"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func setupTestApplicationRouter(handler *handlers.ApplicationHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/applications", handler.CreateApplication).Methods("POST")
	router.HandleFunc("/applications/job/{id}", handler.GetApplicationsByJobID).Methods("GET")
	router.HandleFunc("/applications/{id}", handler.GetApplicationByID).Methods("GET")
	router.HandleFunc("/applications/candidate/{id}", handler.GetApplicationByCandidateID).Methods("GET")
	return router
}

func TestApplicationHandler_CreateApplication(t *testing.T) {
	mockRepo := NewMockApplicationRepo()
	mockKafka := NewMockKafkaPublisher()
	service := services.ApplicationsService{AppRepo: mockRepo}
	handler := handlers.ApplicationHandler{
		ApplicationService: service,
		KafkaPublisher:     mockKafka,
	}

	tests := []struct {
		name           string
		payload        bson.M
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful application creation",
			payload: bson.M{
				"jobId":       1,
				"candidateId": 1,
				"resumeUrl":   "http://example.com/resume.pdf",
				"skills":      []string{"Go", "Python"},
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response bson.M
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["jobId"] != float64(1) {
					t.Errorf("Expected jobId 1, got %v", response["jobId"])
				}
				if response["candidateId"] != float64(1) {
					t.Errorf("Expected candidateId 1, got %v", response["candidateId"])
				}
			},
		},
		{
			name: "missing required fields",
			payload: bson.M{
				"resumeUrl": "http://example.com/resume.pdf",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Body.String() != "JobID and CandidateID are required\n" {
					t.Errorf("Expected error about missing fields, got %v", rr.Body.String())
				}
			},
		},
		{
			name: "duplicate application",
			payload: bson.M{
				"jobId":       1,
				"candidateId": 1,
				"resumeUrl":   "http://example.com/resume.pdf",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if rr.Body.String() != "candidate has already applied for this job\n" {
					t.Errorf("Expected duplicate application error, got %v", rr.Body.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/applications", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			router := setupTestApplicationRouter(&handler)
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestApplicationHandler_GetApplicationsByJobID(t *testing.T) {
	mockRepo := NewMockApplicationRepo()
	mockKafka := NewMockKafkaPublisher()
	service := services.ApplicationsService{AppRepo: mockRepo}
	handler := handlers.ApplicationHandler{
		ApplicationService: service,
		KafkaPublisher:     mockKafka,
	}

	// Create test applications
	app1 := bson.M{
		"application_id": "1",
		"job_id":         uint(1),
		"candidate_id":   uint(1),
		"resume_url":     "http://example.com/resume1.pdf",
		"status": bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		},
	}
	app2 := bson.M{
		"application_id": "2",
		"job_id":         uint(1),
		"candidate_id":   uint(2),
		"resume_url":     "http://example.com/resume2.pdf",
		"status": bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		},
	}

	mockRepo.CreateApplication(app1)
	mockRepo.CreateApplication(app2)

	tests := []struct {
		name           string
		jobID          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get applications for job",
			jobID:          "1",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "invalid job ID",
			jobID:          "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "non-existent job ID",
			jobID:          "999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/applications/job/"+tt.jobID, nil)
			rr := httptest.NewRecorder()

			router := setupTestApplicationRouter(&handler)
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []bson.M
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if len(response) != tt.expectedCount {
					t.Errorf("Expected %d applications, got %d", tt.expectedCount, len(response))
				}
			}
		})
	}
}

func TestApplicationHandler_GetApplicationByID(t *testing.T) {
	mockRepo := NewMockApplicationRepo()
	mockKafka := NewMockKafkaPublisher()
	service := services.ApplicationsService{AppRepo: mockRepo}
	handler := handlers.ApplicationHandler{
		ApplicationService: service,
		KafkaPublisher:     mockKafka,
	}

	// Create test application
	app := bson.M{
		"application_id": "1",
		"job_id":         uint(1),
		"candidate_id":   uint(1),
		"resume_url":     "http://example.com/resume.pdf",
		"status": bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		},
	}

	mockRepo.CreateApplication(app)

	tests := []struct {
		name           string
		applicationID  string
		expectedStatus int
		shouldExist    bool
	}{
		{
			name:           "get existing application",
			applicationID:  "1",
			expectedStatus: http.StatusOK,
			shouldExist:    true,
		},
		{
			name:           "get non-existent application",
			applicationID:  "999",
			expectedStatus: http.StatusOK,
			shouldExist:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/applications/"+tt.applicationID, nil)
			rr := httptest.NewRecorder()

			router := setupTestApplicationRouter(&handler)
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.shouldExist {
				var response bson.M
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if response["applicationId"] != tt.applicationID {
					t.Errorf("Expected application ID %s, got %v", tt.applicationID, response["applicationId"])
				}
			}
		})
	}
}

func TestApplicationHandler_GetApplicationsByCandidateID(t *testing.T) {
	mockRepo := NewMockApplicationRepo()
	mockKafka := NewMockKafkaPublisher()
	service := services.ApplicationsService{AppRepo: mockRepo}
	handler := handlers.ApplicationHandler{
		ApplicationService: service,
		KafkaPublisher:     mockKafka,
	}

	// Create test applications
	app1 := bson.M{
		"application_id": "1",
		"job_id":         uint(1),
		"candidate_id":   uint(1),
		"resume_url":     "http://example.com/resume1.pdf",
		"status": bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		},
	}
	app2 := bson.M{
		"application_id": "2",
		"job_id":         uint(2),
		"candidate_id":   uint(1),
		"resume_url":     "http://example.com/resume2.pdf",
		"status": bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		},
	}

	mockRepo.CreateApplication(app1)
	mockRepo.CreateApplication(app2)

	tests := []struct {
		name           string
		candidateID    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get applications for candidate",
			candidateID:    "1",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "invalid candidate ID",
			candidateID:    "invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:           "non-existent candidate ID",
			candidateID:    "999",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/applications/candidate/"+tt.candidateID, nil)
			rr := httptest.NewRecorder()

			router := setupTestApplicationRouter(&handler)
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response []bson.M
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}
				if len(response) != tt.expectedCount {
					t.Errorf("Expected %d applications, got %d", tt.expectedCount, len(response))
				}
			}
		})
	}
}
