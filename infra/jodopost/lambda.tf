resource "aws_lambda_function" "receiver" {
  function_name    = "jodopost-receiver-${var.environment}"
  role             = aws_iam_role.receiver_role.arn

  runtime       = "provided.al2023"
  architectures = ["x86_64"]
  handler       = "bootstrap"

  # code binaries built and stored on S3
  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  # versioned using git commit hash - update here to release new version
  s3_key           = "jodopost/receiver/a02b502.zip"

  timeout          = 10
  memory_size      = 256

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      QUEUE_URL = aws_sqs_queue.main_queue.url
    }
  }

  depends_on = [aws_cloudwatch_log_group.receiver_log_group]
}

resource "aws_lambda_function" "processor" {
  function_name    = "jodopost-processor-${var.environment}"
  role             = aws_iam_role.processor_role.arn
  
  runtime       = "provided.al2023"
  architectures = ["x86_64"]
  handler       = "bootstrap"

  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  s3_key           = "jodopost/processor/a02b502.zip"

  timeout          = 30
  memory_size      = 256

  tracing_config {
    mode = "Active"
  }

  environment {
    variables = {
      BUCKET_NAME = aws_s3_bucket.data_bucket.id
    }
  }

  depends_on = [aws_cloudwatch_log_group.processor_log_group]
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.main_queue.arn
  function_name    = aws_lambda_function.processor.arn
  batch_size       = 10
  enabled          = true

  scaling_config {
    maximum_concurrency = 10
  }
}