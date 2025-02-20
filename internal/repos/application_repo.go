package repos

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"jobs-svc/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AppRepo struct {
	Collection *mongo.Collection
}

func (repo *AppRepo) CreateApplication(application *models.Application) error {
	if application.ApplicationID == "" {
		application.ApplicationID = primitive.NewObjectID().Hex() // MongoDB ObjectID as string
	}

	if application.JobID == "" || application.CandidateID == "" {
		return errors.New("JobID and CandidateID are required")
	}

	application.Status.LastUpdated = time.Now()

	_, err := repo.Collection.InsertOne(context.TODO(), application)
	return err
}

func (repo *AppRepo) GetApplicationsByJobID(jobID string) ([]models.Application, error) {
	var applications []models.Application
	filter := bson.M{"job_id": jobID}
	cursor, err := repo.Collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var application models.Application
		if err := cursor.Decode(&application); err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}

	return applications, nil
}
