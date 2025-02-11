package handlers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"jobs-svc/internal/models"
	"jobs-svc/internal/services"
	"net/http"
	"strconv"
)

type ApplicationHandler struct {
	ApplicationService services.ApplicationsService
}

func (h *ApplicationHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var application models.Application
	if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.ApplicationService.CreateApplication(&application); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(application)
}

func (h *ApplicationHandler) GetApplicationsByJobID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, _ := strconv.Atoi(vars["id"])
	applications, err := h.ApplicationService.GetApplicationsByJobID(uint(jobID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(applications)
}
