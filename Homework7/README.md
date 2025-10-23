# Homework 7

## Phase 1
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

**40 users**
<img width="2928" height="1800" alt="total_requests_per_second_1761254882 821" src="https://github.com/user-attachments/assets/4fc3651e-92fe-48ef-abe2-04bfbfb0c91c" />


