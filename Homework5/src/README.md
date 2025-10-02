# Homework 5

## Part 1: Reading
The CAP theorem is most informative for me. It completely changed my perspective that there is no perfect distributed system can satisfy **Consistency**, **Availability**, and **Partition tolerance**. Each distributed system has to make trade-offs among these 3 properties based on its use cases.

I'd actually encountered this concept in practice before, but didn't understand the theory behind it. We had a master-slave database store "likes" for each story in a social media feeds where users would occasionally see outdated "likes" count when reading from slave while writing to master. The replication lag wasn't a bug - it was the inevitable consequence of our architectural choice to keep read services available even during network issues. This experience makes CAP theorem much more concrete and meaningful to me.

## Part 2: Product API Service

### API Endpoints

**Base URL:** `http://localhost:8080`

#### 1. Get Product
```bash
GET /v1/products/{productId}
```

**Example:**
```bash
curl http://localhost:8080/v1/products/1
```

**Response:**
```json
{
  "productId": 1,
  "sku": "ABC-123",
  "manufacturer": "ACME Corp",
  "categoryId": 100,
  "weight": 500,
  "someOtherId": 999
}
```

#### 2. Create/Update Product
```bash
POST /v1/products/{productId}/details
```

**Example:**
```bash
curl -X POST http://localhost:8080/v1/products/1/details \
  -H "Content-Type: application/json" \
  -d '{
    "productId": 1,
    "sku": "LAPTOP-001",
    "manufacturer": "TechCorp",
    "categoryId": 10,
    "weight": 2000,
    "someOtherId": 555
  }'
```

**Response:** `204 No Content` on success

### How to Run

1. **Start the service:**
```bash
go run .
```

2. **Or with Docker:**
```bash
docker build -t product-api .
docker run -p 8080:8080 product-api
```

### Usage Example
```bash
# Create a product
curl -X POST http://localhost:8080/v1/products/1/details \
  -H "Content-Type: application/json" \
  -d '{"productId": 1, "sku": "TEST-001", "manufacturer": "TestCorp"}'

# Get the product
curl http://localhost:8080/v1/products/1
```