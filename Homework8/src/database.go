package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitRDSDatabase initializes GORM database connection and creates tables if needed
func InitRDSDatabase() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbPort == "" {
		dbPort = "3306"
	}

	// Validate required environment variables
	if dbHost == "" {
		return nil, fmt.Errorf("DB_HOST environment variable not set")
	}
	if dbUser == "" {
		return nil, fmt.Errorf("DB_USER environment variable not set")
	}
	if dbPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD environment variable not set")
	}
	if dbName == "" {
		return nil, fmt.Errorf("DB_NAME environment variable not set")
	}

	// Build DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	log.Printf("Connecting to database at %s:%s...", dbHost, dbPort)

	// Configure GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error), // Only log errors in production
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool for 100 concurrent users
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")

	// Initialize schema (create tables if they don't exist)
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates tables using GORM AutoMigrate
func initSchema(db *gorm.DB) error {
	log.Println("Initializing database schema...")

	migrator := db.Migrator()
	models := []interface{}{&Product{}, &ShoppingCart{}, &ShoppingCartItem{}}

	allTablesExist := true
	for _, model := range models {
		if !migrator.HasTable(model) {
			allTablesExist = false
			log.Printf("Table for %T does not exist", model)
			break
		}
	}

	if allTablesExist {
		return nil
	}

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Add unique constraint for cart+product combination if not exists
	if !db.Migrator().HasConstraint(&ShoppingCartItem{}, "unique_cart_product") {
		db.Exec("ALTER TABLE shopping_cart_items ADD CONSTRAINT unique_cart_product UNIQUE (shopping_cart_id, product_id)")
	}

	log.Println("✓ Database schema created successfully")
	return nil
}

func initDynamoDB() (dynamodbClient *dynamodb.Client, err error) {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %v", err)
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)
	return dynamoClient, nil
}
