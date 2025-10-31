#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TERRAFORM_DIR="$SCRIPT_DIR/../terraform"

# Get ALB URL from Terraform output
if [ -z "$1" ]; then
    echo "Getting ALB URL from Terraform..."
    cd "$TERRAFORM_DIR"
    ALB_DNS=$(terraform output -raw alb_dns_name 2>/dev/null)
    if [ -z "$ALB_DNS" ]; then
        echo "Error: Could not get ALB DNS from Terraform"
        echo "Usage: $0 [alb-dns-name]"
        exit 1
    fi
    echo "Using ALB: $ALB_DNS"
else
    ALB_DNS="$1"
fi

HOST="http://${ALB_DNS}"

if ! command -v locust &> /dev/null; then
    echo "Error: locust not installed. Run: pip install locust"
    exit 1
fi


echo "Running MySQL load test against $HOST..."
cd "$SCRIPT_DIR"

locust -f mysql_load_test.py \
    --headless \
    --users 10 \
    --spawn-rate 2 \
    --run-time 300s \
    --host="${HOST}" \
    --html=mysql_load_test_report.html \
    --csv=mysql_load_test

echo "Done. Results: mysql_test_results.json"

