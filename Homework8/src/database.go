package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// InitDatabase initializes database connection and creates tables if needed
func InitDatabase() (*sql.DB, error) {
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
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	log.Printf("Connecting to database at %s:%s...", dbHost, dbPort)

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool for 100 concurrent users
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✓ Database connection established")

	// Initialize schema (create tables if they don't exist)
	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// initSchema creates tables if they don't exist
func initSchema(db *sql.DB) error {
	log.Println("Initializing database schema...")

	// Check if tables exist
	var tableCount int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = DATABASE() 
		AND table_name IN ('products', 'shopping_carts', 'shopping_cart_items')
	`).Scan(&tableCount)

	if err != nil {
		return fmt.Errorf("failed to check tables: %w", err)
	}

	if tableCount == 3 {
		log.Println("✓ Database schema already exists")
		return nil
	}

	log.Println("Creating database schema...")

	// Read and execute schema
	schema := `
-- Products table
CREATE TABLE IF NOT EXISTS products (
    product_id INT PRIMARY KEY AUTO_INCREMENT,
    sku VARCHAR(100) NOT NULL UNIQUE,
    manufacturer VARCHAR(200) NOT NULL,
    category_id INT NOT NULL,
    weight INT NOT NULL COMMENT 'Weight in grams',
    some_other_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Shopping carts table
CREATE TABLE IF NOT EXISTS shopping_carts (
    shopping_cart_id INT PRIMARY KEY AUTO_INCREMENT,
    customer_id INT NOT NULL,
    status ENUM('active', 'abandoned') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Shopping cart items table
CREATE TABLE IF NOT EXISTS shopping_cart_items (
    cart_item_id INT PRIMARY KEY AUTO_INCREMENT,
    shopping_cart_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (shopping_cart_id) REFERENCES shopping_carts(shopping_cart_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE RESTRICT,
    UNIQUE KEY unique_cart_product (shopping_cart_id, product_id),
    CHECK (quantity > 0),
    INDEX idx_cart (shopping_cart_id)
);

-- Insert sample products
INSERT INTO products (sku, manufacturer, category_id, weight, some_other_id) 
VALUES
    ('LAPTOP-001', 'TechCorp', 1, 2500, 101),
    ('MOUSE-002', 'PeripheralsCo', 2, 150, 102),
    ('KEYBOARD-003', 'PeripheralsCo', 2, 800, 103),
    ('MONITOR-004', 'DisplayTech', 3, 5000, 104),
    ('HEADSET-005', 'AudioMax', 4, 300, 105)
ON DUPLICATE KEY UPDATE product_id=product_id;
`

	// Execute schema
	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	log.Println("✓ Database schema created successfully")
	return nil
}
