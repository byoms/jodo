resource "aws_sqs_queue" "dlq" {
  name                              = "jodopost-dlq-${var.environment}"
  kms_master_key_id                 = aws_kms_key.pipeline_key.id
  message_retention_seconds         = 1209600  # 14 days retention
  kms_data_key_reuse_period_seconds = 300
}

resource "aws_sqs_queue" "main_queue" {
  name                              = "jodopost-buffer-queue-${var.environment}"
  kms_master_key_id                 = aws_kms_key.pipeline_key.id
  kms_data_key_reuse_period_seconds = 300
  visibility_timeout_seconds        = 180     # 6x processor Lambda timeout
  message_retention_seconds         = 345600  # 4 days

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.dlq.arn
    maxReceiveCount     = 3
  })
}