###############################################################################
# Viewing Worker — SQS Queue + Dead-Letter Queue
###############################################################################

resource "aws_sqs_queue" "viewing_worker_dlq" {
  name = "viewing-worker-dlq"
  tags = var.tags
}

resource "aws_sqs_queue" "viewing_worker" {
  name = "viewing-worker"

  visibility_timeout_seconds = 60 # Must be >= Lambda timeout (30s), recommended 2x

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.viewing_worker_dlq.arn
    maxReceiveCount     = 3
  })

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "viewing_worker_dlq" {
  alarm_name          = "viewing-worker-dlq-not-empty"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 60
  statistic           = "Sum"
  threshold           = 0
  alarm_description   = "Messages in the viewing-worker DLQ — requires investigation"

  dimensions = {
    QueueName = aws_sqs_queue.viewing_worker_dlq.name
  }

  tags = var.tags
}
