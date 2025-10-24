package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// OrderMessage represents the structure of order messages
type OrderMessage struct {
	OrderID    string `json:"order_id"`
	CustomerID int    `json:"customer_id"`
	Items      []Item `json:"items"`
	CreatedAt  string `json:"created_at"`
	Status     string `json:"status"`
}

// Item represents a single item in an order
type Item struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// handleSNSEvent processes SNS messages triggered by Lambda
func handleSNSEvent(ctx context.Context, snsEvent events.SNSEvent) error {
	for _, record := range snsEvent.Records {
		snsRecord := record.SNS

		// Parse the order message from SNS
		var order OrderMessage
		if err := json.Unmarshal([]byte(snsRecord.Message), &order); err != nil {
			log.Printf("Error parsing order: %v", err)
			continue
		}

		// Simulate payment processing (3 seconds delay)
		if err := processPayment(order); err != nil {
			log.Printf("Payment failed for order %s: %v", order.OrderID, err)
			continue
		}
	}

	return nil
}

// processPayment simulates payment processing with 3-second delay
func processPayment(order OrderMessage) error {
	// Simulate payment processing time (3 seconds)
	time.Sleep(3 * time.Second)

	// Silent processing - no logs for production
	return nil
}

func main() {
	lambda.Start(handleSNSEvent)
	fmt.Println("Lambda function started")
}
