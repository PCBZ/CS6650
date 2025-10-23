# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "order-service"
}

variable "service_name" {
  type    = string
  default = "order-processing-service"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 1
}

# CPU and Memory settings for Fargate
variable "cpu" {
  type        = string
  default     = "256"  # 512 CPU units (0.5 vCPU)
  description = "vCPU units for Fargate task"
}

variable "memory" {
  type        = string
  default     = "512"  # 1024 MB
  description = "Memory (MiB) for Fargate task"
}

# Auto Scaling settings (disabled for Phase 1 testing)
variable "min_capacity" {
  type        = number
  default     = 1
  description = "Minimum number of ECS tasks"
}

variable "max_capacity" {
  type        = number
  default     = 1
  description = "Maximum number of ECS tasks (same as min to disable scaling)"
}

variable "target_cpu_utilization" {
  type        = number
  default     = 70
  description = "Target CPU utilization percentage for auto scaling (not used when min=max)"
}

variable "health_check_path" {
  type        = string
  default     = "/health"
  description = "Health check path for ALB"
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Order Processor settings
variable "processor_count" {
  type        = number
  default     = 1
  description = "Number of order processor tasks to run"
}

variable "processor_cpu" {
  type        = string
  default     = "256"
  description = "vCPU units for processor task"
}

variable "processor_memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB) for processor task"
}

variable "num_workers" {
  type        = number
  default     = 1
  description = "Number of SQS polling workers per processor task"
}
