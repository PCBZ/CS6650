output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "ECS service name"
  value       = module.ecs.service_name
}

output "alb_dns_name" {
  description = "DNS name of the Application Load Balancer"
  value       = module.alb.alb_dns_name
}

output "api_url" {
  description = "Base URL for the API"
  value       = "http://${module.alb.alb_dns_name}"
}

output "health_check_url" {
  description = "Health check URL"
  value       = "http://${module.alb.alb_dns_name}/health"
}

output "search_api_example" {
  description = "Example search API URL"
  value       = "http://${module.alb.alb_dns_name}/v1/products/search?q=Alpha"
}

# Messaging outputs
output "sns_topic_arn" {
  description = "ARN of the SNS topic for order processing"
  value       = module.messaging.sns_topic_arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue for order processing"
  value       = module.messaging.sqs_queue_url
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue for order processing"
  value       = module.messaging.sqs_queue_arn
}

# Order Processor outputs
output "processor_service_name" {
  description = "Order processor service name"
  value       = module.ecs.processor_service_name
}

output "processor_service_arn" {
  description = "Order processor service ARN"
  value       = module.ecs.processor_service_arn
}

output "processor_ecr_url" {
  description = "ECR repository URL for order processor"
  value       = module.ecr_processor.repository_url
}

# Lambda outputs
output "lambda_function_name" {
  description = "Lambda function name for order processing"
  value       = module.lambda.lambda_function_name
}

output "lambda_function_arn" {
  description = "Lambda function ARN for order processing"
  value       = module.lambda.lambda_function_arn
}
