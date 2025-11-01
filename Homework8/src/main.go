package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CartAPI defines the interface for shopping cart operations
type CartAPI interface {
	CreateShoppingCart(c *gin.Context)
	GetShoppingCart(c *gin.Context)
	AddItemToCart(c *gin.Context)
}

// Global stores
var productStore = NewProductStore()
var orderStore = NewOrderStore()
var rdsDB *gorm.DB           // RDS database
var dynamoClient interface{} // DynamoDB client

func initDB() {
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		log.Println("Database configuration detected, initializing both MySQL and DynamoDB...")
		var err error
		rdsDB, err = InitRDSDatabase()
		if err != nil {
			log.Printf("⚠️  Warning: Database initialization failed: %v", err)
		}
		dynamoClient, err = initDynamoDB()
		if err != nil {
			log.Printf("⚠️  Warning: DynamoDB initialization failed: %v", err)
		}
	}
}

func main() {
	initDB()

	// Generate sample data
	// dataGen := NewDataGenerator()
	// products := dataGen.GenerateProducts()

	// Store the generated products
	// for _, product := range products {
	// 	productStore.SetProduct(product.ProductID, product)
	// }

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

	// Select database implementation based on DB_TYPE
	var cartAPI CartAPI
	dbType := os.Getenv("DB_TYPE")
	switch dbType {
	case "dynamodb":
		if client, ok := dynamoClient.(*dynamodb.Client); ok {
			cartAPI = NewShoppingCartAPIDynamo(client)
			log.Println("✅ Using DynamoDB implementation for shopping cart")
		} else {
			log.Printf("⚠️  Warning: DynamoDB client not properly initialized, falling back to MySQL")
			cartAPI = NewShoppingCartAPI(rdsDB)
		}
	case "mysql", "":
		cartAPI = NewShoppingCartAPI(rdsDB)
		log.Println("✅ Using MySQL implementation for shopping cart")
	default:
		log.Printf("⚠️  Warning: Unknown DB_TYPE '%s', falling back to MySQL", dbType)
		cartAPI = NewShoppingCartAPI(rdsDB)
	}

	// Setup router with API handlers
	router := SetupRouter(orderAPI, productAPI, cartAPI)

	// Start server on port 8080
	fmt.Println("🚀 Starting order processing service on port 8080...")
	if err := router.Run(":8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
