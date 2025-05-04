package main

import (
	"jobs-svc/middleware"
	"log"
	"net/http"
	"os"

	"jobs-svc/internal/handlers"
	"jobs-svc/internal/kafka"
	"jobs-svc/internal/models"
	"jobs-svc/internal/repos"
	"jobs-svc/internal/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
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
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		log.Fatal("KAFKA_BROKERS environment variable is not set")
	}

	kafkaConfig := &kafka.Config{
		Brokers:          []string{kafkaBrokers},
		SecurityProtocol: os.Getenv("KAFKA_SECURITY_PROTOCOL"),
		SASLMechanism:    os.Getenv("KAFKA_SASL_MECHANISM"),
		ConfluentKey:     os.Getenv("CONFLUENT_KEY"),
		ConfluentSecret:  os.Getenv("CONFLUENT_SECRET"),
	}

	// Validate required Kafka credentials
	if kafkaConfig.SecurityProtocol != "" {
		if kafkaConfig.ConfluentKey == "" {
			log.Fatal("KAFKA_USERNAME is required when using SASL authentication")
		}
		if kafkaConfig.ConfluentSecret == "" {
			log.Fatal("KAFKA_PASSWORD is required when using SASL authentication")
		}
		if kafkaConfig.SASLMechanism == "" {
			log.Fatal("KAFKA_SASL_MECHANISM is required when using SASL authentication")
		}
	}

	kafkaPublisher, err := kafka.NewPublisher(kafkaConfig)
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
		KafkaPublisher:     kafkaPublisher,
	}

	router := mux.NewRouter()
	//router.Use(middleware.CORSMiddleware)

	// job related routes
	//router.HandleFunc("/jobs", jobHandler.CreateJob).Methods("POST")
	router.Handle("/jobs", middleware.AuthMiddleware("create_job")(http.HandlerFunc(jobHandler.CreateJob))).Methods("POST")
	router.HandleFunc("/jobs", jobHandler.GetJobs).Methods("GET")
	router.HandleFunc("/jobs/summary", jobHandler.GetJobsByIDs).Methods("POST")
	router.HandleFunc("/jobs/{id}", jobHandler.GetJobByID).Methods("GET")
	router.HandleFunc("/jobs/recruiter/{id}", jobHandler.GetJobsByRecruiterID).Methods("GET")
	router.HandleFunc("/jobs/{id}", jobHandler.UpdateJob).Methods("PUT")
	router.HandleFunc("/jobs/{id}", jobHandler.DeleteJob).Methods("DELETE")

	// app related routes
	router.HandleFunc("/applications", applicationHandler.CreateApplication).Methods("POST")
	router.HandleFunc("/applications/job/{id}", applicationHandler.GetApplicationsByJobID).Methods("GET")
	router.HandleFunc("/applications/{id}", applicationHandler.GetApplicationByID).Methods("GET")
	router.HandleFunc("/applications/candidate/{id}", applicationHandler.GetApplicationByCandidateID).Methods("GET")

	corsRouter := middleware.CORSMiddleware(router)

	log.Println("Server started on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", corsRouter))
}
