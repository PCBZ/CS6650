variable "service_name" {
  type        = string
  description = "Base name for ALB resources"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for ALB"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs for ALB"
}

variable "security_group_ids" {
  type        = list(string)
  description = "Security group IDs for ALB"
}

variable "container_port" {
  type        = number
  description = "Port that the container listens on"
}

variable "health_check_path" {
  type        = string
  description = "Health check path"
  default     = "/health"
}