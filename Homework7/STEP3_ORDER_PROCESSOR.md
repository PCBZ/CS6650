# Step 3: Order Processor Service Implementation

## Overview
Implemented a production-grade order processor service that consumes messages from SQS and processes orders asynchronously. This service runs as a separate ECS Fargate task and scales independently from the web service.

## Architecture

```
┌─────────────┐      ┌─────────┐      ┌─────────┐      ┌──────────────┐
│   Client    │─────▶│   ALB   │─────▶│ Web API │─────▶│  SNS Topic   │
└─────────────┘      └─────────┘      └─────────┘      └──────┬───────┘
                                       (202 Accepted)           │
                                                                 │ Fan-out
                                                                 ▼
                                                         ┌───────────────┐
                                                         │   SQS Queue   │
                                                         └───────┬───────┘
                                                                 │
                                                                 │ Poll (long polling)
                                                                 ▼
                                                     ┌───────────────────────┐
                                                     │  Order Processor(s)   │
                                                     │  - Multiple workers   │
                                                     │  - Concurrent tasks   │
                                                     │  - Payment simulation │
                                                     └───────────────────────┘
```

## Components Created

### 1. Go Order Processor (`src/orderProcessor.go`)
**Key Features:**
- **Long Polling**: Uses 20-second long polling to efficiently receive messages from SQS
- **Concurrent Processing**: Each processor task runs multiple workers (default: 5)
- **Batch Processing**: Receives up to 10 messages per poll
- **Graceful Shutdown**: Handles SIGTERM/SIGINT for clean shutdown
- **Error Handling**: Robust error handling with message deletion on success

**Core Components:**
```go
type OrderProcessor struct {
    sqsClient *sqs.Client
    queueURL  string
    wg        sync.WaitGroup
    ctx       context.Context
    cancel    context.CancelFunc
}
```

**Processing Flow:**
1. **Receive**: Long poll SQS for up to 10 messages (20s wait time)
2. **Parse**: Unwrap SNS message and extract order details
3. **Process**: Simulate payment processing (100-500ms)
4. **Update**: Log order status update (COMPLETED)
5. **Delete**: Remove message from queue on success

**Configuration via Environment Variables:**
- `SQS_QUEUE_URL`: Required - URL of the SQS queue to consume from
- `NUM_WORKERS`: Optional (default: 5) - Number of polling workers per task
- `AWS_REGION`: Required - AWS region for SDK

### 2. Processor Dockerfile (`src/Dockerfile.processor`)
**Multi-stage Build:**
- **Builder stage**: Compiles Go binary with static linking
- **Runtime stage**: Minimal Alpine image with ca-certificates
- **Optimization**: Small image size, fast startup

### 3. Infrastructure Updates

#### ECS Module Enhancements (`terraform/modules/ecs/`)

**New Variables:**
```hcl
- sqs_queue_url          # SQS queue URL to consume from
- processor_image        # ECR image for processor
- processor_image_digest # Force redeployment on changes
- enable_processor       # Enable/disable processor service
- processor_count        # Number of processor tasks (default: 2)
- processor_cpu          # vCPU units (default: 256)
- processor_memory       # Memory in MiB (default: 512)
- num_workers            # Workers per task (default: 5)
```

**New Resources:**
```hcl
resource "aws_ecs_task_definition" "processor" {
  # Task definition for order processor
  # - No port mappings (background worker)
  # - Environment: SQS_QUEUE_URL, NUM_WORKERS, AWS_REGION
  # - Uses same execution/task roles as web service
}

resource "aws_ecs_service" "processor" {
  # ECS service for processor
  # - No load balancer attachment
  # - Runs in private subnets
  # - Desired count: 2 (configurable)
}
```

**New Outputs:**
```hcl
- processor_service_name # Name of processor ECS service
- processor_service_arn  # ARN of processor service
```

#### Root Terraform (`terraform/main.tf`)

**New ECR Repository:**
```hcl
module "ecr_processor" {
  source          = "./modules/ecr"
  repository_name = "${var.ecr_repository_name}-processor"
}
```

**Processor Docker Build & Push:**
```hcl
resource "docker_image" "processor" {
  name = "${module.ecr_processor.repository_url}:latest"
  build {
    context    = "../src"
    dockerfile = "Dockerfile.processor"
    platform   = "linux/amd64"
  }
}

resource "docker_registry_image" "processor" {
  name = docker_image.processor.name
}
```

**Updated ECS Module Call:**
```hcl
module "ecs" {
  # ...existing configuration...
  
  # Order processor configuration
  enable_processor       = true
  processor_image        = "${module.ecr_processor.repository_url}:latest"
  processor_image_digest = docker_image.processor.repo_digest
  sqs_queue_url          = module.messaging.sqs_queue_url
  processor_count        = var.processor_count
  processor_cpu          = var.processor_cpu
  processor_memory       = var.processor_memory
  num_workers            = var.num_workers
  
  depends_on = [..., docker_registry_image.processor]
}
```

**New Root Variables:**
```hcl
variable "processor_count"  { default = 2 }     # Number of processor tasks
variable "processor_cpu"    { default = "256" } # vCPU units
variable "processor_memory" { default = "512" } # Memory in MiB
variable "num_workers"      { default = 5 }     # Workers per task
```

**New Root Outputs:**
```hcl
output "processor_service_name" { value = module.ecs.processor_service_name }
output "processor_service_arn"  { value = module.ecs.processor_service_arn }
output "processor_ecr_url"      { value = module.ecr_processor.repository_url }
```

## Key Design Decisions

### 1. **Separate ECS Service**
- Processor runs as an independent service in the same cluster
- Enables independent scaling based on queue depth
- No load balancer needed (background worker)

### 2. **Long Polling**
- 20-second wait time reduces empty responses and API costs
- More efficient than short polling
- Better resource utilization

### 3. **Multiple Workers per Task**
- Default: 5 workers per task × 2 tasks = 10 concurrent workers
- Each worker can process up to 10 messages in parallel
- Total potential throughput: 100 concurrent order processings

### 4. **Visibility Timeout**
- Set to 30 seconds (generous for 100-500ms processing time)
- Allows for retry if processing fails or task crashes
- Messages return to queue if not deleted within timeout

### 5. **Error Handling**
- Malformed messages: Deleted immediately (won't block queue)
- Payment failures: Logged and deleted (in production, would retry or DLQ)
- AWS errors: Exponential backoff (5s sleep)

### 6. **Graceful Shutdown**
- Listens for SIGTERM/SIGINT
- Cancels context to stop all workers
- Waits for in-flight processing to complete
- Clean ECS task termination

## Performance Characteristics

### Throughput Capacity
With default configuration (2 tasks × 5 workers):
- **Max concurrent processing**: 10 workers × 10 messages = 100 orders
- **Processing time**: 100-500ms per order
- **Theoretical throughput**: ~200-1000 orders/second (if queue is full)

### Scaling
- **Horizontal**: Increase `processor_count` for more tasks
- **Vertical**: Increase `num_workers` for more workers per task
- **Auto-scaling**: Can be added based on SQS queue depth (ApproximateNumberOfMessagesVisible)

### Resource Usage
- **CPU**: 256 units (0.25 vCPU) per task - light CPU load
- **Memory**: 512 MiB per task - sufficient for in-memory processing
- **Network**: Minimal - only SQS API calls

## Testing & Deployment

### Deployment Steps
```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### Verify Deployment
```bash
# Check processor service status
aws ecs describe-services \
  --cluster order-processing-service-cluster \
  --services order-processing-service-processor

# Check processor logs
aws logs tail /aws/ecs/order-processing-service \
  --follow \
  --filter-pattern "Worker"
```

### Send Test Messages
```bash
# Use async endpoint to send orders
curl -X POST http://<ALB_DNS>/orders/async \
  -H "Content-Type: application/json" \
  -d '{
    "customerId": 1,
    "productId": 101,
    "quantity": 2,
    "totalAmount": "19.98"
  }'

# Response: 202 Accepted (immediate)
# Check logs to see processor handling the order
```

### Monitoring
Key metrics to watch:
- **SQS ApproximateNumberOfMessagesVisible**: Queue depth
- **SQS NumberOfMessagesReceived**: Messages consumed
- **SQS NumberOfMessagesDeleted**: Successfully processed
- **ECS CPUUtilization**: Processor CPU usage
- **CloudWatch Logs**: Order processing logs

## Message Flow Example

### 1. Order Submission (Async Endpoint)
```
POST /orders/async
→ 202 Accepted
→ SNS Publish
```

### 2. SNS to SQS
```
SNS Topic → SQS Queue (fan-out)
Message: {
  "Message": "{\"orderId\":123,\"customerId\":1,\"productId\":101,...}"
}
```

### 3. Processor Consumption
```
Worker 1: Polling SQS (long poll, 20s)
↓
Receive 10 messages
↓
Spawn 10 goroutines (parallel processing)
↓
For each message:
  - Parse SNS wrapper
  - Extract order
  - Process payment (100-500ms)
  - Update status: COMPLETED
  - Delete message from SQS
```

### 4. Logs
```
Worker 1: Received 10 messages
Worker 1: Processing order 123 (Customer: 1, Product: 101, Quantity: 2, Amount: 19.98)
Order 123 status updated to: COMPLETED
Worker 1: Successfully processed order 123 in 234ms
```

## IAM Permissions Required

The processor uses the same `LabRole` as the web service, which should include:
- **SQS**: `ReceiveMessage`, `DeleteMessage`, `GetQueueAttributes`
- **CloudWatch Logs**: `CreateLogStream`, `PutLogEvents`
- **ECR**: `GetAuthorizationToken`, `BatchCheckLayerAvailability`, `GetDownloadUrlForLayer`, `BatchGetImage`

## Configuration Summary

| Component | Resource | Configuration |
|-----------|----------|---------------|
| Processor Tasks | ECS Fargate | 2 tasks × 256 CPU × 512 MB |
| Workers per Task | Go goroutines | 5 workers |
| Total Workers | - | 10 concurrent workers |
| Messages per Poll | SQS | Up to 10 messages |
| Long Poll Duration | SQS | 20 seconds |
| Visibility Timeout | SQS | 30 seconds |
| Processing Time | Simulated | 100-500ms per order |
| Max Throughput | Theoretical | ~200-1000 orders/sec |

## Next Steps (Step 4)
1. Update Locust tests to include async endpoint testing
2. Verify 100% acceptance rate with async processing
3. Monitor CloudWatch logs for processor activity
4. Compare sync vs async performance under load
5. Demonstrate independent scaling of web service and processor

## Files Modified/Created

### Created:
- `src/orderProcessor.go` - Main processor service
- `src/Dockerfile.processor` - Processor Docker image

### Modified:
- `terraform/modules/ecs/variables.tf` - Added processor variables
- `terraform/modules/ecs/main.tf` - Added processor task & service
- `terraform/modules/ecs/outputs.tf` - Added processor outputs
- `terraform/main.tf` - Added processor ECR, docker build/push, updated ECS module call
- `terraform/variables.tf` - Added processor configuration variables
- `terraform/outputs.tf` - Added processor outputs

## Summary
Step 3 successfully implements a robust, scalable order processor service that:
- ✅ Consumes messages from SQS using long polling
- ✅ Processes orders concurrently with multiple workers
- ✅ Simulates payment processing (100-500ms)
- ✅ Handles graceful shutdown and error conditions
- ✅ Deploys as a separate ECS Fargate service
- ✅ Scales independently from the web service
- ✅ Integrates seamlessly with existing SNS/SQS infrastructure

The system now supports true asynchronous order processing with immediate response (202) to clients and background processing of orders in a scalable, fault-tolerant manner.
