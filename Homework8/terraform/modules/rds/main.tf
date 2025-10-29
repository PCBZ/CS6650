# RDS MySQL Instance
resource "aws_db_instance" "mysql" {
  identifier = "${var.service_name}-mysql"
  
  # Engine configuration
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
  
  # Database configuration
  db_name  = var.database_name
  username = var.database_username
  password = var.database_password
  port     = 3306
  
  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  
  # Backup configuration
  backup_retention_period = 0
  skip_final_snapshot     = true
  deletion_protection     = false
  
  # Maintenance
  auto_minor_version_upgrade = true
  
  tags = {
    Name = "${var.service_name}-mysql-db"
  }
}

# DB Subnet Group (must span at least 2 AZs)
resource "aws_db_subnet_group" "this" {
  name       = "${var.service_name}-db-subnet-group"
  subnet_ids = var.private_subnet_ids
  
  tags = {
    Name = "${var.service_name}-db-subnet-group"
  }
}

# Security Group for RDS
resource "aws_security_group" "rds" {
  name        = "${var.service_name}-rds-sg"
  description = "Security group for RDS MySQL - Allow access from ECS tasks only"
  vpc_id      = var.vpc_id
  
  ingress {
    description     = "MySQL from ECS tasks"
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = var.ecs_security_group_ids
  }
  
  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "${var.service_name}-rds-security-group"
  }
}
