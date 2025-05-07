package repos

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ApplicationRepo struct {
	Collection *mongo.Collection
}

// CreateUniqueIndex creates a unique compound index on candidate_id and job_id
// for checking if candidate trying to apply to same job again
func (repo *ApplicationRepo) CreateUniqueIndex() error {
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

func (repo *ApplicationRepo) CreateApplication(application bson.M) error {
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

func (repo *ApplicationRepo) GetApplicationsByJobID(jobID uint) ([]bson.M, error) {
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

func (repo *ApplicationRepo) GetApplicationByID(applicationID string) (bson.M, error) {
	var application bson.M
	filter := bson.M{"application_id": applicationID}
	err := repo.Collection.FindOne(context.TODO(), filter).Decode(&application)
	if err != nil {
		return nil, err
	}

	return application, nil
}

func (repo *ApplicationRepo) GetApplicationsByCandidateID(candidateID uint) ([]bson.M, error) {
	log.Printf("Searching for applications with candidate_id: %d", candidateID)

	filter := bson.M{"candidate_id": candidateID}

	cursor, err := repo.Collection.Find(context.TODO(), filter)
	if err != nil {
		log.Printf("Error finding applications: %v", err)
		return nil, fmt.Errorf("failed to find applications: %v", err)
	}
	defer cursor.Close(context.TODO())

	var applications []bson.M
	if err = cursor.All(context.TODO(), &applications); err != nil {
		log.Printf("Error decoding applications: %v", err)
		return nil, fmt.Errorf("failed to decode applications: %v", err)
	}

	log.Printf("Found %d applications for candidate_id: %d", len(applications), candidateID)
	return applications, nil
}

func (repo *ApplicationRepo) GetApplicationByCandidateID(candidateID uint) (bson.M, error) {
	log.Printf("Searching for application with candidate_id: %d", candidateID)

	filter := bson.M{"candidate_id": candidateID}

	var application bson.M
	err := repo.Collection.FindOne(context.TODO(), filter).Decode(&application)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		log.Printf("Error finding application: %v", err)
		return nil, fmt.Errorf("failed to find application: %v", err)
	}

	return application, nil
}
