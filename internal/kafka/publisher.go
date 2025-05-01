package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"jobs-svc/internal/models"
	"go.mongodb.org/mongo-driver/bson"
)

type JobKafkaMessage struct {
	JobID       uint     `json:"jobId"`
	Title       string   `json:"title"`
	Overview    string   `json:"overview"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
	Experience  string   `json:"experience"`
}

type ApplicationKafkaMessage struct {
	ApplicationID string `json:"applicationId"`
	JobID        string `json:"jobId"`
	ResumeURL    string `json:"resumeUrl"`
}

type Publisher struct {
	producer sarama.SyncProducer
}

func NewPublisher(brokers []string) (*Publisher, error) {
	log.Printf("Initializing Kafka publisher with brokers: %v", brokers)
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	
	// Add debug logging
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)
	
	// Additional configuration for better debugging
	config.Producer.Return.Errors = true
	config.Producer.Compression = sarama.CompressionNone
	config.Producer.Retry.Backoff = 500 * time.Millisecond
	config.Producer.MaxMessageBytes = 1000000
	config.Version = sarama.V2_8_1_0  // Specify Kafka version explicitly

	log.Printf("Attempting to connect to Kafka brokers with config: %+v", config)
	
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	log.Printf("Successfully connected to Kafka brokers: %v", brokers)
	return &Publisher{
		producer: producer,
	}, nil
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}

func (p *Publisher) PublishJob(job interface{}) error {
	log.Printf("Attempting to publish job to Kafka")
	
	// Type assert to models.Job
	jobData, ok := job.(*models.Job)
	if !ok {
		return fmt.Errorf("invalid job data type: expected *models.Job, got %T", job)
	}

	// Split skills string into array
	skills := strings.Split(jobData.Skills, ",")
	for i, skill := range skills {
		skills[i] = strings.TrimSpace(skill)
	}

	kafkaMessage := JobKafkaMessage{
		JobID:       jobData.ID,
		Title:       jobData.Title,
		Overview:    jobData.Overview,
		Description: jobData.Description,
		Skills:      skills,
		Experience:  jobData.Experience,
	}

	log.Printf("Created Kafka message: %+v", kafkaMessage)

	jobBytes, err := json.Marshal(kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "jobs_topic",
		Value: sarama.ByteEncoder(jobBytes),
	}

	log.Printf("Sending job message to Kafka topic: jobs_topic")
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send job message: %v", err)
	}

	log.Printf("Job published successfully to partition %d at offset %d", partition, offset)
	return nil
}

func (p *Publisher) PublishApplication(application interface{}) error {
	log.Printf("Attempting to publish application to Kafka")
	
	// Type assert to bson.M
	appData, ok := application.(bson.M)
	if !ok {
		return fmt.Errorf("invalid application data type: expected bson.M, got %T", application)
	}

	kafkaMessage := ApplicationKafkaMessage{
		ApplicationID: appData["applicationId"].(string),
		JobID:        appData["jobId"].(string),
		ResumeURL:    appData["resumeUrl"].(string),
	}

	log.Printf("Created Kafka message: %+v", kafkaMessage)

	appBytes, err := json.Marshal(kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal application: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: "candidate_topic",
		Value: sarama.ByteEncoder(appBytes),
	}

	log.Printf("Sending application message to Kafka topic: candidate_topic")
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send application message: %v", err)
	}

	log.Printf("Application published successfully to partition %d at offset %d", partition, offset)
	return nil
} 