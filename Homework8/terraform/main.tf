# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

# ECR repository for order processor
module "ecr_processor" {
  source          = "./modules/ecr"
  repository_name = "${var.ecr_repository_name}-processor"
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# SNS/SQS for async order processing
module "messaging" {
  source       = "./modules/messaging"
  service_name = var.service_name
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Application Load Balancer
module "alb" {
  source             = "./modules/alb"
  service_name       = var.service_name
  vpc_id             = module.network.vpc_id
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.alb_security_group_id]
  container_port     = var.container_port
  health_check_path  = var.health_check_path
}

module "ecs" {
  source                 = "./modules/ecs"
  service_name           = var.service_name
  image                  = "${module.ecr.repository_url}:latest"
  image_digest           = docker_image.app.repo_digest
  container_port         = var.container_port
  subnet_ids             = module.network.private_subnet_ids
  security_group_ids     = [module.network.security_group_id]
  execution_role_arn     = data.aws_iam_role.lab_role.arn
  task_role_arn          = data.aws_iam_role.lab_role.arn
  log_group_name         = module.logging.log_group_name
  ecs_count              = var.ecs_count
  cpu                    = var.cpu
  memory                 = var.memory
  region                 = var.aws_region
  target_group_arn       = module.alb.target_group_arn
  min_capacity           = var.min_capacity
  max_capacity           = var.max_capacity
  target_cpu_utilization = var.target_cpu_utilization
  enable_autoscaling     = false
  sns_topic_arn          = module.messaging.sns_topic_arn
  
  # Order processor configuration
  enable_processor       = true
  processor_image        = "${module.ecr_processor.repository_url}:latest"
  processor_image_digest = docker_image.processor.repo_digest
  sqs_queue_url          = module.messaging.sqs_queue_url
  processor_count        = var.processor_count
  processor_cpu          = var.processor_cpu
  processor_memory       = var.processor_memory

  depends_on = [module.alb, docker_registry_image.app, docker_registry_image.processor]
}

# Lambda function for order processing (alternative to ECS processor)
module "lambda" {
  source             = "./modules/lambda"
  service_name       = var.service_name
  region             = var.aws_region
  sns_topic_arn      = module.messaging.sns_topic_arn
  enable_lambda      = var.enable_lambda
  log_retention_days = var.log_retention_days
}

// Build & push the Go app image into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "latest"
  name = "${module.ecr.repository_url}:latest"
  
  # Force rebuild on every apply by using a timestamp trigger
  triggers = {
    build_time = timestamp()
  }

  build {
    # relative path from terraform/ → src/
    context = "../src"
    # Dockerfile defaults to "Dockerfile" in that context
    platform = "linux/amd64"
    # Remove intermediate containers to ensure clean builds
    remove = true
    # Force fresh build without cache
    no_cache = true
  }
}

resource "docker_registry_image" "app" {
  # this will push :latest → ECR
  name = docker_image.app.name
}

// Build & push the order processor image into ECR
resource "docker_image" "processor" {
  name = "${module.ecr_processor.repository_url}:latest"
  
  triggers = {
    build_time = timestamp()
  }

  build {
    context    = "../src"
    dockerfile = "Dockerfile.processor"
    platform   = "linux/amd64"
    remove     = true
    no_cache   = true
  }
}

resource "docker_registry_image" "processor" {
  name = docker_image.processor.name
}
