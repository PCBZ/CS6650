# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "product-service"
}

variable "service_name" {
  type    = string
  default = "product-api-service"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 2
}

# CPU and Memory settings for Fargate
variable "cpu" {
  type        = string
  default     = "256"  # 256 CPU units (0.25 vCPU)
  description = "vCPU units for Fargate task"
}

variable "memory" {
  type        = string
  default     = "512"  # 512 MB
  description = "Memory (MiB) for Fargate task"
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}