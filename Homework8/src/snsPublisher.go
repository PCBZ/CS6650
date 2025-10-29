package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

// SNSPublisher handles publishing messages to SNS
type SNSPublisher struct {
	client   *sns.SNS
	topicArn string
}

// NewSNSPublisher creates a new SNS publisher
func NewSNSPublisher() (*SNSPublisher, error) {
	// Get SNS topic ARN from environment variable
	topicArn := os.Getenv("SNS_TOPIC_ARN")
	if topicArn == "" {
		return nil, fmt.Errorf("SNS_TOPIC_ARN environment variable not set")
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &SNSPublisher{
		client:   sns.New(sess),
		topicArn: topicArn,
	}, nil
}

// PublishOrder publishes an order to SNS topic
func (p *SNSPublisher) PublishOrder(order *Order) error {
	// Convert order to JSON
	orderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order: %w", err)
	}

	// Publish to SNS
	input := &sns.PublishInput{
		Message:  aws.String(string(orderJSON)),
		TopicArn: aws.String(p.topicArn),
		MessageAttributes: map[string]*sns.MessageAttributeValue{
			"OrderID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(order.OrderID),
			},
			"CustomerID": {
				DataType:    aws.String("Number"),
				StringValue: aws.String(fmt.Sprintf("%d", order.CustomerID)),
			},
		},
	}

	result, err := p.client.Publish(input)
	if err != nil {
		return fmt.Errorf("failed to publish to SNS: %w", err)
	}

	fmt.Printf("📤 Published order %s to SNS (MessageID: %s)\n", order.OrderID, *result.MessageId)
	return nil
}
