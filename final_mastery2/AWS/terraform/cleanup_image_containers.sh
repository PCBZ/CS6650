#!/usr/bin/env bash
# cleanup_image_containers.sh
# Find and force-remove any running containers that are using a given Docker image,
# then (optionally) remove the image. Useful when LocalStack/ECS started containers
# that prevent Terraform's docker provider from removing/replacing an image.

set -euo pipefail

IMAGE=${1:-"000000000000.dkr.ecr.us-west-2.localhost.localstack.cloud:4566/order-service:latest"}

echo "Searching for containers using image: $IMAGE"
CONTAINERS=$(docker ps -q --filter "ancestor=$IMAGE")

if [ -z "$CONTAINERS" ]; then
  echo "No running containers found using image: $IMAGE"
  exit 0
fi

echo "Found containers: $CONTAINERS"
for c in $CONTAINERS; do
  echo "Stopping & removing container $c"
  docker rm -f "$c"
done

echo "You can now remove the image if desired:"
echo "  docker rmi -f $IMAGE"

exit 0
