"""
MySQL Shopping Cart API Load Testing
CS6650 Distributed Systems - Database Operations Performance Analysis

Test Protocol:
- Run exactly 150 operations: 50 create, 50 add items, 50 get cart
- Complete test sequence within 5 minutes
- Save results to: mysql_test_results.json

Operations:
1. POST /v1/shopping-carts (create cart) - 50 times
2. POST /v1/shopping-carts/{id}/items (add items) - 50 times
3. GET /v1/shopping-carts/{id} (retrieve cart) - 50 times
"""

from locust import HttpUser, task, events
from locust.env import Environment
from locust.stats import stats_printer, stats_history
from locust import runners
import random
import json
import time
from datetime import datetime, timezone
import os


# Global storage for test results
test_results = []
cart_ids = []  # Store created cart IDs
operation_counts = {
    "create_cart": 0,
    "add_items": 0,
    "get_cart": 0
}
MAX_OPERATIONS = 50  # Maximum operations per type


class MySQLShoppingCartUser(HttpUser):
    """
    MySQL Shopping Cart API load testing
    Tests database operations with specific operation counts
    """
    wait_time = lambda self: 0  # No wait time between requests
    
    def on_start(self):
        """Called when a user starts"""
        self.customer_id = random.randint(10000, 99999)
    
    @task(3)
    def create_shopping_cart(self):
        """
        POST /v1/shopping-carts - Create new shopping cart
        Run 50 times
        """
        global operation_counts, cart_ids
        
        # Check if we've reached the limit
        if operation_counts["create_cart"] >= MAX_OPERATIONS:
            return
            
        operation_counts["create_cart"] += 1
        
        cart_data = {
            "customer_id": self.customer_id
        }
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            with self.client.post(
                "/v1/shopping-carts",
                json=cart_data,
                headers={"Content-Type": "application/json"},
                catch_response=True
            ) as response:
                response_time = (time.time() - start_time) * 1000
                
                result = {
                    "operation": "create_cart",
                    "response_time": round(response_time, 2),
                    "success": response.status_code == 201,
                    "status_code": response.status_code,
                    "timestamp": timestamp
                }
                
                if response.status_code == 201:
                    response_json = response.json()
                    cart_id = response_json.get("shopping_cart_id")
                    if cart_id:
                        cart_ids.append(cart_id)
                    print(f"✅ Created cart {cart_id} in {response_time:.1f}ms")
                    response.success()
                else:
                    print(f"❌ Create cart failed: {response.status_code}")
                    response.failure(f"Status: {response.status_code}")
                
                test_results.append(result)
                
        except Exception as e:
            result = {
                "operation": "create_cart",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
            test_results.append(result)
            print(f"💥 Create cart exception: {e}")
    
    @task(2)
    def add_item_to_cart(self):
        """
        POST /v1/shopping-carts/{id}/items - Add items to cart
        Run 50 times
        """
        global operation_counts, cart_ids
        
        # Check if we've reached the limit
        if operation_counts["add_items"] >= MAX_OPERATIONS:
            return
        
        # Need at least one cart to add items
        if not cart_ids:
            return
            
        operation_counts["add_items"] += 1
        
        # Pick a random cart from created carts
        cart_id = random.choice(cart_ids)
        
        item_data = {
            "product_id": random.randint(1, 5),  # Products 1-5 from sample data
            "quantity": random.randint(1, 5)
        }
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            with self.client.post(
                f"/v1/shopping-carts/{cart_id}/items",
                json=item_data,
                headers={"Content-Type": "application/json"},
                catch_response=True
            ) as response:
                response_time = (time.time() - start_time) * 1000
                
                result = {
                    "operation": "add_items",
                    "response_time": round(response_time, 2),
                    "success": response.status_code == 200,
                    "status_code": response.status_code,
                    "timestamp": timestamp
                }
                
                if response.status_code == 200:
                    response_json = response.json()
                    quantity = response_json.get("quantity", 0)
                    print(f"✅ Added item to cart {cart_id}, qty: {quantity} in {response_time:.1f}ms")
                    response.success()
                else:
                    print(f"❌ Add item failed: {response.status_code}")
                    response.failure(f"Status: {response.status_code}")
                
                test_results.append(result)
                
        except Exception as e:
            result = {
                "operation": "add_items",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
            test_results.append(result)
            print(f"💥 Add item exception: {e}")
    
    @task(1)
    def get_shopping_cart(self):
        """
        GET /v1/shopping-carts/{id} - Retrieve cart with items
        Run 50 times
        """
        global operation_counts, cart_ids
        
        # Check if we've reached the limit
        if operation_counts["get_cart"] >= MAX_OPERATIONS:
            return
        
        # Need at least one cart to retrieve
        if not cart_ids:
            return
            
        operation_counts["get_cart"] += 1
        
        # Pick a random cart from created carts
        cart_id = random.choice(cart_ids)
        
        start_time = time.time()
        timestamp = datetime.now(timezone.utc).isoformat()
        
        try:
            with self.client.get(
                f"/v1/shopping-carts/{cart_id}",
                catch_response=True
            ) as response:
                response_time = (time.time() - start_time) * 1000
                
                result = {
                    "operation": "get_cart",
                    "response_time": round(response_time, 2),
                    "success": response.status_code == 200,
                    "status_code": response.status_code,
                    "timestamp": timestamp
                }
                
                if response.status_code == 200:
                    response_json = response.json()
                    item_count = len(response_json.get("items", []))
                    print(f"✅ Retrieved cart {cart_id} with {item_count} items in {response_time:.1f}ms")
                    response.success()
                else:
                    print(f"❌ Get cart failed: {response.status_code}")
                    response.failure(f"Status: {response.status_code}")
                
                test_results.append(result)
                
        except Exception as e:
            result = {
                "operation": "get_cart",
                "response_time": 0,
                "success": False,
                "status_code": 0,
                "timestamp": timestamp,
                "error": str(e)
            }
            test_results.append(result)
            print(f"💥 Get cart exception: {e}")


@events.quitting.add_listener
def save_results(environment, **kwargs):
    """
    Save test results to JSON file when test completes
    """
    output_file = "mysql_test_results.json"
    
    print(f"\n{'='*60}")
    print("📊 MySQL Shopping Cart Load Test Results")
    print(f"{'='*60}")
    print(f"Total operations: {len(test_results)}")
    print(f"  - create_cart: {operation_counts['create_cart']}")
    print(f"  - add_items: {operation_counts['add_items']}")
    print(f"  - get_cart: {operation_counts['get_cart']}")
    print(f"\nSaving results to: {output_file}")
    
    # Calculate summary statistics
    summary = {
        "test_info": {
            "total_operations": len(test_results),
            "target_operations": MAX_OPERATIONS * 3,
            "completed_operations": operation_counts,
            "test_duration_seconds": environment.stats.total.get_response_time_percentile(1.0),
        },
        "results": test_results
    }
    
    # Add per-operation statistics
    for op_type in ["create_cart", "add_items", "get_cart"]:
        op_results = [r for r in test_results if r["operation"] == op_type]
        if op_results:
            response_times = [r["response_time"] for r in op_results if r["success"]]
            success_count = sum(1 for r in op_results if r["success"])
            
            summary[f"{op_type}_stats"] = {
                "total": len(op_results),
                "successful": success_count,
                "failed": len(op_results) - success_count,
                "success_rate": f"{(success_count / len(op_results) * 100):.1f}%",
                "avg_response_time": round(sum(response_times) / len(response_times), 2) if response_times else 0,
                "min_response_time": round(min(response_times), 2) if response_times else 0,
                "max_response_time": round(max(response_times), 2) if response_times else 0,
            }
    
    # Write to file
    with open(output_file, 'w') as f:
        json.dump(summary, f, indent=2)
    
    print(f"✅ Results saved successfully!")
    print(f"{'='*60}\n")


@events.test_stop.add_listener
def on_test_stop(environment, **kwargs):
    """
    Called when test stops - print summary
    """
    print(f"\n{'='*60}")
    print("🛑 Test Stopped")
    print(f"Operations completed:")
    print(f"  - create_cart: {operation_counts['create_cart']}/{MAX_OPERATIONS}")
    print(f"  - add_items: {operation_counts['add_items']}/{MAX_OPERATIONS}")
    print(f"  - get_cart: {operation_counts['get_cart']}/{MAX_OPERATIONS}")
    print(f"{'='*60}\n")
