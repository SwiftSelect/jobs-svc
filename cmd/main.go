package main

import (
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"jobs-svc/internal/handlers"
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
	"jobs-svc/internal/services"
	"log"
	"net/http"
)

func main() {

	log.Println("Starting connection to jobs database...")
	//TODO: move to config file and use env
	conn := "postgres://postgres:password@localhost:5432/jobs_db?sslmode=disable"
	db, err := sql.Open("postgres", conn)
	defer db.Close()
	if err != nil {
		log.Fatal("failed to connect to the database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping the database: %v", err)
	}

	log.Println("Database connection established successfully")

	// connect with orm
	ormdb, err := gorm.Open("postgres", db)
	if err != nil {
		log.Fatalf("failed to initialize orm: %v", err)
	}
	defer ormdb.Close()
	log.Printf("ORM connected to database")

	ormdb.AutoMigrate(&models.Job{}, &models.Application{})

	jobRepo := repos.JobRepo{DB: ormdb}
	applicationRepo := repos.AppRepo{DB: ormdb}

	jobService := services.JobService{JobRepo: jobRepo}
	applicationService := services.ApplicationsService{AppRepo: applicationRepo}

	jobHandler := handlers.JobHandler{JobService: jobService}
	applicationHandler := handlers.ApplicationHandler{ApplicationService: applicationService}

	router := mux.NewRouter()

	router.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	router.HandleFunc("/jobs", jobHandler.GetJobs).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobHandler.GetJobByID).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	router.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")

	router.HandleFunc("/applications", applicationHandler.CreateApplication).Methods("POST")
	router.HandleFunc("/applications/job/{id}", applicationHandler.GetApplicationsByJobID).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}
