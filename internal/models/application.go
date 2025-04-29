package models

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

// application model for MongoDB
type Application struct {
	ApplicationID string `bson:"application_id" json:"applicationId"`
	CandidateID   string `bson:"candidate_id" json:"candidateId"`
	JobID         string `bson:"job_id" json:"jobId"`
	ResumeURL     string `bson:"resumeUrl" json:"resumeUrl"`
	Status        Status `bson:"status" json:"status"`
	Email         string `bson:"email" json:"email"`
	Phone         string `bson:"phone" json:"phone"`
}

// type Resume struct {
// 	Text       string       `bson:"text" json:"text"`
// 	Skills     []string     `bson:"skills" json:"skills"`
// 	Experience []Experience `bson:"experience" json:"experience"`
// }

// type Experience struct {
// 	Company  string `bson:"company" json:"company"`
// 	Role     string `bson:"role" json:"role"`
// 	Duration string `bson:"duration" json:"duration"`
// }

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
