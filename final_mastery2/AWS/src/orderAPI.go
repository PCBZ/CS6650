package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OrderAPI provides stubbed handlers so the service can compile and respond.
// For Homework6 we only need synchronous processing and a basic stats endpoint.
type OrderAPI struct{}

// NewOrderAPI creates a new OrderAPI instance.
func NewOrderAPI() *OrderAPI {
	return &OrderAPI{}
}

// ProcessOrderSync handles POST /v1/orders/sync.
func (api *OrderAPI) ProcessOrderSync(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "order processed",
	})
}

// GetOrderStats handles GET /v1/orders/stats.
func (api *OrderAPI) GetOrderStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"total_orders":     0,
		"processed_orders": 0,
	})
}
