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

