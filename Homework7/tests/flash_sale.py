"""
Flash Sale Load Test  
CS6650 Phase 1 - High concurrency flash sale simulation
Expected: Timeouts, failures, demonstrates synchronous processing bottleneck
"""

from locust import HttpUser, task, between
import random
import json
import time


class FlashSaleUser(HttpUser):
    """
    Flash sale user: 20 users with high-frequency requests simulating flash sale scenario
    """
    wait_time = between(0.1, 0.3)  # 100-300ms wait time (more aggressive)
    
    def on_start(self):
        """Initialize user data when starting"""
        self.customer_id = random.randint(1000, 9999)
        
    @task
    def post_sync_order(self):
        """
        Synchronous order processing test - Flash sale scenario
        POST /v1/orders/sync - Expected to show bottleneck
        """
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": random.randint(1, 100),  # More concentrated product IDs (flash sale items)
                    "quantity": random.randint(1, 3),
                    "price": round(random.uniform(50.0, 200.0), 2)  # Higher value items
                }
                for _ in range(random.randint(1, 2))  # 1-2 items per order
            ]
        }
        
        start_time = time.time()
        
        try:
            with self.client.post(
                "/v1/orders/sync",
                json=order_data,
                headers={"Content-Type": "application/json"},
                catch_response=True
            ) as response:
                response_time = (time.time() - start_time) * 1000
                
                if response.status_code == 200:
                    response_json = response.json()
                    print(f"🔥 Flash sale order {response_json.get('order_id', 'unknown')} completed in {response_time:.0f}ms")
                    response.success()
                else:
                    print(f"⚠️  Flash sale order failed with status {response.status_code}: {response.text}")
                    response.failure(f"Status code: {response.status_code}")
                    
        except Exception as e:
            print(f"💥 Flash sale request failed: {e}")