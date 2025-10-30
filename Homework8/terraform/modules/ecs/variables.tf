variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "image_digest" {
  type        = string
  description = "Docker image digest to force redeployment"
  default     = ""
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Desired Fargate task count"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}

variable "target_group_arn" {
  type        = string
  description = "ALB Target Group ARN"
  default     = null
}

variable "enable_autoscaling" {
  type        = bool
  description = "Enable auto scaling for ECS service"
  default     = true
}

variable "min_capacity" {
  type        = number
  description = "Minimum number of ECS tasks"
  default     = 2
}

variable "max_capacity" {
  type        = number
  description = "Maximum number of ECS tasks"
  default     = 4
}

variable "target_cpu_utilization" {
  type        = number
  description = "Target CPU utilization percentage"
  default     = 70
}

variable "sns_topic_arn" {
  type        = string
  description = "ARN of SNS topic for async order processing"
  default     = ""
}

variable "sqs_queue_url" {
  type        = string
  description = "URL of SQS queue for order processing"
  default     = ""
}

variable "processor_image" {
  type        = string
  description = "ECR image URI for order processor (with tag)"
  default     = ""
}

variable "processor_image_digest" {
  type        = string
  description = "Docker image digest for processor to force redeployment"
  default     = ""
}

variable "enable_processor" {
  type        = bool
  description = "Enable the order processor service"
  default     = true
}

variable "processor_count" {
  type        = number
  description = "Number of order processor tasks to run"
  default     = 2
}

variable "processor_cpu" {
  type        = string
  default     = "256"
  description = "vCPU units for processor"
}

variable "processor_memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB) for processor"
}

# Database configuration
variable "db_host" {
  type        = string
  description = "RDS MySQL endpoint"
  default     = ""
}

variable "db_port" {
  type        = string
  description = "RDS MySQL port"
  default     = "3306"
}

variable "db_name" {
  type        = string
  description = "Database name"
  default     = ""
}

variable "db_user" {
  type        = string
  description = "Database username"
  default     = ""
}

variable "db_password" {
  type        = string
  description = "Database password"
  sensitive   = true
  default     = ""
}
