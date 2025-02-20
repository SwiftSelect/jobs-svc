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
	ApplicationID string `bson:"application_id"`
	CandidateID   string `bson:"candidate_id"`
	JobID         string `bson:"job_id"`
	Resume        Resume `bson:"resume"`
	Status        Status `bson:"status"`
}

type Resume struct {
	Text       string       `bson:"text"`
	Skills     []string     `bson:"skills"`
	Experience []Experience `bson:"experience"`
}

type Experience struct {
	Company  string `bson:"company"`
	Role     string `bson:"role"`
	Duration string `bson:"duration"`
}

type Status struct {
	CurrentStage string    `bson:"current_stage"`
	LastUpdated  time.Time `bson:"last_updated"`
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
