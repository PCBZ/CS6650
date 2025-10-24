variable "service_name" {
  type        = string
  description = "Name of the service"
}

variable "region" {
  type        = string
  description = "AWS region"
}

variable "sns_topic_arn" {
  type        = string
  description = "ARN of the SNS topic to subscribe to"
}

variable "enable_lambda" {
  type        = bool
  default     = false
  description = "Enable Lambda function for order processing"
}

variable "log_retention_days" {
  type        = number
  default     = 7
  description = "Number of days to retain Lambda logs"
}
