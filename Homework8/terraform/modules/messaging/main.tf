# SNS Topic for order processing events
resource "aws_sns_topic" "order_processing" {
  name = "${var.service_name}-order-processing-events"

  tags = {
    Name = "${var.service_name}-order-processing-events"
  }
}

# SQS Queue for order processing
resource "aws_sqs_queue" "order_processing" {
  name                       = "${var.service_name}-order-processing-queue"
  visibility_timeout_seconds = 30     # 30 seconds visibility timeout
  message_retention_seconds  = 345600 # 4 days retention
  receive_wait_time_seconds  = 20     # 20 seconds long polling

  tags = {
    Name = "${var.service_name}-order-processing-queue"
  }
}

# Dead Letter Queue for failed messages
resource "aws_sqs_queue" "order_processing_dlq" {
  name = "${var.service_name}-order-processing-dlq"

  tags = {
    Name = "${var.service_name}-order-processing-dlq"
  }
}

# Redrive policy for main queue to DLQ
resource "aws_sqs_queue_redrive_policy" "order_processing" {
  queue_url = aws_sqs_queue.order_processing.id

  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.order_processing_dlq.arn
    maxReceiveCount     = 3
  })
}

# SNS subscription to SQS
resource "aws_sns_topic_subscription" "order_processing" {
  topic_arn = aws_sns_topic.order_processing.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.order_processing.arn
}

# SQS Queue Policy to allow SNS to send messages
resource "aws_sqs_queue_policy" "order_processing" {
  queue_url = aws_sqs_queue.order_processing.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.order_processing.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.order_processing.arn
          }
        }
      }
    ]
  })
}
