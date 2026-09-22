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
 receiver and processor with an API gateway (necessary component introduced to serve as the entry point)

Lambda functions:
  - Receiver: Accepts the HTTP request through an AWS API Gateway and creates a payload with all the required data that can be pushed to SQS. 
  - Processor: Consumes messages from the same SQS queue and writes them as files to S3


Build steps for deploying to Lambda:

```sh
# example: for the receiver
cd ./app/jodopost
go mod tidy
GOOS=linux GOARCH=amd64 go build -tags lambda.norpc -o ./bootstrap ./receiver/main.go
zip function.zip bootstrap
```

Required environment variables to configured on Lambda:  

  - QUEUE_URL: URL of the SQS queue used to support asynchronous processing (receiver)
  - BUCKET_NAME: Name of S3 bucket used to store data (processor)


### Details

Stored on S3 at a key like: `requests/2026/09/22/POST/abc-123-def-<uuid>.json`

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
  - Idempotency: SQS is at-least-once delivery, so the same message could be processed twice. The S3 key includes a UUID, so duplicate processing would create duplicate objects rather than overwrite
  - DLQ: after 3 failed attempts, a message goes to RequestDLQ instead of retrying forever
  - Archiving is best-effort, not blocking: `archiveRequest` failures (SQS being down, throttled, etc.) are logged but don't fail the caller's HTTP response.

