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
	store *OrderStore
}

// generateSimpleUUID creates a simple UUID-like string
func generateSimpleUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// NewOrderAPI returns a new instance of OrderAPI
func NewOrderAPI() *OrderAPI {
	return &OrderAPI{
		store: orderStore,
	}
}

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

// GetOrderStats handles GET /orders/stats - for monitoring during load tests
func (api *OrderAPI) GetOrderStats(c *gin.Context) {
	stats := api.store.GetStats()
	c.JSON(http.StatusOK, stats)
}
