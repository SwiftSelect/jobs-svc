package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"log"
	"net/http"
	"strconv"
)

type JobHandler struct {
	JobService     services.JobService
	KafkaPublisher *kafka.Publisher
}

func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.JobService.GetJobs()
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(jobs); err != nil {
		http.Error(w, "Failed to encode jobs response", http.StatusInternalServerError)
	}
}

func (h *JobHandler) GetJobByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists || idStr == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	job, err := h.JobService.GetJobByID(uint(jobID))
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// calculate and add days posted ago field
	response := map[string]interface{}{
		"id":               job.ID,
		"title":            job.Title,
		"overview":         job.Overview,
		"description":      job.Description,
		"company":          job.Company,
		"skills":           job.Skills,
		"experience":       job.Experience,
		"location":         job.Location,
		"status":           job.Status,
		"postedDate":       job.PostedDate,
		"salaryRange":      job.SalaryRange,
		"recruiterId":      job.RecruiterId,
		"daysPostedAgo":    job.DaysPostedAgo(),
		"benefitsAndPerks": job.BenefitsAndPerks,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode job response", http.StatusInternalServerError)
	}
}

func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	if err := h.JobService.CreateJob(&job); err != nil {
		http.Error(w, "Failed to create job", http.StatusInternalServerError)
		return
	}

	// publish jobs to kafka
	if err := h.KafkaPublisher.PublishJob(&job); err != nil {
		log.Printf("Failed to publish job to Kafka: %v", err)
		// Continue with the response even if Kafka publish fails
	}

	// Respond with created job
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(job); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists || idStr == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	var job models.Job
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	job.ID = uint(jobID)
	if err := h.JobService.UpdateJob(&job); err != nil {
		http.Error(w, "Failed to update job", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(job); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, exists := vars["id"]
	if !exists || idStr == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid Job ID", http.StatusBadRequest)
		return
	}

	if err := h.JobService.DeleteJob(uint(jobID)); err != nil {
		http.Error(w, "Failed to delete job", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Job deleted successfully"))
}
