package main

import (
	"fmt"
	"log"
)

// Global stores
var productStore = NewProductStore()
var orderStore = NewOrderStore()

func main() {
	// Generate sample data
	dataGen := NewDataGenerator()
	products := dataGen.GenerateProducts()

	// Store the generated products
	for _, product := range products {
		productStore.SetProduct(product.ProductID, product)
	}

	// Initialize SNS publisher for async order processing
	publisher, err := NewSNSPublisher()
	if err != nil {
		log.Printf("⚠️  Warning: SNS publisher initialization failed: %v", err)
		log.Printf("⚠️  Async endpoint will not be available")
		publisher = nil
	} else {
		log.Println("✅ SNS publisher initialized successfully")
	}

	// Initialize API handlers
	orderAPI := NewOrderAPI(publisher)
	productAPI := NewProductAPI()

	// Setup router with API handlers
	router := SetupRouter(orderAPI, productAPI)

	// Start server on port 8080
	fmt.Println("🚀 Starting order processing service on port 8080...")
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
