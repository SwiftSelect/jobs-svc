package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"

	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
)

// MockApplicationRepo is a mock implementation of ApplicationRepoInterface
type MockApplicationRepo struct {
	mock.Mock
}

func (m *MockApplicationRepo) CreateUniqueIndex() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockApplicationRepo) CreateApplication(application bson.M) error {
	args := m.Called(application)
	return args.Error(0)
}

func (m *MockApplicationRepo) GetApplicationsByJobID(jobID uint) ([]bson.M, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bson.M), args.Error(1)
}

func (m *MockApplicationRepo) GetApplicationByID(applicationID string) (bson.M, error) {
	args := m.Called(applicationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(bson.M), args.Error(1)
}

func (m *MockApplicationRepo) GetApplicationsByCandidateID(candidateID uint) ([]bson.M, error) {
	args := m.Called(candidateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]bson.M), args.Error(1)
}

func (m *MockApplicationRepo) GetApplicationByCandidateID(candidateID uint) (bson.M, error) {
	args := m.Called(candidateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(bson.M), args.Error(1)
}

// MockKafkaPublisher is a mock implementation of kafka.PublisherInterface
type MockKafkaPublisher struct {
	mock.Mock
}

func (m *MockKafkaPublisher) PublishJob(job *models.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockKafkaPublisher) PublishApplication(application bson.M) error {
	args := m.Called(application)
	return args.Error(0)
}

func (m *MockKafkaPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestApplicationHandler_CreateApplication(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func(*MockApplicationRepo, *MockKafkaPublisher)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful application",
			requestBody: map[string]interface{}{
				"jobId":       123,
				"candidateId": 456,
				"resumeUrl":   "https://example.com/resume.pdf",
				"email":       "test@example.com",
				"phone":       "1234567890",
			},
			mockSetup: func(m *MockApplicationRepo, k *MockKafkaPublisher) {
				m.On("CreateApplication", mock.AnythingOfType("primitive.M")).Return(nil)
				k.On("PublishApplication", mock.AnythingOfType("primitive.M")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)

				// Check required fields
				assert.Equal(t, float64(123), response["jobId"])
				assert.Equal(t, float64(456), response["candidateId"])
				assert.Equal(t, "https://example.com/resume.pdf", response["resumeUrl"])
				assert.Equal(t, "test@example.com", response["email"])
				assert.Equal(t, "1234567890", response["phone"])

				// Check status
				status, ok := response["status"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "Applied", status["currentStage"])
				assert.NotEmpty(t, status["lastUpdated"])

				// Check applicationId is present and not empty
				assert.NotEmpty(t, response["applicationId"])
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"candidateId": 456,
			},
			mockSetup:      func(m *MockApplicationRepo, k *MockKafkaPublisher) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "JobID and CandidateID are required\n", w.Body.String())
			},
		},
		{
			name: "repository error",
			requestBody: map[string]interface{}{
				"jobId":       123,
				"candidateId": 456,
				"resumeUrl":   "https://example.com/resume.pdf",
				"email":       "test@example.com",
				"phone":       "1234567890",
			},
			mockSetup: func(m *MockApplicationRepo, k *MockKafkaPublisher) {
				m.On("CreateApplication", mock.AnythingOfType("primitive.M")).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "Failed to create application\n", w.Body.String())
			},
		},
		{
			name: "duplicate application",
			requestBody: map[string]interface{}{
				"jobId":       123,
				"candidateId": 456,
				"resumeUrl":   "https://example.com/resume.pdf",
				"email":       "test@example.com",
				"phone":       "1234567890",
			},
			mockSetup: func(m *MockApplicationRepo, k *MockKafkaPublisher) {
				m.On("CreateApplication", mock.AnythingOfType("primitive.M")).Return(errors.New("candidate has already applied for this job"))
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "candidate has already applied for this job\n", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockApplicationRepo)
			mockKafka := new(MockKafkaPublisher)
			tt.mockSetup(mockRepo, mockKafka)

			service := services.ApplicationsService{AppRepo: mockRepo}
			handler := ApplicationHandler{
				ApplicationService: service,
				KafkaPublisher:     mockKafka,
			}

			// Create test request
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/applications", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			handler.CreateApplication(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockRepo.AssertExpectations(t)
			mockKafka.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_GetApplicationsByJobID(t *testing.T) {
	now := time.Now()
	mockApplications := []bson.M{
		{
			"application_id": "app1",
			"job_id":         123,
			"candidate_id":   456,
			"resume_url":     "https://example.com/resume1.pdf",
			"email":          "test1@example.com",
			"phone":          "1234567890",
			"status": bson.M{
				"current_stage": "Applied",
				"last_updated":  now,
			},
		},
		{
			"application_id": "app2",
			"job_id":         123,
			"candidate_id":   789,
			"resume_url":     "https://example.com/resume2.pdf",
			"email":          "test2@example.com",
			"phone":          "0987654321",
			"status": bson.M{
				"current_stage": "Reviewed",
				"last_updated":  now,
			},
		},
	}

	tests := []struct {
		name           string
		jobID          string
		mockSetup      func(*MockApplicationRepo)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:  "successful retrieval",
			jobID: "123",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByJobID", uint(123)).Return(mockApplications, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)

				// Check first application
				assert.Equal(t, "app1", response[0]["applicationId"])
				assert.Equal(t, float64(123), response[0]["jobId"])
				assert.Equal(t, float64(456), response[0]["candidateId"])
				assert.Equal(t, "https://example.com/resume1.pdf", response[0]["resumeUrl"])
				assert.Equal(t, "test1@example.com", response[0]["email"])
				assert.Equal(t, "1234567890", response[0]["phone"])
				status1 := response[0]["status"].(map[string]interface{})
				assert.Equal(t, "Applied", status1["currentStage"])
				assert.NotEmpty(t, status1["lastUpdated"])

				// Check second application
				assert.Equal(t, "app2", response[1]["applicationId"])
				assert.Equal(t, float64(123), response[1]["jobId"])
				assert.Equal(t, float64(789), response[1]["candidateId"])
				assert.Equal(t, "https://example.com/resume2.pdf", response[1]["resumeUrl"])
				assert.Equal(t, "test2@example.com", response[1]["email"])
				assert.Equal(t, "0987654321", response[1]["phone"])
				status2 := response[1]["status"].(map[string]interface{})
				assert.Equal(t, "Reviewed", status2["currentStage"])
				assert.NotEmpty(t, status2["lastUpdated"])
			},
		},
		{
			name:  "no applications found",
			jobID: "123",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByJobID", uint(123)).Return([]bson.M{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Empty(t, response)
			},
		},
		{
			name:  "repository error",
			jobID: "123",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByJobID", uint(123)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "database error\n", w.Body.String())
			},
		},
		{
			name:           "invalid job ID",
			jobID:          "invalid",
			mockSetup:      func(m *MockApplicationRepo) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "Invalid job ID format\n", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockApplicationRepo)
			tt.mockSetup(mockRepo)

			service := services.ApplicationsService{AppRepo: mockRepo}
			handler := ApplicationHandler{ApplicationService: service}

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/jobs/"+tt.jobID+"/applications", nil)
			w := httptest.NewRecorder()

			// Setup router with vars
			router := mux.NewRouter()
			router.HandleFunc("/jobs/{id}/applications", handler.GetApplicationsByJobID).Methods(http.MethodGet)
			req = mux.SetURLVars(req, map[string]string{"id": tt.jobID})

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_GetApplicationByID(t *testing.T) {
	now := time.Now()
	mockApplication := bson.M{
		"application_id": "app123",
		"job_id":         123,
		"candidate_id":   456,
		"resume_url":     "https://example.com/resume.pdf",
		"email":          "test@example.com",
		"phone":          "1234567890",
		"status": bson.M{
			"current_stage": "applied",
			"last_updated":  now,
		},
	}

	tests := []struct {
		name           string
		applicationID  string
		mockSetup      func(*MockApplicationRepo)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:          "successful retrieval",
			applicationID: "app123",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationByID", "app123").Return(mockApplication, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)

				// Check required fields
				assert.Equal(t, "app123", response["applicationId"])
				assert.Equal(t, float64(123), response["jobId"])
				assert.Equal(t, float64(456), response["candidateId"])
				assert.Equal(t, "https://example.com/resume.pdf", response["resumeUrl"])
				assert.Equal(t, "test@example.com", response["email"])
				assert.Equal(t, "1234567890", response["phone"])

				// Check status
				status, ok := response["status"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "applied", status["currentStage"])
				assert.NotEmpty(t, status["lastUpdated"])
			},
		},
		{
			name:          "application not found",
			applicationID: "nonexistent",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationByID", "nonexistent").Return(nil, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Empty(t, response)
			},
		},
		{
			name:          "repository error",
			applicationID: "app123",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationByID", "app123").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "database error\n", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockApplicationRepo)
			tt.mockSetup(mockRepo)

			service := services.ApplicationsService{AppRepo: mockRepo}
			handler := ApplicationHandler{ApplicationService: service}

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/applications/"+tt.applicationID, nil)
			w := httptest.NewRecorder()

			// Setup router with vars
			router := mux.NewRouter()
			router.HandleFunc("/applications/{id}", handler.GetApplicationByID).Methods(http.MethodGet)
			req = mux.SetURLVars(req, map[string]string{"id": tt.applicationID})

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestApplicationHandler_GetApplicationsByCandidateID(t *testing.T) {
	now := time.Now()
	mockApplications := []bson.M{
		{
			"application_id": "app1",
			"job_id":         123,
			"candidate_id":   456,
			"resume_url":     "https://example.com/resume1.pdf",
			"email":          "test1@example.com",
			"phone":          "1234567890",
			"status": bson.M{
				"current_stage": "applied",
				"last_updated":  now,
			},
		},
		{
			"application_id": "app2",
			"job_id":         789,
			"candidate_id":   456,
			"resume_url":     "https://example.com/resume2.pdf",
			"email":          "test1@example.com",
			"phone":          "1234567890",
			"status": bson.M{
				"current_stage": "reviewed",
				"last_updated":  now,
			},
		},
	}

	tests := []struct {
		name           string
		candidateID    string
		mockSetup      func(*MockApplicationRepo)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:        "successful retrieval",
			candidateID: "456",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByCandidateID", uint(456)).Return(mockApplications, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Len(t, response, 2)

				// Check first application
				assert.Equal(t, "app1", response[0]["applicationId"])
				assert.Equal(t, float64(123), response[0]["jobId"])
				assert.Equal(t, float64(456), response[0]["candidateId"])
				assert.Equal(t, "https://example.com/resume1.pdf", response[0]["resumeUrl"])
				assert.Equal(t, "test1@example.com", response[0]["email"])
				assert.Equal(t, "1234567890", response[0]["phone"])
				status1 := response[0]["status"].(map[string]interface{})
				assert.Equal(t, "applied", status1["currentStage"])
				assert.NotEmpty(t, status1["lastUpdated"])

				// Check second application
				assert.Equal(t, "app2", response[1]["applicationId"])
				assert.Equal(t, float64(789), response[1]["jobId"])
				assert.Equal(t, float64(456), response[1]["candidateId"])
				assert.Equal(t, "https://example.com/resume2.pdf", response[1]["resumeUrl"])
				assert.Equal(t, "test1@example.com", response[1]["email"])
				assert.Equal(t, "1234567890", response[1]["phone"])
				status2 := response[1]["status"].(map[string]interface{})
				assert.Equal(t, "reviewed", status2["currentStage"])
				assert.NotEmpty(t, status2["lastUpdated"])
			},
		},
		{
			name:        "no applications found",
			candidateID: "456",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByCandidateID", uint(456)).Return([]bson.M{}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response []map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Empty(t, response)
			},
		},
		{
			name:        "repository error",
			candidateID: "456",
			mockSetup: func(m *MockApplicationRepo) {
				m.On("GetApplicationsByCandidateID", uint(456)).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "Failed to get applications\n", w.Body.String())
			},
		},
		{
			name:           "invalid candidate ID",
			candidateID:    "invalid",
			mockSetup:      func(m *MockApplicationRepo) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, "Invalid candidate ID format\n", w.Body.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(MockApplicationRepo)
			tt.mockSetup(mockRepo)

			service := services.ApplicationsService{AppRepo: mockRepo}
			handler := ApplicationHandler{ApplicationService: service}

			// Create test request
			req := httptest.NewRequest(http.MethodGet, "/candidates/"+tt.candidateID+"/applications", nil)
			w := httptest.NewRecorder()

			// Setup router with vars
			router := mux.NewRouter()
			router.HandleFunc("/candidates/{id}/applications", handler.GetApplicationsByCandidateID).Methods(http.MethodGet)
			req = mux.SetURLVars(req, map[string]string{"id": tt.candidateID})

			// Execute
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)

			mockRepo.AssertExpectations(t)
		})
	}
}
