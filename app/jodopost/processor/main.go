package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"jodopost/internal/audit"
)

// RequestBody defines the expected shape of the incoming JSON payload
type RequestBody struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ResponseBody defines what we send back to the caller
type ResponseBody struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

var (
	sqsClient *sqs.Client
	queueURL  string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	sqsClient = sqs.NewFromConfig(cfg)
	queueURL = os.Getenv("QUEUE_URL")
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if request.HTTPMethod != "POST" {
		resp := errorResponse(405, "Method not allowed")
		archiveRequest(ctx, request, timestamp, resp)
		return resp, nil
	}

	var body RequestBody
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		log.Printf("error parsing body: %v", err)
		resp := errorResponse(400, "Invalid JSON body")
		archiveRequest(ctx, request, timestamp, resp)
		return resp, nil
	}

	if body.Name == "" {
		resp := errorResponse(400, "Field 'name' is required")
		archiveRequest(ctx, request, timestamp, resp)
		return resp, nil
	}

	log.Printf("Processing request for: %s (%s)", body.Name, body.Email)

	respBody := ResponseBody{
		Message: fmt.Sprintf("Hello, %s! Your request was processed.", body.Name),
		Success: true,
	}
	respJSON, _ := json.Marshal(respBody)

	resp := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(respJSON),
	}

	archiveRequest(ctx, request, timestamp, resp)

	return resp, nil
}

// archiveRequest builds the audit record and pushes it to SQS for async S3 storage.
func archiveRequest(ctx context.Context, request events.APIGatewayProxyRequest, timestamp string, resp events.APIGatewayProxyResponse) {
	record := audit.Record{
		RequestID:      request.RequestContext.RequestID,
		Timestamp:      timestamp,
		Path:           request.Path,
		HTTPMethod:     request.HTTPMethod,
		Headers:        request.Headers,
		QueryParams:    request.QueryStringParameters,
		RequestBody:    request.Body,
		ResponseStatus: resp.StatusCode,
		ResponseBody:   resp.Body,
	}

	recordJSON, err := json.Marshal(record)
	if err != nil {
		log.Printf("error marshaling audit record: %v", err)
		return
	}

	if err := sendToQueue(ctx, string(recordJSON)); err != nil {
		log.Printf("error sending audit record to SQS: %v", err)
	}
}

func sendToQueue(ctx context.Context, messageBody string) error {
	if queueURL == "" {
		return fmt.Errorf("QUEUE_URL environment variable not set")
	}
	_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: &messageBody,
	})
	return err
}

func errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	body, _ := json.Marshal(map[string]string{"error": message})
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func main() {
	lambda.Start(handler)
}