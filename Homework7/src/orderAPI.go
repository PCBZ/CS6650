package main

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// OrderAPI contains all order-related business logic
type OrderAPI struct {
	store     *OrderStore
	publisher *SNSPublisher
}

// generateSimpleUUID creates a simple UUID-like string
func generateSimpleUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// NewOrderAPI returns a new instance of OrderAPI
func NewOrderAPI(publisher *SNSPublisher) *OrderAPI {
	return &OrderAPI{
		store:     orderStore,
		publisher: publisher,
	}
}

var paymentSemaphore = make(chan struct{}, 5) // Limit to 5 concurrent sync orders

// ProcessOrderSync handles POST /orders/sync - synchronous order processing
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

	makePayment()

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

// ProcessOrderAsync handles POST /orders/async - asynchronous order processing
func (api *OrderAPI) ProcessOrderAsync(c *gin.Context) {
	// Parse request body
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError(err.Error()))
		return
	}

	// Generate order ID if not provided
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

func makePayment() {
	// Acquire semaphore slot (blocks if 5 concurrent payments already running)
	paymentSemaphore <- struct{}{}
	defer func() {
		<-paymentSemaphore // Release semaphore slot
	}()

	// Simulate payment processing time (3 seconds)
	time.Sleep(3 * time.Second)
}

// GetOrderStats handles GET /orders/stats - for monitoring during load tests
func (api *OrderAPI) GetOrderStats(c *gin.Context) {
	stats := api.store.GetStats()
	c.JSON(http.StatusOK, stats)
}
