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
     "candidateId": "33",
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

## Testing

The service includes comprehensive unit tests for both the service layer and HTTP handlers. The tests use mock implementations of dependencies to ensure isolated testing.

### Running Tests

To run all tests:
```bash
go test ./...
```

To run tests with coverage:
```bash
go test ./... -cover
```

To run tests for a specific package:
```bash
go test ./internal/handlers -v
```

### Test Structure

The test suite is organized into the following components:

1. **Mock Implementations**
   - `MockJobRepo`: Implements `JobRepoInterface` for testing job-related operations
   - `MockApplicationRepo`: Implements `ApplicationRepoInterface` for testing application-related operations
   - `MockKafkaPublisher`: Implements `PublisherInterface` for testing Kafka integration

2. **Service Layer Tests**
   - Job Service (`internal/services/tests/job_service_test.go`)
     - Tests for all CRUD operations
     - Tests for job retrieval by recruiter ID
     - Tests for batch job retrieval by IDs
   
   - Application Service (`internal/services/tests/application_service_test.go`)
     - Tests for application creation
     - Tests for application retrieval by various criteria
     - Tests for duplicate application prevention

3. **Handler Layer Tests**
   - Job Handler (`internal/handlers/tests/job_handler_test.go`)
     - Tests for all job-related HTTP endpoints
     - Tests for request validation
     - Tests for response formatting
     - Tests for Kafka event publishing

   - Application Handler (`internal/handlers/tests/application_handler_test.go`)
     - Tests for application submission
       - Successful application creation
       - Handling missing required fields
       - Preventing duplicate applications
       - Proper error responses
     - Tests for application retrieval
       - Getting applications by job ID
       - Getting applications by candidate ID
       - Getting single application by ID
       - Proper handling of not-found cases
     - Tests for response formatting
       - Proper status code handling
       - Consistent error message format
       - Proper timestamp handling

### Test Coverage

Key areas covered by tests include:

1. **Input Validation**
   - Required field validation
   - Data type validation
   - Duplicate application prevention

2. **Error Handling**
   - Database errors
   - Not found scenarios
   - Invalid input data
   - Duplicate applications

3. **Response Formatting**
   - Consistent JSON structure
   - Proper status codes
   - Error message formatting
   - Timestamp formatting

4. **Business Logic**
   - Application status management
   - Application retrieval logic
   - Job-application relationships

### Writing Tests

When adding new features, follow these testing guidelines:

1. **Use Mocks**: Create mock implementations for external dependencies
2. **Test Edge Cases**: Include tests for error conditions and edge cases
3. **Consistent Format**: Follow the existing test structure and naming conventions
4. **Clear Setup**: Each test should have clear setup and teardown phases
5. **Meaningful Assertions**: Make specific assertions about the expected outcomes


```