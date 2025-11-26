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