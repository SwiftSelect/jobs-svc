package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"jobs-svc/internal/models"

	"github.com/IBM/sarama"
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
	JobID         string `json:"jobId"`
	ResumeURL     string `json:"resumeUrl"`
	CandidateID   string `json:"candidateId"`
}

type Config struct {
	Brokers          []string
	SecurityProtocol string
	SASLMechanism    string
	ConfluentKey     string
	ConfluentSecret  string
}

type Publisher struct {
	producer sarama.SyncProducer
}

var (
	KAFKA_CANDIDATE_TOPIC = getEnvOrDefault("KAFKA_CANDIDATE_TOPIC", "candidate_topic")
	KAFKA_JOB_TOPIC       = getEnvOrDefault("KAFKA_JOBS_TOPIC", "jobs_topic")
)

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("Warning: %s not set, using default value: %s", key, defaultValue)
		return defaultValue
	}
	return value
}

func NewPublisher(config *Config) (*Publisher, error) {
	if config == nil {
		return nil, fmt.Errorf("kafka config cannot be nil")
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("at least one broker must be specified")
	}

	log.Printf("Initializing Kafka publisher with config: %+v", config)

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Retry.Max = 5

	// Configure security if provided
	if config.SecurityProtocol != "" {
		if config.ConfluentKey == "" || config.ConfluentSecret == "" || config.SASLMechanism == "" {
			return nil, fmt.Errorf("username, password, and SASL mechanism are required when using SASL authentication")
		}

		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.User = config.ConfluentKey
		saramaConfig.Net.SASL.Password = config.ConfluentSecret
		saramaConfig.Net.SASL.Mechanism = sarama.SASLMechanism(config.SASLMechanism)
		saramaConfig.Net.TLS.Enable = config.SecurityProtocol == "SASL_SSL"
	}

	// Add debug logging
	sarama.Logger = log.New(os.Stdout, "[Sarama] ", log.LstdFlags)

	// Additional configuration for better debugging
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Compression = sarama.CompressionNone
	saramaConfig.Producer.Retry.Backoff = 500 * time.Millisecond
	saramaConfig.Producer.MaxMessageBytes = 1000000
	saramaConfig.Version = sarama.V2_8_1_0 // Specify Kafka version explicitly

	log.Printf("Attempting to connect to Kafka brokers with config: %+v", saramaConfig)

	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	log.Printf("Successfully connected to Kafka brokers: %v", config.Brokers)
	return &Publisher{
		producer: producer,
	}, nil
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}

func (p *Publisher) PublishJob(job *models.Job) error {
	log.Printf("Attempting to publish job to Kafka topic: %s", KAFKA_JOB_TOPIC)

	// Split skills string into array
	skills := strings.Split(job.Skills, ",")
	for i, skill := range skills {
		skills[i] = strings.TrimSpace(skill)
	}

	kafkaMessage := JobKafkaMessage{
		JobID:       job.ID,
		Title:       job.Title,
		Overview:    job.Overview,
		Description: job.Description,
		Skills:      skills,
		Experience:  job.Experience,
	}

	log.Printf("Created Kafka message: %+v", kafkaMessage)

	jobBytes, err := json.Marshal(kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: KAFKA_JOB_TOPIC,
		Value: sarama.ByteEncoder(jobBytes),
	}

	log.Printf("Sending job message to Kafka topic: %s", KAFKA_JOB_TOPIC)
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send job message: %v", err)
	}

	log.Printf("Job published successfully to partition %d at offset %d", partition, offset)
	return nil
}

func (p *Publisher) PublishApplication(application bson.M) error {
	log.Printf("Publishing application to Kafka topic: %s", KAFKA_CANDIDATE_TOPIC)

	// Safely extract fields with type assertions, handling both camelCase and snake_case
	applicationID, _ := application["applicationId"].(string)
	if applicationID == "" {
		applicationID, _ = application["application_id"].(string)
	}

	jobID, _ := application["jobId"].(string)
	if jobID == "" {
		jobID, _ = application["job_id"].(string)
	}

	resumeURL, _ := application["resumeUrl"].(string)
	if resumeURL == "" {
		resumeURL, _ = application["resume_url"].(string)
	}

	// Handle candidateId as either string or number
	var candidateID string
	if strID, ok := application["candidateId"].(string); ok {
		candidateID = strID
	} else if numID, ok := application["candidateId"].(int); ok {
		candidateID = strconv.Itoa(numID)
	} else if numID, ok := application["candidateId"].(int64); ok {
		candidateID = strconv.FormatInt(numID, 10)
	} else if numID, ok := application["candidateId"].(float64); ok {
		candidateID = strconv.FormatInt(int64(numID), 10)
	} else if strID, ok := application["candidate_id"].(string); ok {
		candidateID = strID
	} else if numID, ok := application["candidate_id"].(int); ok {
		candidateID = strconv.Itoa(numID)
	} else if numID, ok := application["candidate_id"].(int64); ok {
		candidateID = strconv.FormatInt(numID, 10)
	} else if numID, ok := application["candidate_id"].(float64); ok {
		candidateID = strconv.FormatInt(int64(numID), 10)
	}

	// Validate required fields
	if applicationID == "" {
		return fmt.Errorf("applicationId is required")
	}
	if jobID == "" {
		return fmt.Errorf("jobId is required")
	}
	if candidateID == "" {
		return fmt.Errorf("candidateId is required")
	}

	kafkaMessage := ApplicationKafkaMessage{
		ApplicationID: applicationID,
		JobID:         jobID,
		ResumeURL:     resumeURL,
		CandidateID:   candidateID,
	}

	appBytes, err := json.Marshal(kafkaMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal application: %v", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: KAFKA_CANDIDATE_TOPIC,
		Value: sarama.ByteEncoder(appBytes),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send application message: %v", err)
	}

	log.Printf("Application published successfully to partition %d at offset %d", partition, offset)
	return nil
}
