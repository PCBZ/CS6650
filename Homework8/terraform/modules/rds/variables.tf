variable "service_name" {
  type        = string
  description = "Name of the service"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID where RDS will be deployed"
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "List of private subnet IDs for RDS (must be at least 2 subnets in different AZs)"
}

variable "ecs_security_group_ids" {
  type        = list(string)
  description = "List of ECS security group IDs that need access to RDS"
}

variable "database_name" {
  type        = string
  default     = "orders_db"
  description = "Name of the initial database"
}

variable "database_username" {
  type        = string
  default     = "admin"
  description = "Master username for RDS"
  sensitive   = true
}

variable "database_password" {
  type        = string
  description = "Master password for RDS"
  sensitive   = true
}
