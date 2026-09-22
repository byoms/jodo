package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"jodopost/internal/audit"
)

var (
	s3Client   *s3.Client
	bucketName string
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
	bucketName = os.Getenv("BUCKET_NAME")
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) (events.SQSEventResponse, error) {
	var failures []events.SQSBatchItemFailure

	for _, record := range sqsEvent.Records {
		if err := processMessage(ctx, record); err != nil {
			log.Printf("failed to process message %s: %v", record.MessageId, err)
			failures = append(failures, events.SQSBatchItemFailure{
				ItemIdentifier: record.MessageId,
			})
			continue
		}
		log.Printf("successfully archived message %s", record.MessageId)
	}

	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}

func processMessage(ctx context.Context, record events.SQSMessage) error {
	var rec audit.Record
	if err := json.Unmarshal([]byte(record.Body), &rec); err != nil {
		return fmt.Errorf("invalid message body: %w", err)
	}

	if bucketName == "" {
		return fmt.Errorf("BUCKET_NAME environment variable not set")
	}

	now, err := time.Parse(time.RFC3339, rec.Timestamp)
	if err != nil {
		now = time.Now().UTC()
	}

	method := rec.HTTPMethod
	if method == "" {
		method = "UNKNOWN"
	}

	key := fmt.Sprintf("requests/%04d/%02d/%02d/%s/%s-%s.json",
		now.Year(), now.Month(), now.Day(), method,
		strings.TrimSpace(rec.RequestID), uuid.New().String())

	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal audit record: %w", err)
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &bucketName,
		Key:         &key,
		Body:        strings.NewReader(string(data)),
		ContentType: awsString("application/json"),
	})
	if err != nil {
		return fmt.Errorf("failed to write to S3: %w", err)
	}

	return nil
}

func awsString(s string) *string { return &s }

func main() {
	lambda.Start(handler)
}