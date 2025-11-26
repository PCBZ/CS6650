package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

// OrderAPI contains all order-related business logic
type OrderAPI struct {
	orderStore *OrderStore
	stats      OrderStats
	totalOrders      int64
	completedOrders  int64
	averageTime      int64
}

// NewOrderAPI returns a new instance of OrderAPI
func NewOrderAPI() *OrderAPI {
	return &OrderAPI{
		orderStore: NewOrderStore(),
		stats:      OrderStats{},
	}
}

// ProcessOrderSync handles POST /orders/sync - synchronous order processing
func (api *OrderAPI) ProcessOrderSync(c *gin.Context) {
	startTime := time.Now()

	// Parse request body
	var order Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError(err.Error()))
		return
	}

	// Validate required fields (excluding OrderID which is generated)
	if order.CustomerID <= 0 {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Customer ID must be positive"))
		return
	}
	if len(order.Items) == 0 {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Order must have at least one item"))
		return
	}
	for _, item := range order.Items {
		if item.ProductID <= 0 {
			c.JSON(http.StatusBadRequest, NewInvalidInputError("Product ID must be positive"))
			return
		}
		if item.Quantity <= 0 {
			c.JSON(http.StatusBadRequest, NewInvalidInputError("Quantity must be positive"))
			return
		}
		if item.Price < 0 {
			c.JSON(http.StatusBadRequest, NewInvalidInputError("Price cannot be negative"))
			return
		}
	}

	// Generate order ID
	orderID := fmt.Sprintf("order-%d", time.Now().UnixNano())
	order.OrderID = orderID
	order.Status = "completed"
	order.CreatedAt = time.Now()

	// Store the order
	api.orderStore.SetOrder(orderID, &order)

	// Update stats
	atomic.AddInt64(&api.totalOrders, 1)
	atomic.AddInt64(&api.completedOrders, 1)

	processingTime := time.Since(startTime).Milliseconds()
	atomic.StoreInt64(&api.averageTime, processingTime)

	response := OrderResponse{
		OrderID:   orderID,
		Status:    "completed",
		Message:   "Order processed successfully",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}

// GetOrderStats handles GET /orders/stats - returns order processing statistics
func (api *OrderAPI) GetOrderStats(c *gin.Context) {
	// Get current stats from store
	storeStats := api.orderStore.GetStats()

	// Combine with API stats
	stats := OrderStats{
		TotalOrders:      storeStats.TotalOrders + int(atomic.LoadInt64(&api.totalOrders)),
		PendingOrders:    storeStats.PendingOrders,
		ProcessingOrders: storeStats.ProcessingOrders,
		CompletedOrders:  storeStats.CompletedOrders + int(atomic.LoadInt64(&api.completedOrders)),
		AverageTime:      float64(atomic.LoadInt64(&api.averageTime)),
	}

	c.JSON(http.StatusOK, stats)
}