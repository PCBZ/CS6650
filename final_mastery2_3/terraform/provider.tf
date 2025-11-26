# Specify where to find the AWS & Docker providers
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.7.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 2.0"
    }
  }
}

# Configure AWS credentials & region
provider "aws" {
  region                      = var.aws_region
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true

  endpoints {
    apigateway               = var.localstack_endpoint
    applicationautoscaling   = var.localstack_endpoint
    cloudformation           = var.localstack_endpoint
    cloudwatch               = var.localstack_endpoint
    dynamodb                 = var.localstack_endpoint
    ec2                      = var.localstack_endpoint
    ecr                      = var.localstack_endpoint
    ecs                      = var.localstack_endpoint
    elbv2                    = var.localstack_endpoint
    iam                      = var.localstack_endpoint
    lambda                   = var.localstack_endpoint
    logs                     = var.localstack_endpoint
    rds                      = var.localstack_endpoint
    s3                       = var.localstack_endpoint
    secretsmanager           = var.localstack_endpoint
    sns                      = var.localstack_endpoint
    sqs                      = var.localstack_endpoint
    ssm                      = var.localstack_endpoint
    sts                      = var.localstack_endpoint
  }
}

# Fetch an ECR auth token so Terraform's Docker provider can log in
# data "aws_ecr_authorization_token" "registry" {}

# Configure Docker provider to authenticate against ECR automatically
provider "docker" {
  registry_auth {
    address  = "000000000000.dkr.ecr.us-west-2.localhost.localstack.cloud:4566"
    username = "test"
    password = "test"
  }
}