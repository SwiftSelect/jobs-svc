package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"net/http"
)

type ApplicationHandler struct {
	ApplicationService services.ApplicationsService
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var application models.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if application.JobID == "" || application.CandidateID == "" {
		http.Error(w, "JobID and CandidateID are required", http.StatusBadRequest)
		return
	}

	if err := h.ApplicationService.CreateApplication(&application); err != nil {
		http.Error(w, "Failed to create application", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}

func (h *ApplicationHandler) GetApplicationsByJobID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"] // MongoDB uses string IDs
	applications, err := h.ApplicationService.GetApplicationsByJobID(jobID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(applications)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
