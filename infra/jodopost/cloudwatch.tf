# cloudwatch log groups for the various components

resource "aws_cloudwatch_log_group" "receiver_log_group" {
  name              = "/aws/lambda/jodopost-receiver-${var.environment}"
  retention_in_days = var.log_retention_in_days
  kms_key_id        = aws_kms_key.pipeline_key.arn
}

resource "aws_cloudwatch_log_group" "processor_log_group" {
  name              = "/aws/lambda/jodopost-processor-${var.environment}"
  retention_in_days = var.log_retention_in_days
  kms_key_id        = aws_kms_key.pipeline_key.arn
}

resource "aws_cloudwatch_log_group" "api_gw_log_group" {
  name              = "/aws/apigateway/jodopost-api-${var.environment}"
  retention_in_days = var.log_retention_in_days
  kms_key_id        = aws_kms_key.pipeline_key.arn
}