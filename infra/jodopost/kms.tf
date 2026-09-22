resource "aws_kms_key" "pipeline_key" {
  description             = "KMS key for Jodopost components"
  deletion_window_in_days = 30
  enable_key_rotation     = true
}

resource "aws_kms_alias" "pipeline_key_alias" {
  name          = "alias/jodopost-key-${var.environment}"
  target_key_id = aws_kms_key.pipeline_key.key_id
}
