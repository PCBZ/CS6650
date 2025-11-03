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

## DynamoDB
### Design
```terraform
resource "aws_dynamodb_table" "shopping_carts" {
  name         = "shopping_carts"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
}

resource "aws_dynamodb_table" "shopping_cart_items" {
  name         = "shopping_cart_items"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "cart_id"
  range_key    = "product_id"
  attribute {
    name = "cart_id"
    type = "N"
  }
  attribute {
    name = "product_id"
    type = "N"
  }
}
```
- **2 tables** use `cart` and `cart-item` associated with `cart_id` to get relationship.
- **Partition Key** use `cart_id` as a random ID being appropriate for even distribution.
- **Secondary Index** use `customer_id` for `cart` table to allow querying by customer id.
- **Sort Key** use `item_id` for `cart-item` table to allow multiple items in a cart.

### API Implementation
#### Create Shopping Cart
```go
cartID := generateSnowflakeID()
cart := map[string]types.AttributeValue{
	"cart_id":     &types.AttributeValueMemberN{Value: strconv.FormatInt(cartID, 10)},
	"customer_id": &types.AttributeValueMemberN{Value: strconv.Itoa(req.CustomerID)},
	"status":      &types.AttributeValueMemberS{Value: "active"},
	"created_at":  &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
}

_, err := api.dynamoClient.PutItem(context.Background(), &dynamodb.PutItemInput{
	TableName: aws.String("shopping_carts"),
	Item:      cart,
})
```
#### Get Shopping Cart
```go
// Get cart
	result, err := api.dynamoClient.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String("shopping_carts"),
		Key: map[string]types.AttributeValue{
			"cart_id": &types.AttributeValueMemberN{Value: cartID},
		},
	})

	if err != nil {
		log.Printf("Failed to get shopping cart: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart"})
		return
	}

	if result.Item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	// Get cart items using Query
	itemsResult, err := api.dynamoClient.Query(context.Background(), &dynamodb.QueryInput{
		TableName:              aws.String("shopping_cart_items"),
		KeyConditionExpression: aws.String("cart_id = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberN{Value: cartID},
		},
	})
```

### Eventual Consistency Testing
I ran the consistency test for 100 iterations. Each iteration contains 3 tests, checking read-after-write consistency for different scenarios.
The final results are as follows:
```bash
Total Checks: 2200
Inconsistencies: 15
Consistency Rate: 99.32%
```

### Testing
<img width="1200" height="600" alt="dynamodb_response_distribution" src="https://github.com/user-attachments/assets/b2ea8a40-2c53-497b-8be7-be195ca220d0" />

**Comparison**:  
- **Response Time**: Both MySQL and DynamoDB have similar typical response delay, but DynamoDB has more extremly situation.
- **get_cart**: MySQL significantly better, tighter distribution

<img width="1214" height="542" alt="image" src="https://github.com/user-attachments/assets/563c4baa-4b8b-465c-932f-f9e77282aaa7" />
No throttle events
<img width="1130" height="639" alt="image" src="https://github.com/user-attachments/assets/194df2e1-6915-4717-83df-be199d3af8f7" />
<img width="1133" height="646" alt="image" src="https://github.com/user-attachments/assets/65e45cf6-1b2e-4d1d-a50e-ffbb4195ae75" />
No partition key is accessed overload.

## Database Comparison & Analysis
### Data Comparison
| Metric | MySQL | DynamoDB | Winner | Margin |
|:-------|------:|--------:|:-------|-------:|
| **Avg Response Time (ms)** | 45.07 | 45.23 | MySQL | 0.36% |
| **P50 Response Time (ms)** | 41.25 | 40.34 | DynamoDB | 2.30% |
| **P95 Response Time (ms)** | 72.95 | 71.04 | DynamoDB | 2.62% |
| **P99 Response Time (ms)** | 105.03 | 122.89 | MySQL | 14.53% |
| **Success Rate (%)** | 100.00% | 100.00% | Tie | 0.00% |

| Operation | MySQL Avg (ms) | DynamoDB Avg (ms) | Faster By |
|-----------|----------------|-------------------|-----------|
| ADD_ITEMS | 49.83 | 44.07 | DynamoDB (11.56%) |
| CREATE_CART | 44.02 | 47.04 | MySQL (6.40%) |
| GET_CART | 41.34 | 44.59 | MySQL (7.27%) |

### Consistency Model Impact Assessment
#### Investigation Requirements
- Consistency: MySQL 100%, Dynamo DB 99.32%
- When a customer purchases an item, inventory decrements might not immediately propagate across all DynamoDB replicas. This creates the classic overselling problem: multiple customers could simultaneously see the same "last item" as available
- MySQL guarantees strong consistency, customers always get the correct information of the system, while losing performance; DynamoDB returns the value that may be outdated, but it ensures the availability and partition tolerance.
- Under heavy-load system, customers may keep receiving the error code or timeout, seeing endless loading of the webUI, if the system uses MySQL; while customers may get the outdated data of products on the web, if the sytem uses DynamoDB.
### Resource Efficiency Analysis
- MySQL requires setting max connections, pool size, and timeout. Scaling requires manual intervention; DynamoDB is managed by AWS, scaling automatically to handle any traffic spike.
- MySQL provides high predictability as it has fixed resource model. And capacity planning is complex, because it requires to foresee the traffic and match the settings. DynamoDB varies with usage pattern. However, AWS helps you scaling it automatically.
- MySQL requires significant operational overhead including daily monitoring of server resources, connection pool tuning, query optimization, backup management, and handling replication/failover scenarios, demanding either dedicated DBA expertise or substantial developer time on infrastructure concerns. DynamoDB eliminates virtually all operational complexity by providing a fully managed service where AWS handles scaling, maintenance, and infrastructure, allowing developers to focus purely on application logic and occasional cost monitoring, though this convenience comes with less predictable costs and reduced fine-grained control over performance tuning.

### Real-World Scenario Recommendations
**Scenario A: Startup MVP**: DynamoDB  
- 1 developer can launch DynamoDB fast but not for MySQL.
- No need to operate overhead. 
**Scenario B: Growing Business**: MySQL
- 5 developers can handle operational complexity.
- Predictable scaling allows better budget forecasting for steady growth.
**Scenario C: High-Traffic Events**: MySQL  
- Predictable infrastructure controls revenue even in a spike traffic.
- Can invest in infrastructure means it offers budget and technical resources to properly architect MySQL for high-scale scenarios.
**Scenario D: Global Platform**: DynamoDB + MySQL
- DynamoDB provides multi-regional features for different location's users.
- Enterprise services like business analytic require complex query, which is suitable for MySQL.
- AWS handles DynamoDB 24/7, internal team focuses MySQL expertise on smaller, critical systems.

## Your Evidence-Based Architecture Recommendations
1. **Shopping Cart Winner**: MySQL
   Response time is lower. Retrieving cart items need less communication with database.
2. **When to choose the other**:
   - Unpredictable traffic patterns, eg. sudden spikes.
   - More concurrent requests happens: DynamoDB is easy to scale.
3. **Polyglot Strategy**:
   - **Shopping carts**: MySQL
   - **User sessions**: DynamoDB, it is usually simple check and being visited in high frequency.
   - **Product catalog**: MySQL, supporting complex query; read operations are more than write.
   - **Order history**: MySQL, supporting complex query.

## Learning Reflection
### What Surprised You?
The eventual consistency speed is unexpected. Even it start immediately after it writes successfully.
### What Failed Initially?
MySQL schema does not initialize. It is quite complex to write SQL to handle API requests.
DynamoDB's query is much more different from conventional SQL.

### Key Insights Gained
Based on learning journey, I suggest using MySQL to handle complex query, and using DynamoDB to fast deployment. Because it is difficult to finish the logic with DynamoDB when a complex query is needed. It takes quite a long time to initialize MySQL.

