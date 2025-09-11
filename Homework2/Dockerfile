# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

# Set destination for COPY
WORKDIR /app

# Download all dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY *.go ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-go-web-service

EXPOSE 8080

# Run
CMD ["/docker-go-web-service"]