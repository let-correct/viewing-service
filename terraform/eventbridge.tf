###############################################################################
# Data source — the shared let-correct event bus owned by agent-service
###############################################################################

data "aws_cloudwatch_event_bus" "let_correct" {
  name = "let-correct"
}

###############################################################################
# EventBridge Rule — filter AppointmentCreated / AppointmentCancelled events
# from agent-service into the viewing-service SQS queue.
###############################################################################

resource "aws_cloudwatch_event_rule" "appointments" {
  name           = "viewing-service-appointments"
  description    = "Routes appointment events from agent-service to viewing-service"
  event_bus_name = data.aws_cloudwatch_event_bus.let_correct.name

  event_pattern = jsonencode({
    source      = ["agent-service"]
    detail-type = ["AppointmentCreated", "AppointmentCancelled"]
  })

  tags = var.tags
}

resource "aws_cloudwatch_event_target" "appointments_sqs" {
  rule           = aws_cloudwatch_event_rule.appointments.name
  event_bus_name = data.aws_cloudwatch_event_bus.let_correct.name
  arn            = aws_sqs_queue.viewing_worker.arn
}

resource "aws_sqs_queue_policy" "viewing_worker_eventbridge" {
  queue_url = aws_sqs_queue.viewing_worker.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "events.amazonaws.com" }
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.viewing_worker.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_cloudwatch_event_rule.appointments.arn
          }
        }
      }
    ]
  })
}
