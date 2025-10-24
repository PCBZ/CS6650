# Homework 7

## Phase 1
### Sync EndPoint
```go
func (api *OrderAPI) ProcessOrderSync(c *gin.Context) {
	// Parse request body
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError(err.Error()))
		return
	} // Generate order ID if not provided
	if order.OrderID == "" {
		order.OrderID = generateSimpleUUID()
	}

	// Set initial status and timestamp
	order.Status = "pending"
	order.CreatedAt = time.Now()

	// Validate the order
	if err := order.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Store order as pending
	api.store.SetOrder(order.OrderID, &order)

	// Update status to processing
	order.Status = "processing"
	api.store.SetOrder(order.OrderID, &order)

	// **CRITICAL: Simulate payment verification (3 seconds delay)**
	// This is where the bottleneck happens during flash sales!
	time.Sleep(3 * time.Second)

	// Update status to completed
	order.Status = "completed"
	api.store.SetOrder(order.OrderID, &order)

	// Return success response
	response := OrderResponse{
		OrderID:   order.OrderID,
		Status:    order.Status,
		Message:   "Order processed successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
```

### VPC Setting
```terraform
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "${var.service_name}-vpc"
  }
}

# Create public subnets
resource "aws_subnet" "public" {
  count                   = 2
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.${count.index + 1}.0/24"
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true

  tags = {
    Name = "${var.service_name}-public-subnet-${count.index + 1}"
    Type = "public"
  }
}

# Create private subnets
resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = "${var.service_name}-private-subnet-${count.index + 1}"
    Type = "private"
  }
}
```

### ALB Setting
```terraform
# Target Group for Fargate tasks
resource "aws_lb_target_group" "this" {
  name                 = "${var.service_name}-tg"
  port                 = var.container_port
  protocol             = "HTTP"
  vpc_id               = var.vpc_id
  target_type          = "ip"  # Required for Fargate
  deregistration_delay = 30     # Reduce from default 300s to 30s

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = var.health_check_path
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 2
  }
```

### Normal Operation
<img width="2928" height="1800" alt="total_requests_per_second_1761167543 74" src="https://github.com/user-attachments/assets/4e1439e5-c142-4d5b-83dd-6814a70eb0ca" />

### Test Flash Sale
Reaching **700 users** triggered failed requests.  
<img width="2928" height="1800" alt="total_requests_per_second_1761192597 164" src="https://github.com/user-attachments/assets/61920804-9de0-4e50-8e20-ff2ad1501890" />

## Phase 2
With 700 users: Maximum thourghput = 233 orders/sec, orders lost.
If demanding 700 users, it will lose 147 orders/sec.

## Phase 3
### Step 1. Add SNS/SQS Infrastructure
```terraform
# SNS Topic for order processing events
resource "aws_sns_topic" "order_processing" {
  name = "${var.service_name}-order-processing-events"

  tags = {
    Name = "${var.service_name}-order-processing-events"
  }
}

# SQS Queue for order processing
resource "aws_sqs_queue" "order_processing" {
  name                       = "${var.service_name}-order-processing-queue"
  visibility_timeout_seconds = 30     # 30 seconds visibility timeout
  message_retention_seconds  = 345600 # 4 days retention
  receive_wait_time_seconds  = 20     # 20 seconds long polling

  tags = {
    Name = "${var.service_name}-order-processing-queue"
  }
}
```

### Step 2. Add Async EndPoint
```go
// ProcessOrderAsync handles POST /orders/async - asynchronous order processing
func (api *OrderAPI) ProcessOrderAsync(c *gin.Context) {
	
  ***

	// Publish order to SNS topic (non-blocking)
	if api.publisher != nil {
		if err := api.publisher.PublishOrder(&order); err != nil {
			// If publishing fails, mark order as failed
			order.Status = "failed"
			api.store.SetOrder(order.OrderID, &order)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to queue order for processing",
			})
			return
		}
	} else {
		// If publisher is not initialized, return error
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Async processing is not available",
		})
		return
	}

	// Return 202 Accepted immediately (order is queued, not processed yet)
	response := OrderResponse{
		OrderID:   order.OrderID,
		Status:    order.Status,
		Message:   "Order accepted and queued for processing",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusAccepted, response)
}
```

### Step 3. Add Order Processor
```go
// receiveAndProcessMessages receives messages from SQS and processes them
func (op *OrderProcessor) receiveAndProcessMessages(workerID int) {
	// Long polling configuration
	result, err := op.sqsClient.ReceiveMessage(op.ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(op.queueURL),
		MaxNumberOfMessages: 10, // Process up to 10 messages at a time
		WaitTimeSeconds:     20, // Long polling - wait up to 20 seconds for messages
		VisibilityTimeout:   30, // Give 30 seconds to process before message becomes visible again
		MessageAttributeNames: []string{
			string(types.QueueAttributeNameAll),
		},
	})

	// Process messages concurrently
	var processingWg sync.WaitGroup
	for _, message := range result.Messages {
		processingWg.Add(1)
		go func(msg types.Message) {
			defer processingWg.Done()
			op.processMessage(workerID, msg)
		}(message)
	}

	processingWg.Wait()
}
```

```go
// processMessage processes a single SQS message
func (op *OrderProcessor) processMessage(workerID int, message types.Message) {
	// Parse the SNS message wrapper
	var snsMessage struct {
		Message string `json:"Message"`
	}

	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		log.Printf("Error parsing SNS wrapper: %v", err)
		op.deleteMessage(message.ReceiptHandle)
		return
	}

	// Parse the actual order message
	var order OrderMessage
	if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
		log.Printf("Error parsing order message: %v", err)
		op.deleteMessage(message.ReceiptHandle)
		return
	}

	// Simulate payment processing
	if err := op.processPayment(order); err != nil {
		log.Printf("Payment failed for order %d: %v", order.OrderID, err)
		op.deleteMessage(message.ReceiptHandle)
		return
	}

	// Update order status (in production, this would update a database)
	op.updateOrderStatus(order.OrderID, "COMPLETED")

	// Delete message from queue after successful processing
	if err := op.deleteMessage(message.ReceiptHandle); err != nil {
		log.Printf("Error deleting message for order %d: %v", order.OrderID, err)
		return
	}
}
```
**5 users**
<img width="2928" height="1800" alt="total_requests_per_second_1761254470 797" src="https://github.com/user-attachments/assets/7555775f-1be7-4c0e-8c57-60dd316466e5" />

**700 users**
<img width="2928" height="1800" alt="total_requests_per_second_1761254882 821" src="https://github.com/user-attachments/assets/4fc3651e-92fe-48ef-abe2-04bfbfb0c91c" />

## Phase 4
<img width="1211" height="416" alt="image" src="https://github.com/user-attachments/assets/b9f35cea-ff02-422e-844e-0319328364ce" />
Queue Growth Rate = 46 messages/sec
For 700 users requests, it will never empty the queue; If it stops at 39k messages in the queue, it will consume 32.8 hours

## Phase 5
**5 goroutines**
<img width="1196" height="108" alt="image" src="https://github.com/user-attachments/assets/2cda1843-3852-4305-bd10-e93980f99a52" />

<img width="2928" height="1800" alt="total_requests_per_second_1761263001 924" src="https://github.com/user-attachments/assets/b3a32fa4-6189-4e81-b7fa-82458babaeb9" />  
**10 goroutines**  
[<img width="2928" height="1800" alt="total_requests_per_second_1761265177 282" src="https://github.com/user-attachments/assets/30913d82-47c2-4a7a-9142-ffbc4aac929b" />](http://order-processing-service-alb-700972238.us-west-2.elb.amazonaws.com)

<img width="1201" height="146" alt="image" src="https://github.com/user-attachments/assets/c309060f-eba6-431b-bb63-924c1249db59" />


| goroutine count | orders/sec | queue growth rate(messages/sec) |
| --------------- | ---------- | ------------------------------- |
| 5  | 70 | 1.4 |
| 20 | 60 | steady ｜

