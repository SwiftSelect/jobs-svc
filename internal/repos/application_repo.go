package repos

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppRepo struct {
	Collection *mongo.Collection
}

// CreateUniqueIndex creates a unique compound index on candidate_id and job_id
// for checking if candidate trying to apply to same job again
func (repo *AppRepo) CreateUniqueIndex() error {
	// Create a compound index
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "candidate_id", Value: 1},
			{Key: "job_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}

	// Create the index
	_, err := repo.Collection.Indexes().CreateOne(context.TODO(), indexModel)
	if err != nil {
		return err
	}

	return nil
}

func (repo *AppRepo) CreateApplication(application bson.M) error {
	filter := bson.M{
		"candidate_id": application["candidate_id"],
		"job_id":       application["job_id"],
	}

	var existingApp bson.M
	err := repo.Collection.FindOne(context.TODO(), filter).Decode(&existingApp)
	if err == nil {
		return errors.New("candidate has already applied for this job")
	} else if err != mongo.ErrNoDocuments {
		return err
	}

	if application["application_id"] == nil {
		application["application_id"] = primitive.NewObjectID().Hex()
	}

	// Validate required fields
	if application["job_id"] == nil || application["candidate_id"] == nil {
		return errors.New("JobID and CandidateID are required")
	}

	// Set status if not provided
	if application["status"] == nil {
		application["status"] = bson.M{
			"current_stage": "applied",
			"last_updated":  time.Now(),
		}
	} else {
		// Update last_updated if status exists
		if status, ok := application["status"].(bson.M); ok {
			status["last_updated"] = time.Now()
		}
	}

	log.Println("Inserting application:", application)
	_, err = repo.Collection.InsertOne(context.TODO(), application)
	log.Println("Error inserting application:", err)
	return err
}

func (repo *AppRepo) GetApplicationsByJobID(jobID string) ([]bson.M, error) {
	var applications []bson.M
	filter := bson.M{"job_id": jobID}
	cursor, err := repo.Collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	if err = cursor.All(context.TODO(), &applications); err != nil {
		return nil, err
	}

	return applications, nil
}
