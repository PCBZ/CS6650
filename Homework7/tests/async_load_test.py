"""
E-commerce API Async Load Testing Suite
CS6650 Distributed Systems - Asynchronous Order Processing Performance Analysis

Tests the async endpoint with SNS/SQS messaging to demonstrate:
1. Immediate 202 Accepted responses (no blocking)
2. 100% acceptance rate even under heavy load
3. Background processing via Order Processor service
"""

from locust import HttpUser, task, between
import random
import json
import time


class OrderLoadTestUser(HttpUser):
    """
    Async order processing load testing
    Tests the critical async endpoint that returns 202 immediately
    """
    # wait_time = between(0.1, 0.5)  # Random wait time 100-500ms between requests
    
    def on_start(self):
        """Called when a user starts - initialize customer data"""
        self.customer_id = random.randint(1000, 9999)
        
    @task
    def post_async_order(self):
        """
        Test asynchronous order processing endpoint
        POST /v1/orders/async - Should return 202 immediately without blocking
        """
        # Generate random order data matching the Order model structure
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": random.randint(1, 1000),
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
                for _ in range(random.randint(1, 3))  # 1-3 items per order
            ]
        }
        
        # Record start time for response time tracking
        start_time = time.time()
        
        try:
            with self.client.post(
                "/v1/orders/async",
                json=order_data,
                headers={"Content-Type": "application/json"},
                catch_response=True
            ) as response:
                # Calculate response time
                response_time = (time.time() - start_time) * 1000
                
                if response.status_code == 202:
                    response_json = response.json()
                    print(f"✅ Order {response_json.get('order_id', 'unknown')} accepted in {response_time:.0f}ms")
                    response.success()
                else:
                    print(f"❌ Order failed with status {response.status_code}: {response.text}")
                    response.failure(f"Status code: {response.status_code}")
                    
        except Exception as e:
            print(f"💥 Request failed with exception: {e}")
