"""
Normal Operations Load Test
CS6650 Phase 1 - Normal business traffic simulation
Expected: 100% success rate, stable response times
"""

from locust import HttpUser, task, between
import random
import json
import time


class NormalOperationsUser(HttpUser):
    """
    Normal operations user: 5 users simulating regular business traffic
    """
    wait_time = between(0.1, 0.5)  # 200-600ms wait time
    
    def on_start(self):
        """Initialize user data when starting"""
        self.customer_id = random.randint(1000, 9999)
        
    @task
    def post_sync_order(self):
        """
        Synchronous order processing test
        POST /v1/orders/sync
        """
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": random.randint(1, 1000),
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(10.0, 100.0), 2)
                }
                for _ in range(random.randint(1, 3))
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
                    print(f"✅ Order {response_json.get('order_id', 'unknown')} completed in {response_time:.0f}ms")
                    response.success()
                else:
                    print(f"❌ Order failed with status {response.status_code}: {response.text}")
                    response.failure(f"Status code: {response.status_code}")
                    
        except Exception as e:
            print(f"💥 Request failed with exception: {e}")