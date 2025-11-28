package main

import "fmt"

// Global product store
var productStore = NewProductStore()

func main() {
	// Generate sample data
	dataGen := NewDataGenerator()
	products := dataGen.GenerateProducts()

	// Store the generated products
	for _, product := range products {
		productStore.SetProduct(product.ProductID, product)
	}

	router := SetupRouter()

	// Start server on port 8080
	fmt.Println("Starting server on port 8080...")
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
