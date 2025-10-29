output "cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.this.name
}

output "service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.this.name
}

output "cluster_arn" {
  description = "ECS cluster ARN"
  value       = aws_ecs_cluster.this.arn
}

output "service_arn" {
  description = "ECS service ARN"
  value       = aws_ecs_service.this.arn
}

output "processor_service_name" {
  description = "Order processor service name"
  value       = var.enable_processor ? aws_ecs_service.processor[0].name : null
}

output "processor_service_arn" {
  description = "Order processor service ARN"
  value       = var.enable_processor ? aws_ecs_service.processor[0].arn : null
}

