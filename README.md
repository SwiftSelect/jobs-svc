# Jobs Service

A microservice for managing job postings and applications, part of larger SwiftSelect microservices architecture. This service handles job creation, management, and application processing, with Kafka integration for event-driven communication.

## Features

- Job Management
  - Create, read, update, and delete job postings
  - Job search and filtering
  - Job status tracking

- Application Management
  - Submit and track job applications
  - Application status updates
  - Candidate information management

- Event-Driven Architecture
  - Kafka integration for job and application events
  - Real-time updates through message queues

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose
- PostgreSQL
- MongoDB
- Kafka
- SwiftSelect auth service running locally

## Setup

1. Clone the repository:
```bash
git clone <repository-url>
cd jobs-svc
```

2. Create a `.env` file with the following variables:
```env
POSTGRES_USER=your_user
POSTGRES_PASSWORD=your_password
POSTGRES_DB=jobs_db
KAFKA_BROKERS=localhost:9092
```

3. Start the required services using Docker Compose:
```bash
docker-compose up -d
```

4. Create Kafka topics (Optional step, kafka will autocreate topics if they do not exist):
```bash
# Create jobs topic
docker exec -it jobs-svc-kafka-1 kafka-topics --create --topic jobs_topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1

# Create applications topic
docker exec -it jobs-svc-kafka-1 kafka-topics --create --topic candidate_topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1
```

## Running the Application

1. Start the service:
```bash
cd cmd
go run main.go
```

The service will start on `http://localhost:8080`

## API Endpoints

### Jobs

- `POST /jobs` - Create a new job posting
  ```json
  {
    "title": "Software Engineer",
    "overview": "Looking for a skilled software engineer",
    "description": "Detailed job description...",
    "company": "Tech Corp",
    "skills": "React, Node.js, TypeScript",
    "experience": "5+ years",
    "location": "Remote",
    "salaryRange": "$120,000 - $160,000",
    "benefitsAndPerks": "Health, Dental, Vision, 401k"
  }
  ```

- `GET /jobs` - List all jobs
- `GET /jobs/{id}` - Get job by ID
- `PUT /jobs/{id}` - Update job
- `DELETE /jobs/{id}` - Delete job

### Applications

- `POST /applications` - Submit a new application
  ```json
  {
    "candidateId": "123",
    "jobId": "456",
    "resumeUrl": "http://example.com/resume.pdf",
    "skills": ["React", "Node.js", "AWS"],
    "yearsOfExperience": 3
  }
  ```

- `GET /applications/job/{jobId}` - Get applications for a specific job

## Kafka Integration

The service publishes events to two Kafka topics:

1. `jobs_topic` - Job-related events
   ```json
   {
     "jobId": 123,
     "title": "Software Engineer",
     "overview": "Job overview...",
     "description": "Detailed description...",
     "company": "Tech Corp",
     "skills": ["React", "Node.js", "TypeScript"],
     "experience": "5+ years",
     "location": "Remote",
     "status": "Open",
     "postedDate": "2024-04-29T00:00:00Z",
     "salaryRange": "$120,000 - $160,000"
   }
   ```

2. `candidate_topic` - Application-related events
   ```json
   {
     "applicationId": "abc123",
     "jobId": "456",
     "resumeUrl": "http://example.com/resume.pdf"
   }
   ```

To monitor Kafka messages:
```bash
# Monitor jobs topic
docker exec -it jobs-svc-kafka-1 kafka-console-consumer --bootstrap-server localhost:9092 --topic jobs_topic --from-beginning

# Monitor applications topic
docker exec -it jobs-svc-kafka-1 kafka-console-consumer --bootstrap-server localhost:9092 --topic candidate_topic --from-beginning
```

## Development

### Project Structure

```
jobs-svc/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── kafka/            # Kafka integration
├── docker-compose.yml    # Docker configuration
└── README.md            # This file
```