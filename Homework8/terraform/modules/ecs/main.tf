# ECS Cluster
resource "aws_ecs_cluster" "this" {
  name = "${var.service_name}-cluster"
}

# Task Definition
resource "aws_ecs_task_definition" "this" {
  family                   = "${var.service_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory

  execution_role_arn = var.execution_role_arn
  task_role_arn      = var.task_role_arn

  # Force task definition update when image changes
  tags = {
    ImageDigest = var.image_digest
    LastUpdated = timestamp()
  }

  container_definitions = jsonencode([{
    name      = "${var.service_name}-container"
    image     = var.image
    essential = true

    environment = [
      {
        name  = "AWS_REGION"
        value = var.region
      },
      {
        name  = "SNS_TOPIC_ARN"
        value = var.sns_topic_arn
      },
      {
        name  = "DB_HOST"
        value = var.db_host
      },
      {
        name  = "DB_PORT"
        value = var.db_port
      },
      {
        name  = "DB_NAME"
        value = var.db_name
      },
      {
        name  = "DB_USER"
        value = var.db_user
      },
      {
        name  = "DB_PASSWORD"
        value = var.db_password
      }
    ]

    portMappings = [{
      containerPort = var.container_port
    }]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = var.log_group_name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

# ECS Service
resource "aws_ecs_service" "this" {
  name            = var.service_name
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = var.min_capacity
  launch_type     = "FARGATE"

  # Speed up destroy process
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 0

  network_configuration {
    subnets         = var.subnet_ids
    security_groups = var.security_group_ids
    assign_public_ip = false
  }

  # Load balancer configuration
  dynamic "load_balancer" {
    for_each = var.target_group_arn != null ? [1] : []
    content {
      target_group_arn = var.target_group_arn
      container_name   = "${var.service_name}-container"
      container_port   = var.container_port
    }
  }

  # No lifecycle ignore rules needed when auto-scaling is disabled
  # lifecycle {
  #   ignore_changes = [desired_count]
  # }
}

# Auto Scaling Target
resource "aws_appautoscaling_target" "this" {
  count              = var.enable_autoscaling ? 1 : 0
  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.this.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

# Auto Scaling Policy - Scale Up
resource "aws_appautoscaling_policy" "scale_up" {
  count              = var.enable_autoscaling ? 1 : 0
  name               = "${var.service_name}-scale-up"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.this[0].resource_id
  scalable_dimension = aws_appautoscaling_target.this[0].scalable_dimension
  service_namespace  = aws_appautoscaling_target.this[0].service_namespace

  target_tracking_scaling_policy_configuration {
    target_value       = var.target_cpu_utilization
    scale_in_cooldown  = 300
    scale_out_cooldown = 300

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
  }
}

# =====================================================================
# Order Processor Service (SQS Consumer)
# =====================================================================

# Order Processor Task Definition
resource "aws_ecs_task_definition" "processor" {
  count                    = var.enable_processor ? 1 : 0
  family                   = "${var.service_name}-processor-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.processor_cpu
  memory                   = var.processor_memory

  execution_role_arn = var.execution_role_arn
  task_role_arn      = var.task_role_arn

  # Force task definition update when image changes
  tags = {
    ImageDigest = var.processor_image_digest
    LastUpdated = timestamp()
  }

  container_definitions = jsonencode([{
    name      = "${var.service_name}-processor-container"
    image     = var.processor_image
    essential = true

    environment = [
      {
        name  = "AWS_REGION"
        value = var.region
      },
      {
        name  = "SQS_QUEUE_URL"
        value = var.sqs_queue_url
      },
      {
        name  = "NUM_WORKERS"
        value = "100"
      }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = var.log_group_name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "processor"
      }
    }
  }])
}

# Order Processor ECS Service (no load balancer, just background worker)
resource "aws_ecs_service" "processor" {
  count           = var.enable_processor ? 1 : 0
  name            = "${var.service_name}-processor"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.processor[0].arn
  desired_count   = var.processor_count
  launch_type     = "FARGATE"

  # Speed up destroy process
  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 0

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = var.security_group_ids
    assign_public_ip = false
  }

  # No load balancer needed for background processor
}

