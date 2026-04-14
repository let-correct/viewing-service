###############################################################################
# Viewing Worker Lambda
###############################################################################

resource "aws_lambda_function" "viewing_worker" {
  function_name = "viewing-worker"
  role          = aws_iam_role.viewing_worker.arn

  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.viewing_worker_ecr.repository_url}:latest"
  architectures = ["arm64"]

  memory_size = 128
  timeout     = 30

  environment {
    variables = {
      LOG_LEVEL     = var.log_level
      EVENT_BUS_NAME = data.aws_cloudwatch_event_bus.let_correct.name
    }
  }

  logging_config {
    log_format = "JSON"
    log_group  = aws_cloudwatch_log_group.viewing_worker.name
  }

  tags = var.tags

  depends_on = [
    aws_iam_role_policy_attachment.viewing_worker_basic_execution,
    aws_cloudwatch_log_group.viewing_worker,
  ]
}

resource "aws_lambda_event_source_mapping" "viewing_worker_sqs" {
  event_source_arn        = aws_sqs_queue.viewing_worker.arn
  function_name           = aws_lambda_function.viewing_worker.arn
  batch_size              = 10
  function_response_types = ["ReportBatchItemFailures"]
}

###############################################################################
# IAM
###############################################################################

data "aws_iam_policy_document" "viewing_worker_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "viewing_worker" {
  name               = "viewing-worker-execution-role"
  assume_role_policy = data.aws_iam_policy_document.viewing_worker_assume_role.json
  tags               = var.tags
}

resource "aws_iam_role_policy_attachment" "viewing_worker_basic_execution" {
  role       = aws_iam_role.viewing_worker.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "viewing_worker_permissions" {
  name = "viewing-worker-permissions"
  role = aws_iam_role.viewing_worker.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["events:PutEvents"]
        Resource = data.aws_cloudwatch_event_bus.let_correct.arn
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
        ]
        Resource = aws_sqs_queue.viewing_worker.arn
      },
    ]
  })
}

###############################################################################
# CloudWatch Log Group
###############################################################################

resource "aws_cloudwatch_log_group" "viewing_worker" {
  name              = "/aws/lambda/viewing_worker"
  retention_in_days = 30
  tags              = var.tags
}
