#!/bin/bash

# Build script for AWS Lambda Go function
# Builds for Linux AMD64 and packages as a zip

set -e

echo "🔨 Building Lambda function..."

# Clean previous builds
rm -f bootstrap bootstrap.zip

# Initialize go modules and download dependencies
go mod tidy
go mod download

# Build for Linux AMD64 (Lambda runtime)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -tags lambda.norpc -o bootstrap main.go

# Create deployment package
zip bootstrap.zip bootstrap

echo "✅ Build complete: bootstrap.zip"
echo "📦 File size: $(ls -lh bootstrap.zip | awk '{print $5}')"
