package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase initializes GORM database connection and creates tables if needed
func InitDatabase() (*gorm.DB, error) {
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
	sqlDB.SetMaxOpenConns(25)
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

	// Check if tables exist
	var tableCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name IN ('products', 'shopping_carts', 'shopping_cart_items')
	`).Scan(&tableCount)

	if tableCount == 3 {
		log.Println("✓ Database schema already exists")
		return nil
	}

	log.Println("Creating database schema using GORM AutoMigrate...")

	// Auto-migrate tables (this will create tables if they don't exist)
	// Note: GORM doesn't support ENUM types directly, so we'll create tables first
	// then let GORM manage them
	if err := db.AutoMigrate(&Product{}, &ShoppingCart{}, &ShoppingCartItem{}); err != nil {
		return fmt.Errorf("failed to auto-migrate: %w", err)
	}

	// Add unique constraint for cart+product combination if not exists
	if !db.Migrator().HasConstraint(&ShoppingCartItem{}, "unique_cart_product") {
		db.Exec("ALTER TABLE shopping_cart_items ADD CONSTRAINT unique_cart_product UNIQUE (shopping_cart_id, product_id)")
	}

	// Insert sample products using GORM
	sampleProducts := []Product{
		{
			SKU:          "LAPTOP-001",
			Manufacturer: "TechCorp",
			CategoryID:   1,
			Weight:       2500,
			SomeOtherID:  101,
			Category:     "Electronics",
			Description:  "High-performance laptop with 16GB RAM and 512GB SSD",
			Brand:        "TechCorp Pro",
		},
		{
			SKU:          "MOUSE-002",
			Manufacturer: "PeripheralsCo",
			CategoryID:   2,
			Weight:       150,
			SomeOtherID:  102,
			Category:     "Accessories",
			Description:  "Wireless ergonomic mouse with precision tracking",
			Brand:        "PeripheralsCo Comfort",
		},
		{
			SKU:          "KEYBOARD-003",
			Manufacturer: "PeripheralsCo",
			CategoryID:   2,
			Weight:       800,
			SomeOtherID:  103,
			Category:     "Accessories",
			Description:  "Mechanical keyboard with RGB lighting",
			Brand:        "PeripheralsCo Gaming",
		},
		{
			SKU:          "MONITOR-004",
			Manufacturer: "DisplayTech",
			CategoryID:   3,
			Weight:       5000,
			SomeOtherID:  104,
			Category:     "Displays",
			Description:  "27-inch 4K UHD monitor with HDR support",
			Brand:        "DisplayTech Ultra",
		},
		{
			SKU:          "HEADSET-005",
			Manufacturer: "AudioMax",
			CategoryID:   4,
			Weight:       300,
			SomeOtherID:  105,
			Category:     "Audio",
			Description:  "Noise-cancelling wireless headset with microphone",
			Brand:        "AudioMax Pro",
		},
	}

	for _, product := range sampleProducts {
		// Use FirstOrCreate to avoid duplicates
		db.Where(Product{SKU: product.SKU}).FirstOrCreate(&product)
	}

	log.Println("✓ Database schema created successfully")
	return nil
}
