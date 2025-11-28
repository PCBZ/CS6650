output "subnet_ids" {
  description = "IDs of the public subnets for ALB"
  value       = aws_subnet.public[*].id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets for ECS tasks"
  value       = aws_subnet.private[*].id
}

output "vpc_id" {
  description = "ID of the created VPC"
  value       = aws_vpc.main.id
}

output "security_group_id" {
  description = "Security group ID for ECS"
  value       = aws_security_group.this.id
}

output "alb_security_group_id" {
  description = "Security group ID for ALB"
  value       = aws_security_group.alb.id
}
