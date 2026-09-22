output "api_endpoint" {
  description = "HTTP API Gateway endpoint URL"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}messages"
}

output "sqs_queue_url" {
  description = "Primary SQS Queue URL"
  value       = aws_sqs_queue.main_queue.url
}

output "sqs_dlq_url" {
  description = "Dead Letter Queue URL"
  value       = aws_sqs_queue.dlq.url
}

output "s3_bucket_name" {
  description = "Destination S3 Bucket Name"
  value       = aws_s3_bucket.data_bucket.id
}