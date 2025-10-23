output "sns_topic_arn" {
  description = "ARN of the SNS topic for order processing events"
  value       = aws_sns_topic.order_processing.arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue for order processing"
  value       = aws_sqs_queue.order_processing.url
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue for order processing"
  value       = aws_sqs_queue.order_processing.arn
}

output "sqs_dlq_url" {
  description = "URL of the Dead Letter Queue"
  value       = aws_sqs_queue.order_processing_dlq.url
}

output "sqs_dlq_arn" {
  description = "ARN of the Dead Letter Queue"
  value       = aws_sqs_queue.order_processing_dlq.arn
}