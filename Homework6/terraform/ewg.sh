#!/bin/bash

ECR_URL=975050147762.dkr.ecr.us-west-2.amazonaws.com/product-service
AWS_REGION=us-west-2

echo "🔍 Checking ECR image: $ECR_URL:latest"

# 登录ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $(echo $ECR_URL | cut -d'/' -f1)

# 使用buildx检查（最准确）
echo "📋 Image architecture info:"
docker buildx imagetools inspect $ECR_URL:latest