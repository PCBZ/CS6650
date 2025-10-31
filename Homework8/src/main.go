package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/gorm"
)

// Global stores
var productStore = NewProductStore()
var orderStore = NewOrderStore()
var db *gorm.DB // Global database connection

func main() {
	// Initialize database if DB_HOST is configured
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		log.Println("Database configuration detected, initializing...")
		var err error
		db, err = InitDatabase()
		if err != nil {
			log.Printf("⚠️  Warning: Database initialization failed: %v", err)
			log.Printf("⚠️  Continuing with in-memory storage only")
			db = nil
		} else {
			// Get underlying sql.DB for cleanup
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				defer sqlDB.Close()
			}
			log.Println("✅ Database initialized successfully")
		}
	} else {
		log.Println("ℹ️  No database configuration found, using in-memory storage")
	}

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
	cartAPI := NewShoppingCartAPI(db)

	// Setup router with API handlers
	router := SetupRouter(orderAPI, productAPI, cartAPI)

	// Start server on port 8080
	fmt.Println("🚀 Starting order processing service on port 8080...")
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
