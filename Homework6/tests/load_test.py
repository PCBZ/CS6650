"""
Product API Load Testing Suite with Locust
CS6650 Distributed Systems - Performance Analysis

This script tests different scenarios and compares HttpUser vs FastHttpUser
to understand performance characteristics of our Product API.
"""

from locust import HttpUser, FastHttpUser, task, between, events
import random
import json
import time
from datetime import datetime


# Test data for creating products
SAMPLE_PRODUCTS = [
    {
        "product_id": 1001,
        "sku": "LPT-GAM-001",
        "manufacturer": "TechMax Corporation",
        "category_id": 100,
        "weight": 2500,
        "some_other_id": 5001
    },
    {
        "product_id": 1002,
        "sku": "HPH-WRL-002",
        "manufacturer": "AudioPro Industries",
        "category_id": 200,
        "weight": 300,
        "some_other_id": 5002
    },
    {
        "product_id": 1003,
        "sku": "SPH-FLG-003",
        "manufacturer": "MobileTech Systems",
        "category_id": 150,
        "weight": 180,
        "some_other_id": 5003
    },
    {
        "product_id": 1004,
        "sku": "TAB-SLT-004",
        "manufacturer": "TabletCorp Limited",
        "category_id": 150,
        "weight": 650,
        "some_other_id": 5004
    },
    {
        "product_id": 1005,
        "sku": "SWT-FIT-005",
        "manufacturer": "WearableTech Solutions",
        "category_id": 300,
        "weight": 85,
        "some_other_id": 5005
    },
    {
        "product_id": 1006,
        "sku": "SPK-PRT-006",
        "manufacturer": "SoundMax Electronics",
        "category_id": 200,
        "weight": 450,
        "some_other_id": 5006
    },
    {
        "product_id": 1007,
        "sku": "HUB-USC-007",
        "manufacturer": "ConnectPro Technologies",
        "category_id": 400,
        "weight": 120,
        "some_other_id": 5007
    },
    {
        "product_id": 1008,
        "sku": "MOS-WRL-008",
        "manufacturer": "PeripheralTech Inc",
        "category_id": 400,
        "weight": 95,
        "some_other_id": 5008
    },
    {
        "product_id": 1009,
        "sku": "KBD-MCH-009",
        "manufacturer": "KeyboardMax Company",
        "category_id": 400,
        "weight": 980,
        "some_other_id": 5009
    },
    {
        "product_id": 1010,
        "sku": "MON-4K-010",
        "manufacturer": "DisplayTech Corporation",
        "category_id": 500,
        "weight": 5200,
        "some_other_id": 5010
    },
    {
        "product_id": 1011,
        "sku": "CAM-DSL-011",
        "manufacturer": "PhotoPro Equipment",
        "category_id": 600,
        "weight": 850,
        "some_other_id": 5011
    },
    {
        "product_id": 1012,
        "sku": "PRT-LAS-012",
        "manufacturer": "PrintMax Solutions",
        "category_id": 700,
        "weight": 12000,
        "some_other_id": 5012
    },
    {
        "product_id": 1013,
        "sku": "RTR-WIF-013",
        "manufacturer": "NetworkTech Systems",
        "category_id": 800,
        "weight": 320,
        "some_other_id": 5013
    },
    {
        "product_id": 1014,
        "sku": "HDD-EXT-014",
        "manufacturer": "StoragePro Devices",
        "category_id": 900,
        "weight": 180,
        "some_other_id": 5014
    },
    {
        "product_id": 1015,
        "sku": "CHG-QCH-015",
        "manufacturer": "PowerTech Electronics",
        "category_id": 400,
        "weight": 150,
        "some_other_id": 5015
    }
]

class ProductAPIUser(HttpUser):
    """
    HttpUser for baseline testing
    """
    wait_time = between(1, 3)  # Wait 1-3 seconds between requests
    created_products = []  # Store created product IDs
    
    def on_start(self):
        """Initialize user session"""
        # Create some initial products for testing
        for _ in range(5):
            self.create_product()
    
    @task(8)  # Weight: 10 (most common)
    def get_product_by_id(self):
        """GET /v1/products/{id} - Product detail views"""
        if self.created_products:
            product_id = random.choice(self.created_products)
            with self.client.get(f"/v1/products/{product_id}", catch_response=True) as response:
                if response.status_code == 200:
                    response.success()
                elif response.status_code == 404:
                    response.success()  # 404 is acceptable for deleted products
                else:
                    response.failure(f"Got status code {response.status_code}")
        else:
            # If no products exist, try to get a random ID
            random_id = str(random.randint(1, 100))
            self.client.get(f"/v1/products/{random_id}")
    
    @task(3)  # Weight: 3 (less common post)
    def create_product(self):
        """POST /v1/products - Product creation"""
        product_data = random.choice(SAMPLE_PRODUCTS).copy()

        product_id = product_data["product_id"]
        
        with self.client.post(
            f"/v1/products/{product_id}/details",
            json=product_data,
            headers={"Content-Type": "application/json"},
            catch_response=True
        ) as response:
            if response.status_code == 204:
                self.created_products.append(product_id)
                response.success()
            else:
                response.failure(f"Got status code {response.status_code}")


# class FastProductAPIUser(FastHttpUser):
#     """
#     FastHttpUser for comparison testing
#     Uses httpx under the hood for better async performance
#     """
#     wait_time = between(1, 3)
#     created_products = []
    
#     def on_start(self):
#         """Initialize user session"""
        
#         # Create some initial products for testing
#         for _ in range(5):
#             self.create_product()
    
#     @task(8)
#     def get_product_by_id(self):
#         """GET /v1/products/{id}"""
#         if self.created_products:
#             product_id = random.choice(self.created_products)
#             with self.client.get(f"/v1/products/{product_id}", catch_response=True) as response:
#                 if response.status_code == 200:
#                     response.success()
#                 elif response.status_code == 404:
#                     response.success()
#                 else:
#                     response.failure(f"Got status code {response.status_code}")
#         else:
#             random_id = str(random.randint(1, 100))
#             self.client.get(f"/v1/products/{random_id}")
    
#     @task(3)
#     def create_product(self):
#         """POST v1/products"""
#         product_data = random.choice(SAMPLE_PRODUCTS).copy()

#         product_id = product_data["product_id"]
        
#         with self.client.post(
#             f"/v1/products/{product_id}/details",
#             json=product_data,
#             headers={"Content-Type": "application/json"},
#             catch_response=True
#         ) as response:
#             if response.status_code == 204:
#                 self.created_products.append(product_id)
#                 response.success()
#             else:
#                 response.failure(f"Got status code {response.status_code}")