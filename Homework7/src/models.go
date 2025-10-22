package main

import (
	"fmt"
	"time"
)

type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
	Category     string `json:"category"`
	Description  string `json:"description"`
	Brand        string `json:"brand"`
}

// Validate checks if the product data is valid
func (p *Product) Validate() error {
	if p.ProductID <= 0 {
		return fmt.Errorf("product ID must be positive")
	}
	if p.SKU == "" {
		return fmt.Errorf("SKU cannot be empty")
	}
	if p.Weight < 0 {
		return fmt.Errorf("weight cannot be negative")
	}
	return nil
}

// ErrorResponse represents a generic error response
type ErrorResponse struct {
	Message string `json:"message"`
	Code    string `json:"code"`
}

// Error implements the error interface
func (e ErrorResponse) Error() string {
	return e.Message
}

// NewInvalidInputError creates an error response for invalid input
func NewInvalidInputError(message string) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    "INVALID_INPUT",
	}
}

// SearchResponse represents the response for product search
type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time,omitempty"`
}

// NewProductNotFoundError creates an error response for when a product is not found
func NewProductNotFoundError(productID int32) ErrorResponse {
	return ErrorResponse{
		Message: "Product not found",
		Code:    "PRODUCT_NOT_FOUND",
	}
}

// === ORDER MODELS ===

// Item represents an item in an order
type Item struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// Order represents an order in the system
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// Validate validates an order
func (o *Order) Validate() error {
	if o.OrderID == "" {
		return NewInvalidInputError("Order ID is required")
	}
	if o.CustomerID <= 0 {
		return NewInvalidInputError("Customer ID must be positive")
	}
	if len(o.Items) == 0 {
		return NewInvalidInputError("Order must have at least one item")
	}
	for _, item := range o.Items {
		if item.ProductID <= 0 {
			return NewInvalidInputError("Product ID must be positive")
		}
		if item.Quantity <= 0 {
			return NewInvalidInputError("Quantity must be positive")
		}
		if item.Price < 0 {
			return NewInvalidInputError("Price cannot be negative")
		}
	}
	return nil
}

// OrderResponse represents the response for order operations
type OrderResponse struct {
	OrderID   string `json:"order_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// OrderStats tracks order processing statistics
type OrderStats struct {
	TotalOrders      int     `json:"total_orders"`
	PendingOrders    int     `json:"pending_orders"`
	ProcessingOrders int     `json:"processing_orders"`
	CompletedOrders  int     `json:"completed_orders"`
	AverageTime      float64 `json:"average_processing_time_ms"`
}
