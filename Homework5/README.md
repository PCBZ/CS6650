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

#### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/PCBZ/CS6650
   cd Homework5/src
   ```
2. **Install dependencies**
   ```bash
   go mod download
   ```
3. **Direct go run**
   ```bash
   go run .
   ```
4. **Or Run with Docker**
   ```bash
   docker build -t product-api .
   docker run -p 8080:8080 product-api
   ```

#### AWS Deployment

1. **Setup credential**
   ```bash
   aws configure
   ```
   fill up the credential information

2. **Navigate to terraform path**
   ```bash
   cd terraform
   ```

3. **Terraform initialization**
   ```bash
   terraform init
   ```

4. **Terraform deployment**
   ```bash
   terraform apply
   ```

### Usage Example
#### Get a product
##### Get sucessfully (200)
<img width="1070" height="470" alt="image" src="https://github.com/user-attachments/assets/49559e31-dfee-4363-b2cc-11a3e4c8d9f5" />

##### Get failed with item not found (404)
<img width="1065" height="452" alt="image" src="https://github.com/user-attachments/assets/e4d343a4-225b-4bf5-be0c-8672c3004ada" />

##### Get failed with invalid parameters (400)
<img width="1086" height="490" alt="image" src="https://github.com/user-attachments/assets/1d284af8-f1a8-43f4-8b65-fe83317f0f56" />


#### Create a product
##### Create sucessfully (204)
<img width="1065" height="427" alt="image" src="https://github.com/user-attachments/assets/2a2a7b61-edbe-4ac6-9224-aae3d50c07d6" />

##### Create failed with invalid parameters (400)
<img width="1070" height="470" alt="image" src="https://github.com/user-attachments/assets/4e7d78af-be0e-4cec-acca-6b3547f72f0f" />

##### Create failed with



# Get the product
curl http://localhost:8080/v1/products/1
```
