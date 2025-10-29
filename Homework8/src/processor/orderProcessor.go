package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// OrderProcessor handles consuming and processing orders from SQS
type OrderProcessor struct {
	sqsClient *sqs.Client
	queueURL  string
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
}

// OrderMessage represents the structure of messages received from SQS
type OrderMessage struct {
	OrderID     int    `json:"orderId"`
	CustomerID  int    `json:"customerId"`
	ProductID   int    `json:"productId"`
	Quantity    int    `json:"quantity"`
	TotalAmount string `json:"totalAmount"`
	Timestamp   string `json:"timestamp"`
}

// NewOrderProcessor creates a new order processor
func NewOrderProcessor(queueURL string) (*OrderProcessor, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &OrderProcessor{
		sqsClient: sqs.NewFromConfig(cfg),
		queueURL:  queueURL,
		ctx:       ctx,
		cancel:    cancel,
	}, nil
}

// Start begins polling SQS for messages
func (op *OrderProcessor) Start(numWorkers int) {
	log.Printf("Starting order processor with %d workers", numWorkers)

	for i := 0; i < numWorkers; i++ {
		op.wg.Add(1)
		go op.pollMessages(i)
	}
}

// Stop gracefully shuts down the processor
func (op *OrderProcessor) Stop() {
	log.Println("Stopping order processor...")
	op.cancel()
	op.wg.Wait()
	log.Println("Order processor stopped")
}

// pollMessages continuously polls SQS for messages
func (op *OrderProcessor) pollMessages(workerID int) {
	defer op.wg.Done()

	for {
		select {
		case <-op.ctx.Done():
			return
		default:
			op.receiveAndProcessMessages(workerID)
		}
	}
}

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

	if err != nil {
		log.Printf("Worker %d: Error receiving messages: %v", workerID, err)
		return
	}

	if len(result.Messages) == 0 {
		// No messages received during long poll
		return
	}

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

var paymentSemaphore = make(chan struct{}, 5) // Limit to 5 concurrent payments

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

// processPayment simulates payment processing
func (op *OrderProcessor) processPayment(order OrderMessage) error {
	// Acquire semaphore slot (blocks if 5 concurrent payments already running)
	paymentSemaphore <- struct{}{}
	defer func() {
		<-paymentSemaphore // Release semaphore slot
	}()

	// Simulate payment processing time (3 seconds)
	time.Sleep(3 * time.Second)

	return nil
}

// updateOrderStatus updates the order status (in production, this would update a database)
func (op *OrderProcessor) updateOrderStatus(orderID int, status string) {
	// In a real application, this would update a database
	// Silent update for production - no verbose logging
}

// deleteMessage deletes a message from the SQS queue
func (op *OrderProcessor) deleteMessage(receiptHandle *string) error {
	_, err := op.sqsClient.DeleteMessage(op.ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(op.queueURL),
		ReceiptHandle: receiptHandle,
	})

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

// Main function for the order processor service
func main() {
	// Get queue URL from environment variable
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable is required")
	}

	// Get number of workers from environment variable (default: 1)
	numWorkers := 1
	if workersEnv := os.Getenv("NUM_WORKERS"); workersEnv != "" {
		fmt.Sscanf(workersEnv, "%d", &numWorkers)
	}

	// Create order processor
	processor, err := NewOrderProcessor(queueURL)
	if err != nil {
		log.Fatalf("Failed to create order processor: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start processing
	processor.Start(numWorkers)

	// Wait for termination signal
	<-sigChan
	log.Println("Received shutdown signal")

	// Graceful shutdown
	processor.Stop()
	log.Println("Order processor service terminated")
}
