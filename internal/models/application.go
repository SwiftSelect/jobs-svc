package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// application model for MongoDB
type Application struct {
	ApplicationID string `bson:"application_id" json:"applicationId"`
	CandidateID   uint   `bson:"candidate_id" json:"candidateId"`
	JobID         uint   `bson:"job_id" json:"jobId"`
	ResumeURL     string `bson:"resumeUrl" json:"resumeUrl"`
	Status        Status `bson:"status" json:"status"`
	Email         string `bson:"email" json:"email"`
	Phone         string `bson:"phone" json:"phone"`
}

type Status struct {
	CurrentStage string    `bson:"current_stage" json:"currentStage"`
	LastUpdated  time.Time `bson:"last_updated" json:"lastUpdated"`
}

var MongoDB *mongo.Database

func ConnectMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("Failed to create MongoDB client:", err)
	}

	MongoDB = client.Database(os.Getenv("MONGO_DB"))
	fmt.Println("Connected to MongoDB")
}
