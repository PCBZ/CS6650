output "ecs_cluster_name" {
  description = "productAPI-cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "productAPI-service"
  value       = module.ecs.service_name
}