package handlers

import (
	"encoding/json"
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"log"
	"net/http"
	"strconv"

	"jobs-svc/internal/clients"

	"github.com/gorilla/mux"
)

type JobHandler struct {
	JobService     services.JobService
	KafkaPublisher kafka.PublisherInterface
}

type JobResponse struct {
	models.Job
	DaysPostedAgo int `json:"daysPostedAgo"`
}

func (h *JobHandler) GetJobs(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.JobService.GetJobs()
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	// Initialize response as an empty array
	response := make([]models.Job, 0)
	if jobs != nil {
		response = *jobs
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode jobs response", http.StatusInternalServerError)
	}
}

func (h *JobHandler) GetJobsByRecruiterID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	recruiterID := vars["id"]
	recruiterIDInt, err := strconv.Atoi(recruiterID)
	if err != nil {
		http.Error(w, "Invalid recruiter ID", http.StatusBadRequest)
		return
	}

	jobs, err := h.JobService.GetJobsByRecruiterID(uint(recruiterIDInt))
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	// Initialize response as an empty array
	response := make([]JobResponse, 0)

	// calculate and add days posted ago field
	if jobs != nil {
		response = make([]JobResponse, len(*jobs))
		for i, job := range *jobs {
			response[i] = JobResponse{
				Job:           job,
				DaysPostedAgo: job.DaysPostedAgo(),
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
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
		"companyId":        job.CompanyID,
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

	// Get user info from context (set by auth middleware)
	userInfo, ok := r.Context().Value("userInfo").(*clients.UserResponse)
	if !ok || userInfo == nil {
		http.Error(w, "User information not found", http.StatusUnauthorized)
		return
	}

	// Get company ID from user's organization
	companyID, err := clients.GetCompanyID(userInfo)
	if err != nil {
		http.Error(w, "Failed to get company ID: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set the company ID from the auth service response
	job.CompanyID = uint(companyID)

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

func (h *JobHandler) GetJobsByIDs(w http.ResponseWriter, r *http.Request) {
	var jobIDs []string
	if err := json.NewDecoder(r.Body).Decode(&jobIDs); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if len(jobIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]models.JobSummary{})
		// http.Error(w, "No job IDs provided", http.StatusBadRequest)
		return
	}

	// Convert string IDs to uint
	uintIDs := make([]uint, len(jobIDs))
	for i, idStr := range jobIDs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid job ID format", http.StatusBadRequest)
			return
		}
		uintIDs[i] = uint(id)
	}

	jobs, err := h.JobService.GetJobsByIDs(uintIDs)
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	// Initialize response as an empty array
	jobSummaries := make([]models.JobSummary, 0)

	// Convert to summary format
	if jobs != nil {
		jobSummaries = make([]models.JobSummary, len(*jobs))
		for i, job := range *jobs {
			jobSummaries[i] = models.JobSummary{
				ID:            job.ID,
				Title:         job.Title,
				Company:       job.Company,
				DaysPostedAgo: job.DaysPostedAgo(),
				Location:      job.Location,
				SalaryRange:   job.SalaryRange,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(jobSummaries); err != nil {
		http.Error(w, "Failed to encode jobs response", http.StatusInternalServerError)
	}
}
