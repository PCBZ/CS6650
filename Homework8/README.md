# Homework 8

## MySQL Intergration
### Infrastructure Extension
```terraform
# Engine configuration
  engine               = "mysql"
  engine_version       = "8.0"
  instance_class       = "db.t3.micro"
  allocated_storage    = 20
```

### Schema Design
- **Tables:** products, shopping_carts, shopping_cart_items

**Key Indexes and Why:**

1. **PRIMARY KEY Indexes (Automatic)**
   - Fast lookup by ID for all tables
   
2. **Foreign Key Indexes (Automatic)**
   - `shopping_cart_items.shopping_cart_id` - Enables fast JOIN for GET API
   - `shopping_cart_items.product_id` - Validates product exists
   
3. **Explicit Index**
   - `shopping_cart_items.idx_cart` on `shopping_cart_id` - Ensures <50ms cart retrieval with up to 50 items
   - **Critical for performance:** Index seek (~5-20ms) vs full table scan (~100ms+)

**Concurrency Strategy:**
- Row-level locking: Each cart item is a separate row
- InnoDB engine handles concurrent INSERT/UPDATE/DELETE on different rows

### Shopping Cart API Implementation

#### Connection Pooling Configuration
```go
// Configure connection pool for 100 concurrent users
sqlDB.SetMaxOpenConns(25)
sqlDB.SetMaxIdleConns(5)
sqlDB.SetConnMaxLifetime(5 * time.Minute)
```

#### 1. POST /v1/shopping-carts
```go
func (api *ShoppingCartAPI) CreateShoppingCart(c *gin.Context) {
    cart := ShoppingCart{
        CustomerID: req.CustomerID,
        Status:     "active",
    }
    
    // Transaction handling
    api.db.Transaction(func(tx *gorm.DB) error {
        return tx.Create(&cart).Error
    })
    
    c.JSON(http.StatusCreated, CreateCartResponse{
        ShoppingCartID: cart.ShoppingCartID,
        CustomerID:     cart.CustomerID,
        Status:         cart.Status,
    })
}
```

#### 2. GET /v1/shopping-carts/:id
```go
func (api *ShoppingCartAPI) GetShoppingCart(c *gin.Context) {
    var cart ShoppingCart
    
    // Efficient JOIN using Preload
    result := api.db.Preload("Items").First(&cart, cartID)
    
    if errors.Is(result.Error, gorm.ErrRecordNotFound) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
        return
    }
    
    c.JSON(http.StatusOK, cart)
}
```

#### 3. POST /v1/shopping-carts/:id/items
```go
func (api *ShoppingCartAPI) AddItemToCart(c *gin.Context) {
    // Multi-table operation with transaction
    api.db.Transaction(func(tx *gorm.DB) error {
        // Verify cart exists
        var cart ShoppingCart
        if err := tx.First(&cart, cartID).Error; err != nil {
            return err
        }
        
        // Verify product exists
        var product Product
        if err := tx.First(&product, req.ProductID).Error; err != nil {
            return err
        }
        
        // Check if item exists
        var existingItem ShoppingCartItem
        result := tx.Where("shopping_cart_id = ? AND product_id = ?", 
            cartID, req.ProductID).First(&existingItem)
        
        if result.Error == nil {
            // Update quantity if exists
            existingItem.Quantity += req.Quantity
            return tx.Save(&existingItem).Error
        } else {
            // Create new item
            newItem := ShoppingCartItem{
                ShoppingCartID: cartID,
                ProductID:      req.ProductID,
                Quantity:       req.Quantity,
            }
            return tx.Create(&newItem).Error
        }
    })
}
```

### Performance Test
<img width="1200" height="600" alt="response_distribution" src="https://github.com/user-attachments/assets/081adb2e-4a48-4d08-a0fa-d1f103f7a940" />

### Learning Note
A small number of create_cart API requests are slow, more than 100ms.

### CloudWatch
<img width="743" height="357" alt="image" src="https://github.com/user-attachments/assets/44714373-f224-4195-9428-e26d593bf684" />
<img width="1143" height="361" alt="image" src="https://github.com/user-attachments/assets/678a64b8-89f2-4487-9241-0d3c193670a0" />

In-memory response time data:
50th：36ms
95th：76ms

Database(MySql) response time data:
50th: 40ms-50ms
95th: 70ms

It shows 2 implementations are similar. In-memory response time is a little better than database(MySql)




