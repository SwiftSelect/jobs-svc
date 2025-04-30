package main

import (
	"jobs-svc/middleware"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"jobs-svc/internal/handlers"
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
	"jobs-svc/internal/services"
)

func LoadEnv() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	LoadEnv()

	models.ConnectPostgres()
	models.ConnectMongo()

	log.Println("Starting connection to databases...")
	jobsDB := models.PostgresDB
	appsDB := models.MongoDB

	err := jobsDB.AutoMigrate(&models.Job{})
	if err != nil {
		return
	}
	log.Println("Jobs database migration completed.")

	// kafka init publisher
	kafkaBrokers := []string{"localhost:9092"}
	if os.Getenv("KAFKA_BROKERS") != "" {
		kafkaBrokers = []string{os.Getenv("KAFKA_BROKERS")}
	}

	kafkaPublisher, err := kafka.NewPublisher(kafkaBrokers)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka publisher: %v", err)
	}
	defer kafkaPublisher.Close()
	log.Println("Kafka publisher initialized successfully")

	// init repositories, services,handlers
	jobRepo := repos.JobRepo{DB: jobsDB}
	applicationRepo := repos.AppRepo{Collection: appsDB.Collection("applications")}

	// Create unique index for applications
	if err := applicationRepo.CreateUniqueIndex(); err != nil {
		log.Fatal("Failed to create unique index for applications:", err)
	}
	log.Println("Created unique index on candidate_id and job_id")

	jobService := services.JobService{JobRepo: jobRepo}
	applicationService := services.ApplicationsService{AppRepo: applicationRepo}

	jobHandler := handlers.JobHandler{
		JobService:     jobService,
		KafkaPublisher: kafkaPublisher,
	}
	applicationHandler := handlers.ApplicationHandler{
		ApplicationService: applicationService,
		KafkaPublisher:    kafkaPublisher,
	}

	router := mux.NewRouter()
	//router.Use(middleware.CORSMiddleware)

	// job related routes
	//router.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	router.Handle("/jobs", middleware.AuthMiddleware("create_job")(http.HandlerFunc(jobHandler.CreateJob))).Methods("POST")
	router.HandleFunc("/jobs", jobHandler.GetJobs).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobHandler.GetJobByID).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	router.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")

	// app related routes
	router.HandleFunc("/applications", applicationHandler.CreateApplication).Methods("POST")
	router.HandleFunc("/applications/job/{id}", applicationHandler.GetApplicationsByJobID).Methods("GET")

	corsRouter := middleware.CORSMiddleware(router)

	log.Println("Server started on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", corsRouter))
}
