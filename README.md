## Assignment

Build a production-ready AWS setup with the following flow:
POST request > Lambda > SQS > asynchronous processing > S3

Any POST request sent to an endpoint should be handled by a Lambda function.

For each request, capture and eventually store the following information in S3:

  - Request headers
  - Endpoint/path
  - Query parameters
  - Request body
  - Response
  - Timestamp

The persistence to S3 must happen asynchronously using SQS, so the API request should
 not depend on the S3 write completing.

### Requirements

  - Use Terraform to provision the infrastructure.
  - Write the solution as if it were intended for a production environment.
  - Include all application/Lambda code required for the flow.
  - Do not include credentials, secrets, Terraform state files, or other sensitive information in the repository.

---


## Code

This project is given the name: "Jodopost". The solution comprises of creating 2 lambda functions -
 receiver and processor with an API gateway (necessary component introduced to serve as the entry point).
 The code is under the `app/` directory and is implemented using Golang.

Lambda functions:
  - Receiver: Accepts the HTTP request through an AWS API Gateway and creates a payload with all the required data that can be pushed to SQS. 
  - Processor: Consumes messages from the same SQS queue and writes them as files to S3


#### Build 

Sample steps for demonstration:  

```sh
# example: for the receiver
cd ./app/jodopost
go mod tidy
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o ./bootstrap ./receiver/main.go
zip a02b502.zip bootstrap
```

#### Deployment 

The zip file name contains the git commit hash as the version identifier. They are stored on S3
 and the S3 path is configured on Lambda. With a new version published to S3, the configuration
 can be updated via terraform to deploy the new version

#### Configuration

Required environment variables to configured on Lambda:  

  - QUEUE_URL: URL of the SQS queue used to support asynchronous processing (receiver)
  - BUCKET_NAME: Name of S3 bucket used to store data (processor)


### Details

That data is stored on S3 at a key like: `requests/2026/09/22/POST/abc-123-def-<uuid>.json`

Each archived object in S3 will look like this:

```json
{
  "request_id": "abc-123-def",
  "timestamp": "2026-09-22T10:15:30Z",
  "path": "/",
  "http_method": "POST",
  "headers": {
    "Content-Type": "application/json",
    "User-Agent": "curl/8.4.0",
    "X-Forwarded-For": "203.0.113.5"
  },
  "query_params": {
    "source": "web"
  },
  "request_body": "{\"name\": \"Alice\", \"email\": \"alice@example.com\"}",
  "response_status": 200,
  "response_body": "{\"message\": \"Hello, Alice! Your request was processed.\", \"success\": true}"
}
```


#### Features

  - Partial batch failures: The consumer uses `ReportBatchItemFailures` so if one message in a batch of 10 fails, only that message gets retried — not the whole batch
  - DLQ: after 3 failed attempts, a message goes to RequestDLQ instead of retrying forever
  - Archiving is best-effort, not blocking: `archiveRequest` failures (SQS being down, throttled, etc.) are logged but don't fail the caller's HTTP response.


#### Future scope

  - Idempotency: SQS is at-least-once delivery, so the same message could be processed twice. The S3 key includes a UUID, so duplicate processing would create duplicate objects rather than overwrite
  - Implement CI: Build pipeline that is triggered to create Lambda build artifacts when merged to the release branch.

---

## Infra

The infrastructure of all the essential components are created using Terraform. The code is under
 the `infra/` directory.

#### Details

  - An API gateway API is created to serve as the entrypoint for the HTTP request
  - There are 2 Lambda functions, namely: receiver and processor. 
  - The receiver accepts the request from the API gateway and creates a payload with the required information and pushes it to an SQS queue
  - The processor consumes messages in the queue and stores the information as a file on S3
  - A dedicated KMS key is used for encryption required on the various components in this project
  - CloudWatch log groups are created for additional visibility in troubleshooting

#### Future scope

  - Enable authentication/authorization at API gateway and TLS certificates
  - S3 data lifecycle policy
  - VPC networking for Lambda
  - VPC endpoints and security groups for security and cost
  - ARM based code builds for cost optimization

