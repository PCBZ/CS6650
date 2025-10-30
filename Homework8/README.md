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