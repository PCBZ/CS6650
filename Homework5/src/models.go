package main

import (
	"fmt"
)

type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
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

// NewInvalidInputError creates an error response for invalid input
func NewInvalidInputError(message string) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    "INVALID_INPUT",
	}
}

// NewProductNotFoundError creates an error response for when a product is not found
func NewProductNotFoundError(productID int32) ErrorResponse {
	return ErrorResponse{
		Message: "Product not found",
		Code:    "PRODUCT_NOT_FOUND",
	}
}
