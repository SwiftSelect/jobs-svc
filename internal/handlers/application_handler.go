package handlers

import (
	"encoding/json"
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/services"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ApplicationHandler struct {
	ApplicationService services.ApplicationsService
	KafkaPublisher     *kafka.Publisher
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var application bson.M
	log.Println("Creating application..")
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	log.Printf("Received application data: %+v", application)

	if application["applicationId"] == nil || application["applicationId"] == "" {
		application["applicationId"] = primitive.NewObjectID().Hex()
	}

	// snake_case for mongo
	mongoDoc := convertToSnakeCase(application)
	log.Printf("Converted to snake_case: %+v", mongoDoc)

	// validate required fields
	if mongoDoc["job_id"] == nil || mongoDoc["candidate_id"] == nil {
		log.Printf("Missing required fields. job_id: %v, candidate_id: %v", mongoDoc["job_id"], mongoDoc["candidate_id"])
		http.Error(w, "JobID and CandidateID are required", http.StatusBadRequest)
		return
	}

	// validate status struct
	if mongoDoc["status"] == nil {
		mongoDoc["status"] = bson.M{
			"current_stage": "Applied",
			"last_updated":  time.Now(),
		}
	} else if status, ok := mongoDoc["status"].(bson.M); ok {
		if status["current_stage"] == nil {
			status["current_stage"] = "Applied"
		}
		if status["last_updated"] == nil {
			status["last_updated"] = time.Now()
		}
	}

	log.Printf("Final document to be inserted: %+v", mongoDoc)
	if err := h.ApplicationService.CreateApplication(mongoDoc); err != nil {
		log.Printf("Error creating application: %v", err)
		if err.Error() == "candidate has already applied for this job" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Failed to create application", http.StatusInternalServerError)
		}
		return
	}

	// camelCase for res
	response := convertToCamelCase(mongoDoc)
	log.Printf("Converted to camelCase for Kafka: %+v", response)

	// publish application to kafka
	if err := h.KafkaPublisher.PublishApplication(response); err != nil {
		log.Printf("Failed to publish application to Kafka: %v", err)
		// Continue with the response even if Kafka publish fails
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ApplicationHandler) GetApplicationsByJobID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobIDStr := vars["id"]
	jobID, err := strconv.ParseUint(jobIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid job ID format", http.StatusBadRequest)
		return
	}

	applications, err := h.ApplicationService.GetApplicationsByJobID(uint(jobID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var responses []bson.M
	for _, app := range applications {
		responses = append(responses, convertToCamelCase(app))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ApplicationHandler) GetApplicationByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["id"]
	application, err := h.ApplicationService.GetApplicationByID(applicationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to camelCase for response
	response := convertToCamelCase(application)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *ApplicationHandler) GetApplicationByCandidateID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	candidateIDStr := vars["id"]
	candidateID, err := strconv.ParseUint(candidateIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid candidate ID format", http.StatusBadRequest)
		return
	}

	applications, err := h.ApplicationService.GetApplicationsByCandidateID(uint(candidateID))
	if err != nil {
		log.Printf("Error getting applications: %v", err)
		http.Error(w, "Failed to get applications", http.StatusInternalServerError)
		return
	}

	// Convert to camelCase for response
	var responses []bson.M
	for _, app := range applications {
		responses = append(responses, convertToCamelCase(app))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ApplicationHandler) GetApplicationsByCandidateID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	candidateIDStr := vars["id"]
	candidateID, err := strconv.ParseUint(candidateIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid candidate ID format", http.StatusBadRequest)
		return
	}

	applications, err := h.ApplicationService.GetApplicationsByCandidateID(uint(candidateID))
	if err != nil {
		log.Printf("Error getting applications: %v", err)
		http.Error(w, "Failed to get applications", http.StatusInternalServerError)
		return
	}

	// Convert to camelCase for response
	var responses []bson.M
	for _, app := range applications {
		responses = append(responses, convertToCamelCase(app))
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func convertToSnakeCase(doc bson.M) bson.M {
	result := make(bson.M)

	for k, v := range doc {
		if nested, ok := v.(map[string]any); ok {
			result[toSnakeCase(k)] = convertToSnakeCase(nested)
			continue
		}

		if arr, ok := v.([]any); ok {
			var newArr []any
			for _, item := range arr {
				if nested, ok := item.(map[string]any); ok {
					newArr = append(newArr, convertToSnakeCase(nested))
				} else {
					newArr = append(newArr, item)
				}
			}
			result[toSnakeCase(k)] = newArr
			continue
		}

		result[toSnakeCase(k)] = v
	}

	return result
}

func convertToCamelCase(doc bson.M) bson.M {
	result := make(bson.M)

	for k, v := range doc {
		switch val := v.(type) {
		case bson.M:
			result[toCamelCase(k)] = convertToCamelCase(val)
		case []interface{}:
			var newArr []interface{}
			for _, item := range val {
				if nested, ok := item.(bson.M); ok {
					newArr = append(newArr, convertToCamelCase(nested))
				} else {
					newArr = append(newArr, item)
				}
			}
			result[toCamelCase(k)] = newArr
		default:
			// Special handling for specific fields
			switch k {
			case "candidate_id":
				result["candidateId"] = v
			case "job_id":
				result["jobId"] = v
			case "application_id":
				result["applicationId"] = v
			case "resume_url":
				result["resumeUrl"] = v
			default:
				result[toCamelCase(k)] = v
			}
		}
	}

	return result
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toCamelCase(s string) string {
	var result strings.Builder
	words := strings.Split(s, "_")
	for i, word := range words {
		if i == 0 {
			result.WriteString(strings.ToLower(word))
		} else {
			if len(word) > 0 {
				result.WriteString(strings.ToUpper(string(word[0])))
				if len(word) > 1 {
					result.WriteString(strings.ToLower(word[1:]))
				}
			}
		}
	}
	return result.String()
}
