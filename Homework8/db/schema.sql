-- Drop tables in reverse dependency order (for clean re-runs)
DROP TABLE IF EXISTS shopping_cart_items;
DROP TABLE IF EXISTS shopping_carts;
DROP TABLE IF EXISTS products;

-- ============================================================================
-- PRODUCTS TABLE
-- ============================================================================
CREATE TABLE products (
    product_id INT PRIMARY KEY AUTO_INCREMENT,
    sku VARCHAR(100) NOT NULL UNIQUE,
    manufacturer VARCHAR(200) NOT NULL,
    category_id INT NOT NULL,
    weight INT NOT NULL COMMENT 'Weight in grams',
    some_other_id INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ============================================================================
-- SHOPPING CARTS TABLE
-- ============================================================================
CREATE TABLE shopping_carts (
    shopping_cart_id INT PRIMARY KEY AUTO_INCREMENT,
    customer_id INT NOT NULL,
    status ENUM('active', 'abandoned') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- ============================================================================
-- SHOPPING CART ITEMS TABLE
-- ============================================================================
CREATE TABLE shopping_cart_items (
    cart_item_id INT PRIMARY KEY AUTO_INCREMENT,
    shopping_cart_id INT NOT NULL,
    product_id INT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    added_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign keys for referential integrity
    FOREIGN KEY (shopping_cart_id) 
        REFERENCES shopping_carts(shopping_cart_id) 
        ON DELETE CASCADE 
        COMMENT 'Delete items when cart is deleted',
    FOREIGN KEY (product_id) 
        REFERENCES products(product_id)
        ON DELETE RESTRICT
        COMMENT 'Prevent product deletion if in carts',
    
    -- Constraints
    UNIQUE KEY unique_cart_product (shopping_cart_id, product_id)
        COMMENT 'One entry per product per cart - prevent duplicates',
    CHECK (quantity > 0),
    
    -- Index for <50ms retrieval requirement
    INDEX idx_cart (shopping_cart_id) 
        COMMENT 'Fast retrieval of all items in cart for GET API'
);